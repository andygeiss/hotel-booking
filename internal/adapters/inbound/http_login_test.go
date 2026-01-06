package inbound_test

import (
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
)

// ============================================================================
// Test Assets
// ============================================================================

//go:embed testdata/assets/templates/*.tmpl testdata/assets/static/css/*.css
var loginTestAssets embed.FS

// ============================================================================
// HttpViewLogin Tests
// ============================================================================

func Test_HttpViewLogin_With_Request_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(loginTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewLogin(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewLogin_With_Request_Should_Render_Template(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(loginTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewLogin(e)
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
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(loginTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewLogin(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	contentType := rec.Header().Get("Content-Type")
	assert.That(t, "content type must be text/html", containsString(contentType, "text/html"), true)
}
