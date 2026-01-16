package inbound

import (
	"net/http"
	"os"
	"time"

	"github.com/andygeiss/cloud-native-utils/redirecting"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
	"github.com/google/uuid"
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
	Rooms      []RoomOption
	Error      string
}

// HttpViewReservationForm defines an HTTP handler function for rendering the new reservation form.
func HttpViewReservationForm(e *templating.Engine) http.HandlerFunc {
	appName := os.Getenv("APP_NAME")
	title := appName + " - New Reservation"

	// Sample rooms - in production, this would come from a room service
	rooms := []RoomOption{
		{ID: "room-101", Name: "Standard Room 101", Price: "$99.00"},
		{ID: "room-102", Name: "Standard Room 102", Price: "$99.00"},
		{ID: "room-201", Name: "Deluxe Room 201", Price: "$149.00"},
		{ID: "room-202", Name: "Deluxe Room 202", Price: "$149.00"},
		{ID: "room-301", Name: "Suite 301", Price: "$249.00"},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check authentication
		sessionID, _ := ctx.Value(security.ContextSessionID).(string)
		email, _ := ctx.Value(security.ContextEmail).(string)
		if sessionID == "" || email == "" {
			redirecting.Redirect(w, r, "/ui/login")
			return
		}

		// Get user name if available
		name, _ := ctx.Value(security.ContextName).(string)

		data := HttpViewReservationFormResponse{
			AppName:    appName,
			Title:      title,
			SessionID:  sessionID,
			MinDate:    time.Now().Format("2006-01-02"),
			GuestName:  name,
			GuestEmail: email,
			Rooms:      rooms,
		}

		HttpView(e, "reservation_form", data)(w, r)
	}
}

// HttpCreateReservation handles the POST request to create a new reservation.
func HttpCreateReservation(e *templating.Engine, reservationService *booking.ReservationService) http.HandlerFunc {
	appName := os.Getenv("APP_NAME")
	title := appName + " - New Reservation"

	// Sample rooms with prices - in production, this would come from a room service
	rooms := []RoomOption{
		{ID: "room-101", Name: "Standard Room 101", Price: "$99.00"},
		{ID: "room-102", Name: "Standard Room 102", Price: "$99.00"},
		{ID: "room-201", Name: "Deluxe Room 201", Price: "$149.00"},
		{ID: "room-202", Name: "Deluxe Room 202", Price: "$149.00"},
		{ID: "room-301", Name: "Suite 301", Price: "$249.00"},
	}

	roomPrices := map[string]int64{
		"room-101": 9900,  // $99.00 in cents
		"room-102": 9900,
		"room-201": 14900, // $149.00 in cents
		"room-202": 14900,
		"room-301": 24900, // $249.00 in cents
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check authentication
		sessionID, _ := ctx.Value(security.ContextSessionID).(string)
		email, _ := ctx.Value(security.ContextEmail).(string)
		if sessionID == "" || email == "" {
			redirecting.Redirect(w, r, "/ui/login")
			return
		}

		// Parse form
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		roomID := r.FormValue("room_id")
		checkInStr := r.FormValue("check_in")
		checkOutStr := r.FormValue("check_out")
		guestName := r.FormValue("guest_name")
		guestEmail := r.FormValue("guest_email")
		guestPhone := r.FormValue("guest_phone")

		// Helper to render form with error
		renderWithError := func(errMsg string) {
			data := HttpViewReservationFormResponse{
				AppName:    appName,
				Title:      title,
				SessionID:  sessionID,
				MinDate:    time.Now().Format("2006-01-02"),
				GuestName:  guestName,
				GuestEmail: guestEmail,
				Rooms:      rooms,
				Error:      errMsg,
			}
			HttpView(e, "reservation_form", data)(w, r)
		}

		// Validate required fields
		if roomID == "" || checkInStr == "" || checkOutStr == "" || guestName == "" || guestEmail == "" {
			renderWithError("Please fill in all required fields")
			return
		}

		// Parse dates
		checkIn, err := time.Parse("2006-01-02", checkInStr)
		if err != nil {
			renderWithError("Invalid check-in date format")
			return
		}

		checkOut, err := time.Parse("2006-01-02", checkOutStr)
		if err != nil {
			renderWithError("Invalid check-out date format")
			return
		}

		// Calculate total amount
		nights := int(checkOut.Sub(checkIn).Hours() / 24)
		pricePerNight, ok := roomPrices[roomID]
		if !ok {
			renderWithError("Invalid room selected")
			return
		}
		totalAmount := booking.NewMoney(pricePerNight*int64(nights), "USD")

		// Create guest info
		guests := []booking.GuestInfo{
			booking.NewGuestInfo(guestName, guestEmail, guestPhone),
		}

		// Create reservation
		reservationID := booking.ReservationID(uuid.New().String())
		guestID := booking.GuestID(email)
		dateRange := booking.NewDateRange(checkIn, checkOut)

		_, err = reservationService.CreateReservation(
			ctx,
			reservationID,
			guestID,
			booking.RoomID(roomID),
			dateRange,
			totalAmount,
			guests,
		)
		if err != nil {
			renderWithError(err.Error())
			return
		}

		// Redirect to reservations list
		redirecting.Redirect(w, r, "/ui/reservations")
	}
}
