package booking

import (
	"testing"
)

// Test Payment Creation

func Test_PaymentID_With_String_Value_Should_Be_Assignable(t *testing.T) {
	var id PaymentID = "pay-123"
	if id != "pay-123" {
		t.Errorf("expected pay-123, got %s", id)
	}
}

func Test_NewPayment_Should_Create_With_Pending_Status(t *testing.T) {
	payment := NewPayment(
		"pay-001",
		"res-001",
		NewMoney(30000, "USD"),
		"credit_card",
	)

	if payment == nil {
		t.Fatal("expected payment, got nil")
	}
	if payment.Status != PaymentPending {
		t.Errorf("expected status %s, got %s", PaymentPending, payment.Status)
	}
	if len(payment.Attempts) != 0 {
		t.Errorf("expected 0 attempts, got %d", len(payment.Attempts))
	}
}

// Test Payment Authorization

func Test_Payment_Authorize_From_Pending_Should_Change_Status(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	transactionID := "txn-12345"

	err := payment.Authorize(transactionID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if payment.Status != PaymentAuthorized {
		t.Errorf("expected status %s, got %s", PaymentAuthorized, payment.Status)
	}
	if payment.TransactionID != transactionID {
		t.Errorf("expected transaction ID %s, got %s", transactionID, payment.TransactionID)
	}
	if len(payment.Attempts) != 1 {
		t.Errorf("expected 1 attempt, got %d", len(payment.Attempts))
	}
}

func Test_Payment_Authorize_Already_Authorized_Should_Return_Error(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Authorize("txn-12345")

	err := payment.Authorize("txn-67890")

	if err != ErrAlreadyAuthorized {
		t.Errorf("expected ErrAlreadyAuthorized, got %v", err)
	}
}

func Test_Payment_Authorize_From_Captured_Should_Return_Error(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Authorize("txn-12345")
	_ = payment.Capture()

	err := payment.Authorize("txn-67890")

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func Test_Payment_Authorize_From_Failed_Should_Succeed(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Fail("insufficient_funds", "Insufficient funds")

	err := payment.Authorize("txn-12345")

	if err != nil {
		t.Errorf("expected no error for retry, got %v", err)
	}
	if payment.Status != PaymentAuthorized {
		t.Errorf("expected status %s, got %s", PaymentAuthorized, payment.Status)
	}
}

// Test Payment Capture

func Test_Payment_Capture_From_Authorized_Should_Change_Status(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Authorize("txn-12345")

	err := payment.Capture()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if payment.Status != PaymentCaptured {
		t.Errorf("expected status %s, got %s", PaymentCaptured, payment.Status)
	}
}

func Test_Payment_Capture_Without_Authorization_Should_Return_Error(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")

	err := payment.Capture()

	if err != ErrNotAuthorized {
		t.Errorf("expected ErrNotAuthorized, got %v", err)
	}
}

func Test_Payment_Capture_Already_Captured_Should_Return_Error(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Authorize("txn-12345")
	_ = payment.Capture()

	err := payment.Capture()

	if err != ErrAlreadyCaptured {
		t.Errorf("expected ErrAlreadyCaptured, got %v", err)
	}
}

// Test Payment Failure

func Test_Payment_Fail_From_Pending_Should_Change_Status(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	errorCode := "card_declined"
	errorMsg := "Card was declined"

	err := payment.Fail(errorCode, errorMsg)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if payment.Status != PaymentFailed {
		t.Errorf("expected status %s, got %s", PaymentFailed, payment.Status)
	}
	if len(payment.Attempts) != 1 {
		t.Errorf("expected 1 attempt, got %d", len(payment.Attempts))
	}
	if payment.Attempts[0].ErrorCode != errorCode {
		t.Errorf("expected error code %s, got %s", errorCode, payment.Attempts[0].ErrorCode)
	}
	if payment.Attempts[0].ErrorMsg != errorMsg {
		t.Errorf("expected error message %s, got %s", errorMsg, payment.Attempts[0].ErrorMsg)
	}
}

func Test_Payment_Fail_From_Authorized_Should_Change_Status(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Authorize("txn-12345")

	err := payment.Fail("timeout", "Gateway timeout")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if payment.Status != PaymentFailed {
		t.Errorf("expected status %s, got %s", PaymentFailed, payment.Status)
	}
}

func Test_Payment_Fail_From_Captured_Should_Return_Error(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Authorize("txn-12345")
	_ = payment.Capture()

	err := payment.Fail("error", "Cannot fail captured payment")

	if err == nil {
		t.Error("expected error, got nil")
	}
}

// Test Payment Refund

func Test_Payment_Refund_From_Captured_Should_Change_Status(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Authorize("txn-12345")
	_ = payment.Capture()

	err := payment.Refund()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if payment.Status != PaymentRefunded {
		t.Errorf("expected status %s, got %s", PaymentRefunded, payment.Status)
	}
}

func Test_Payment_Refund_Without_Capture_Should_Return_Error(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Authorize("txn-12345")

	err := payment.Refund()

	if err != ErrCannotRefund {
		t.Errorf("expected ErrCannotRefund, got %v", err)
	}
}

func Test_Payment_Refund_Already_Refunded_Should_Return_Error(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Authorize("txn-12345")
	_ = payment.Capture()
	_ = payment.Refund()

	err := payment.Refund()

	if err != ErrAlreadyRefunded {
		t.Errorf("expected ErrAlreadyRefunded, got %v", err)
	}
}

// Test Helper Methods

func Test_Payment_IsSuccessful_With_Captured_Payment_Should_Return_True(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Authorize("txn-12345")
	_ = payment.Capture()

	if !payment.IsSuccessful() {
		t.Error("expected IsSuccessful to return true for captured payment")
	}
}

func Test_Payment_IsSuccessful_With_Pending_Payment_Should_Return_False(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")

	if payment.IsSuccessful() {
		t.Error("expected IsSuccessful to return false for pending payment")
	}
}

func Test_Payment_CanBeRetried_With_Failed_Payment_Should_Return_True(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Fail("card_declined", "Card declined")

	if !payment.CanBeRetried() {
		t.Error("expected CanBeRetried to return true for failed payment")
	}
}

func Test_Payment_CanBeRetried_With_Captured_Payment_Should_Return_False(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")
	_ = payment.Authorize("txn-12345")
	_ = payment.Capture()

	if payment.CanBeRetried() {
		t.Error("expected CanBeRetried to return false for captured payment")
	}
}

func Test_Payment_CanBeRetried_After_3_Failed_Attempts_Should_Return_False(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")

	// Fail 3 times
	_ = payment.Fail("error1", "First failure")
	_ = payment.Authorize("txn-1")
	_ = payment.Fail("error2", "Second failure")
	_ = payment.Authorize("txn-2")
	_ = payment.Fail("error3", "Third failure")

	if payment.CanBeRetried() {
		t.Error("expected CanBeRetried to return false after 3 failed attempts")
	}
}

// Test Payment Attempts Tracking

func Test_Payment_Attempts_Should_Track_History(t *testing.T) {
	payment := NewPayment("pay-001", "res-001", NewMoney(30000, "USD"), "credit_card")

	_ = payment.Fail("card_declined", "Card declined")
	_ = payment.Authorize("txn-12345")
	_ = payment.Capture()

	if len(payment.Attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", len(payment.Attempts))
	}

	// Check attempt order
	if payment.Attempts[0].Status != PaymentFailed {
		t.Errorf("expected first attempt to be failed, got %s", payment.Attempts[0].Status)
	}
	if payment.Attempts[1].Status != PaymentAuthorized {
		t.Errorf("expected second attempt to be authorized, got %s", payment.Attempts[1].Status)
	}
	if payment.Attempts[2].Status != PaymentCaptured {
		t.Errorf("expected third attempt to be captured, got %s", payment.Attempts[2].Status)
	}
}
