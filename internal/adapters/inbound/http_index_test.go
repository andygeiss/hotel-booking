package inbound_test

import (
	"context"
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
)

// ============================================================================
// Test Assets
// ============================================================================

//go:embed testdata/assets/templates/*.tmpl testdata/assets/static/css/*.css
var indexTestAssets embed.FS

// ============================================================================
// HttpViewIndex Tests
// ============================================================================

func Test_HttpViewIndex_Without_Session_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(indexTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewIndex(e)
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
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(indexTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewIndex(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	// Add empty session ID to context
	ctx := context.WithValue(req.Context(), security.ContextSessionID, "")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
}

func Test_HttpViewIndex_With_Valid_Session_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(indexTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewIndex(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)

	// Add session context values
	ctx := req.Context()
	ctx = context.WithValue(ctx, security.ContextSessionID, "test-session-123")
	ctx = context.WithValue(ctx, security.ContextEmail, "test@example.com")
	ctx = context.WithValue(ctx, security.ContextIssuer, "https://issuer.example.com")
	ctx = context.WithValue(ctx, security.ContextName, "Test User")
	ctx = context.WithValue(ctx, security.ContextSubject, "user-subject-456")
	ctx = context.WithValue(ctx, security.ContextVerified, true)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewIndex_With_Valid_Session_Should_Render_User_Data(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(indexTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewIndex(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)

	// Add session context values
	ctx := req.Context()
	ctx = context.WithValue(ctx, security.ContextSessionID, "test-session-123")
	ctx = context.WithValue(ctx, security.ContextEmail, "test@example.com")
	ctx = context.WithValue(ctx, security.ContextIssuer, "https://issuer.example.com")
	ctx = context.WithValue(ctx, security.ContextName, "Test User")
	ctx = context.WithValue(ctx, security.ContextSubject, "user-subject-456")
	ctx = context.WithValue(ctx, security.ContextVerified, true)
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
