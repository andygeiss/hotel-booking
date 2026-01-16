package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/andygeiss/cloud-native-utils/event"
)

// Event topics for reservation domain events
const (
	EventTopicReservationCreated   = "booking.reservation_created"
	EventTopicReservationConfirmed = "booking.reservation_confirmed"
	EventTopicReservationActivated = "booking.reservation_activated"
	EventTopicReservationCompleted = "booking.reservation_completed"
	EventTopicReservationCancelled = "booking.reservation_cancelled"
)

// EventReservationCreated represents a reservation created event
type EventReservationCreated struct {
	ReservationID ReservationID `json:"reservation_id"`
	GuestID       GuestID       `json:"guest_id"`
	RoomID        RoomID        `json:"room_id"`
	CheckIn       time.Time     `json:"check_in"`
	CheckOut      time.Time     `json:"check_out"`
	TotalAmount   Money         `json:"total_amount"`
}

// NewEventReservationCreated creates a new EventReservationCreated instance
func NewEventReservationCreated() *EventReservationCreated {
	return &EventReservationCreated{}
}

// Topic returns the topic for the event
func (e *EventReservationCreated) Topic() string {
	return EventTopicReservationCreated
}

// WithReservationID sets the ReservationID field
func (e *EventReservationCreated) WithReservationID(id ReservationID) *EventReservationCreated {
	e.ReservationID = id
	return e
}

// WithGuestID sets the GuestID field
func (e *EventReservationCreated) WithGuestID(id GuestID) *EventReservationCreated {
	e.GuestID = id
	return e
}

// WithRoomID sets the RoomID field
func (e *EventReservationCreated) WithRoomID(id RoomID) *EventReservationCreated {
	e.RoomID = id
	return e
}

// WithCheckIn sets the CheckIn field
func (e *EventReservationCreated) WithCheckIn(checkIn time.Time) *EventReservationCreated {
	e.CheckIn = checkIn
	return e
}

// WithCheckOut sets the CheckOut field
func (e *EventReservationCreated) WithCheckOut(checkOut time.Time) *EventReservationCreated {
	e.CheckOut = checkOut
	return e
}

// WithTotalAmount sets the TotalAmount field
func (e *EventReservationCreated) WithTotalAmount(amount Money) *EventReservationCreated {
	e.TotalAmount = amount
	return e
}

// EventReservationConfirmed represents a reservation confirmed event
type EventReservationConfirmed struct {
	ReservationID ReservationID `json:"reservation_id"`
	ConfirmedAt   time.Time     `json:"confirmed_at"`
}

// NewEventReservationConfirmed creates a new EventReservationConfirmed instance
func NewEventReservationConfirmed() *EventReservationConfirmed {
	return &EventReservationConfirmed{}
}

// Topic returns the topic for the event
func (e *EventReservationConfirmed) Topic() string {
	return EventTopicReservationConfirmed
}

// WithReservationID sets the ReservationID field
func (e *EventReservationConfirmed) WithReservationID(id ReservationID) *EventReservationConfirmed {
	e.ReservationID = id
	return e
}

// WithConfirmedAt sets the ConfirmedAt field
func (e *EventReservationConfirmed) WithConfirmedAt(t time.Time) *EventReservationConfirmed {
	e.ConfirmedAt = t
	return e
}

// EventReservationCancelled represents a reservation cancelled event
type EventReservationCancelled struct {
	ReservationID ReservationID `json:"reservation_id"`
	Reason        string        `json:"reason"`
	CancelledAt   time.Time     `json:"cancelled_at"`
}

// NewEventReservationCancelled creates a new EventReservationCancelled instance
func NewEventReservationCancelled() *EventReservationCancelled {
	return &EventReservationCancelled{}
}

// Topic returns the topic for the event
func (e *EventReservationCancelled) Topic() string {
	return EventTopicReservationCancelled
}

// WithReservationID sets the ReservationID field
func (e *EventReservationCancelled) WithReservationID(id ReservationID) *EventReservationCancelled {
	e.ReservationID = id
	return e
}

// WithReason sets the Reason field
func (e *EventReservationCancelled) WithReason(reason string) *EventReservationCancelled {
	e.Reason = reason
	return e
}

// WithCancelledAt sets the CancelledAt field
func (e *EventReservationCancelled) WithCancelledAt(t time.Time) *EventReservationCancelled {
	e.CancelledAt = t
	return e
}

// ReservationService handles reservation workflows
type ReservationService struct {
	reservationRepo     ReservationRepository
	availabilityChecker AvailabilityChecker
	publisher           event.EventPublisher
}

// NewReservationService creates a new ReservationService with dependencies
func NewReservationService(
	repo ReservationRepository,
	checker AvailabilityChecker,
	pub event.EventPublisher,
) *ReservationService {
	return &ReservationService{
		reservationRepo:     repo,
		availabilityChecker: checker,
		publisher:           pub,
	}
}

// CreateReservation creates a new pending reservation after checking availability
func (s *ReservationService) CreateReservation(
	ctx context.Context,
	id ReservationID,
	guestID GuestID,
	roomID RoomID,
	dateRange DateRange,
	amount Money,
	guests []GuestInfo,
) (*Reservation, error) {
	// 1. Check room availability
	available, err := s.availabilityChecker.IsRoomAvailable(ctx, roomID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}
	if !available {
		return nil, fmt.Errorf("room %s is not available for the selected dates", roomID)
	}

	// 2. Create reservation aggregate
	reservation, err := NewReservation(id, guestID, roomID, dateRange, amount, guests)
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// 3. Persist to repository
	if err := s.reservationRepo.Create(ctx, id, *reservation); err != nil {
		return nil, fmt.Errorf("failed to persist reservation: %w", err)
	}

	// 4. Publish domain event
	evt := NewEventReservationCreated().
		WithReservationID(id).
		WithGuestID(guestID).
		WithRoomID(roomID).
		WithCheckIn(dateRange.CheckIn).
		WithCheckOut(dateRange.CheckOut).
		WithTotalAmount(amount)

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return reservation, nil
}

// ConfirmReservation transitions a reservation to confirmed status
func (s *ReservationService) ConfirmReservation(ctx context.Context, id ReservationID) error {
	// 1. Load reservation from repository
	reservation, err := s.reservationRepo.Read(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read reservation: %w", err)
	}

	// 2. Confirm reservation (aggregate business logic)
	if err := reservation.Confirm(); err != nil {
		return fmt.Errorf("failed to confirm reservation: %w", err)
	}

	// 3. Update repository
	if err := s.reservationRepo.Update(ctx, id, *reservation); err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	// 4. Publish domain event
	evt := NewEventReservationConfirmed().
		WithReservationID(id).
		WithConfirmedAt(time.Now())

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// CancelReservation cancels a reservation with business rule validation
func (s *ReservationService) CancelReservation(ctx context.Context, id ReservationID, reason string) error {
	// 1. Load reservation from repository
	reservation, err := s.reservationRepo.Read(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read reservation: %w", err)
	}

	// 2. Cancel reservation (aggregate business logic validates rules)
	if err := reservation.Cancel(reason); err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}

	// 3. Update repository
	if err := s.reservationRepo.Update(ctx, id, *reservation); err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	// 4. Publish domain event
	evt := NewEventReservationCancelled().
		WithReservationID(id).
		WithReason(reason).
		WithCancelledAt(time.Now())

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// GetReservation retrieves a reservation by ID
func (s *ReservationService) GetReservation(ctx context.Context, id ReservationID) (*Reservation, error) {
	reservation, err := s.reservationRepo.Read(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to read reservation: %w", err)
	}
	return reservation, nil
}

// ListReservationsByGuest retrieves all reservations for a guest
func (s *ReservationService) ListReservationsByGuest(ctx context.Context, guestID GuestID) ([]*Reservation, error) {
	// List all reservations
	allReservations, err := s.reservationRepo.ReadAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list reservations: %w", err)
	}

	// Filter by guest ID
	var guestReservations []*Reservation
	for _, res := range allReservations {
		if res.GuestID == guestID {
			r := res // Create a copy to avoid pointer issues
			guestReservations = append(guestReservations, &r)
		}
	}

	return guestReservations, nil
}
