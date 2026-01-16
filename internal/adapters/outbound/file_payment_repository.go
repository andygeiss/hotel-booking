package outbound

import (
	"github.com/andygeiss/cloud-native-utils/resource"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// FilePaymentRepository implements the PaymentRepository port using JSON file storage
type FilePaymentRepository struct {
	resource.JsonFileAccess[booking.PaymentID, booking.Payment]
}

// NewFilePaymentRepository creates a new file-based payment repository
func NewFilePaymentRepository(filename string) booking.PaymentRepository {
	return resource.NewJsonFileAccess[booking.PaymentID, booking.Payment](filename)
}
