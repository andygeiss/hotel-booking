package inbound

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
)

func Test_WithBearerAuth_Missing_Authorization_Header_Should_Return_401(t *testing.T) {
	// Arrange
	handler := WithBearerAuth(nil, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 401", rec.Code, http.StatusUnauthorized)
	assert.That(t, "content type must be application/json", rec.Header().Get("Content-Type"), "application/json")
	assert.That(t, "body must contain error message", strings.Contains(rec.Body.String(), "Missing Authorization header"), true)
}

func Test_WithBearerAuth_Invalid_Authorization_Scheme_Should_Return_401(t *testing.T) {
	// Arrange
	handler := WithBearerAuth(nil, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 401", rec.Code, http.StatusUnauthorized)
	assert.That(t, "content type must be application/json", rec.Header().Get("Content-Type"), "application/json")
	assert.That(t, "body must contain error message", strings.Contains(rec.Body.String(), "Invalid Authorization scheme"), true)
}

func Test_WithBearerAuth_Invalid_Token_Should_Return_401(t *testing.T) {
	// Arrange - use nil verifier which will cause verification to fail
	handler := WithBearerAuth(nil, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 401", rec.Code, http.StatusUnauthorized)
	assert.That(t, "content type must be application/json", rec.Header().Get("Content-Type"), "application/json")
	assert.That(t, "body must contain error message", strings.Contains(rec.Body.String(), "Invalid or expired token"), true)
}

func Test_WithBearerAuth_Empty_Bearer_Token_Should_Return_401(t *testing.T) {
	// Arrange
	handler := WithBearerAuth(nil, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 401", rec.Code, http.StatusUnauthorized)
}

func Test_WithBearerAuth_Error_Response_Is_Valid_JSON_RPC(t *testing.T) {
	// Arrange
	handler := WithBearerAuth(nil, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert - check JSON-RPC error structure
	body := rec.Body.String()
	assert.That(t, "body must contain jsonrpc version", strings.Contains(body, `"jsonrpc":"2.0"`), true)
	assert.That(t, "body must contain error object", strings.Contains(body, `"error":`), true)
	assert.That(t, "body must contain error code", strings.Contains(body, `"code":-32001`), true)
	assert.That(t, "body must contain id null", strings.Contains(body, `"id":null`), true)
}
