package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// EventHandlers manages cross-context event subscriptions.
// It wires up the event-driven communication between bounded contexts.
type EventHandlers struct {
	bookingService     *BookingService
	reservationService *reservation.Service
	paymentService     *payment.Service
}

// NewEventHandlers creates a new event handlers instance.
func NewEventHandlers(
	bookingSvc *BookingService,
	reservationSvc *reservation.Service,
	paymentSvc *payment.Service,
) *EventHandlers {
	return &EventHandlers{
		bookingService:     bookingSvc,
		reservationService: reservationSvc,
		paymentService:     paymentSvc,
	}
}

// RegisterHandlers registers all cross-context event subscriptions with the dispatcher.
func (h *EventHandlers) RegisterHandlers(ctx context.Context, dispatcher messaging.Dispatcher) error {
	// Payment context subscribes to reservation.created
	// When a reservation is created, initiate payment authorization
	if err := dispatcher.Subscribe(ctx, reservation.EventTopicCreated, service.Wrap(h.handleReservationCreated)); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", reservation.EventTopicCreated, err)
	}

	// Orchestration subscribes to payment.authorized
	// When payment is authorized, capture it
	if err := dispatcher.Subscribe(ctx, payment.EventTopicAuthorized, service.Wrap(h.handlePaymentAuthorized)); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", payment.EventTopicAuthorized, err)
	}

	// Reservation context subscribes to payment.captured
	// When payment is captured, confirm the reservation
	if err := dispatcher.Subscribe(ctx, payment.EventTopicCaptured, service.Wrap(h.handlePaymentCaptured)); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", payment.EventTopicCaptured, err)
	}

	// Orchestration subscribes to payment.failed
	// When payment fails, cancel the reservation as compensation
	if err := dispatcher.Subscribe(ctx, payment.EventTopicFailed, service.Wrap(h.handlePaymentFailed)); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", payment.EventTopicFailed, err)
	}

	return nil
}

// handleReservationCreated processes reservation.created events.
// It triggers payment authorization in the payment context.
func (h *EventHandlers) handleReservationCreated(msg messaging.Message) (messaging.MessageState, error) {
	var evt reservation.EventCreated
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return messaging.MessageStateFailed, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	ctx := context.Background()

	// Generate a payment ID from the reservation's underlying UUID (strip the "res-" prefix
	// so payment IDs read as "pay-<uuid>" rather than "pay-res-<uuid>").
	paymentID := payment.PaymentID(fmt.Sprintf("pay-%s", strings.TrimPrefix(string(evt.ReservationID), "res-")))

	// Authorize payment for the reservation
	_, err := h.paymentService.AuthorizePaymentForReservation(
		ctx,
		paymentID,
		shared.ReservationID(evt.ReservationID),
		evt.TotalAmount,
		"default", // Payment method - could be passed in event
	)
	if err != nil {
		// The payment service already publishes payment.failed event
		// which will trigger compensation
		return messaging.MessageStateFailed, fmt.Errorf("failed to authorize payment: %w", err)
	}

	return messaging.MessageStateCompleted, nil
}

// handlePaymentAuthorized processes payment.authorized events.
// It triggers payment capture.
func (h *EventHandlers) handlePaymentAuthorized(msg messaging.Message) (messaging.MessageState, error) {
	var evt payment.EventAuthorized
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return messaging.MessageStateFailed, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	ctx := context.Background()

	// Capture the authorized payment
	if err := h.bookingService.OnPaymentAuthorized(ctx, evt.PaymentID, evt.ReservationID); err != nil {
		return messaging.MessageStateFailed, fmt.Errorf("failed to handle payment authorized: %w", err)
	}

	return messaging.MessageStateCompleted, nil
}

// handlePaymentCaptured processes payment.captured events.
// It triggers reservation confirmation.
func (h *EventHandlers) handlePaymentCaptured(msg messaging.Message) (messaging.MessageState, error) {
	var evt payment.EventCaptured
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return messaging.MessageStateFailed, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	ctx := context.Background()

	// Confirm the reservation
	if err := h.bookingService.OnPaymentCaptured(ctx, evt.ReservationID); err != nil {
		return messaging.MessageStateFailed, fmt.Errorf("failed to confirm reservation: %w", err)
	}

	return messaging.MessageStateCompleted, nil
}

// handlePaymentFailed processes payment.failed events.
// It triggers reservation cancellation as compensation.
func (h *EventHandlers) handlePaymentFailed(msg messaging.Message) (messaging.MessageState, error) {
	var evt payment.EventFailed
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return messaging.MessageStateFailed, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	ctx := context.Background()

	// Cancel the reservation as compensation
	reason := fmt.Sprintf("payment_failed: %s - %s", evt.ErrorCode, evt.ErrorMsg)
	if err := h.bookingService.OnPaymentFailed(ctx, evt.ReservationID, reason); err != nil {
		return messaging.MessageStateFailed, fmt.Errorf("failed to cancel reservation: %w", err)
	}

	return messaging.MessageStateCompleted, nil
}
