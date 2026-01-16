package outbound

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// MockNotificationService implements NotificationService by logging to console
type MockNotificationService struct {
	logger *slog.Logger
}

// NewMockNotificationService creates a new mock notification service
func NewMockNotificationService(logger *slog.Logger) *MockNotificationService {
	return &MockNotificationService{
		logger: logger,
	}
}

// SendReservationConfirmation logs a confirmation message
func (s *MockNotificationService) SendReservationConfirmation(
	ctx context.Context,
	reservation *booking.Reservation,
) error {
	if len(reservation.Guests) == 0 {
		return fmt.Errorf("no guests found in reservation")
	}

	primaryGuest := reservation.Guests[0]

	s.logger.Info("sending reservation confirmation email",
		"reservation_id", reservation.ID,
		"guest_email", primaryGuest.Email,
		"guest_name", primaryGuest.Name,
		"room_id", reservation.RoomID,
		"check_in", reservation.DateRange.CheckIn.Format("2006-01-02"),
		"check_out", reservation.DateRange.CheckOut.Format("2006-01-02"),
		"total_amount", reservation.TotalAmount.FormatAmount(),
	)

	return nil
}

// SendCancellationNotice logs a cancellation message
func (s *MockNotificationService) SendCancellationNotice(
	ctx context.Context,
	reservation *booking.Reservation,
	reason string,
) error {
	if len(reservation.Guests) == 0 {
		return fmt.Errorf("no guests found in reservation")
	}

	primaryGuest := reservation.Guests[0]

	s.logger.Info("sending cancellation notice email",
		"reservation_id", reservation.ID,
		"guest_email", primaryGuest.Email,
		"guest_name", primaryGuest.Name,
		"reason", reason,
	)

	return nil
}

// SendPaymentReceipt logs a payment receipt message
func (s *MockNotificationService) SendPaymentReceipt(
	ctx context.Context,
	payment *booking.Payment,
) error {
	s.logger.Info("sending payment receipt email",
		"payment_id", payment.ID,
		"reservation_id", payment.ReservationID,
		"amount", payment.Amount.FormatAmount(),
		"payment_method", payment.PaymentMethod,
		"transaction_id", payment.TransactionID,
	)

	return nil
}
