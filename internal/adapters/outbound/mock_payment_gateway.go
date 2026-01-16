package outbound

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// MockPaymentGateway simulates a payment gateway for testing and demonstration
type MockPaymentGateway struct {
	ShouldFail      bool
	FailureRate     float64 // 0.0 to 1.0, probability of random failures
	transactions    map[string]booking.Money
}

// NewMockPaymentGateway creates a new mock payment gateway
func NewMockPaymentGateway() *MockPaymentGateway {
	return &MockPaymentGateway{
		ShouldFail:   false,
		FailureRate:  0.0,
		transactions: make(map[string]booking.Money),
	}
}

// Authorize simulates authorizing a payment
func (g *MockPaymentGateway) Authorize(ctx context.Context, payment *booking.Payment) (string, error) {
	// Check for forced or random failure
	if g.ShouldFail || (g.FailureRate > 0 && rand.Float64() < g.FailureRate) {
		return "", fmt.Errorf("payment authorization failed: insufficient funds")
	}

	// Generate mock transaction ID
	transactionID := fmt.Sprintf("txn_%s_%d", payment.ID, payment.Amount.Amount)

	// Store authorized amount
	g.transactions[transactionID] = payment.Amount

	return transactionID, nil
}

// Capture simulates capturing an authorized payment
func (g *MockPaymentGateway) Capture(ctx context.Context, transactionID string, amount booking.Money) error {
	// Check for forced or random failure
	if g.ShouldFail || (g.FailureRate > 0 && rand.Float64() < g.FailureRate) {
		return fmt.Errorf("payment capture failed: gateway timeout")
	}

	// Verify transaction exists
	authorizedAmount, exists := g.transactions[transactionID]
	if !exists {
		return fmt.Errorf("transaction %s not found", transactionID)
	}

	// Verify amount matches
	if authorizedAmount.Amount != amount.Amount || authorizedAmount.Currency != amount.Currency {
		return fmt.Errorf("capture amount mismatch: authorized %v, requested %v", authorizedAmount, amount)
	}

	// Mark as captured (in a real implementation, this would update external state)
	// For mock purposes, we just keep it in the map

	return nil
}

// Refund simulates refunding a captured payment
func (g *MockPaymentGateway) Refund(ctx context.Context, transactionID string, amount booking.Money) error {
	// Check for forced or random failure
	if g.ShouldFail || (g.FailureRate > 0 && rand.Float64() < g.FailureRate) {
		return fmt.Errorf("payment refund failed: gateway error")
	}

	// Verify transaction exists
	_, exists := g.transactions[transactionID]
	if !exists {
		return fmt.Errorf("transaction %s not found", transactionID)
	}

	// Remove from transactions (simulate refund)
	delete(g.transactions, transactionID)

	return nil
}

// SetShouldFail configures the mock to always fail (for testing error paths)
func (g *MockPaymentGateway) SetShouldFail(shouldFail bool) {
	g.ShouldFail = shouldFail
}

// SetFailureRate sets the probability of random failures (0.0 to 1.0)
func (g *MockPaymentGateway) SetFailureRate(rate float64) {
	g.FailureRate = rate
}

// Reset clears all transaction state
func (g *MockPaymentGateway) Reset() {
	g.transactions = make(map[string]booking.Money)
	g.ShouldFail = false
	g.FailureRate = 0.0
}
