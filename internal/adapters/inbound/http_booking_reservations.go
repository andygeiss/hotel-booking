package inbound

import (
	"context"
	"net/http"

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
	Layout
	Reservations []ReservationListItem
}

// HttpViewReservations defines an HTTP handler function for rendering the reservations list.
func HttpViewReservations(v *View, app AppInfo, reservationService *reservation.Service) http.HandlerFunc {
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

		data := reservationsResponse(ctx, reservationService, appName, title, sessionID, email)

		// Naming the fragment here is what makes the page dual-mode: a plain
		// request gets the document, an htmx one gets the list block alone.
		v.Render(w, r, http.StatusOK, "reservations", "reservation-list", data)
	}
}

// reservationsResponse builds one guest's list. HttpCancelReservation renders
// the same data to answer a fragment swap, and building it in one place is what
// stops the two views of the list from drifting apart.
func reservationsResponse(ctx context.Context, reservationService *reservation.Service, appName, title, sessionID, email string) HttpViewReservationsResponse {
	// Get reservations for the current user (using email as guest ID)
	reservations, err := reservationService.ListReservationsByGuest(ctx, reservation.GuestID(email))
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

	return HttpViewReservationsResponse{
		AppName:      appName,
		Title:        title,
		SessionID:    sessionID,
		ShowNew:      true,
		Reservations: items,
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
