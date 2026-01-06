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
var viewTestAssets embed.FS

// ============================================================================
// HttpView Tests
// ============================================================================

func Test_HttpView_With_Valid_Template_Should_Return_200(t *testing.T) {
	// Arrange
	e := templating.NewEngine(viewTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	data := struct {
		AppName string
		Title   string
	}{
		AppName: "TestApp",
		Title:   "Test Title",
	}

	handler := inbound.HttpView(e, "login", data)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpView_With_Data_Should_Render_Data_In_Template(t *testing.T) {
	// Arrange
	e := templating.NewEngine(viewTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	data := struct {
		AppName string
		Title   string
	}{
		AppName: "MyCustomApp",
		Title:   "Custom Title",
	}

	handler := inbound.HttpView(e, "login", data)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain custom app name", containsString(bodyStr, "MyCustomApp"), true)
}
