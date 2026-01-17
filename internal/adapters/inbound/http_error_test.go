package inbound_test

import (
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/hotel-booking/internal/adapters/inbound"
)

// ============================================================================
// Test Assets
// ============================================================================

//go:embed testdata/assets/templates/*.tmpl testdata/assets/static/css/*.css
var errorTestAssets embed.FS

// ============================================================================
// HttpViewError Tests
// ============================================================================

func Test_HttpViewError_With_Request_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(errorTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewError(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/error", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewError_With_Request_Should_Render_Default_Error_Title(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(errorTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewError(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/error", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain default error title", containsString(bodyStr, "An Error Occurred"), true)
}

func Test_HttpViewError_With_Request_Should_Render_Default_Error_Message(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(errorTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewError(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/error", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain default error message", containsString(bodyStr, "Something went wrong"), true)
}

func Test_HttpViewError_With_Custom_Error_Title_Should_Render_Custom_Title(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(errorTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewError(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/error?title=Custom+Error+Title", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain custom error title", containsString(bodyStr, "Custom Error Title"), true)
}

func Test_HttpViewError_With_Custom_Error_Message_Should_Render_Custom_Message(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(errorTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewError(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/error?message=Custom+error+message+here", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain custom error message", containsString(bodyStr, "Custom error message here"), true)
}

func Test_HttpViewError_With_Error_Details_Should_Render_Details(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(errorTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewError(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/error?details=Stack+trace+info", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain error details", containsString(bodyStr, "Stack trace info"), true)
}

func Test_HttpViewError_With_All_Parameters_Should_Render_All(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(errorTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewError(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/error?title=Not+Found&message=Page+not+found&details=Error+404", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain custom title", containsString(bodyStr, "Not Found"), true)
	assert.That(t, "body must contain custom message", containsString(bodyStr, "Page not found"), true)
	assert.That(t, "body must contain custom details", containsString(bodyStr, "Error 404"), true)
}

func Test_HttpViewError_With_Request_Should_Contain_App_Name(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(errorTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewError(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/error", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain app name", containsString(bodyStr, "TestApp"), true)
}

func Test_HttpViewError_With_Request_Should_Return_HTML_Content_Type(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(errorTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewError(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/error", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	contentType := rec.Header().Get("Content-Type")
	assert.That(t, "content type must be text/html", containsString(contentType, "text/html"), true)
}
