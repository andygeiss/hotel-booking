package outbound

import (
	"context"
	"fmt"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// RepositoryAvailabilityChecker implements AvailabilityChecker by querying the reservation repository
type RepositoryAvailabilityChecker struct {
	reservationRepo booking.ReservationRepository
}

// NewRepositoryAvailabilityChecker creates a new availability checker
func NewRepositoryAvailabilityChecker(repo booking.ReservationRepository) *RepositoryAvailabilityChecker {
	return &RepositoryAvailabilityChecker{
		reservationRepo: repo,
	}
}

// IsRoomAvailable checks if a room is available for the given date range
func (c *RepositoryAvailabilityChecker) IsRoomAvailable(
	ctx context.Context,
	roomID booking.RoomID,
	dateRange booking.DateRange,
) (bool, error) {
	overlapping, err := c.GetOverlappingReservations(ctx, roomID, dateRange)
	if err != nil {
		return false, fmt.Errorf("failed to check overlaps: %w", err)
	}

	// Room is available if there are no overlapping reservations
	return len(overlapping) == 0, nil
}

// GetOverlappingReservations returns all reservations that overlap with the given date range
func (c *RepositoryAvailabilityChecker) GetOverlappingReservations(
	ctx context.Context,
	roomID booking.RoomID,
	dateRange booking.DateRange,
) ([]*booking.Reservation, error) {
	// Get all reservations
	allReservations, err := c.reservationRepo.ReadAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read reservations: %w", err)
	}

	// Create a temporary reservation to use IsOverlapping method
	tempReservation := &booking.Reservation{
		RoomID:    roomID,
		DateRange: dateRange,
		Status:    booking.StatusPending, // Status doesn't matter for overlap check
	}

	// Filter for overlapping reservations
	var overlapping []*booking.Reservation
	for _, res := range allReservations {
		r := res // Create a copy
		if tempReservation.IsOverlapping(&r) {
			overlapping = append(overlapping, &r)
		}
	}

	return overlapping, nil
}
