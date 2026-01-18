package reservation_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/cloud-native-utils/mcp"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// ============================================================================
// Mock Implementations for Tools Tests
// ============================================================================

type toolsMockReservationRepository struct {
	reservations map[reservation.ReservationID]reservation.Reservation
	createErr    error
	readErr      error
	updateErr    error
}

func newToolsMockReservationRepository() *toolsMockReservationRepository {
	return &toolsMockReservationRepository{
		reservations: make(map[reservation.ReservationID]reservation.Reservation),
	}
}

func (m *toolsMockReservationRepository) Create(ctx context.Context, id reservation.ReservationID, res reservation.Reservation) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.reservations[id] = res
	return nil
}

func (m *toolsMockReservationRepository) Read(ctx context.Context, id reservation.ReservationID) (*reservation.Reservation, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	res, ok := m.reservations[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &res, nil
}

func (m *toolsMockReservationRepository) Update(ctx context.Context, id reservation.ReservationID, res reservation.Reservation) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.reservations[id] = res
	return nil
}

func (m *toolsMockReservationRepository) Delete(ctx context.Context, id reservation.ReservationID) error {
	delete(m.reservations, id)
	return nil
}

func (m *toolsMockReservationRepository) ReadAll(ctx context.Context) ([]reservation.Reservation, error) {
	result := make([]reservation.Reservation, 0, len(m.reservations))
	for _, res := range m.reservations {
		result = append(result, res)
	}
	return result, nil
}

type toolsMockAvailabilityChecker struct {
	available bool
	err       error
}

func (m *toolsMockAvailabilityChecker) IsRoomAvailable(ctx context.Context, roomID reservation.RoomID, dateRange reservation.DateRange) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.available, nil
}

func (m *toolsMockAvailabilityChecker) GetOverlappingReservations(ctx context.Context, roomID reservation.RoomID, dateRange reservation.DateRange) ([]*reservation.Reservation, error) {
	return nil, nil
}

type toolsMockEventPublisher struct {
	published []event.Event
	err       error
}

func (m *toolsMockEventPublisher) Publish(ctx context.Context, evt event.Event) error {
	if m.err != nil {
		return m.err
	}
	m.published = append(m.published, evt)
	return nil
}

// ============================================================================
// Test Helpers
// ============================================================================

func createToolsTestService(repo *toolsMockReservationRepository, checker *toolsMockAvailabilityChecker, publisher *toolsMockEventPublisher) *reservation.Service {
	return reservation.NewService(repo, checker, publisher)
}

func toolsValidDateRange() reservation.DateRange {
	checkIn := time.Now().Add(48 * time.Hour).Truncate(24 * time.Hour)
	checkOut := checkIn.Add(72 * time.Hour)
	return reservation.NewDateRange(checkIn, checkOut)
}

func toolsValidGuests() []reservation.GuestInfo {
	return []reservation.GuestInfo{
		reservation.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
	}
}

func toolsValidMoney() shared.Money {
	return shared.NewMoney(10000, "USD")
}

// ============================================================================
// RegisterTools Tests
// ============================================================================

func Test_RegisterTools_Should_Register_All_Tools(t *testing.T) {
	// Arrange
	repo := newToolsMockReservationRepository()
	checker := &toolsMockAvailabilityChecker{available: true}
	publisher := &toolsMockEventPublisher{}
	service := createToolsTestService(repo, checker, publisher)
	server := mcp.NewServer("test-server", "1.0.0")

	// Act
	reservation.RegisterTools(server, service, checker)

	// Assert
	tools := server.Tools()
	assert.That(t, "must register 4 tools", len(tools), 4)

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Definition.Name] = true
	}
	assert.That(t, "get_reservation must be registered", toolNames["get_reservation"], true)
	assert.That(t, "list_reservations must be registered", toolNames["list_reservations"], true)
	assert.That(t, "cancel_reservation must be registered", toolNames["cancel_reservation"], true)
	assert.That(t, "check_availability must be registered", toolNames["check_availability"], true)
}

// ============================================================================
// GetReservation Tool Tests
// ============================================================================

func Test_GetReservationTool_Should_Return_Reservation_JSON(t *testing.T) {
	// Arrange
	repo := newToolsMockReservationRepository()
	checker := &toolsMockAvailabilityChecker{available: true}
	publisher := &toolsMockEventPublisher{}
	service := createToolsTestService(repo, checker, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	reservation.RegisterTools(server, service, checker)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")

	// Create a reservation first
	_, _ = service.CreateReservation(ctx, id, "guest-001", "room-101", toolsValidDateRange(), toolsValidMoney(), toolsValidGuests())

	// Get the tool and call it
	tools := server.Tools()
	var getTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "get_reservation" {
			getTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "get_reservation",
		Arguments: map[string]any{"id": "res-001"},
	}

	// Act
	result, err := getTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must have 1 content block", len(result.Content), 1)
	assert.That(t, "content must contain reservation ID", strings.Contains(result.Content[0].Text, "res-001"), true)
}

func Test_GetReservationTool_When_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newToolsMockReservationRepository()
	checker := &toolsMockAvailabilityChecker{available: true}
	publisher := &toolsMockEventPublisher{}
	service := createToolsTestService(repo, checker, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	reservation.RegisterTools(server, service, checker)

	ctx := context.Background()

	tools := server.Tools()
	var getTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "get_reservation" {
			getTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "get_reservation",
		Arguments: map[string]any{"id": "non-existent"},
	}

	// Act
	_, err := getTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

// ============================================================================
// ListReservations Tool Tests
// ============================================================================

func Test_ListReservationsTool_Should_Return_Guest_Reservations(t *testing.T) {
	// Arrange
	repo := newToolsMockReservationRepository()
	checker := &toolsMockAvailabilityChecker{available: true}
	publisher := &toolsMockEventPublisher{}
	service := createToolsTestService(repo, checker, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	reservation.RegisterTools(server, service, checker)

	ctx := context.Background()
	guestID := reservation.GuestID("guest-001")

	// Create reservations
	_, _ = service.CreateReservation(ctx, "res-001", guestID, "room-101", toolsValidDateRange(), toolsValidMoney(), toolsValidGuests())
	_, _ = service.CreateReservation(ctx, "res-002", guestID, "room-102", toolsValidDateRange(), toolsValidMoney(), toolsValidGuests())

	tools := server.Tools()
	var listTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "list_reservations" {
			listTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "list_reservations",
		Arguments: map[string]any{"guest_email": "guest-001"},
	}

	// Act
	result, err := listTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must have 1 content block", len(result.Content), 1)
	assert.That(t, "content must contain res-001", strings.Contains(result.Content[0].Text, "res-001"), true)
	assert.That(t, "content must contain res-002", strings.Contains(result.Content[0].Text, "res-002"), true)
}

// ============================================================================
// CancelReservation Tool Tests
// ============================================================================

func Test_CancelReservationTool_Should_Cancel_Reservation(t *testing.T) {
	// Arrange
	repo := newToolsMockReservationRepository()
	checker := &toolsMockAvailabilityChecker{available: true}
	publisher := &toolsMockEventPublisher{}
	service := createToolsTestService(repo, checker, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	reservation.RegisterTools(server, service, checker)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")

	// Create a reservation first
	_, _ = service.CreateReservation(ctx, id, "guest-001", "room-101", toolsValidDateRange(), toolsValidMoney(), toolsValidGuests())

	tools := server.Tools()
	var cancelTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "cancel_reservation" {
			cancelTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "cancel_reservation",
		Arguments: map[string]any{"id": "res-001", "reason": "Guest requested"},
	}

	// Act
	result, err := cancelTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must have 1 content block", len(result.Content), 1)
	assert.That(t, "content must confirm cancellation", strings.Contains(result.Content[0].Text, "cancelled successfully"), true)

	// Verify reservation is cancelled
	res, _ := service.GetReservation(ctx, id)
	assert.That(t, "status must be cancelled", res.Status, reservation.StatusCancelled)
}

func Test_CancelReservationTool_When_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newToolsMockReservationRepository()
	checker := &toolsMockAvailabilityChecker{available: true}
	publisher := &toolsMockEventPublisher{}
	service := createToolsTestService(repo, checker, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	reservation.RegisterTools(server, service, checker)

	ctx := context.Background()

	tools := server.Tools()
	var cancelTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "cancel_reservation" {
			cancelTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "cancel_reservation",
		Arguments: map[string]any{"id": "non-existent", "reason": "Test"},
	}

	// Act
	_, err := cancelTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

// ============================================================================
// CheckAvailability Tool Tests
// ============================================================================

func Test_CheckAvailabilityTool_When_Available_Should_Return_Available(t *testing.T) {
	// Arrange
	repo := newToolsMockReservationRepository()
	checker := &toolsMockAvailabilityChecker{available: true}
	publisher := &toolsMockEventPublisher{}
	service := createToolsTestService(repo, checker, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	reservation.RegisterTools(server, service, checker)

	ctx := context.Background()

	tools := server.Tools()
	var checkTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "check_availability" {
			checkTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name: "check_availability",
		Arguments: map[string]any{
			"room_id":   "room-101",
			"check_in":  "2024-06-01T14:00:00Z",
			"check_out": "2024-06-05T11:00:00Z",
		},
	}

	// Act
	result, err := checkTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must have 1 content block", len(result.Content), 1)
	assert.That(t, "content must indicate available", strings.Contains(result.Content[0].Text, "is available"), true)
}

func Test_CheckAvailabilityTool_When_Not_Available_Should_Return_Not_Available(t *testing.T) {
	// Arrange
	repo := newToolsMockReservationRepository()
	checker := &toolsMockAvailabilityChecker{available: false}
	publisher := &toolsMockEventPublisher{}
	service := createToolsTestService(repo, checker, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	reservation.RegisterTools(server, service, checker)

	ctx := context.Background()

	tools := server.Tools()
	var checkTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "check_availability" {
			checkTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name: "check_availability",
		Arguments: map[string]any{
			"room_id":   "room-101",
			"check_in":  "2024-06-01T14:00:00Z",
			"check_out": "2024-06-05T11:00:00Z",
		},
	}

	// Act
	result, err := checkTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must have 1 content block", len(result.Content), 1)
	assert.That(t, "content must indicate not available", strings.Contains(result.Content[0].Text, "not available"), true)
}

func Test_CheckAvailabilityTool_When_Invalid_Date_Format_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newToolsMockReservationRepository()
	checker := &toolsMockAvailabilityChecker{available: true}
	publisher := &toolsMockEventPublisher{}
	service := createToolsTestService(repo, checker, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	reservation.RegisterTools(server, service, checker)

	ctx := context.Background()

	tools := server.Tools()
	var checkTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "check_availability" {
			checkTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name: "check_availability",
		Arguments: map[string]any{
			"room_id":   "room-101",
			"check_in":  "invalid-date",
			"check_out": "2024-06-05T11:00:00Z",
		},
	}

	// Act
	_, err := checkTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "error must mention invalid date", strings.Contains(err.Error(), "invalid check_in date format"), true)
}

func Test_CheckAvailabilityTool_When_Checker_Fails_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newToolsMockReservationRepository()
	checker := &toolsMockAvailabilityChecker{err: errors.New("service unavailable")}
	publisher := &toolsMockEventPublisher{}
	service := createToolsTestService(repo, checker, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	reservation.RegisterTools(server, service, checker)

	ctx := context.Background()

	tools := server.Tools()
	var checkTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "check_availability" {
			checkTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name: "check_availability",
		Arguments: map[string]any{
			"room_id":   "room-101",
			"check_in":  "2024-06-01T14:00:00Z",
			"check_out": "2024-06-05T11:00:00Z",
		},
	}

	// Act
	_, err := checkTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}
