package inbound

import (
	"net/http"
	"os"

	"github.com/andygeiss/cloud-native-utils/redirecting"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// ReservationListItem represents a reservation item for the list view.
type ReservationListItem struct {
	ID          string
	RoomID      string
	CheckIn     string
	CheckOut    string
	Status      string
	StatusClass string
	TotalAmount string
	CanCancel   bool
}

// HttpViewReservationsResponse specifies the view data for the reservations list.
type HttpViewReservationsResponse struct {
	AppName      string
	Title        string
	SessionID    string
	Reservations []ReservationListItem
}

// HttpViewReservations defines an HTTP handler function for rendering the reservations list.
func HttpViewReservations(e *templating.Engine, reservationService *booking.ReservationService) http.HandlerFunc {
	appName := os.Getenv("APP_NAME")
	title := appName + " - Reservations"

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check authentication
		sessionID, _ := ctx.Value(security.ContextSessionID).(string)
		email, _ := ctx.Value(security.ContextEmail).(string)
		if sessionID == "" || email == "" {
			redirecting.Redirect(w, r, "/ui/login")
			return
		}

		// Get reservations for the current user (using email as guest ID)
		guestID := booking.GuestID(email)
		reservations, err := reservationService.ListReservationsByGuest(ctx, guestID)
		if err != nil {
			// If repository doesn't exist yet, treat as empty list
			reservations = []*booking.Reservation{}
		}

		// Convert domain reservations to view items
		items := make([]ReservationListItem, 0, len(reservations))
		for _, res := range reservations {
			items = append(items, ReservationListItem{
				ID:          string(res.ID),
				RoomID:      string(res.RoomID),
				CheckIn:     res.DateRange.CheckIn.Format("2006-01-02"),
				CheckOut:    res.DateRange.CheckOut.Format("2006-01-02"),
				Status:      string(res.Status),
				StatusClass: reservationStatusClass(res.Status),
				TotalAmount: res.TotalAmount.FormatAmount(),
				CanCancel:   res.CanBeCancelled(),
			})
		}

		data := HttpViewReservationsResponse{
			AppName:      appName,
			Title:        title,
			SessionID:    sessionID,
			Reservations: items,
		}

		HttpView(e, "reservations", data)(w, r)
	}
}

// reservationStatusClass returns the CSS class for a reservation status.
func reservationStatusClass(status booking.ReservationStatus) string {
	switch status {
	case booking.StatusPending:
		return "warning"
	case booking.StatusConfirmed:
		return "info"
	case booking.StatusActive:
		return "primary"
	case booking.StatusCompleted:
		return "success"
	case booking.StatusCancelled:
		return "danger"
	default:
		return "secondary"
	}
}
