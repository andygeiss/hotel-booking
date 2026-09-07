package inbound

import (
	"net/http"
	"time"

	"github.com/andygeiss/cloud-native-utils/web"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// RoomOption represents a room option for the form dropdown.
type RoomOption struct {
	ID    string
	Name  string
	Price string
}

// HttpViewReservationFormResponse specifies the view data for the reservation form.
type HttpViewReservationFormResponse struct {
	Layout
	MinDate    string
	RoomID     string
	CheckIn    string
	CheckOut   string
	GuestName  string
	GuestEmail string
	GuestPhone string
	Error      string
	Rooms      []RoomOption
}

func getDefaultRooms() []RoomOption {
	return []RoomOption{
		{ID: "room-101", Name: "Standard Room 101", Price: "$99.00"},
		{ID: "room-102", Name: "Standard Room 102", Price: "$99.00"},
		{ID: "room-201", Name: "Deluxe Room 201", Price: "$149.00"},
		{ID: "room-202", Name: "Deluxe Room 202", Price: "$149.00"},
		{ID: "room-301", Name: "Suite 301", Price: "$249.00"},
	}
}

func getRoomPrices() map[string]int64 {
	return map[string]int64{
		"room-101": 9900,
		"room-102": 9900,
		"room-201": 14900,
		"room-202": 14900,
		"room-301": 24900,
	}
}

// HttpViewReservationForm defines an HTTP handler function for rendering the new reservation form.
func HttpViewReservationForm(v *View, app AppInfo) http.HandlerFunc {
	appName := app.Name
	title := appName + " - New Reservation"

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sessionID, _ := ctx.Value(web.ContextSessionID).(string)
		email, _ := ctx.Value(web.ContextEmail).(string)
		if sessionID == "" || email == "" {
			http.Redirect(w, r, "/ui/login", http.StatusSeeOther)
			return
		}

		name, _ := ctx.Value(web.ContextName).(string)

		data := HttpViewReservationFormResponse{
			AppName:    appName,
			Title:      title,
			SessionID:  sessionID,
			ShowNew:    true,
			Rooms:      getDefaultRooms(),
			MinDate:    time.Now().Format("2006-01-02"),
			GuestName:  name,
			GuestEmail: email,
		}

		v.Render(w, r, http.StatusOK, "reservation_form", "reservation-form", data)
	}
}

type reservationFormInput struct {
	checkIn    time.Time
	checkOut   time.Time
	roomID     string
	guestName  string
	guestEmail string
	guestPhone string
}

func parseReservationForm(r *http.Request) (*reservationFormInput, string) {
	if err := r.ParseForm(); err != nil {
		return nil, "Invalid form data"
	}

	roomID := r.FormValue("room_id")
	checkInStr := r.FormValue("check_in")
	checkOutStr := r.FormValue("check_out")
	guestName := r.FormValue("guest_name")
	guestEmail := r.FormValue("guest_email")
	guestPhone := r.FormValue("guest_phone")

	if roomID == "" || checkInStr == "" || checkOutStr == "" || guestName == "" || guestEmail == "" {
		return nil, "Please fill in all required fields"
	}

	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		return nil, "Invalid check-in date format"
	}

	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		return nil, "Invalid check-out date format"
	}

	if _, ok := getRoomPrices()[roomID]; !ok {
		return nil, "Invalid room selected"
	}

	return &reservationFormInput{
		checkIn:    checkIn,
		checkOut:   checkOut,
		roomID:     roomID,
		guestName:  guestName,
		guestEmail: guestEmail,
		guestPhone: guestPhone,
	}, ""
}

// HttpCreateReservation handles the POST request to create a new reservation.
func HttpCreateReservation(v *View, app AppInfo, reservationService *reservation.Service) http.HandlerFunc {
	appName := app.Name
	title := appName + " - New Reservation"

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sessionID, _ := ctx.Value(web.ContextSessionID).(string)
		email, _ := ctx.Value(web.ContextEmail).(string)
		if sessionID == "" || email == "" {
			http.Redirect(w, r, "/ui/login", http.StatusSeeOther)
			return
		}

		input, errMsg := parseReservationForm(r)
		if errMsg != "" {
			renderReservationFormWithError(v, w, r, appName, title, sessionID, errMsg)
			return
		}

		nights := int(input.checkOut.Sub(input.checkIn).Hours() / 24)
		totalAmount := shared.NewMoney(getRoomPrices()[input.roomID]*int64(nights), "USD")
		guests := []reservation.GuestInfo{reservation.NewGuestInfo(input.guestName, input.guestEmail, input.guestPhone)}

		_, err := reservationService.CreateReservation(ctx, shared.NewReservationID(), reservation.GuestID(email), reservation.RoomID(input.roomID), reservation.NewDateRange(input.checkIn, input.checkOut), totalAmount, guests)
		if err != nil {
			renderReservationFormWithError(v, w, r, appName, title, sessionID, err.Error())
			return
		}

		http.Redirect(w, r, "/ui/reservations", http.StatusSeeOther)
	}
}

// renderReservationFormWithError answers a rejected submission with 422 and the
// form fragment, every field still holding what was typed.
//
// 422 rather than 200 because the request was understood and refused, and htmx
// only swaps a 4xx because the layout's htmx-config lists this one code.
//
// HX-Push-Url: false is for the boosted case. A boosted swap otherwise pushes
// this POST's URL into history, and the next refresh or Back re-issues it as a
// GET against a route that only accepts POST — a 405 in the user's face. A 422
// is not a redirect, so nothing else prevents that push.
func renderReservationFormWithError(v *View, w http.ResponseWriter, r *http.Request, appName, title, sessionID, errMsg string) {
	data := HttpViewReservationFormResponse{
		AppName:    appName,
		Title:      title,
		SessionID:  sessionID,
		ShowNew:    true,
		Rooms:      getDefaultRooms(),
		MinDate:    time.Now().Format("2006-01-02"),
		RoomID:     r.FormValue("room_id"),
		CheckIn:    r.FormValue("check_in"),
		CheckOut:   r.FormValue("check_out"),
		GuestName:  r.FormValue("guest_name"),
		GuestEmail: r.FormValue("guest_email"),
		GuestPhone: r.FormValue("guest_phone"),
		Error:      errMsg,
	}
	if r.Header.Get("HX-Boosted") == "true" {
		w.Header().Set("HX-Push-Url", "false")
	}
	v.Render(w, r, http.StatusUnprocessableEntity, "reservation_form", "reservation-form", data)
}
