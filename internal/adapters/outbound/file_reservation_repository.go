package outbound

import (
	"github.com/andygeiss/cloud-native-utils/resource"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// FileReservationRepository implements the ReservationRepository port using JSON file storage
type FileReservationRepository struct {
	resource.JsonFileAccess[booking.ReservationID, booking.Reservation]
}

// NewFileReservationRepository creates a new file-based reservation repository
func NewFileReservationRepository(filename string) booking.ReservationRepository {
	return resource.NewJsonFileAccess[booking.ReservationID, booking.Reservation](filename)
}
