package booking

import (
	"testing"
	"time"
)

// Test Value Objects

func Test_ReservationID_With_String_Value_Should_Be_Assignable(t *testing.T) {
	var id ReservationID = "res-123"
	if id != "res-123" {
		t.Errorf("expected res-123, got %s", id)
	}
}

func Test_Money_FormatAmount_Should_Convert_Cents_To_Dollars(t *testing.T) {
	money := NewMoney(12050, "USD")
	formatted := money.FormatAmount()
	expected := "120.50 USD"
	if formatted != expected {
		t.Errorf("expected %s, got %s", expected, formatted)
	}
}

func Test_Money_Currency_Should_Be_Uppercase(t *testing.T) {
	money := NewMoney(10000, "usd")
	if money.Currency != "USD" {
		t.Errorf("expected USD, got %s", money.Currency)
	}
}

// Test Reservation Creation

func Test_NewReservation_With_Valid_Data_Should_Create_Instance(t *testing.T) {
	checkIn := time.Now().AddDate(0, 0, 7)  // 7 days from now
	checkOut := time.Now().AddDate(0, 0, 10) // 10 days from now
	dateRange := NewDateRange(checkIn, checkOut)
	guests := []GuestInfo{NewGuestInfo("John Doe", "john@example.com", "+1234567890")}

	reservation, err := NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		NewMoney(30000, "USD"),
		guests,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if reservation == nil {
		t.Fatal("expected reservation, got nil")
	}
	if reservation.Status != StatusPending {
		t.Errorf("expected status %s, got %s", StatusPending, reservation.Status)
	}
}

func Test_NewReservation_With_CheckOut_Before_CheckIn_Should_Return_Error(t *testing.T) {
	checkIn := time.Now().AddDate(0, 0, 10)
	checkOut := time.Now().AddDate(0, 0, 7) // Before check-in
	dateRange := NewDateRange(checkIn, checkOut)
	guests := []GuestInfo{NewGuestInfo("John Doe", "john@example.com", "+1234567890")}

	_, err := NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		NewMoney(30000, "USD"),
		guests,
	)

	if err != ErrInvalidDateRange {
		t.Errorf("expected ErrInvalidDateRange, got %v", err)
	}
}

func Test_NewReservation_With_CheckIn_In_Past_Should_Return_Error(t *testing.T) {
	checkIn := time.Now().AddDate(0, 0, -7)  // 7 days ago
	checkOut := time.Now().AddDate(0, 0, -5) // 5 days ago
	dateRange := NewDateRange(checkIn, checkOut)
	guests := []GuestInfo{NewGuestInfo("John Doe", "john@example.com", "+1234567890")}

	_, err := NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		NewMoney(30000, "USD"),
		guests,
	)

	if err != ErrCheckInPast {
		t.Errorf("expected ErrCheckInPast, got %v", err)
	}
}

func Test_NewReservation_With_Same_Day_CheckIn_CheckOut_Should_Return_Error(t *testing.T) {
	sameDay := time.Now().AddDate(0, 0, 7)
	dateRange := NewDateRange(sameDay, sameDay)
	guests := []GuestInfo{NewGuestInfo("John Doe", "john@example.com", "+1234567890")}

	_, err := NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		NewMoney(30000, "USD"),
		guests,
	)

	if err != ErrMinimumStay {
		t.Errorf("expected ErrMinimumStay, got %v", err)
	}
}

func Test_NewReservation_With_No_Guests_Should_Return_Error(t *testing.T) {
	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := NewDateRange(checkIn, checkOut)

	_, err := NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		NewMoney(30000, "USD"),
		[]GuestInfo{}, // Empty guests
	)

	if err != ErrNoGuests {
		t.Errorf("expected ErrNoGuests, got %v", err)
	}
}

// Test State Transitions

func Test_Reservation_Confirm_From_Pending_Should_Change_Status(t *testing.T) {
	reservation := createValidReservation(t)

	err := reservation.Confirm()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if reservation.Status != StatusConfirmed {
		t.Errorf("expected status %s, got %s", StatusConfirmed, reservation.Status)
	}
}

func Test_Reservation_Confirm_From_Confirmed_Should_Return_Error(t *testing.T) {
	reservation := createValidReservation(t)
	_ = reservation.Confirm()

	err := reservation.Confirm()

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func Test_Reservation_Activate_From_Confirmed_Should_Change_Status(t *testing.T) {
	reservation := createValidReservation(t)
	_ = reservation.Confirm()

	err := reservation.Activate()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if reservation.Status != StatusActive {
		t.Errorf("expected status %s, got %s", StatusActive, reservation.Status)
	}
}

func Test_Reservation_Activate_From_Pending_Should_Return_Error(t *testing.T) {
	reservation := createValidReservation(t)

	err := reservation.Activate()

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func Test_Reservation_Complete_From_Active_Should_Change_Status(t *testing.T) {
	reservation := createValidReservation(t)
	_ = reservation.Confirm()
	_ = reservation.Activate()

	err := reservation.Complete()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if reservation.Status != StatusCompleted {
		t.Errorf("expected status %s, got %s", StatusCompleted, reservation.Status)
	}
}

func Test_Reservation_Complete_From_Pending_Should_Return_Error(t *testing.T) {
	reservation := createValidReservation(t)

	err := reservation.Complete()

	if err == nil {
		t.Error("expected error, got nil")
	}
}

// Test Cancellation

func Test_Reservation_Cancel_From_Pending_Should_Change_Status(t *testing.T) {
	reservation := createValidReservation(t)
	reason := "guest changed plans"

	err := reservation.Cancel(reason)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if reservation.Status != StatusCancelled {
		t.Errorf("expected status %s, got %s", StatusCancelled, reservation.Status)
	}
	if reservation.CancellationReason != reason {
		t.Errorf("expected reason %s, got %s", reason, reservation.CancellationReason)
	}
}

func Test_Reservation_Cancel_Already_Cancelled_Should_Return_Error(t *testing.T) {
	reservation := createValidReservation(t)
	_ = reservation.Cancel("first cancellation")

	err := reservation.Cancel("second cancellation")

	if err != ErrAlreadyCancelled {
		t.Errorf("expected ErrAlreadyCancelled, got %v", err)
	}
}

func Test_Reservation_Cancel_Completed_Should_Return_Error(t *testing.T) {
	reservation := createValidReservation(t)
	_ = reservation.Confirm()
	_ = reservation.Activate()
	_ = reservation.Complete()

	err := reservation.Cancel("too late")

	if err != ErrCannotCancelCompleted {
		t.Errorf("expected ErrCannotCancelCompleted, got %v", err)
	}
}

func Test_Reservation_Cancel_Active_Should_Return_Error(t *testing.T) {
	reservation := createValidReservation(t)
	_ = reservation.Confirm()
	_ = reservation.Activate()

	err := reservation.Cancel("guest already checked in")

	if err != ErrCannotCancelActive {
		t.Errorf("expected ErrCannotCancelActive, got %v", err)
	}
}

func Test_Reservation_CanBeCancelled_Within_24_Hours_Should_Return_False(t *testing.T) {
	// Create reservation with check-in in 12 hours
	checkIn := time.Now().Add(12 * time.Hour)
	checkOut := checkIn.AddDate(0, 0, 3)
	dateRange := NewDateRange(checkIn, checkOut)
	guests := []GuestInfo{NewGuestInfo("John Doe", "john@example.com", "+1234567890")}

	reservation, _ := NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		NewMoney(30000, "USD"),
		guests,
	)

	if reservation.CanBeCancelled() {
		t.Error("expected CanBeCancelled to return false within 24 hours")
	}
}

func Test_Reservation_CanBeCancelled_More_Than_24_Hours_Should_Return_True(t *testing.T) {
	reservation := createValidReservation(t)

	if !reservation.CanBeCancelled() {
		t.Error("expected CanBeCancelled to return true more than 24 hours before check-in")
	}
}

// Test Overlapping

func Test_Reservation_IsOverlapping_With_Same_Room_And_Overlapping_Dates_Should_Return_True(t *testing.T) {
	reservation1 := createValidReservation(t)

	// Create overlapping reservation (starts during first reservation)
	checkIn := reservation1.DateRange.CheckIn.AddDate(0, 0, 1)
	checkOut := reservation1.DateRange.CheckOut.AddDate(0, 0, 1)
	dateRange := NewDateRange(checkIn, checkOut)
	guests := []GuestInfo{NewGuestInfo("Jane Doe", "jane@example.com", "+1234567890")}

	reservation2, _ := NewReservation(
		"res-002",
		"guest-002",
		reservation1.RoomID, // Same room
		dateRange,
		NewMoney(30000, "USD"),
		guests,
	)

	if !reservation1.IsOverlapping(reservation2) {
		t.Error("expected reservations to overlap")
	}
}

func Test_Reservation_IsOverlapping_With_Different_Rooms_Should_Return_False(t *testing.T) {
	reservation1 := createValidReservation(t)

	checkIn := reservation1.DateRange.CheckIn
	checkOut := reservation1.DateRange.CheckOut
	dateRange := NewDateRange(checkIn, checkOut)
	guests := []GuestInfo{NewGuestInfo("Jane Doe", "jane@example.com", "+1234567890")}

	reservation2, _ := NewReservation(
		"res-002",
		"guest-002",
		"room-102", // Different room
		dateRange,
		NewMoney(30000, "USD"),
		guests,
	)

	if reservation1.IsOverlapping(reservation2) {
		t.Error("expected reservations with different rooms not to overlap")
	}
}

func Test_Reservation_IsOverlapping_With_SameDay_Checkout_Checkin_Should_Return_False(t *testing.T) {
	checkIn1 := time.Now().AddDate(0, 0, 7)
	checkOut1 := time.Now().AddDate(0, 0, 10)
	dateRange1 := NewDateRange(checkIn1, checkOut1)
	guests := []GuestInfo{NewGuestInfo("John Doe", "john@example.com", "+1234567890")}

	reservation1, _ := NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange1,
		NewMoney(30000, "USD"),
		guests,
	)

	// Second reservation starts exactly when first ends
	checkIn2 := checkOut1
	checkOut2 := checkOut1.AddDate(0, 0, 3)
	dateRange2 := NewDateRange(checkIn2, checkOut2)

	reservation2, _ := NewReservation(
		"res-002",
		"guest-002",
		"room-101",
		dateRange2,
		NewMoney(30000, "USD"),
		guests,
	)

	if reservation1.IsOverlapping(reservation2) {
		t.Error("expected same-day checkout/check-in not to overlap")
	}
}

func Test_Reservation_IsOverlapping_With_Cancelled_Reservation_Should_Return_False(t *testing.T) {
	reservation1 := createValidReservation(t)
	_ = reservation1.Cancel("cancelled")

	checkIn := reservation1.DateRange.CheckIn
	checkOut := reservation1.DateRange.CheckOut
	dateRange := NewDateRange(checkIn, checkOut)
	guests := []GuestInfo{NewGuestInfo("Jane Doe", "jane@example.com", "+1234567890")}

	reservation2, _ := NewReservation(
		"res-002",
		"guest-002",
		reservation1.RoomID,
		dateRange,
		NewMoney(30000, "USD"),
		guests,
	)

	if reservation1.IsOverlapping(reservation2) {
		t.Error("expected cancelled reservation not to count as overlapping")
	}
}

// Test Helper Methods

func Test_Reservation_Nights_Should_Calculate_Correctly(t *testing.T) {
	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10) // 3 nights
	dateRange := NewDateRange(checkIn, checkOut)
	guests := []GuestInfo{NewGuestInfo("John Doe", "john@example.com", "+1234567890")}

	reservation, _ := NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		NewMoney(30000, "USD"),
		guests,
	)

	nights := reservation.Nights()
	if nights != 3 {
		t.Errorf("expected 3 nights, got %d", nights)
	}
}

func Test_Reservation_DaysUntilCheckIn_Should_Calculate_Correctly(t *testing.T) {
	checkIn := time.Now().AddDate(0, 0, 7) // 7 days from now
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := NewDateRange(checkIn, checkOut)
	guests := []GuestInfo{NewGuestInfo("John Doe", "john@example.com", "+1234567890")}

	reservation, _ := NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		NewMoney(30000, "USD"),
		guests,
	)

	days := reservation.DaysUntilCheckIn()
	// Should be approximately 7 days (allow for truncation)
	if days < 6 || days > 7 {
		t.Errorf("expected ~7 days, got %d", days)
	}
}

// Helper function to create a valid reservation for testing
func createValidReservation(t *testing.T) *Reservation {
	t.Helper()
	checkIn := time.Now().AddDate(0, 0, 7)  // 7 days from now
	checkOut := time.Now().AddDate(0, 0, 10) // 10 days from now
	dateRange := NewDateRange(checkIn, checkOut)
	guests := []GuestInfo{NewGuestInfo("John Doe", "john@example.com", "+1234567890")}

	reservation, err := NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		NewMoney(30000, "USD"),
		guests,
	)

	if err != nil {
		t.Fatalf("failed to create test reservation: %v", err)
	}

	return reservation
}
