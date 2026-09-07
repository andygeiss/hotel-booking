package inbound

import (
	"net/http"

	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/cloud-native-utils/web"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
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
func HttpViewReservations(e *templating.Engine, app AppInfo, reservationService *reservation.Service) http.HandlerFunc {
	appName := app.Name
	title := appName + " - Reservations"

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check authentication
		sessionID, _ := ctx.Value(web.ContextSessionID).(string)
		email, _ := ctx.Value(web.ContextEmail).(string)
		if sessionID == "" || email == "" {
			http.Redirect(w, r, "/ui/login", http.StatusSeeOther)
			return
		}

		// Get reservations for the current user (using email as guest ID)
		guestID := reservation.GuestID(email)
		reservations, err := reservationService.ListReservationsByGuest(ctx, guestID)
		if err != nil {
			// If repository doesn't exist yet, treat as empty list
			reservations = []*reservation.Reservation{}
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
func reservationStatusClass(status reservation.ReservationStatus) string {
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
