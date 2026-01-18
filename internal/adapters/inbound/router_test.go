package inbound_test

import (
	"context"
	"embed"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/mcp"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/hotel-booking/internal/adapters/inbound"
	"github.com/andygeiss/hotel-booking/internal/adapters/outbound"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
)

// Note: containsString helper is defined in http_index_test.go and shared across the test package

// ============================================================================
// Test Assets
// ============================================================================
// We embed testdata/assets and use fs.Sub to remap paths so that
// "assets/templates/*.tmpl" resolves correctly for Route().

//go:embed testdata/assets
var routerTestAssetsRaw embed.FS

func getRouterTestFS(t *testing.T) fs.FS {
	t.Helper()
	// Remap testdata/assets -> assets (root level)
	sub, err := fs.Sub(routerTestAssetsRaw, "testdata")
	if err != nil {
		t.Fatalf("failed to create sub filesystem: %v", err)
	}
	return sub
}

// mockReservationRepository is a simple in-memory mock for testing.
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

func createTestReservationService(t *testing.T) *reservation.Service {
	t.Helper()
	reservationRepo := newMockReservationRepository()
	availabilityChecker := outbound.NewRepositoryAvailabilityChecker(reservationRepo)
	eventPublisher := outbound.NewEventPublisher(messaging.NewInternalDispatcher())
	return reservation.NewService(reservationRepo, availabilityChecker, eventPublisher)
}

// ============================================================================
// Route Tests
// ============================================================================

func Test_Route_Should_Return_Non_Nil_Mux(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	reservationService := createTestReservationService(t)

	// Act
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                getRouterTestFS(t),
		Logger:             logger,
		ReservationService: reservationService,
	})

	// Assert
	assert.That(t, "mux must not be nil", mux != nil, true)
}

func Test_Route_Liveness_Endpoint_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	reservationService := createTestReservationService(t)
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                getRouterTestFS(t),
		Logger:             logger,
		ReservationService: reservationService,
	})

	req := httptest.NewRequest(http.MethodGet, "/liveness", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_Route_Readiness_Endpoint_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	reservationService := createTestReservationService(t)
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                getRouterTestFS(t),
		Logger:             logger,
		ReservationService: reservationService,
	})

	req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_Route_UI_Endpoint_Without_Session_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	reservationService := createTestReservationService(t)
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                getRouterTestFS(t),
		Logger:             logger,
		ReservationService: reservationService,
	})

	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
	location := rec.Header().Get("Location")
	assert.That(t, "location must contain login", containsString(location, "/ui/login"), true)
}

func Test_Route_Login_Endpoint_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	reservationService := createTestReservationService(t)
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                getRouterTestFS(t),
		Logger:             logger,
		ReservationService: reservationService,
	})

	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_Route_Login_Endpoint_Should_Return_HTML_Content(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	reservationService := createTestReservationService(t)
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                getRouterTestFS(t),
		Logger:             logger,
		ReservationService: reservationService,
	})

	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain app name", containsString(bodyStr, "TestApp"), true)
}

func Test_Route_Session_Endpoint_Without_Valid_Session_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	reservationService := createTestReservationService(t)
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                getRouterTestFS(t),
		Logger:             logger,
		ReservationService: reservationService,
	})

	// Note: Session endpoints with invalid/unknown session IDs redirect to login
	// The WithAuth middleware handles session validation and redirects unauthenticated requests
	req := httptest.NewRequest(http.MethodGet, "/ui/test-session-123/", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
}

func Test_Route_Session_Endpoint_Without_Trailing_Slash_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	reservationService := createTestReservationService(t)
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                getRouterTestFS(t),
		Logger:             logger,
		ReservationService: reservationService,
	})

	// Note: Session endpoints with invalid/unknown session IDs redirect to login
	// The WithAuth middleware handles session validation and redirects unauthenticated requests
	req := httptest.NewRequest(http.MethodGet, "/ui/test-session-123", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
}

func Test_Route_Unknown_Endpoint_Should_Return_404(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	reservationService := createTestReservationService(t)
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                getRouterTestFS(t),
		Logger:             logger,
		ReservationService: reservationService,
	})

	req := httptest.NewRequest(http.MethodGet, "/unknown/path", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 404", rec.Code, http.StatusNotFound)
}

// ============================================================================
// MCP Endpoint Tests
// ============================================================================

func Test_Route_MCP_Endpoint_Without_MCPServer_Should_Return_404(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	reservationService := createTestReservationService(t)
	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                getRouterTestFS(t),
		Logger:             logger,
		ReservationService: reservationService,
		// MCPServer is nil - endpoint should not be registered
	})

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 404", rec.Code, http.StatusNotFound)
}

func Test_Route_MCP_Endpoint_With_MCPServer_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	reservationService := createTestReservationService(t)
	mcpServer := mcp.NewServer("test-server", "1.0.0")

	mux := inbound.Route(inbound.RouterConfig{
		Ctx:                ctx,
		EFS:                getRouterTestFS(t),
		Logger:             logger,
		ReservationService: reservationService,
		MCPServer:          mcpServer,
		// Verifier is nil - no auth required for unit test
	})

	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(initReq))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}
