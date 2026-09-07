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

//go:embed testdata/assets/templates/*.tmpl testdata/assets/static/css/*.css
var manifestTestAssets embed.FS

func Test_HttpViewManifest_With_Request_Should_Return_200(t *testing.T) {
	// Arrange
	e := templating.NewEngine(manifestTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewManifest(e, testApp())
	req := httptest.NewRequest(http.MethodGet, "/manifest.json", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewManifest_With_Request_Should_Render_Template(t *testing.T) {
	// Arrange
	e := templating.NewEngine(manifestTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewManifest(e, testApp())
	req := httptest.NewRequest(http.MethodGet, "/manifest.json", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain app name", containsString(bodyStr, "TestApp"), true)
	assert.That(t, "body must contain description", containsString(bodyStr, "Test Description"), true)
}

func Test_HttpViewManifest_With_Request_Should_Return_Manifest_Content_Type(t *testing.T) {
	// Arrange
	e := templating.NewEngine(manifestTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewManifest(e, testApp())
	req := httptest.NewRequest(http.MethodGet, "/manifest.json", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	contentType := rec.Header().Get("Content-Type")
	assert.That(t, "content type must be application/manifest+json", contentType, "application/manifest+json")
}
