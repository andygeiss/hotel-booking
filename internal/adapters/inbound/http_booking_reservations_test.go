package inbound_test

import (
	"context"
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/hotel-booking/internal/adapters/inbound"
	"github.com/andygeiss/hotel-booking/internal/adapters/outbound"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// ============================================================================
// Test Assets
// ============================================================================

//go:embed testdata/assets/templates/*.tmpl testdata/assets/static/css/*.css
var reservationsTestAssets embed.FS

// ============================================================================
// Helper Functions
// ============================================================================

func createReservationsTestService(repo *mockReservationRepository) *reservation.Service {
	availabilityChecker := outbound.NewRepositoryAvailabilityChecker(repo)
	eventPublisher := outbound.NewEventPublisher(messaging.NewInternalDispatcher())
	return reservation.NewService(repo, availabilityChecker, eventPublisher)
}

func createTestReservation(id, guestEmail, roomID string, checkIn, checkOut time.Time) *reservation.Reservation {
	dateRange := reservation.NewDateRange(checkIn, checkOut)
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	amount := shared.NewMoney(int64(nights)*9900, "USD")
	guests := []reservation.GuestInfo{reservation.NewGuestInfo("Test Guest", guestEmail, "+1234567890")}
	r, _ := reservation.NewReservation(
		shared.ReservationID(id),
		reservation.GuestID(guestEmail),
		reservation.RoomID(roomID),
		dateRange,
		amount,
		guests,
	)
	return r
}

func addAuthContext(req *http.Request, sessionID, email string) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, security.ContextSessionID, sessionID)
	ctx = context.WithValue(ctx, security.ContextEmail, email)
	ctx = context.WithValue(ctx, security.ContextIssuer, "https://issuer.example.com")
	ctx = context.WithValue(ctx, security.ContextName, "Test User")
	ctx = context.WithValue(ctx, security.ContextSubject, "user-subject-456")
	ctx = context.WithValue(ctx, security.ContextVerified, true)
	return req.WithContext(ctx)
}

// ============================================================================
// HttpViewReservations Tests
// ============================================================================

func Test_HttpViewReservations_Without_Session_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(reservationsTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createReservationsTestService(repo)

	handler := inbound.HttpViewReservations(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
	location := rec.Header().Get("Location")
	assert.That(t, "location must contain login", containsString(location, "/ui/login"), true)
}

func Test_HttpViewReservations_With_Empty_Session_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(reservationsTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createReservationsTestService(repo)

	handler := inbound.HttpViewReservations(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations", nil)
	req = addAuthContext(req, "", "")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
}

func Test_HttpViewReservations_With_Valid_Session_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(reservationsTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createReservationsTestService(repo)

	handler := inbound.HttpViewReservations(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewReservations_With_Valid_Session_Should_Render_App_Name(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(reservationsTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createReservationsTestService(repo)

	handler := inbound.HttpViewReservations(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain app name", containsString(bodyStr, "TestApp"), true)
}

func Test_HttpViewReservations_With_Valid_Session_Should_Render_Session_ID(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(reservationsTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createReservationsTestService(repo)

	handler := inbound.HttpViewReservations(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain session ID", containsString(bodyStr, "test-session-123"), true)
}

func Test_HttpViewReservations_With_Reservations_Should_Render_Reservation_List(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(reservationsTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createReservationsTestService(repo)

	// Create a test reservation
	checkIn := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	checkOut := checkIn.AddDate(0, 0, 3)
	res := createTestReservation("res-001", "test@example.com", "room-101", checkIn, checkOut)
	repo.reservations[shared.ReservationID("res-001")] = *res

	handler := inbound.HttpViewReservations(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain reservation ID", containsString(bodyStr, "res-001"), true)
	assert.That(t, "body must contain room ID", containsString(bodyStr, "room-101"), true)
}

func Test_HttpViewReservations_Should_Only_Show_Current_User_Reservations(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(reservationsTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createReservationsTestService(repo)

	// Create reservations for different users
	checkIn := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	checkOut := checkIn.AddDate(0, 0, 3)
	res1 := createTestReservation("res-001", "test@example.com", "room-101", checkIn, checkOut)
	res2 := createTestReservation("res-002", "other@example.com", "room-102", checkIn, checkOut)
	repo.reservations[shared.ReservationID("res-001")] = *res1
	repo.reservations[shared.ReservationID("res-002")] = *res2

	handler := inbound.HttpViewReservations(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain user's reservation", containsString(bodyStr, "res-001"), true)
	assert.That(t, "body must not contain other user's reservation", containsString(bodyStr, "res-002"), false)
}

func Test_HttpViewReservations_With_No_Reservations_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(reservationsTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createReservationsTestService(repo)

	handler := inbound.HttpViewReservations(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

// ============================================================================
// ReservationStatusClass Tests
// ============================================================================

func Test_ReservationStatusClass_Pending_Should_Return_Warning(t *testing.T) {
	// Arrange
	status := reservation.StatusPending

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be warning", result, "warning")
}

func Test_ReservationStatusClass_Confirmed_Should_Return_Info(t *testing.T) {
	// Arrange
	status := reservation.StatusConfirmed

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be info", result, "info")
}

func Test_ReservationStatusClass_Active_Should_Return_Primary(t *testing.T) {
	// Arrange
	status := reservation.StatusActive

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be primary", result, "primary")
}

func Test_ReservationStatusClass_Completed_Should_Return_Success(t *testing.T) {
	// Arrange
	status := reservation.StatusCompleted

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be success", result, "success")
}

func Test_ReservationStatusClass_Cancelled_Should_Return_Danger(t *testing.T) {
	// Arrange
	status := reservation.StatusCancelled

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be danger", result, "danger")
}

func Test_ReservationStatusClass_Unknown_Should_Return_Secondary(t *testing.T) {
	// Arrange
	status := reservation.ReservationStatus("unknown")

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be secondary", result, "secondary")
}

func testReservationStatusClass(status reservation.ReservationStatus) string {
	switch status {
	case reservation.StatusPending:
		return "warning"
	case reservation.StatusConfirmed:
		return "info"
	case reservation.StatusActive:
		return "primary"
	case reservation.StatusCompleted:
		return "success"
	case reservation.StatusCancelled:
		return "danger"
	default:
		return "secondary"
	}
}
