// Package shared contains types that are shared across bounded contexts.
// This is the "Shared Kernel" in DDD terminology - types that multiple
// bounded contexts need to reference but don't own.
package shared

import (
	"fmt"
	"strings"

	"github.com/andygeiss/cloud-native-utils/security"
)

// ReservationID is a strongly-typed identifier for reservations.
// Shared because Payment needs to reference it.
type ReservationID string

// NewReservationID generates a new ReservationID with the canonical "res-" prefix.
func NewReservationID() ReservationID {
	return ReservationID("res-" + security.GenerateID())
}

// Money represents a monetary value in the smallest currency unit (cents).
// Shared because both Reservation and Payment use it.
type Money struct {
	Currency string // ISO 4217 currency code (e.g., "USD", "EUR")
	Amount   int64  // Amount in cents/smallest unit
}

// NewMoney creates a Money value object with validation.
func NewMoney(amount int64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: strings.ToUpper(currency),
	}
}

// FormatAmount returns a human-readable amount (converts cents to dollars).
func (m Money) FormatAmount() string {
	dollars := float64(m.Amount) / 100.0
	return fmt.Sprintf("%.2f %s", dollars, m.Currency)
}
