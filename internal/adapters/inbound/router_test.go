package inbound_test

import (
	"context"
	"embed"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
)

// ============================================================================
// Test Assets
// ============================================================================
// We embed testdata/assets and use fs.Sub to remap paths so that
// "assets/templates/*.tmpl" resolves correctly for Route().

//go:embed testdata/assets
var routerTestAssetsRaw embed.FS

func getRouterTestFS(t *testing.T) fs.FS {
	t.Helper()
	// Remap testdata/assets -> assets (root level)
	sub, err := fs.Sub(routerTestAssetsRaw, "testdata")
	if err != nil {
		t.Fatalf("failed to create sub filesystem: %v", err)
	}
	return sub
}

// ============================================================================
// Route Tests
// ============================================================================

func Test_Route_Should_Return_Non_Nil_Mux(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()

	// Act
	mux := inbound.Route(ctx, getRouterTestFS(t), logger)

	// Assert
	assert.That(t, "mux must not be nil", mux != nil, true)
}

func Test_Route_Liveness_Endpoint_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	mux := inbound.Route(ctx, getRouterTestFS(t), logger)

	req := httptest.NewRequest(http.MethodGet, "/liveness", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_Route_Readiness_Endpoint_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	mux := inbound.Route(ctx, getRouterTestFS(t), logger)

	req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_Route_UI_Endpoint_Without_Session_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	mux := inbound.Route(ctx, getRouterTestFS(t), logger)

	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
	location := rec.Header().Get("Location")
	assert.That(t, "location must contain login", containsString(location, "/ui/login"), true)
}

func Test_Route_Login_Endpoint_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	mux := inbound.Route(ctx, getRouterTestFS(t), logger)

	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_Route_Login_Endpoint_Should_Return_HTML_Content(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	mux := inbound.Route(ctx, getRouterTestFS(t), logger)

	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain app name", containsString(bodyStr, "TestApp"), true)
}

func Test_Route_Session_Endpoint_Without_Valid_Session_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	mux := inbound.Route(ctx, getRouterTestFS(t), logger)

	// Note: Session endpoints with invalid/unknown session IDs return 200 with index page
	// The WithAuth middleware handles session validation
	req := httptest.NewRequest(http.MethodGet, "/ui/test-session-123/", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_Route_Session_Endpoint_Without_Trailing_Slash_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	mux := inbound.Route(ctx, getRouterTestFS(t), logger)

	// Note: Session endpoints with invalid/unknown session IDs return 200 with index page
	// The WithAuth middleware handles session validation
	req := httptest.NewRequest(http.MethodGet, "/ui/test-session-123", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_Route_Unknown_Endpoint_Should_Return_404(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := slog.Default()
	mux := inbound.Route(ctx, getRouterTestFS(t), logger)

	req := httptest.NewRequest(http.MethodGet, "/unknown/path", nil)
	rec := httptest.NewRecorder()

	// Act
	mux.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 404", rec.Code, http.StatusNotFound)
}
