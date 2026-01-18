package payment

import (
	"context"
	"encoding/json"

	"github.com/andygeiss/cloud-native-utils/mcp"
)

// RegisterTools registers all payment MCP tools with the server.
func RegisterTools(server *mcp.Server, service *Service) {
	server.RegisterTool(newGetPaymentTool(service))
	server.RegisterTool(newCapturePaymentTool(service))
	server.RegisterTool(newRefundPaymentTool(service))
}

// newGetPaymentTool creates a new get_payment tool.
func newGetPaymentTool(service *Service) mcp.Tool {
	return mcp.NewTool(
		"get_payment",
		"Get payment details by ID. Returns payment status, amount, and transaction info.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"id": mcp.NewStringProperty("The payment ID"),
			},
			[]string{"id"},
		),
		func(ctx context.Context, params mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
			id, _ := params.Arguments["id"].(string)
			payment, err := service.GetPayment(ctx, PaymentID(id))
			if err != nil {
				return mcp.ToolsCallResult{}, err
			}
			data, _ := json.MarshalIndent(payment, "", "  ")
			return mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.NewTextContent(string(data))},
			}, nil
		},
	)
}

// newCapturePaymentTool creates a new tool for capturing.
func newCapturePaymentTool(service *Service) mcp.Tool {
	return mcp.NewTool(
		"capture_payment",
		"Capture an authorized payment. Payment must be in 'authorized' status.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"id": mcp.NewStringProperty("The payment ID"),
			},
			[]string{"id"},
		),
		func(ctx context.Context, params mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
			id, _ := params.Arguments["id"].(string)
			err := service.CapturePayment(ctx, PaymentID(id))
			if err != nil {
				return mcp.ToolsCallResult{}, err
			}
			return mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.NewTextContent("Payment captured successfully")},
			}, nil
		},
	)
}

// newRefundPaymentTool creates a tool for refunding.
func newRefundPaymentTool(service *Service) mcp.Tool {
	return mcp.NewTool(
		"refund_payment",
		"Refund a captured payment. Payment must be in 'captured' status.",
		mcp.NewObjectSchema(
			map[string]mcp.Property{
				"id": mcp.NewStringProperty("The payment ID"),
			},
			[]string{"id"},
		),
		func(ctx context.Context, params mcp.ToolsCallParams) (mcp.ToolsCallResult, error) {
			id, _ := params.Arguments["id"].(string)
			err := service.RefundPayment(ctx, PaymentID(id))
			if err != nil {
				return mcp.ToolsCallResult{}, err
			}
			return mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.NewTextContent("Payment refunded successfully")},
			}, nil
		},
	)
}
