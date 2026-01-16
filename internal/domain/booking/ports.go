package booking

import (
	"context"

	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/cloud-native-utils/resource"
)

// Outbound Ports - Repositories

// ReservationRepository provides CRUD operations for reservations
type ReservationRepository resource.Access[ReservationID, Reservation]

// PaymentRepository provides CRUD operations for payments
type PaymentRepository resource.Access[PaymentID, Payment]

// Outbound Ports - External Services

// PaymentGateway handles payment processing with external providers
type PaymentGateway interface {
	// Authorize holds funds without capturing them
	Authorize(ctx context.Context, payment *Payment) (transactionID string, err error)
	// Capture finalizes an authorized payment
	Capture(ctx context.Context, transactionID string, amount Money) error
	// Refund returns funds to the customer
	Refund(ctx context.Context, transactionID string, amount Money) error
}

// AvailabilityChecker validates room availability for reservations
type AvailabilityChecker interface {
	// IsRoomAvailable checks if a room is available for the given date range
	IsRoomAvailable(ctx context.Context, roomID RoomID, dateRange DateRange) (bool, error)
	// GetOverlappingReservations returns all reservations that overlap with the given date range
	GetOverlappingReservations(ctx context.Context, roomID RoomID, dateRange DateRange) ([]*Reservation, error)
}

// NotificationService handles sending notifications to guests
type NotificationService interface {
	// SendReservationConfirmation sends a confirmation email to the guest
	SendReservationConfirmation(ctx context.Context, reservation *Reservation) error
	// SendCancellationNotice sends a cancellation notice to the guest
	SendCancellationNotice(ctx context.Context, reservation *Reservation, reason string) error
	// SendPaymentReceipt sends a payment receipt to the guest
	SendPaymentReceipt(ctx context.Context, payment *Payment) error
}

// EventPublisher publishes domain events
type EventPublisher event.EventPublisher
