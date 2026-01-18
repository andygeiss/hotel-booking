package reservation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andygeiss/cloud-native-utils/mcp"
)

// RegisterTools registers all reservation MCP tools with the server.
func RegisterTools(server *mcp.Server, service *Service, checker AvailabilityChecker) {
	server.RegisterTool(newGetReservationTool(service))
	server.RegisterTool(newListReservationsTool(service))
	server.RegisterTool(newCancelReservationTool(service))
	server.RegisterTool(newCheckAvailabilityTool(checker))
}

// newGetReservationTool creates a new tool for getting.
func newGetReservationTool(service *Service) mcp.Tool {
	return mcp.NewTool(
		"get_reservation",
		"Get reservation details by ID. Returns reservation status, guest info, dates, and amount.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"id": mcp.NewStringProperty("The reservation ID"),
			},
			[]string{"id"},
		),
		func(ctx context.Context, params mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
			id, _ := params.Arguments["id"].(string)
			reservation, err := service.GetReservation(ctx, ReservationID(id))
			if err != nil {
				return mcp.ToolsCallResult{}, err
			}
			data, _ := json.MarshalIndent(reservation, "", "  ")
			return mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.NewTextContent(string(data))},
			}, nil
		},
	)
}

// newListReservationsTool creates a tool for listing reservations.
func newListReservationsTool(service *Service) mcp.Tool {
	return mcp.NewTool(
		"list_reservations",
		"List all reservations for a guest by their email address.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"guest_email": mcp.NewStringProperty("The guest's email address"),
			},
			[]string{"guest_email"},
		),
		func(ctx context.Context, params mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
			email, _ := params.Arguments["guest_email"].(string)
			reservations, err := service.ListReservationsByGuest(ctx, GuestID(email))
			if err != nil {
				return mcp.ToolsCallResult{}, err
			}
			data, _ := json.MarshalIndent(reservations, "", "  ")
			return mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.NewTextContent(string(data))},
			}, nil
		},
	)
}

// newCancelReservationTool creates a tool for canceling reservations.
func newCancelReservationTool(service *Service) mcp.Tool {
	return mcp.NewTool(
		"cancel_reservation",
		"Cancel a reservation. Requires a reason. Cannot cancel within 24 hours of check-in.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"id":     mcp.NewStringProperty("The reservation ID"),
				"reason": mcp.NewStringProperty("Reason for cancellation"),
			},
			[]string{"id", "reason"},
		),
		func(ctx context.Context, params mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
			id, _ := params.Arguments["id"].(string)
			reason, _ := params.Arguments["reason"].(string)
			err := service.CancelReservation(ctx, ReservationID(id), reason)
			if err != nil {
				return mcp.ToolsCallResult{}, err
			}
			return mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.NewTextContent("Reservation cancelled successfully")},
			}, nil
		},
	)
}

// newCheckAvailabilityTool creates a tool for checking room availability.
func newCheckAvailabilityTool(checker AvailabilityChecker) mcp.Tool {
	return mcp.NewTool(
		"check_availability",
		"Check if a room is available for the specified date range.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"room_id":   mcp.NewStringProperty("The room ID"),
				"check_in":  mcp.NewStringProperty("Check-in date (RFC3339 format, e.g. 2024-01-15T14:00:00Z)"),
				"check_out": mcp.NewStringProperty("Check-out date (RFC3339 format, e.g. 2024-01-17T11:00:00Z)"),
			},
			[]string{"room_id", "check_in", "check_out"},
		),
		func(ctx context.Context, params mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
			roomID, _ := params.Arguments["room_id"].(string)
			checkInStr, _ := params.Arguments["check_in"].(string)
			checkOutStr, _ := params.Arguments["check_out"].(string)

			checkIn, err := time.Parse(time.RFC3339, checkInStr)
			if err != nil {
				return mcp.ToolsCallResult{}, fmt.Errorf("invalid check_in date format: %w", err)
			}
			checkOut, err := time.Parse(time.RFC3339, checkOutStr)
			if err != nil {
				return mcp.ToolsCallResult{}, fmt.Errorf("invalid check_out date format: %w", err)
			}

			dateRange := NewDateRange(checkIn, checkOut)
			available, err := checker.IsRoomAvailable(ctx, RoomID(roomID), dateRange)
			if err != nil {
				return mcp.ToolsCallResult{}, err
			}

			status := "available"
			if !available {
				status = "not available"
			}
			result := fmt.Sprintf("Room %s is %s for %s to %s",
				roomID, status, checkInStr, checkOutStr)
			return mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.NewTextContent(result)},
			}, nil
		},
	)
}
