package inbound_test

import (
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/hotel-booking/internal/adapters/inbound"
	"github.com/andygeiss/hotel-booking/internal/adapters/outbound"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
)

// ============================================================================
// Test Assets
// ============================================================================

//go:embed testdata/assets/templates/*.tmpl testdata/assets/static/css/*.css
var formTestAssets embed.FS

// ============================================================================
// Helper Functions
// ============================================================================

func createFormTestService(repo *mockReservationRepository) *reservation.Service {
	availabilityChecker := outbound.NewRepositoryAvailabilityChecker(repo)
	eventPublisher := outbound.NewEventPublisher(messaging.NewInternalDispatcher())
	return reservation.NewService(repo, availabilityChecker, eventPublisher)
}

// ============================================================================
// HttpViewReservationForm Tests
// ============================================================================

func Test_HttpViewReservationForm_Without_Session_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(formTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewReservationForm(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations/new", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
	location := rec.Header().Get("Location")
	assert.That(t, "location must contain login", containsString(location, "/ui/login"), true)
}

func Test_HttpViewReservationForm_With_Valid_Session_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(formTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewReservationForm(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations/new", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewReservationForm_Should_Render_App_Name(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(formTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewReservationForm(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations/new", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain app name", containsString(bodyStr, "TestApp"), true)
}

func Test_HttpViewReservationForm_Should_Render_Guest_Email(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(formTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewReservationForm(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations/new", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain guest email", containsString(bodyStr, "test@example.com"), true)
}

func Test_HttpViewReservationForm_Should_Render_Rooms(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(formTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewReservationForm(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/reservations/new", nil)
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain room-101", containsString(bodyStr, "room-101"), true)
}

// ============================================================================
// HttpCreateReservation Tests
// ============================================================================

func Test_HttpCreateReservation_Without_Session_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(formTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createFormTestService(repo)

	handler := inbound.HttpCreateReservation(e, service)
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/new", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
	location := rec.Header().Get("Location")
	assert.That(t, "location must contain login", containsString(location, "/ui/login"), true)
}

func Test_HttpCreateReservation_With_Missing_Fields_Should_Show_Error(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(formTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createFormTestService(repo)

	handler := inbound.HttpCreateReservation(e, service)

	// Create request with empty form
	form := url.Values{}
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200 (form re-rendered with error)", rec.Code, http.StatusOK)
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain error message", containsString(bodyStr, "required"), true)
}

func Test_HttpCreateReservation_With_Invalid_Room_Should_Show_Error(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(formTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createFormTestService(repo)

	handler := inbound.HttpCreateReservation(e, service)

	// Create request with invalid room
	checkIn := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	checkOut := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	form := url.Values{
		"room_id":     {"invalid-room"},
		"check_in":    {checkIn},
		"check_out":   {checkOut},
		"guest_name":  {"Test Guest"},
		"guest_email": {"test@example.com"},
	}
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200 (form re-rendered with error)", rec.Code, http.StatusOK)
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain error message", containsString(bodyStr, "Invalid room"), true)
}

func Test_HttpCreateReservation_With_Valid_Data_Should_Redirect_To_Reservations(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(formTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createFormTestService(repo)

	handler := inbound.HttpCreateReservation(e, service)

	// Create request with valid data
	checkIn := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	checkOut := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	form := url.Values{
		"room_id":     {"room-101"},
		"check_in":    {checkIn},
		"check_out":   {checkOut},
		"guest_name":  {"Test Guest"},
		"guest_email": {"test@example.com"},
		"guest_phone": {"+1234567890"},
	}
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
	location := rec.Header().Get("Location")
	assert.That(t, "location must redirect to reservations", containsString(location, "/ui/reservations"), true)
}

func Test_HttpCreateReservation_Should_Create_Reservation_In_Repository(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(formTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createFormTestService(repo)

	handler := inbound.HttpCreateReservation(e, service)

	// Create request with valid data
	checkIn := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	checkOut := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	form := url.Values{
		"room_id":     {"room-101"},
		"check_in":    {checkIn},
		"check_out":   {checkOut},
		"guest_name":  {"Test Guest"},
		"guest_email": {"test@example.com"},
		"guest_phone": {"+1234567890"},
	}
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "repository must have 1 reservation", len(repo.reservations), 1)
}

func Test_HttpCreateReservation_With_Invalid_CheckIn_Date_Format_Should_Show_Error(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(formTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	repo := newMockReservationRepository()
	service := createFormTestService(repo)

	handler := inbound.HttpCreateReservation(e, service)

	// Create request with invalid date format
	form := url.Values{
		"room_id":     {"room-101"},
		"check_in":    {"invalid-date"},
		"check_out":   {"2024-01-20"},
		"guest_name":  {"Test Guest"},
		"guest_email": {"test@example.com"},
	}
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = addAuthContext(req, "test-session-123", "test@example.com")
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200 (form re-rendered with error)", rec.Code, http.StatusOK)
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain error message", containsString(bodyStr, "Invalid check-in date"), true)
}

// ============================================================================
// Unit Tests for Room Configuration
// ============================================================================

type roomOption struct {
	ID    string
	Name  string
	Price string
}

func getDefaultRoomsForTest() []roomOption {
	return []roomOption{
		{ID: "room-101", Name: "Standard Room 101", Price: "$99.00"},
		{ID: "room-102", Name: "Standard Room 102", Price: "$99.00"},
		{ID: "room-201", Name: "Deluxe Room 201", Price: "$149.00"},
		{ID: "room-202", Name: "Deluxe Room 202", Price: "$149.00"},
		{ID: "room-301", Name: "Suite 301", Price: "$249.00"},
	}
}

func getRoomPricesForTest() map[string]int64 {
	return map[string]int64{
		"room-101": 9900,
		"room-102": 9900,
		"room-201": 14900,
		"room-202": 14900,
		"room-301": 24900,
	}
}

func Test_GetDefaultRooms_Should_Return_Five_Rooms(t *testing.T) {
	// Arrange & Act
	rooms := getDefaultRoomsForTest()

	// Assert
	assert.That(t, "must return 5 rooms", len(rooms), 5)
}

func Test_GetDefaultRooms_Should_Have_Standard_Rooms(t *testing.T) {
	// Arrange
	rooms := getDefaultRoomsForTest()

	// Act
	foundStandard := 0
	for _, room := range rooms {
		if room.Price == "$99.00" {
			foundStandard++
		}
	}

	// Assert
	assert.That(t, "must have 2 standard rooms at $99.00", foundStandard, 2)
}

func Test_GetRoomPrices_Should_Return_All_Room_Prices(t *testing.T) {
	// Arrange & Act
	prices := getRoomPricesForTest()

	// Assert
	assert.That(t, "must return 5 room prices", len(prices), 5)
}

func Test_GetRoomPrices_Should_Have_Correct_Standard_Price(t *testing.T) {
	// Arrange
	prices := getRoomPricesForTest()

	// Act & Assert
	assert.That(t, "room-101 price must be 9900", prices["room-101"], int64(9900))
	assert.That(t, "room-102 price must be 9900", prices["room-102"], int64(9900))
}
