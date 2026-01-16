package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/andygeiss/cloud-native-utils/event"
)

// Event topics for payment domain events
const (
	EventTopicPaymentAuthorized = "payment.payment_authorized"
	EventTopicPaymentCaptured   = "payment.payment_captured"
	EventTopicPaymentFailed     = "payment.payment_failed"
	EventTopicPaymentRefunded   = "payment.payment_refunded"
)

// EventPaymentAuthorized represents a payment authorized event
type EventPaymentAuthorized struct {
	PaymentID     PaymentID     `json:"payment_id"`
	ReservationID ReservationID `json:"reservation_id"`
	Amount        Money         `json:"amount"`
	TransactionID string        `json:"transaction_id"`
}

// NewEventPaymentAuthorized creates a new EventPaymentAuthorized instance
func NewEventPaymentAuthorized() *EventPaymentAuthorized {
	return &EventPaymentAuthorized{}
}

// Topic returns the topic for the event
func (e *EventPaymentAuthorized) Topic() string {
	return EventTopicPaymentAuthorized
}

// WithPaymentID sets the PaymentID field
func (e *EventPaymentAuthorized) WithPaymentID(id PaymentID) *EventPaymentAuthorized {
	e.PaymentID = id
	return e
}

// WithReservationID sets the ReservationID field
func (e *EventPaymentAuthorized) WithReservationID(id ReservationID) *EventPaymentAuthorized {
	e.ReservationID = id
	return e
}

// WithAmount sets the Amount field
func (e *EventPaymentAuthorized) WithAmount(amount Money) *EventPaymentAuthorized {
	e.Amount = amount
	return e
}

// WithTransactionID sets the TransactionID field
func (e *EventPaymentAuthorized) WithTransactionID(id string) *EventPaymentAuthorized {
	e.TransactionID = id
	return e
}

// EventPaymentCaptured represents a payment captured event
type EventPaymentCaptured struct {
	PaymentID     PaymentID     `json:"payment_id"`
	ReservationID ReservationID `json:"reservation_id"`
	Amount        Money         `json:"amount"`
	CapturedAt    time.Time     `json:"captured_at"`
}

// NewEventPaymentCaptured creates a new EventPaymentCaptured instance
func NewEventPaymentCaptured() *EventPaymentCaptured {
	return &EventPaymentCaptured{}
}

// Topic returns the topic for the event
func (e *EventPaymentCaptured) Topic() string {
	return EventTopicPaymentCaptured
}

// WithPaymentID sets the PaymentID field
func (e *EventPaymentCaptured) WithPaymentID(id PaymentID) *EventPaymentCaptured {
	e.PaymentID = id
	return e
}

// WithReservationID sets the ReservationID field
func (e *EventPaymentCaptured) WithReservationID(id ReservationID) *EventPaymentCaptured {
	e.ReservationID = id
	return e
}

// WithAmount sets the Amount field
func (e *EventPaymentCaptured) WithAmount(amount Money) *EventPaymentCaptured {
	e.Amount = amount
	return e
}

// WithCapturedAt sets the CapturedAt field
func (e *EventPaymentCaptured) WithCapturedAt(t time.Time) *EventPaymentCaptured {
	e.CapturedAt = t
	return e
}

// EventPaymentFailed represents a payment failed event
type EventPaymentFailed struct {
	PaymentID     PaymentID     `json:"payment_id"`
	ReservationID ReservationID `json:"reservation_id"`
	ErrorCode     string        `json:"error_code"`
	ErrorMsg      string        `json:"error_msg"`
}

// NewEventPaymentFailed creates a new EventPaymentFailed instance
func NewEventPaymentFailed() *EventPaymentFailed {
	return &EventPaymentFailed{}
}

// Topic returns the topic for the event
func (e *EventPaymentFailed) Topic() string {
	return EventTopicPaymentFailed
}

// WithPaymentID sets the PaymentID field
func (e *EventPaymentFailed) WithPaymentID(id PaymentID) *EventPaymentFailed {
	e.PaymentID = id
	return e
}

// WithReservationID sets the ReservationID field
func (e *EventPaymentFailed) WithReservationID(id ReservationID) *EventPaymentFailed {
	e.ReservationID = id
	return e
}

// WithErrorCode sets the ErrorCode field
func (e *EventPaymentFailed) WithErrorCode(code string) *EventPaymentFailed {
	e.ErrorCode = code
	return e
}

// WithErrorMsg sets the ErrorMsg field
func (e *EventPaymentFailed) WithErrorMsg(msg string) *EventPaymentFailed {
	e.ErrorMsg = msg
	return e
}

// EventPaymentRefunded represents a payment refunded event
type EventPaymentRefunded struct {
	PaymentID     PaymentID     `json:"payment_id"`
	ReservationID ReservationID `json:"reservation_id"`
	Amount        Money         `json:"amount"`
	RefundedAt    time.Time     `json:"refunded_at"`
}

// NewEventPaymentRefunded creates a new EventPaymentRefunded instance
func NewEventPaymentRefunded() *EventPaymentRefunded {
	return &EventPaymentRefunded{}
}

// Topic returns the topic for the event
func (e *EventPaymentRefunded) Topic() string {
	return EventTopicPaymentRefunded
}

// WithPaymentID sets the PaymentID field
func (e *EventPaymentRefunded) WithPaymentID(id PaymentID) *EventPaymentRefunded {
	e.PaymentID = id
	return e
}

// WithReservationID sets the ReservationID field
func (e *EventPaymentRefunded) WithReservationID(id ReservationID) *EventPaymentRefunded {
	e.ReservationID = id
	return e
}

// WithAmount sets the Amount field
func (e *EventPaymentRefunded) WithAmount(amount Money) *EventPaymentRefunded {
	e.Amount = amount
	return e
}

// WithRefundedAt sets the RefundedAt field
func (e *EventPaymentRefunded) WithRefundedAt(t time.Time) *EventPaymentRefunded {
	e.RefundedAt = t
	return e
}

// PaymentService handles payment workflows
type PaymentService struct {
	paymentRepo    PaymentRepository
	paymentGateway PaymentGateway
	publisher      event.EventPublisher
}

// NewPaymentService creates a new PaymentService with dependencies
func NewPaymentService(
	repo PaymentRepository,
	gateway PaymentGateway,
	pub event.EventPublisher,
) *PaymentService {
	return &PaymentService{
		paymentRepo:    repo,
		paymentGateway: gateway,
		publisher:      pub,
	}
}

// AuthorizePayment creates a payment and authorizes it with the gateway
func (s *PaymentService) AuthorizePayment(
	ctx context.Context,
	id PaymentID,
	reservationID ReservationID,
	amount Money,
	method string,
) (*Payment, error) {
	// 1. Create payment aggregate
	payment := NewPayment(id, reservationID, amount, method)

	// 2. Authorize with payment gateway
	transactionID, err := s.paymentGateway.Authorize(ctx, payment)
	if err != nil {
		// Mark payment as failed
		_ = payment.Fail("gateway_error", err.Error())

		// Persist failed payment
		if persistErr := s.paymentRepo.Create(ctx, id, *payment); persistErr != nil {
			return nil, fmt.Errorf("failed to persist failed payment: %w", persistErr)
		}

		// Publish failure event
		failEvt := NewEventPaymentFailed().
			WithPaymentID(id).
			WithReservationID(reservationID).
			WithErrorCode("gateway_error").
			WithErrorMsg(err.Error())

		_ = s.publisher.Publish(ctx, failEvt)

		return nil, fmt.Errorf("payment authorization failed: %w", err)
	}

	// 3. Update payment with transaction ID
	if err := payment.Authorize(transactionID); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// 4. Persist to repository
	if err := s.paymentRepo.Create(ctx, id, *payment); err != nil {
		return nil, fmt.Errorf("failed to persist payment: %w", err)
	}

	// 5. Publish success event
	evt := NewEventPaymentAuthorized().
		WithPaymentID(id).
		WithReservationID(reservationID).
		WithAmount(amount).
		WithTransactionID(transactionID)

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return payment, nil
}

// CapturePayment captures an authorized payment
func (s *PaymentService) CapturePayment(ctx context.Context, id PaymentID) error {
	// 1. Load payment from repository
	payment, err := s.paymentRepo.Read(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read payment: %w", err)
	}

	// 2. Capture with payment gateway
	if err := s.paymentGateway.Capture(ctx, payment.TransactionID, payment.Amount); err != nil {
		// Mark as failed
		_ = payment.Fail("capture_failed", err.Error())
		_ = s.paymentRepo.Update(ctx, id, *payment)

		// Publish failure event
		failEvt := NewEventPaymentFailed().
			WithPaymentID(id).
			WithReservationID(payment.ReservationID).
			WithErrorCode("capture_failed").
			WithErrorMsg(err.Error())

		_ = s.publisher.Publish(ctx, failEvt)

		return fmt.Errorf("payment capture failed: %w", err)
	}

	// 3. Update payment status
	if err := payment.Capture(); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// 4. Update repository
	if err := s.paymentRepo.Update(ctx, id, *payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// 5. Publish success event
	evt := NewEventPaymentCaptured().
		WithPaymentID(id).
		WithReservationID(payment.ReservationID).
		WithAmount(payment.Amount).
		WithCapturedAt(time.Now())

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// RefundPayment processes a refund for a captured payment
func (s *PaymentService) RefundPayment(ctx context.Context, id PaymentID) error {
	// 1. Load payment from repository
	payment, err := s.paymentRepo.Read(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read payment: %w", err)
	}

	// 2. Refund with payment gateway
	if err := s.paymentGateway.Refund(ctx, payment.TransactionID, payment.Amount); err != nil {
		return fmt.Errorf("payment refund failed: %w", err)
	}

	// 3. Update payment status
	if err := payment.Refund(); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// 4. Update repository
	if err := s.paymentRepo.Update(ctx, id, *payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// 5. Publish event
	evt := NewEventPaymentRefunded().
		WithPaymentID(id).
		WithReservationID(payment.ReservationID).
		WithAmount(payment.Amount).
		WithRefundedAt(time.Now())

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// GetPayment retrieves a payment by ID
func (s *PaymentService) GetPayment(ctx context.Context, id PaymentID) (*Payment, error) {
	payment, err := s.paymentRepo.Read(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to read payment: %w", err)
	}
	return payment, nil
}
