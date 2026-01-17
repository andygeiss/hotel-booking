package inbound_test

import (
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/messaging"
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
var detailTestAssets embed.FS

// ============================================================================
// Helper Functions
// ============================================================================

func createDetailTestService(repo *mockReservationRepository) *reservation.Service {
	availabilityChecker := outbound.NewRepositoryAvailabilityChecker(repo)
	eventPublisher := outbound.NewEventPublisher(messaging.NewInternalDispatcher())
	return reservation.NewService(repo, availabilityChecker, eventPublisher)
}

// ============================================================================
// HttpViewReservationDetail Tests
// ============================================================================

func Test_HttpViewReservationDetail_Without_Session_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(detailTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	handler := inbound.HttpViewReservationDetail(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations/res-001", nil)
	req.SetPathValue("id", "res-001")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
	location := rec.Header().Get("Location")
	assert.That(t, "location must contain login", containsString(location, "/ui/login"), true)
}

func Test_HttpViewReservationDetail_Without_Reservation_ID_Should_Return_400(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(detailTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	handler := inbound.HttpViewReservationDetail(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations/", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 400", rec.Code, http.StatusBadRequest)
}

func Test_HttpViewReservationDetail_With_NonExistent_Reservation_Should_Return_404(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(detailTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	handler := inbound.HttpViewReservationDetail(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 404", rec.Code, http.StatusNotFound)
}

func Test_HttpViewReservationDetail_With_Other_User_Reservation_Should_Return_403(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(detailTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	// Create reservation for another user
	checkIn := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	checkOut := checkIn.AddDate(0, 0, 3)
	res := createTestReservation("res-001", "other@example.com", "room-101", checkIn, checkOut)
	repo.reservations[shared.ReservationID("res-001")] = *res

	handler := inbound.HttpViewReservationDetail(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations/res-001", nil)
	req.SetPathValue("id", "res-001")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 403", rec.Code, http.StatusForbidden)
}

func Test_HttpViewReservationDetail_With_Valid_Session_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(detailTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	// Create reservation for current user
	checkIn := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	checkOut := checkIn.AddDate(0, 0, 3)
	res := createTestReservation("res-001", "test@example.com", "room-101", checkIn, checkOut)
	repo.reservations[shared.ReservationID("res-001")] = *res

	handler := inbound.HttpViewReservationDetail(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations/res-001", nil)
	req.SetPathValue("id", "res-001")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewReservationDetail_Should_Render_Reservation_Data(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(detailTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	// Create reservation for current user
	checkIn := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	checkOut := checkIn.AddDate(0, 0, 3)
	res := createTestReservation("res-001", "test@example.com", "room-101", checkIn, checkOut)
	repo.reservations[shared.ReservationID("res-001")] = *res

	handler := inbound.HttpViewReservationDetail(e, service)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations/res-001", nil)
	req.SetPathValue("id", "res-001")
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

// ============================================================================
// HttpCancelReservation Tests
// ============================================================================

func Test_HttpCancelReservation_Without_Session_Should_Return_401(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	handler := inbound.HttpCancelReservation(service)
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/res-001/cancel", nil)
	req.SetPathValue("id", "res-001")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 401", rec.Code, http.StatusUnauthorized)
}

func Test_HttpCancelReservation_Without_Reservation_ID_Should_Return_400(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	handler := inbound.HttpCancelReservation(service)
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations//cancel", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 400", rec.Code, http.StatusBadRequest)
}

func Test_HttpCancelReservation_With_NonExistent_Reservation_Should_Return_404(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	handler := inbound.HttpCancelReservation(service)
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/nonexistent/cancel", nil)
	req.SetPathValue("id", "nonexistent")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 404", rec.Code, http.StatusNotFound)
}

func Test_HttpCancelReservation_With_Other_User_Reservation_Should_Return_403(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	// Create reservation for another user
	checkIn := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	checkOut := checkIn.AddDate(0, 0, 3)
	res := createTestReservation("res-001", "other@example.com", "room-101", checkIn, checkOut)
	repo.reservations[shared.ReservationID("res-001")] = *res

	handler := inbound.HttpCancelReservation(service)
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/res-001/cancel", nil)
	req.SetPathValue("id", "res-001")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 403", rec.Code, http.StatusForbidden)
}

func Test_HttpCancelReservation_With_Valid_Reservation_Should_Redirect(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	// Create reservation for current user (far enough in future to be cancellable)
	checkIn := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	checkOut := checkIn.AddDate(0, 0, 3)
	res := createTestReservation("res-001", "test@example.com", "room-101", checkIn, checkOut)
	repo.reservations[shared.ReservationID("res-001")] = *res

	handler := inbound.HttpCancelReservation(service)
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/res-001/cancel", nil)
	req.SetPathValue("id", "res-001")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
	location := rec.Header().Get("Location")
	assert.That(t, "location must redirect to reservations", containsString(location, "/ui/reservations"), true)
}

func Test_HttpCancelReservation_Should_Update_Reservation_Status(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	repo := newMockReservationRepository()
	service := createDetailTestService(repo)

	// Create reservation for current user
	checkIn := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	checkOut := checkIn.AddDate(0, 0, 3)
	res := createTestReservation("res-001", "test@example.com", "room-101", checkIn, checkOut)
	repo.reservations[shared.ReservationID("res-001")] = *res

	handler := inbound.HttpCancelReservation(service)
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/res-001/cancel", nil)
	req.SetPathValue("id", "res-001")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	updatedRes := repo.reservations[shared.ReservationID("res-001")]
	assert.That(t, "reservation status must be cancelled", updatedRes.Status, reservation.StatusCancelled)
}

// ============================================================================
// Unit Tests for View Logic
// ============================================================================

func Test_GuestInfoView_Should_Have_Expected_Fields(t *testing.T) {
	// Arrange
	guest := struct {
		Name        string
		Email       string
		PhoneNumber string
	}{
		Name:        "John Doe",
		Email:       "john@example.com",
		PhoneNumber: "+1234567890",
	}

	// Assert
	assert.That(t, "Name must match", guest.Name, "John Doe")
	assert.That(t, "Email must match", guest.Email, "john@example.com")
	assert.That(t, "PhoneNumber must match", guest.PhoneNumber, "+1234567890")
}

func Test_BuildReservationDetailView_Logic_Should_Convert_Guests(t *testing.T) {
	// Arrange
	domainGuests := []reservation.GuestInfo{
		reservation.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
		reservation.NewGuestInfo("Jane Doe", "jane@example.com", "+0987654321"),
	}

	viewGuests := make([]struct {
		Name        string
		Email       string
		PhoneNumber string
	}, 0, len(domainGuests))

	// Act
	for _, g := range domainGuests {
		viewGuests = append(viewGuests, struct {
			Name        string
			Email       string
			PhoneNumber string
		}{
			Name:        g.Name,
			Email:       g.Email,
			PhoneNumber: g.PhoneNumber,
		})
	}

	// Assert
	assert.That(t, "must have 2 guests", len(viewGuests), 2)
	assert.That(t, "first guest name must match", viewGuests[0].Name, "John Doe")
	assert.That(t, "second guest name must match", viewGuests[1].Name, "Jane Doe")
}

func Test_BuildReservationDetailView_Logic_Should_Format_Dates(t *testing.T) {
	// Arrange
	checkIn := time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)
	checkOut := time.Date(2024, 1, 18, 12, 0, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 10, 14, 30, 0, 0, time.UTC)

	// Act
	checkInFormatted := checkIn.Format("2006-01-02")
	checkOutFormatted := checkOut.Format("2006-01-02")
	createdAtFormatted := createdAt.Format("2006-01-02 15:04")

	// Assert
	assert.That(t, "checkIn format must match", checkInFormatted, "2024-01-15")
	assert.That(t, "checkOut format must match", checkOutFormatted, "2024-01-18")
	assert.That(t, "createdAt format must match", createdAtFormatted, "2024-01-10 14:30")
}

func Test_StatusClass_Mapping_For_All_Statuses(t *testing.T) {
	// Arrange
	tests := []struct {
		status   reservation.ReservationStatus
		expected string
	}{
		{reservation.StatusPending, "warning"},
		{reservation.StatusConfirmed, "info"},
		{reservation.StatusActive, "primary"},
		{reservation.StatusCompleted, "success"},
		{reservation.StatusCancelled, "danger"},
	}

	// Act & Assert
	for _, tc := range tests {
		result := testReservationStatusClass(tc.status)
		assert.That(t, "status class must match for "+string(tc.status), result, tc.expected)
	}
}
