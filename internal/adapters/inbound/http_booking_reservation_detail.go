package inbound

import (
	"net/http"

	"github.com/andygeiss/cloud-native-utils/templating"
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
	AppName     string
	Title       string
	SessionID   string
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
func HttpViewReservationDetail(e *templating.Engine, app AppInfo, reservationService *reservation.Service) http.HandlerFunc {
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

		HttpView(e, "reservation_detail", data)(w, r)
	}
}

// HttpCancelReservation handles the POST request to cancel a reservation.
func HttpCancelReservation(reservationService *reservation.Service) http.HandlerFunc {
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

		// Redirect back to reservations list
		// Use HX-Redirect header for HTMX requests to trigger a full page navigation
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/ui/reservations")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/ui/reservations", http.StatusSeeOther)
	}
}
