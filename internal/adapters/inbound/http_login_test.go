package inbound_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/hotel-booking/internal/adapters/inbound"
)

// ============================================================================
// Test Assets
// ============================================================================

// ============================================================================
// HttpViewLogin Tests
// ============================================================================

func Test_HttpViewLogin_With_Request_Should_Return_200(t *testing.T) {
	// Arrange
	v := newTestView(t)

	handler := inbound.HttpViewLogin(v, testApp())
	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewLogin_With_Request_Should_Render_Template(t *testing.T) {
	// Arrange
	v := newTestView(t)

	handler := inbound.HttpViewLogin(v, testApp())
	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain app name", containsString(bodyStr, "TestApp"), true)
}

func Test_HttpViewLogin_With_Request_Should_Return_HTML_Content_Type(t *testing.T) {
	// Arrange
	v := newTestView(t)

	handler := inbound.HttpViewLogin(v, testApp())
	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	contentType := rec.Header().Get("Content-Type")
	assert.That(t, "content type must be text/html", containsString(contentType, "text/html"), true)
}
