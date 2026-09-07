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
var serviceWorkerTestAssets embed.FS

// ============================================================================
// HttpViewServiceWorker Tests
// ============================================================================

func serviceWorkerBody(t *testing.T) (*httptest.ResponseRecorder, string) {
	t.Helper()

	e := templating.NewEngine(serviceWorkerTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewServiceWorker(e)
	req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	return rec, string(body)
}

func Test_HttpViewServiceWorker_With_Request_Should_Return_200(t *testing.T) {
	// Arrange, Act
	rec, _ := serviceWorkerBody(t)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewServiceWorker_With_Request_Should_Return_JavaScript_Content_Type(t *testing.T) {
	// Arrange, Act
	rec, _ := serviceWorkerBody(t)

	// Assert
	assert.That(t, "content type must be application/javascript", rec.Header().Get("Content-Type"), "application/javascript")
}

func Test_HttpViewServiceWorker_With_Request_Should_Have_No_Cache_Header(t *testing.T) {
	// A cached tombstone would keep the old worker alive on exactly the
	// browsers this route exists to reach.

	// Arrange, Act
	rec, _ := serviceWorkerBody(t)

	// Assert
	assert.That(t, "cache control must prevent caching", rec.Header().Get("Cache-Control"), "no-cache, no-store, must-revalidate")
}

func Test_HttpViewServiceWorker_With_Request_Should_Unregister_Itself(t *testing.T) {
	// Arrange, Act
	_, body := serviceWorkerBody(t)

	// Assert
	assert.That(t, "body must unregister the worker", containsString(body, "self.registration.unregister()"), true)
	assert.That(t, "body must drop every cache", containsString(body, "caches.delete(name)"), true)
}

func Test_HttpViewServiceWorker_With_Request_Should_Not_Handle_Fetch(t *testing.T) {
	// The whole point of the tombstone: no fetch handler, so requests reach
	// the network while the old worker is being torn down. A fetch handler
	// creeping back in is the app caching in the client again.

	// Arrange, Act
	_, body := serviceWorkerBody(t)

	// Assert
	assert.That(t, "body must not handle fetch events", containsString(body, "addEventListener('fetch'"), false)
}
