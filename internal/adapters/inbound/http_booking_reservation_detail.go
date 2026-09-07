package inbound

import (
	"net/http"

	"github.com/andygeiss/cloud-native-utils/web"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// GuestInfoView represents guest information for the view.
type GuestInfoView struct {
	Name        string
	Email       string
	PhoneNumber string
}

// ReservationDetailView represents a reservation for the detail view.
type ReservationDetailView struct {
	ID                 string
	RoomID             string
	CheckIn            string
	CheckOut           string
	Status             string
	StatusClass        string
	TotalAmount        string
	CreatedAt          string
	CancellationReason string
	Guests             []GuestInfoView
	Nights             int
	CanCancel          bool
}

// HttpViewReservationDetailResponse specifies the view data for the reservation detail.
type HttpViewReservationDetailResponse struct {
	Layout
	Reservation ReservationDetailView
}

func buildReservationDetailView(res *reservation.Reservation) ReservationDetailView {
	guests := make([]GuestInfoView, 0, len(res.Guests))
	for _, g := range res.Guests {
		guests = append(guests, GuestInfoView{
			Name:        g.Name,
			Email:       g.Email,
			PhoneNumber: g.PhoneNumber,
		})
	}

	return ReservationDetailView{
		Guests:             guests,
		ID:                 string(res.ID),
		RoomID:             string(res.RoomID),
		CheckIn:            res.DateRange.CheckIn.Format("2006-01-02"),
		CheckOut:           res.DateRange.CheckOut.Format("2006-01-02"),
		Status:             string(res.Status),
		StatusClass:        reservationStatusClass(res.Status),
		TotalAmount:        res.TotalAmount.FormatAmount(),
		CreatedAt:          res.CreatedAt.Format("2006-01-02 15:04"),
		CancellationReason: res.CancellationReason,
		Nights:             res.Nights(),
		CanCancel:          res.CanBeCancelled(),
	}
}

// HttpViewReservationDetail defines an HTTP handler function for rendering a single reservation.
func HttpViewReservationDetail(v *View, app AppInfo, reservationService *reservation.Service) http.HandlerFunc {
	appName := app.Name

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sessionID, _ := ctx.Value(web.ContextSessionID).(string)
		email, _ := ctx.Value(web.ContextEmail).(string)
		if sessionID == "" || email == "" {
			http.Redirect(w, r, "/ui/login", http.StatusSeeOther)
			return
		}

		reservationID := r.PathValue("id")
		if reservationID == "" {
			http.Error(w, "Reservation ID required", http.StatusBadRequest)
			return
		}

		res, err := reservationService.GetReservation(ctx, shared.ReservationID(reservationID))
		if err != nil {
			http.Error(w, "Reservation not found", http.StatusNotFound)
			return
		}

		if string(res.GuestID) != email {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		data := HttpViewReservationDetailResponse{
			AppName:     appName,
			Title:       appName + " - Reservation " + reservationID,
			SessionID:   sessionID,
			Reservation: buildReservationDetailView(res),
		}

		// No fragment: cancelling from here navigates away rather than swapping,
		// so there is no block for htmx to ask for.
		v.Render(w, r, http.StatusOK, "reservation_detail", "", data)
	}
}

// HttpCancelReservation handles the POST request to cancel a reservation.
//
// Three ways out, and which one is taken is decided by isFragment, the same
// test the renderer uses:
//
//   - a fragment swap from the list gets the re-rendered list and 200;
//   - a boosted form, or a browser with no htmx at all, gets 303 See Other,
//     because a direct 200 would push this POST's URL into history and a later
//     refresh or Back would re-issue it as a GET against a POST-only route;
//   - a fragment swap with no target to swap into — the detail page — gets
//     HX-Redirect, because the whole page is stale once the booking is gone.
func HttpCancelReservation(v *View, app AppInfo, reservationService *reservation.Service) http.HandlerFunc {
	appName := app.Name
	title := appName + " - Reservations"

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check authentication
		sessionID, _ := ctx.Value(web.ContextSessionID).(string)
		email, _ := ctx.Value(web.ContextEmail).(string)
		if sessionID == "" || email == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get reservation ID from path
		reservationID := r.PathValue("id")
		if reservationID == "" {
			http.Error(w, "Reservation ID required", http.StatusBadRequest)
			return
		}

		// Verify the reservation belongs to the current user
		res, err := reservationService.GetReservation(ctx, shared.ReservationID(reservationID))
		if err != nil {
			http.Error(w, "Reservation not found", http.StatusNotFound)
			return
		}

		if string(res.GuestID) != email {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Cancel the reservation
		err = reservationService.CancelReservation(ctx, shared.ReservationID(reservationID), "Cancelled by guest")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !isFragment(r) {
			// Plain and boosted forms both land here: POST-redirect-GET.
			http.Redirect(w, r, "/ui/reservations", http.StatusSeeOther)
			return
		}

		// The list page names a target to swap; the detail page does not, and
		// its whole document is stale now, so send it somewhere else instead.
		if r.Header.Get("HX-Target") != "reservation-list" {
			w.Header().Set("HX-Redirect", "/ui/reservations")
			w.WriteHeader(http.StatusOK)
			return
		}

		data := reservationsResponse(ctx, reservationService, appName, title, sessionID, email)
		v.Render(w, r, http.StatusOK, "reservations", "reservation-list", data)
	}
}
