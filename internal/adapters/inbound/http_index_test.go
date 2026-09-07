package inbound_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/web"
	"github.com/andygeiss/hotel-booking/internal/adapters/inbound"
)

// ============================================================================
// Test Assets
// ============================================================================

// ============================================================================
// HttpViewIndex Tests
// ============================================================================

func Test_HttpViewIndex_Without_Session_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	v := newTestView(t)

	handler := inbound.HttpViewIndex(v, testApp())
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
	location := rec.Header().Get("Location")
	assert.That(t, "location must contain login", containsString(location, "/ui/login"), true)
}

func Test_HttpViewIndex_With_Empty_SessionID_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	v := newTestView(t)

	handler := inbound.HttpViewIndex(v, testApp())
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	// Add empty session ID to context
	ctx := context.WithValue(req.Context(), web.ContextSessionID, "")
	ctx = context.WithValue(ctx, web.ContextEmail, "")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
}

func Test_HttpViewIndex_With_SessionID_But_Empty_Email_Should_Redirect_To_Login(t *testing.T) {
	// Arrange - simulates the case after logout where session is deleted but cookie remains

	v := newTestView(t)

	handler := inbound.HttpViewIndex(v, testApp())
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	// Session ID exists (from stale cookie) but email is empty (session deleted server-side)
	ctx := context.WithValue(req.Context(), web.ContextSessionID, "stale-session-id")
	ctx = context.WithValue(ctx, web.ContextEmail, "")
	ctx = context.WithValue(ctx, web.ContextIssuer, "")
	ctx = context.WithValue(ctx, web.ContextName, "")
	ctx = context.WithValue(ctx, web.ContextSubject, "")
	ctx = context.WithValue(ctx, web.ContextVerified, false)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
	location := rec.Header().Get("Location")
	assert.That(t, "location must redirect to login", containsString(location, "/ui/login"), true)
}

func Test_HttpViewIndex_With_Valid_Session_Should_Return_200(t *testing.T) {
	// Arrange
	v := newTestView(t)

	handler := inbound.HttpViewIndex(v, testApp())
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)

	// Add session context values
	ctx := req.Context()
	ctx = context.WithValue(ctx, web.ContextSessionID, "test-session-123")
	ctx = context.WithValue(ctx, web.ContextEmail, "test@example.com")
	ctx = context.WithValue(ctx, web.ContextIssuer, "https://issuer.example.com")
	ctx = context.WithValue(ctx, web.ContextName, "Test User")
	ctx = context.WithValue(ctx, web.ContextSubject, "user-subject-456")
	ctx = context.WithValue(ctx, web.ContextVerified, true)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewIndex_With_Valid_Session_Should_Render_User_Data(t *testing.T) {
	// Arrange
	v := newTestView(t)

	handler := inbound.HttpViewIndex(v, testApp())
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)

	// Add session context values
	ctx := req.Context()
	ctx = context.WithValue(ctx, web.ContextSessionID, "test-session-123")
	ctx = context.WithValue(ctx, web.ContextEmail, "test@example.com")
	ctx = context.WithValue(ctx, web.ContextIssuer, "https://issuer.example.com")
	ctx = context.WithValue(ctx, web.ContextName, "Test User")
	ctx = context.WithValue(ctx, web.ContextSubject, "user-subject-456")
	ctx = context.WithValue(ctx, web.ContextVerified, true)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain user email", containsString(bodyStr, "test@example.com"), true)
	assert.That(t, "body must contain user name", containsString(bodyStr, "Test User"), true)
	assert.That(t, "body must contain session ID", containsString(bodyStr, "test-session-123"), true)
}

// containsString is a helper function to check if a string contains a substring.
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
