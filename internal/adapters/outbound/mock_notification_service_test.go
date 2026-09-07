package outbound_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/hotel-booking/internal/adapters/outbound"
	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// ============================================================================
// MockNotificationService Tests
// ============================================================================

func createTestReservation() *reservation.Reservation {
	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)

	return &reservation.Reservation{
		ID:          "res-001",
		GuestID:     "guest-001",
		RoomID:      "room-101",
		DateRange:   reservation.NewDateRange(checkIn, checkOut),
		Status:      reservation.StatusPending,
		TotalAmount: shared.NewMoney(30000, "USD"),
		Guests: []reservation.GuestInfo{
			reservation.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
		},
	}
}

func createTestPayment() *payment.Payment {
	return payment.NewPayment(
		"pay-001",
		"res-001",
		shared.NewMoney(30000, "USD"),
		"credit_card",
	)
}

func Test_MockNotificationService_SendReservationConfirmation_Should_Succeed(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := outbound.NewMockNotificationService(logger)
	ctx := context.Background()
	res := createTestReservation()

	// Act
	err := svc.SendReservationConfirmation(ctx, res)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
}

func Test_MockNotificationService_SendReservationConfirmation_No_Guests_Should_Return_Error(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := outbound.NewMockNotificationService(logger)
	ctx := context.Background()
	res := createTestReservation()
	res.Guests = []reservation.GuestInfo{}

	// Act
	err := svc.SendReservationConfirmation(ctx, res)

	// Assert
	assert.That(t, "error must not be nil for no guests", err != nil, true)
}

func Test_MockNotificationService_SendCancellationNotice_Should_Succeed(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := outbound.NewMockNotificationService(logger)
	ctx := context.Background()
	res := createTestReservation()

	// Act
	err := svc.SendCancellationNotice(ctx, res, "guest requested cancellation")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
}

func Test_MockNotificationService_SendCancellationNotice_No_Guests_Should_Return_Error(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := outbound.NewMockNotificationService(logger)
	ctx := context.Background()
	res := createTestReservation()
	res.Guests = []reservation.GuestInfo{}

	// Act
	err := svc.SendCancellationNotice(ctx, res, "test")

	// Assert
	assert.That(t, "error must not be nil for no guests", err != nil, true)
}

func Test_MockNotificationService_SendPaymentReceipt_Should_Succeed(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := outbound.NewMockNotificationService(logger)
	ctx := context.Background()
	pay := createTestPayment()

	// Act
	err := svc.SendPaymentReceipt(ctx, pay)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
}
