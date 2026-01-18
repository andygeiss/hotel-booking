package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/logging"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/hotel-booking/internal/adapters/inbound"
	"github.com/andygeiss/hotel-booking/internal/adapters/outbound"
	"github.com/andygeiss/hotel-booking/internal/domain/orchestration"
	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// Benchmarks for Profile-Guided Optimization (PGO).
// Run with: just profile
// This generates cpuprofile.pprof for optimized builds.

// mockReservationRepository is a simple in-memory mock for benchmarking.
type mockReservationRepository struct {
	reservations map[reservation.ReservationID]reservation.Reservation
}

func newMockReservationRepository() *mockReservationRepository {
	return &mockReservationRepository{
		reservations: make(map[reservation.ReservationID]reservation.Reservation),
	}
}

func (m *mockReservationRepository) Create(ctx context.Context, id reservation.ReservationID, res reservation.Reservation) error {
	m.reservations[id] = res
	return nil
}

func (m *mockReservationRepository) Read(ctx context.Context, id reservation.ReservationID) (*reservation.Reservation, error) {
	res, ok := m.reservations[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &res, nil
}

func (m *mockReservationRepository) Update(ctx context.Context, id reservation.ReservationID, res reservation.Reservation) error {
	m.reservations[id] = res
	return nil
}

func (m *mockReservationRepository) Delete(ctx context.Context, id reservation.ReservationID) error {
	delete(m.reservations, id)
	return nil
}

func (m *mockReservationRepository) ReadAll(ctx context.Context) ([]reservation.Reservation, error) {
	result := make([]reservation.Reservation, 0, len(m.reservations))
	for _, res := range m.reservations {
		result = append(result, res)
	}
	return result, nil
}

func createBenchReservationService() *reservation.Service {
	reservationRepo := newMockReservationRepository()
	availabilityChecker := outbound.NewRepositoryAvailabilityChecker(reservationRepo)
	eventPublisher := outbound.NewEventPublisher(messaging.NewInternalDispatcher())
	return reservation.NewService(reservationRepo, availabilityChecker, eventPublisher)
}

func Benchmark_Server_Integration_Liveness_Should_Respond_Fast(b *testing.B) {
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	reservationService := createBenchReservationService()
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                efs,
		Logger:             logger,
		ReservationService: reservationService,
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for b.Loop() {
		resp, _ := client.Get(server.URL + "/liveness")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}
}

func Benchmark_Server_Integration_Static_CSS_Should_Serve_Fast(b *testing.B) {
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	reservationService := createBenchReservationService()
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                efs,
		Logger:             logger,
		ReservationService: reservationService,
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for b.Loop() {
		resp, _ := client.Get(server.URL + "/static/css/base.css")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}
}

func Benchmark_Server_Integration_Login_Page_Should_Render_Fast(b *testing.B) {
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	reservationService := createBenchReservationService()
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                efs,
		Logger:             logger,
		ReservationService: reservationService,
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for b.Loop() {
		resp, _ := client.Get(server.URL + "/ui/login")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}
}

func Benchmark_Server_Integration_MCP_Tools_List_Should_Be_Fast(b *testing.B) {
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	reservationService := createBenchReservationService()
	paymentService := createBenchPaymentService()
	availabilityChecker := outbound.NewRepositoryAvailabilityChecker(newMockReservationRepository())

	// Build MCP server with tools registered.
	mcpServer := buildMCPServer(reservationService, availabilityChecker, paymentService)

	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                efs,
		Logger:             logger,
		ReservationService: reservationService,
		MCPServer:          mcpServer,
		// Verifier is nil - no auth for benchmarks
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Initialize MCP session first.
	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"bench","version":"1.0"}}}`
	resp, _ := client.Post(server.URL+"/mcp", "application/json", strings.NewReader(initReq))
	if resp != nil {
		_ = resp.Body.Close()
	}

	toolsListReq := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`

	for b.Loop() {
		resp, _ := client.Post(server.URL+"/mcp", "application/json", strings.NewReader(toolsListReq))
		if resp != nil {
			_ = resp.Body.Close()
		}
	}
}

// ============================================================================
// Domain Benchmarks - Reservation Context
// ============================================================================

func benchValidDateRange() reservation.DateRange {
	checkIn := time.Now().Add(48 * time.Hour).Truncate(24 * time.Hour)
	checkOut := checkIn.Add(72 * time.Hour)
	return reservation.NewDateRange(checkIn, checkOut)
}

func benchValidGuests() []reservation.GuestInfo {
	return []reservation.GuestInfo{
		reservation.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
	}
}

func benchValidMoney() shared.Money {
	return shared.NewMoney(10000, "USD")
}

func Benchmark_Reservation_Create_Should_Be_Fast(b *testing.B) {
	ctx := context.Background()
	reservationService := createBenchReservationService()
	dateRange := benchValidDateRange()
	guests := benchValidGuests()
	amount := benchValidMoney()

	for b.Loop() {
		id := reservation.ReservationID(fmt.Sprintf("res-%d", b.N))
		_, _ = reservationService.CreateReservation(ctx, id, "guest-001", "room-101", dateRange, amount, guests)
	}
}

func Benchmark_Reservation_Confirm_Should_Be_Fast(b *testing.B) {
	ctx := context.Background()
	reservationService := createBenchReservationService()
	dateRange := benchValidDateRange()
	guests := benchValidGuests()
	amount := benchValidMoney()

	// Pre-create reservations
	for i := 0; i < b.N; i++ {
		id := reservation.ReservationID(fmt.Sprintf("res-%d", i))
		_, _ = reservationService.CreateReservation(ctx, id, "guest-001", "room-101", dateRange, amount, guests)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := reservation.ReservationID(fmt.Sprintf("res-%d", i))
		_ = reservationService.ConfirmReservation(ctx, id)
	}
}

func Benchmark_Reservation_StateTransition_Full_Lifecycle_Should_Be_Fast(b *testing.B) {
	ctx := context.Background()
	reservationService := createBenchReservationService()
	dateRange := benchValidDateRange()
	guests := benchValidGuests()
	amount := benchValidMoney()

	for b.Loop() {
		id := reservation.ReservationID(fmt.Sprintf("res-%d", b.N))
		_, _ = reservationService.CreateReservation(ctx, id, "guest-001", "room-101", dateRange, amount, guests)
		_ = reservationService.ConfirmReservation(ctx, id)
		_ = reservationService.ActivateReservation(ctx, id)
		_ = reservationService.CompleteReservation(ctx, id)
	}
}

// ============================================================================
// Domain Benchmarks - Payment Context
// ============================================================================

// mockPaymentRepository for benchmarking
type mockPaymentRepository struct {
	payments map[payment.PaymentID]payment.Payment
}

func newMockPaymentRepository() *mockPaymentRepository {
	return &mockPaymentRepository{
		payments: make(map[payment.PaymentID]payment.Payment),
	}
}

func (m *mockPaymentRepository) Create(ctx context.Context, id payment.PaymentID, p payment.Payment) error {
	m.payments[id] = p
	return nil
}

func (m *mockPaymentRepository) Read(ctx context.Context, id payment.PaymentID) (*payment.Payment, error) {
	p, ok := m.payments[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &p, nil
}

func (m *mockPaymentRepository) Update(ctx context.Context, id payment.PaymentID, p payment.Payment) error {
	m.payments[id] = p
	return nil
}

func (m *mockPaymentRepository) Delete(ctx context.Context, id payment.PaymentID) error {
	delete(m.payments, id)
	return nil
}

func (m *mockPaymentRepository) ReadAll(ctx context.Context) ([]payment.Payment, error) {
	result := make([]payment.Payment, 0, len(m.payments))
	for _, p := range m.payments {
		result = append(result, p)
	}
	return result, nil
}

// mockPaymentGateway for benchmarking (instant responses)
type mockPaymentGateway struct{}

func (m *mockPaymentGateway) Authorize(ctx context.Context, p *payment.Payment) (string, error) {
	return "tx-bench-12345", nil
}

func (m *mockPaymentGateway) Capture(ctx context.Context, transactionID string, amount shared.Money) error {
	return nil
}

func (m *mockPaymentGateway) Refund(ctx context.Context, transactionID string, amount shared.Money) error {
	return nil
}

func createBenchPaymentService() *payment.Service {
	paymentRepo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{}
	eventPublisher := outbound.NewEventPublisher(messaging.NewInternalDispatcher())
	return payment.NewService(paymentRepo, gateway, eventPublisher)
}

func Benchmark_Payment_Authorize_Should_Be_Fast(b *testing.B) {
	ctx := context.Background()
	paymentService := createBenchPaymentService()
	amount := benchValidMoney()

	for b.Loop() {
		id := payment.PaymentID(fmt.Sprintf("pay-%d", b.N))
		_, _ = paymentService.AuthorizePayment(ctx, id, "res-001", amount, "credit_card")
	}
}

func Benchmark_Payment_Capture_Should_Be_Fast(b *testing.B) {
	ctx := context.Background()
	paymentService := createBenchPaymentService()
	amount := benchValidMoney()

	// Pre-create and authorize payments
	for i := 0; i < b.N; i++ {
		id := payment.PaymentID(fmt.Sprintf("pay-%d", i))
		_, _ = paymentService.AuthorizePayment(ctx, id, "res-001", amount, "credit_card")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := payment.PaymentID(fmt.Sprintf("pay-%d", i))
		_ = paymentService.CapturePayment(ctx, id)
	}
}

func Benchmark_Payment_Full_Lifecycle_Should_Be_Fast(b *testing.B) {
	ctx := context.Background()
	paymentService := createBenchPaymentService()
	amount := benchValidMoney()

	for b.Loop() {
		id := payment.PaymentID(fmt.Sprintf("pay-%d", b.N))
		_, _ = paymentService.AuthorizePayment(ctx, id, "res-001", amount, "credit_card")
		_ = paymentService.CapturePayment(ctx, id)
		_ = paymentService.RefundPayment(ctx, id)
	}
}

// ============================================================================
// Domain Benchmarks - Orchestration Context
// ============================================================================

// mockNotificationService for benchmarking
type mockNotificationService struct{}

func (m *mockNotificationService) SendReservationConfirmation(ctx context.Context, r *reservation.Reservation) error {
	return nil
}

func (m *mockNotificationService) SendCancellationNotice(ctx context.Context, r *reservation.Reservation, reason string) error {
	return nil
}

func (m *mockNotificationService) SendPaymentReceipt(ctx context.Context, p *payment.Payment) error {
	return nil
}

func createBenchBookingService() *orchestration.BookingService {
	reservationService := createBenchReservationService()
	paymentService := createBenchPaymentService()
	notificationService := &mockNotificationService{}
	return orchestration.NewBookingService(reservationService, paymentService, notificationService)
}

func Benchmark_Orchestration_InitiateBooking_Should_Be_Fast(b *testing.B) {
	ctx := context.Background()
	bookingService := createBenchBookingService()
	dateRange := benchValidDateRange()
	guests := benchValidGuests()
	amount := benchValidMoney()

	for b.Loop() {
		id := shared.ReservationID(fmt.Sprintf("res-%d", b.N))
		_, _ = bookingService.InitiateBooking(ctx, id, "guest-001", "room-101", dateRange, amount, guests)
	}
}

func Benchmark_Orchestration_CompleteBooking_Should_Be_Fast(b *testing.B) {
	ctx := context.Background()
	bookingService := createBenchBookingService()
	dateRange := benchValidDateRange()
	guests := benchValidGuests()
	amount := benchValidMoney()

	for b.Loop() {
		resID := shared.ReservationID(fmt.Sprintf("res-%d", b.N))
		payID := payment.PaymentID(fmt.Sprintf("pay-%d", b.N))
		_, _ = bookingService.CompleteBooking(ctx, resID, payID, "guest-001", "room-101", dateRange, amount, guests, "credit_card")
	}
}
