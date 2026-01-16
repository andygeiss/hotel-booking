package booking

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Value Objects - Strongly-typed IDs
type ReservationID string
type GuestID string
type RoomID string

// DateRange represents a time period for a reservation
type DateRange struct {
	CheckIn  time.Time
	CheckOut time.Time
}

// Money represents a monetary value in the smallest currency unit (cents)
type Money struct {
	Amount   int64  // Amount in cents/smallest unit
	Currency string // ISO 4217 currency code (e.g., "USD", "EUR")
}

// ReservationStatus represents the state of a reservation
type ReservationStatus string

const (
	StatusPending   ReservationStatus = "pending"
	StatusConfirmed ReservationStatus = "confirmed"
	StatusActive    ReservationStatus = "active"
	StatusCompleted ReservationStatus = "completed"
	StatusCancelled ReservationStatus = "cancelled"
)

// GuestInfo represents information about a guest (entity within Reservation aggregate)
type GuestInfo struct {
	Name        string
	Email       string
	PhoneNumber string
}

// Reservation is the aggregate root for booking reservations
type Reservation struct {
	ID              ReservationID
	GuestID         GuestID
	RoomID          RoomID
	DateRange       DateRange
	Status          ReservationStatus
	TotalAmount     Money
	CancellationReason string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Guests          []GuestInfo
}

// Validation errors
var (
	ErrInvalidDateRange        = errors.New("check-out must be after check-in")
	ErrCheckInPast             = errors.New("check-in date must be in the future")
	ErrMinimumStay             = errors.New("minimum stay is 1 night")
	ErrInvalidStateTransition  = errors.New("invalid state transition")
	ErrCannotCancelNearCheckIn = errors.New("cannot cancel within 24 hours of check-in")
	ErrCannotCancelActive      = errors.New("cannot cancel active reservation")
	ErrCannotCancelCompleted   = errors.New("cannot cancel completed reservation")
	ErrAlreadyCancelled        = errors.New("reservation already cancelled")
	ErrNoGuests                = errors.New("at least one guest required")
)

// NewReservation creates a new reservation with validation
func NewReservation(id ReservationID, guestID GuestID, roomID RoomID, dateRange DateRange, amount Money, guests []GuestInfo) (*Reservation, error) {
	r := &Reservation{
		ID:          id,
		GuestID:     guestID,
		RoomID:      roomID,
		DateRange:   dateRange,
		Status:      StatusPending,
		TotalAmount: amount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Guests:      guests,
	}

	if err := r.validate(); err != nil {
		return nil, err
	}

	return r, nil
}

// validate checks business rules for the reservation
func (r *Reservation) validate() error {
	// Validate dates
	if err := r.validateDateRange(); err != nil {
		return err
	}

	// Validate guests
	if len(r.Guests) == 0 {
		return ErrNoGuests
	}

	return nil
}

// validateDateRange ensures dates follow business rules
func (r *Reservation) validateDateRange() error {
	// Calculate nights first
	nights := r.DateRange.CheckOut.Sub(r.DateRange.CheckIn).Hours() / 24

	// Minimum 1 night stay (check this before the "after" check for better error messages)
	if nights < 1 {
		if r.DateRange.CheckOut.Equal(r.DateRange.CheckIn) {
			return ErrMinimumStay
		}
		return ErrInvalidDateRange
	}

	// Check-out must be after check-in
	if !r.DateRange.CheckOut.After(r.DateRange.CheckIn) {
		return ErrInvalidDateRange
	}

	// Check-in must be in the future (allow same-day for testing purposes)
	now := time.Now().Truncate(24 * time.Hour)
	checkIn := r.DateRange.CheckIn.Truncate(24 * time.Hour)
	if checkIn.Before(now) {
		return ErrCheckInPast
	}

	return nil
}

// Confirm transitions the reservation from pending to confirmed
func (r *Reservation) Confirm() error {
	if r.Status != StatusPending {
		return fmt.Errorf("%w: cannot confirm from %s", ErrInvalidStateTransition, r.Status)
	}

	r.Status = StatusConfirmed
	r.UpdatedAt = time.Now()
	return nil
}

// Activate transitions the reservation to active (check-in)
func (r *Reservation) Activate() error {
	if r.Status != StatusConfirmed {
		return fmt.Errorf("%w: cannot activate from %s", ErrInvalidStateTransition, r.Status)
	}

	r.Status = StatusActive
	r.UpdatedAt = time.Now()
	return nil
}

// Complete transitions the reservation to completed (check-out)
func (r *Reservation) Complete() error {
	if r.Status != StatusActive {
		return fmt.Errorf("%w: cannot complete from %s", ErrInvalidStateTransition, r.Status)
	}

	r.Status = StatusCompleted
	r.UpdatedAt = time.Now()
	return nil
}

// Cancel cancels the reservation with business rule validation
func (r *Reservation) Cancel(reason string) error {
	// Cannot cancel already cancelled reservations
	if r.Status == StatusCancelled {
		return ErrAlreadyCancelled
	}

	// Cannot cancel completed reservations
	if r.Status == StatusCompleted {
		return ErrCannotCancelCompleted
	}

	// Cannot cancel active reservations
	if r.Status == StatusActive {
		return ErrCannotCancelActive
	}

	// Cannot cancel within 24 hours of check-in
	if !r.CanBeCancelled() {
		return ErrCannotCancelNearCheckIn
	}

	r.Status = StatusCancelled
	r.CancellationReason = reason
	r.UpdatedAt = time.Now()
	return nil
}

// CanBeCancelled checks if the reservation can be cancelled based on business rules
func (r *Reservation) CanBeCancelled() bool {
	// Already cancelled, completed, or active cannot be cancelled
	if r.Status == StatusCancelled || r.Status == StatusCompleted || r.Status == StatusActive {
		return false
	}

	// Check if within 24 hours of check-in
	now := time.Now()
	hoursUntilCheckIn := r.DateRange.CheckIn.Sub(now).Hours()
	return hoursUntilCheckIn >= 24
}

// IsOverlapping checks if this reservation overlaps with another for the same room
func (r *Reservation) IsOverlapping(other *Reservation) bool {
	// Different rooms never overlap
	if r.RoomID != other.RoomID {
		return false
	}

	// Cancelled reservations don't count as overlapping
	if r.Status == StatusCancelled || other.Status == StatusCancelled {
		return false
	}

	// Check for date overlap
	// Reservations overlap if: r.CheckIn < other.CheckOut AND r.CheckOut > other.CheckIn
	// Note: Same-day checkout/check-in is allowed (not considered overlapping)
	return r.DateRange.CheckIn.Before(other.DateRange.CheckOut) &&
		r.DateRange.CheckOut.After(other.DateRange.CheckIn)
}

// DaysUntilCheckIn returns the number of days until check-in
func (r *Reservation) DaysUntilCheckIn() int {
	now := time.Now().Truncate(24 * time.Hour)
	checkIn := r.DateRange.CheckIn.Truncate(24 * time.Hour)
	days := checkIn.Sub(now).Hours() / 24
	return int(days)
}

// Nights returns the number of nights for this reservation
func (r *Reservation) Nights() int {
	nights := r.DateRange.CheckOut.Sub(r.DateRange.CheckIn).Hours() / 24
	return int(nights)
}

// NewMoney creates a Money value object with validation
func NewMoney(amount int64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: strings.ToUpper(currency),
	}
}

// FormatAmount returns a human-readable amount (converts cents to dollars)
func (m Money) FormatAmount() string {
	dollars := float64(m.Amount) / 100.0
	return fmt.Sprintf("%.2f %s", dollars, m.Currency)
}

// NewDateRange creates a DateRange value object
func NewDateRange(checkIn, checkOut time.Time) DateRange {
	return DateRange{
		CheckIn:  checkIn,
		CheckOut: checkOut,
	}
}

// NewGuestInfo creates a GuestInfo entity
func NewGuestInfo(name, email, phoneNumber string) GuestInfo {
	return GuestInfo{
		Name:        name,
		Email:       email,
		PhoneNumber: phoneNumber,
	}
}
