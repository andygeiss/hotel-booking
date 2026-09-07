package inbound_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/hotel-booking/internal/adapters/inbound"
)

func healthyPing(ctx context.Context) error { return nil }

func brokenPing(ctx context.Context) error { return errors.New("database is down") }

func Test_OpsHandler_Healthz_With_Healthy_Databases_Should_Return_200(t *testing.T) {
	// Arrange
	handler := inbound.OpsHandler("v1.2.3", healthyPing)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
	assert.That(t, "content type must be json", rec.Header().Get("Content-Type"), "application/json")
	assert.That(t, "body must report ok", containsString(string(body), `"status":"ok"`), true)
	assert.That(t, "body must name the build", containsString(string(body), `"version":"v1.2.3"`), true)
}

func Test_OpsHandler_Healthz_With_Broken_Databases_Should_Return_503(t *testing.T) {
	// Arrange
	handler := inbound.OpsHandler("v1.2.3", brokenPing)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	assert.That(t, "status code must be 503", rec.Code, http.StatusServiceUnavailable)
	assert.That(t, "body must report unavailable", containsString(string(body), `"status":"unavailable"`), true)
}

func Test_OpsHandler_Pprof_Index_Should_Return_200(t *testing.T) {
	// Arrange
	handler := inbound.OpsHandler("v1.2.3", healthyPing)
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

// The ops endpoints are guarded by nothing but the listener they sit on, so
// the application mux must not answer for them. This test is what stops a
// later edit from moving them onto the public listener.
func Test_Route_Ops_Endpoints_Should_Not_Be_Served_By_The_App_Mux(t *testing.T) {
	// Arrange
	mux := inbound.Route(inbound.RouterConfig{
		App:                testApp(),
		Ctx:                context.Background(),
		EFS:                getRouterTestFS(t),
		Logger:             slog.Default(),
		ReservationService: createTestReservationService(t),
	})

	for _, path := range []string{"/healthz", "/debug/pprof/", "/debug/pprof/profile"} {
		// Act
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		// Assert
		assert.That(t, "status code for "+path+" must be 404", rec.Code, http.StatusNotFound)
	}
}
