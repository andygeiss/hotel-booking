package payment_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/cloud-native-utils/mcp"
	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// ============================================================================
// Mock Implementations for Tools Tests
// ============================================================================

type toolsMockPaymentRepository struct {
	payments  map[payment.PaymentID]payment.Payment
	createErr error
	readErr   error
	updateErr error
}

func newToolsMockPaymentRepository() *toolsMockPaymentRepository {
	return &toolsMockPaymentRepository{
		payments: make(map[payment.PaymentID]payment.Payment),
	}
}

func (m *toolsMockPaymentRepository) Create(ctx context.Context, id payment.PaymentID, p payment.Payment) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.payments[id] = p
	return nil
}

func (m *toolsMockPaymentRepository) Read(ctx context.Context, id payment.PaymentID) (*payment.Payment, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	p, ok := m.payments[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &p, nil
}

func (m *toolsMockPaymentRepository) Update(ctx context.Context, id payment.PaymentID, p payment.Payment) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.payments[id] = p
	return nil
}

func (m *toolsMockPaymentRepository) Delete(ctx context.Context, id payment.PaymentID) error {
	delete(m.payments, id)
	return nil
}

func (m *toolsMockPaymentRepository) ReadAll(ctx context.Context) ([]payment.Payment, error) {
	result := make([]payment.Payment, 0, len(m.payments))
	for _, p := range m.payments {
		result = append(result, p)
	}
	return result, nil
}

type toolsMockPaymentGateway struct {
	authorizeTransactionID string
	authorizeErr           error
	captureErr             error
	refundErr              error
}

func (m *toolsMockPaymentGateway) Authorize(ctx context.Context, p *payment.Payment) (string, error) {
	if m.authorizeErr != nil {
		return "", m.authorizeErr
	}
	return m.authorizeTransactionID, nil
}

func (m *toolsMockPaymentGateway) Capture(ctx context.Context, transactionID string, amount shared.Money) error {
	return m.captureErr
}

func (m *toolsMockPaymentGateway) Refund(ctx context.Context, transactionID string, amount shared.Money) error {
	return m.refundErr
}

type toolsMockEventPublisher struct {
	published []event.Event
	err       error
}

func (m *toolsMockEventPublisher) Publish(ctx context.Context, evt event.Event) error {
	if m.err != nil {
		return m.err
	}
	m.published = append(m.published, evt)
	return nil
}

// ============================================================================
// Test Helpers
// ============================================================================

func createToolsPaymentTestService(repo *toolsMockPaymentRepository, gateway *toolsMockPaymentGateway, publisher *toolsMockEventPublisher) *payment.Service {
	return payment.NewService(repo, gateway, publisher)
}

func toolsPaymentTestMoney() shared.Money {
	return shared.NewMoney(10000, "USD")
}

// ============================================================================
// RegisterTools Tests
// ============================================================================

func Test_RegisterTools_Should_Register_All_Tools(t *testing.T) {
	// Arrange
	repo := newToolsMockPaymentRepository()
	gateway := &toolsMockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &toolsMockEventPublisher{}
	service := createToolsPaymentTestService(repo, gateway, publisher)
	server := mcp.NewServer("test-server", "1.0.0")

	// Act
	payment.RegisterTools(server, service)

	// Assert
	tools := server.Tools()
	assert.That(t, "must register 3 tools", len(tools), 3)

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Definition.Name] = true
	}
	assert.That(t, "get_payment must be registered", toolNames["get_payment"], true)
	assert.That(t, "capture_payment must be registered", toolNames["capture_payment"], true)
	assert.That(t, "refund_payment must be registered", toolNames["refund_payment"], true)
}

// ============================================================================
// GetPayment Tool Tests
// ============================================================================

func Test_GetPaymentTool_Should_Return_Payment_JSON(t *testing.T) {
	// Arrange
	repo := newToolsMockPaymentRepository()
	gateway := &toolsMockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &toolsMockEventPublisher{}
	service := createToolsPaymentTestService(repo, gateway, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	payment.RegisterTools(server, service)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	// Create a payment first
	_, _ = service.AuthorizePayment(ctx, id, "res-001", toolsPaymentTestMoney(), "credit_card")

	// Get the tool and call it
	tools := server.Tools()
	var getTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "get_payment" {
			getTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "get_payment",
		Arguments: map[string]any{"id": "pay-001"},
	}

	// Act
	result, err := getTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must have 1 content block", len(result.Content), 1)
	assert.That(t, "content must contain payment ID", strings.Contains(result.Content[0].Text, "pay-001"), true)
}

func Test_GetPaymentTool_When_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newToolsMockPaymentRepository()
	gateway := &toolsMockPaymentGateway{}
	publisher := &toolsMockEventPublisher{}
	service := createToolsPaymentTestService(repo, gateway, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	payment.RegisterTools(server, service)

	ctx := context.Background()

	tools := server.Tools()
	var getTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "get_payment" {
			getTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "get_payment",
		Arguments: map[string]any{"id": "non-existent"},
	}

	// Act
	_, err := getTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

// ============================================================================
// CapturePayment Tool Tests
// ============================================================================

func Test_CapturePaymentTool_Should_Capture_Payment(t *testing.T) {
	// Arrange
	repo := newToolsMockPaymentRepository()
	gateway := &toolsMockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &toolsMockEventPublisher{}
	service := createToolsPaymentTestService(repo, gateway, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	payment.RegisterTools(server, service)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	// Create and authorize a payment first
	_, _ = service.AuthorizePayment(ctx, id, "res-001", toolsPaymentTestMoney(), "credit_card")

	tools := server.Tools()
	var captureTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "capture_payment" {
			captureTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "capture_payment",
		Arguments: map[string]any{"id": "pay-001"},
	}

	// Act
	result, err := captureTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must have 1 content block", len(result.Content), 1)
	assert.That(t, "content must confirm capture", strings.Contains(result.Content[0].Text, "captured successfully"), true)

	// Verify payment is captured
	p, _ := service.GetPayment(ctx, id)
	assert.That(t, "status must be captured", p.Status, payment.StatusCaptured)
}

func Test_CapturePaymentTool_When_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newToolsMockPaymentRepository()
	gateway := &toolsMockPaymentGateway{}
	publisher := &toolsMockEventPublisher{}
	service := createToolsPaymentTestService(repo, gateway, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	payment.RegisterTools(server, service)

	ctx := context.Background()

	tools := server.Tools()
	var captureTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "capture_payment" {
			captureTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "capture_payment",
		Arguments: map[string]any{"id": "non-existent"},
	}

	// Act
	_, err := captureTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_CapturePaymentTool_When_Gateway_Fails_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newToolsMockPaymentRepository()
	gateway := &toolsMockPaymentGateway{
		authorizeTransactionID: "tx-12345",
		captureErr:             errors.New("capture failed"),
	}
	publisher := &toolsMockEventPublisher{}
	service := createToolsPaymentTestService(repo, gateway, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	payment.RegisterTools(server, service)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	// Create and authorize a payment first
	_, _ = service.AuthorizePayment(ctx, id, "res-001", toolsPaymentTestMoney(), "credit_card")

	tools := server.Tools()
	var captureTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "capture_payment" {
			captureTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "capture_payment",
		Arguments: map[string]any{"id": "pay-001"},
	}

	// Act
	_, err := captureTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

// ============================================================================
// RefundPayment Tool Tests
// ============================================================================

func Test_RefundPaymentTool_Should_Refund_Payment(t *testing.T) {
	// Arrange
	repo := newToolsMockPaymentRepository()
	gateway := &toolsMockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &toolsMockEventPublisher{}
	service := createToolsPaymentTestService(repo, gateway, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	payment.RegisterTools(server, service)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	// Create, authorize, and capture a payment first
	_, _ = service.AuthorizePayment(ctx, id, "res-001", toolsPaymentTestMoney(), "credit_card")
	_ = service.CapturePayment(ctx, id)

	tools := server.Tools()
	var refundTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "refund_payment" {
			refundTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "refund_payment",
		Arguments: map[string]any{"id": "pay-001"},
	}

	// Act
	result, err := refundTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must have 1 content block", len(result.Content), 1)
	assert.That(t, "content must confirm refund", strings.Contains(result.Content[0].Text, "refunded successfully"), true)

	// Verify payment is refunded
	p, _ := service.GetPayment(ctx, id)
	assert.That(t, "status must be refunded", p.Status, payment.StatusRefunded)
}

func Test_RefundPaymentTool_When_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newToolsMockPaymentRepository()
	gateway := &toolsMockPaymentGateway{}
	publisher := &toolsMockEventPublisher{}
	service := createToolsPaymentTestService(repo, gateway, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	payment.RegisterTools(server, service)

	ctx := context.Background()

	tools := server.Tools()
	var refundTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "refund_payment" {
			refundTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "refund_payment",
		Arguments: map[string]any{"id": "non-existent"},
	}

	// Act
	_, err := refundTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_RefundPaymentTool_When_Not_Captured_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newToolsMockPaymentRepository()
	gateway := &toolsMockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &toolsMockEventPublisher{}
	service := createToolsPaymentTestService(repo, gateway, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	payment.RegisterTools(server, service)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	// Only authorize, don't capture
	_, _ = service.AuthorizePayment(ctx, id, "res-001", toolsPaymentTestMoney(), "credit_card")

	tools := server.Tools()
	var refundTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "refund_payment" {
			refundTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "refund_payment",
		Arguments: map[string]any{"id": "pay-001"},
	}

	// Act
	_, err := refundTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_RefundPaymentTool_When_Gateway_Fails_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newToolsMockPaymentRepository()
	gateway := &toolsMockPaymentGateway{
		authorizeTransactionID: "tx-12345",
		refundErr:              errors.New("refund failed"),
	}
	publisher := &toolsMockEventPublisher{}
	service := createToolsPaymentTestService(repo, gateway, publisher)
	server := mcp.NewServer("test-server", "1.0.0")
	payment.RegisterTools(server, service)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	// Create, authorize, and capture a payment first
	_, _ = service.AuthorizePayment(ctx, id, "res-001", toolsPaymentTestMoney(), "credit_card")
	_ = service.CapturePayment(ctx, id)

	tools := server.Tools()
	var refundTool mcp.Tool
	for _, tool := range tools {
		if tool.Definition.Name == "refund_payment" {
			refundTool = tool
			break
		}
	}

	params := mcp.ToolsCallParams{
		Name:      "refund_payment",
		Arguments: map[string]any{"id": "pay-001"},
	}

	// Act
	_, err := refundTool.Handler(ctx, params)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}
