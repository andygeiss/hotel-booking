package inbound

import (
	"net/http"
	"os"
	"time"

	"github.com/andygeiss/cloud-native-utils/templating"
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
	AppName    string
	Title      string
	SessionID  string
	MinDate    string
	GuestName  string
	GuestEmail string
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
func HttpViewReservationForm(e *templating.Engine) http.HandlerFunc {
	appName := os.Getenv("APP_NAME")
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
			Rooms:      getDefaultRooms(),
			AppName:    appName,
			Title:      title,
			SessionID:  sessionID,
			MinDate:    time.Now().Format("2006-01-02"),
			GuestName:  name,
			GuestEmail: email,
		}

		HttpView(e, "reservation_form", data)(w, r)
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
func HttpCreateReservation(e *templating.Engine, reservationService *reservation.Service) http.HandlerFunc {
	appName := os.Getenv("APP_NAME")
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
			renderReservationFormWithError(e, w, r, appName, title, sessionID, errMsg, r.FormValue("guest_name"), r.FormValue("guest_email"))
			return
		}

		nights := int(input.checkOut.Sub(input.checkIn).Hours() / 24)
		totalAmount := shared.NewMoney(getRoomPrices()[input.roomID]*int64(nights), "USD")
		guests := []reservation.GuestInfo{reservation.NewGuestInfo(input.guestName, input.guestEmail, input.guestPhone)}

		_, err := reservationService.CreateReservation(ctx, shared.NewReservationID(), reservation.GuestID(email), reservation.RoomID(input.roomID), reservation.NewDateRange(input.checkIn, input.checkOut), totalAmount, guests)
		if err != nil {
			renderReservationFormWithError(e, w, r, appName, title, sessionID, err.Error(), input.guestName, input.guestEmail)
			return
		}

		http.Redirect(w, r, "/ui/reservations", http.StatusSeeOther)
	}
}

func renderReservationFormWithError(e *templating.Engine, w http.ResponseWriter, r *http.Request, appName, title, sessionID, errMsg, guestName, guestEmail string) {
	data := HttpViewReservationFormResponse{
		Rooms:      getDefaultRooms(),
		AppName:    appName,
		Title:      title,
		SessionID:  sessionID,
		MinDate:    time.Now().Format("2006-01-02"),
		GuestName:  guestName,
		GuestEmail: guestEmail,
		Error:      errMsg,
	}
	HttpView(e, "reservation_form", data)(w, r)
}
