package booking

import (
	"errors"
	"fmt"
	"time"
)

// PaymentID is a strongly-typed identifier for payments
type PaymentID string

// PaymentStatus represents the state of a payment
type PaymentStatus string

const (
	PaymentPending    PaymentStatus = "pending"
	PaymentAuthorized PaymentStatus = "authorized"
	PaymentCaptured   PaymentStatus = "captured"
	PaymentFailed     PaymentStatus = "failed"
	PaymentRefunded   PaymentStatus = "refunded"
)

// PaymentAttempt represents a single payment attempt (entity within Payment aggregate)
type PaymentAttempt struct {
	AttemptedAt time.Time
	Status      PaymentStatus
	ErrorCode   string
	ErrorMsg    string
}

// Payment is the aggregate root for payment processing
type Payment struct {
	ID            PaymentID
	ReservationID ReservationID
	Amount        Money
	Status        PaymentStatus
	PaymentMethod string
	TransactionID string // External payment gateway transaction ID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Attempts      []PaymentAttempt
}

// Payment errors
var (
	ErrInvalidPaymentTransition = errors.New("invalid payment state transition")
	ErrAlreadyAuthorized        = errors.New("payment already authorized")
	ErrNotAuthorized            = errors.New("payment not authorized")
	ErrAlreadyCaptured          = errors.New("payment already captured")
	ErrNotCaptured              = errors.New("payment not captured")
	ErrAlreadyRefunded          = errors.New("payment already refunded")
	ErrCannotRefund             = errors.New("can only refund captured payments")
)

// NewPayment creates a new payment in pending status
func NewPayment(id PaymentID, reservationID ReservationID, amount Money, method string) *Payment {
	return &Payment{
		ID:            id,
		ReservationID: reservationID,
		Amount:        amount,
		Status:        PaymentPending,
		PaymentMethod: method,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Attempts:      []PaymentAttempt{},
	}
}

// Authorize transitions the payment to authorized status
func (p *Payment) Authorize(transactionID string) error {
	if p.Status == PaymentAuthorized {
		return ErrAlreadyAuthorized
	}

	if p.Status != PaymentPending && p.Status != PaymentFailed {
		return fmt.Errorf("%w: cannot authorize from %s", ErrInvalidPaymentTransition, p.Status)
	}

	p.Status = PaymentAuthorized
	p.TransactionID = transactionID
	p.UpdatedAt = time.Now()
	p.addAttempt(PaymentAuthorized, "", "")

	return nil
}

// Capture transitions the payment to captured status (finalizes the payment)
func (p *Payment) Capture() error {
	if p.Status == PaymentCaptured {
		return ErrAlreadyCaptured
	}

	if p.Status != PaymentAuthorized {
		return ErrNotAuthorized
	}

	p.Status = PaymentCaptured
	p.UpdatedAt = time.Now()
	p.addAttempt(PaymentCaptured, "", "")

	return nil
}

// Fail marks the payment as failed with error details
func (p *Payment) Fail(errorCode, errorMsg string) error {
	if p.Status == PaymentCaptured || p.Status == PaymentRefunded {
		return fmt.Errorf("%w: cannot fail from %s", ErrInvalidPaymentTransition, p.Status)
	}

	p.Status = PaymentFailed
	p.UpdatedAt = time.Now()
	p.addAttempt(PaymentFailed, errorCode, errorMsg)

	return nil
}

// Refund transitions the payment to refunded status
func (p *Payment) Refund() error {
	if p.Status == PaymentRefunded {
		return ErrAlreadyRefunded
	}

	if p.Status != PaymentCaptured {
		return ErrCannotRefund
	}

	p.Status = PaymentRefunded
	p.UpdatedAt = time.Now()
	p.addAttempt(PaymentRefunded, "", "")

	return nil
}

// IsSuccessful returns true if the payment was successfully captured
func (p *Payment) IsSuccessful() bool {
	return p.Status == PaymentCaptured
}

// CanBeRetried returns true if the payment can be retried
func (p *Payment) CanBeRetried() bool {
	// Can only retry failed or pending payments
	if p.Status != PaymentFailed && p.Status != PaymentPending {
		return false
	}

	// Limit retry attempts to 3
	failedAttempts := 0
	for _, attempt := range p.Attempts {
		if attempt.Status == PaymentFailed {
			failedAttempts++
		}
	}

	return failedAttempts < 3
}

// addAttempt adds a payment attempt to the history
func (p *Payment) addAttempt(status PaymentStatus, errorCode, errorMsg string) {
	attempt := PaymentAttempt{
		AttemptedAt: time.Now(),
		Status:      status,
		ErrorCode:   errorCode,
		ErrorMsg:    errorMsg,
	}
	p.Attempts = append(p.Attempts, attempt)
}

// NewPaymentAttempt creates a new payment attempt entity
func NewPaymentAttempt(status PaymentStatus, errorCode, errorMsg string) PaymentAttempt {
	return PaymentAttempt{
		AttemptedAt: time.Now(),
		Status:      status,
		ErrorCode:   errorCode,
		ErrorMsg:    errorMsg,
	}
}
