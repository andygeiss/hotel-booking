package orchestration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/andygeiss/hotel-booking/internal/domain/orchestration"
	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// ============================================================================
// Mock Dispatcher
// ============================================================================

type mockDispatcher struct {
	subscriptions map[string][]service.Function[messaging.Message, messaging.MessageState]
	publishedMsgs []messaging.Message
}

func newMockDispatcher() *mockDispatcher {
	return &mockDispatcher{
		subscriptions: make(map[string][]service.Function[messaging.Message, messaging.MessageState]),
		publishedMsgs: []messaging.Message{},
	}
}

func (m *mockDispatcher) Subscribe(ctx context.Context, topic string, handler service.Function[messaging.Message, messaging.MessageState]) error {
	m.subscriptions[topic] = append(m.subscriptions[topic], handler)
	return nil
}

func (m *mockDispatcher) Publish(ctx context.Context, msg messaging.Message) error {
	m.publishedMsgs = append(m.publishedMsgs, msg)
	return nil
}

func (m *mockDispatcher) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockDispatcher) triggerEvent(topic string, data []byte) (messaging.MessageState, error) {
	handlers := m.subscriptions[topic]
	if len(handlers) == 0 {
		return messaging.MessageStateFailed, errors.New("no handlers for topic")
	}
	msg := messaging.NewMessage(topic, data)
	ctx := context.Background()
	return handlers[0](ctx, msg)
}

// ============================================================================
// Test Services Setup (reusing from booking_service_test.go)
// ============================================================================

type eventHandlerTestServices struct {
	reservationRepo    *mockReservationRepository
	availabilityCheck  *mockAvailabilityChecker
	reservationPub     *mockEventPublisher
	reservationService *reservation.Service

	paymentRepo    *mockPaymentRepository
	paymentGateway *mockPaymentGateway
	paymentPub     *mockEventPublisher
	paymentService *payment.Service

	notificationService *mockNotificationService
	bookingService      *orchestration.BookingService
	eventHandlers       *orchestration.EventHandlers
	dispatcher          *mockDispatcher
}

func createEventHandlerTestServices() *eventHandlerTestServices {
	// Reservation context
	reservationRepo := newMockReservationRepository()
	availabilityChecker := &mockAvailabilityChecker{available: true}
	reservationPub := &mockEventPublisher{}
	reservationService := reservation.NewService(reservationRepo, availabilityChecker, reservationPub)

	// Payment context
	paymentRepo := newMockPaymentRepository()
	paymentGateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	paymentPub := &mockEventPublisher{}
	paymentService := payment.NewService(paymentRepo, paymentGateway, paymentPub)

	// Orchestration
	notificationService := &mockNotificationService{}
	bookingService := orchestration.NewBookingService(reservationService, paymentService, notificationService)
	eventHandlers := orchestration.NewEventHandlers(bookingService, reservationService, paymentService)
	dispatcher := newMockDispatcher()

	return &eventHandlerTestServices{
		reservationRepo:     reservationRepo,
		availabilityCheck:   availabilityChecker,
		reservationPub:      reservationPub,
		reservationService:  reservationService,
		paymentRepo:         paymentRepo,
		paymentGateway:      paymentGateway,
		paymentPub:          paymentPub,
		paymentService:      paymentService,
		notificationService: notificationService,
		bookingService:      bookingService,
		eventHandlers:       eventHandlers,
		dispatcher:          dispatcher,
	}
}

func eventHandlerValidDateRange() reservation.DateRange {
	checkIn := time.Now().Add(48 * time.Hour).Truncate(24 * time.Hour)
	checkOut := checkIn.Add(72 * time.Hour)
	return reservation.NewDateRange(checkIn, checkOut)
}

func eventHandlerValidGuests() []reservation.GuestInfo {
	return []reservation.GuestInfo{
		reservation.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
	}
}

func eventHandlerValidMoney() shared.Money {
	return shared.NewMoney(10000, "USD")
}

// ============================================================================
// RegisterHandlers Tests
// ============================================================================

func Test_EventHandlers_RegisterHandlers_Should_Subscribe_To_All_Topics(t *testing.T) {
	// Arrange
	svc := createEventHandlerTestServices()
	ctx := context.Background()

	// Act
	err := svc.eventHandlers.RegisterHandlers(ctx, svc.dispatcher)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must subscribe to reservation.created", len(svc.dispatcher.subscriptions[reservation.EventTopicCreated]), 1)
	assert.That(t, "must subscribe to payment.authorized", len(svc.dispatcher.subscriptions[payment.EventTopicAuthorized]), 1)
	assert.That(t, "must subscribe to payment.captured", len(svc.dispatcher.subscriptions[payment.EventTopicCaptured]), 1)
	assert.That(t, "must subscribe to payment.failed", len(svc.dispatcher.subscriptions[payment.EventTopicFailed]), 1)
}

// ============================================================================
// HandleReservationCreated Tests
// ============================================================================

func Test_HandleReservationCreated_Should_Parse_Event(t *testing.T) {
	// Arrange
	svc := createEventHandlerTestServices()
	ctx := context.Background()
	_ = svc.eventHandlers.RegisterHandlers(ctx, svc.dispatcher)

	dateRange := eventHandlerValidDateRange()
	evt := reservation.EventCreated{
		ReservationID: "res-001",
		GuestID:       "guest-001",
		RoomID:        "room-101",
		CheckIn:       dateRange.CheckIn,
		CheckOut:      dateRange.CheckOut,
		TotalAmount:   eventHandlerValidMoney(),
	}
	data, _ := json.Marshal(evt)

	// Act
	state, err := svc.dispatcher.triggerEvent(reservation.EventTopicCreated, data)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "state must be completed", state, messaging.MessageStateCompleted)
}

func Test_HandleReservationCreated_Should_Trigger_Payment_Authorization(t *testing.T) {
	// Arrange
	svc := createEventHandlerTestServices()
	ctx := context.Background()
	_ = svc.eventHandlers.RegisterHandlers(ctx, svc.dispatcher)

	dateRange := eventHandlerValidDateRange()
	evt := reservation.EventCreated{
		ReservationID: "res-001",
		GuestID:       "guest-001",
		RoomID:        "room-101",
		CheckIn:       dateRange.CheckIn,
		CheckOut:      dateRange.CheckOut,
		TotalAmount:   eventHandlerValidMoney(),
	}
	data, _ := json.Marshal(evt)

	// Act
	_, _ = svc.dispatcher.triggerEvent(reservation.EventTopicCreated, data)

	// Assert
	paymentID := payment.PaymentID("pay-001")
	storedPayment, err := svc.paymentRepo.Read(ctx, paymentID)
	assert.That(t, "payment must exist", err == nil, true)
	assert.That(t, "payment must be authorized", storedPayment.Status, payment.StatusAuthorized)
}

func Test_HandleReservationCreated_With_Invalid_JSON_Should_Return_Failed(t *testing.T) {
	// Arrange
	svc := createEventHandlerTestServices()
	ctx := context.Background()
	_ = svc.eventHandlers.RegisterHandlers(ctx, svc.dispatcher)

	invalidData := []byte("{invalid json}")

	// Act
	state, err := svc.dispatcher.triggerEvent(reservation.EventTopicCreated, invalidData)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "state must be failed", state, messaging.MessageStateFailed)
}

// ============================================================================
// HandlePaymentAuthorized Tests
// ============================================================================

func Test_HandlePaymentAuthorized_Should_Parse_Event(t *testing.T) {
	// Arrange
	svc := createEventHandlerTestServices()
	ctx := context.Background()
	_ = svc.eventHandlers.RegisterHandlers(ctx, svc.dispatcher)

	reservationID := shared.ReservationID("res-001")
	paymentID := payment.PaymentID("pay-001")

	// Setup: create reservation and authorize payment
	_, _ = svc.reservationService.CreateReservation(
		ctx, reservationID, "guest-001", "room-101",
		eventHandlerValidDateRange(), eventHandlerValidMoney(), eventHandlerValidGuests(),
	)
	_, _ = svc.paymentService.AuthorizePayment(ctx, paymentID, reservationID, eventHandlerValidMoney(), "credit_card")

	evt := payment.EventAuthorized{
		PaymentID:     paymentID,
		ReservationID: reservationID,
		Amount:        eventHandlerValidMoney(),
		TransactionID: "tx-12345",
	}
	data, _ := json.Marshal(evt)

	// Act
	state, err := svc.dispatcher.triggerEvent(payment.EventTopicAuthorized, data)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "state must be completed", state, messaging.MessageStateCompleted)
}

func Test_HandlePaymentAuthorized_Should_Capture_Payment(t *testing.T) {
	// Arrange
	svc := createEventHandlerTestServices()
	ctx := context.Background()
	_ = svc.eventHandlers.RegisterHandlers(ctx, svc.dispatcher)

	reservationID := shared.ReservationID("res-001")
	paymentID := payment.PaymentID("pay-001")

	// Setup
	_, _ = svc.reservationService.CreateReservation(
		ctx, reservationID, "guest-001", "room-101",
		eventHandlerValidDateRange(), eventHandlerValidMoney(), eventHandlerValidGuests(),
	)
	_, _ = svc.paymentService.AuthorizePayment(ctx, paymentID, reservationID, eventHandlerValidMoney(), "credit_card")

	evt := payment.EventAuthorized{
		PaymentID:     paymentID,
		ReservationID: reservationID,
		Amount:        eventHandlerValidMoney(),
		TransactionID: "tx-12345",
	}
	data, _ := json.Marshal(evt)

	// Act
	_, _ = svc.dispatcher.triggerEvent(payment.EventTopicAuthorized, data)

	// Assert
	storedPayment, _ := svc.paymentRepo.Read(ctx, paymentID)
	assert.That(t, "payment must be captured", storedPayment.Status, payment.StatusCaptured)
}

// ============================================================================
// HandlePaymentCaptured Tests
// ============================================================================

func Test_HandlePaymentCaptured_Should_Parse_Event(t *testing.T) {
	// Arrange
	svc := createEventHandlerTestServices()
	ctx := context.Background()
	_ = svc.eventHandlers.RegisterHandlers(ctx, svc.dispatcher)

	reservationID := shared.ReservationID("res-001")

	// Setup: create reservation
	_, _ = svc.reservationService.CreateReservation(
		ctx, reservationID, "guest-001", "room-101",
		eventHandlerValidDateRange(), eventHandlerValidMoney(), eventHandlerValidGuests(),
	)

	evt := payment.EventCaptured{
		PaymentID:     "pay-001",
		ReservationID: reservationID,
		Amount:        eventHandlerValidMoney(),
	}
	data, _ := json.Marshal(evt)

	// Act
	state, err := svc.dispatcher.triggerEvent(payment.EventTopicCaptured, data)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "state must be completed", state, messaging.MessageStateCompleted)
}

func Test_HandlePaymentCaptured_Should_Confirm_Reservation(t *testing.T) {
	// Arrange
	svc := createEventHandlerTestServices()
	ctx := context.Background()
	_ = svc.eventHandlers.RegisterHandlers(ctx, svc.dispatcher)

	reservationID := shared.ReservationID("res-001")

	// Setup
	_, _ = svc.reservationService.CreateReservation(
		ctx, reservationID, "guest-001", "room-101",
		eventHandlerValidDateRange(), eventHandlerValidMoney(), eventHandlerValidGuests(),
	)

	evt := payment.EventCaptured{
		PaymentID:     "pay-001",
		ReservationID: reservationID,
		Amount:        eventHandlerValidMoney(),
	}
	data, _ := json.Marshal(evt)

	// Act
	_, _ = svc.dispatcher.triggerEvent(payment.EventTopicCaptured, data)

	// Assert
	storedRes, _ := svc.reservationRepo.Read(ctx, reservationID)
	assert.That(t, "reservation must be confirmed", storedRes.Status, reservation.StatusConfirmed)
}

// ============================================================================
// HandlePaymentFailed Tests
// ============================================================================

func Test_HandlePaymentFailed_Should_Parse_Event(t *testing.T) {
	// Arrange
	svc := createEventHandlerTestServices()
	ctx := context.Background()
	_ = svc.eventHandlers.RegisterHandlers(ctx, svc.dispatcher)

	reservationID := shared.ReservationID("res-001")

	// Setup: create reservation
	_, _ = svc.reservationService.CreateReservation(
		ctx, reservationID, "guest-001", "room-101",
		eventHandlerValidDateRange(), eventHandlerValidMoney(), eventHandlerValidGuests(),
	)

	evt := payment.EventFailed{
		PaymentID:     "pay-001",
		ReservationID: reservationID,
		ErrorCode:     "declined",
		ErrorMsg:      "Card declined",
	}
	data, _ := json.Marshal(evt)

	// Act
	state, err := svc.dispatcher.triggerEvent(payment.EventTopicFailed, data)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "state must be completed", state, messaging.MessageStateCompleted)
}

func Test_HandlePaymentFailed_Should_Cancel_Reservation(t *testing.T) {
	// Arrange
	svc := createEventHandlerTestServices()
	ctx := context.Background()
	_ = svc.eventHandlers.RegisterHandlers(ctx, svc.dispatcher)

	reservationID := shared.ReservationID("res-001")

	// Setup
	_, _ = svc.reservationService.CreateReservation(
		ctx, reservationID, "guest-001", "room-101",
		eventHandlerValidDateRange(), eventHandlerValidMoney(), eventHandlerValidGuests(),
	)

	evt := payment.EventFailed{
		PaymentID:     "pay-001",
		ReservationID: reservationID,
		ErrorCode:     "declined",
		ErrorMsg:      "Card declined",
	}
	data, _ := json.Marshal(evt)

	// Act
	_, _ = svc.dispatcher.triggerEvent(payment.EventTopicFailed, data)

	// Assert
	storedRes, _ := svc.reservationRepo.Read(ctx, reservationID)
	assert.That(t, "reservation must be cancelled", storedRes.Status, reservation.StatusCancelled)
}

func Test_HandlePaymentFailed_With_Invalid_JSON_Should_Return_Failed(t *testing.T) {
	// Arrange
	svc := createEventHandlerTestServices()
	ctx := context.Background()
	_ = svc.eventHandlers.RegisterHandlers(ctx, svc.dispatcher)

	invalidData := []byte("{not valid json}")

	// Act
	state, err := svc.dispatcher.triggerEvent(payment.EventTopicFailed, invalidData)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "state must be failed", state, messaging.MessageStateFailed)
}

// ============================================================================
// Helper mock for event.Event interface check
// ============================================================================

type testEvent struct {
	topic string
}

func (e testEvent) Topic() string { return e.topic }

var _ event.Event = testEvent{} // compile-time interface check
