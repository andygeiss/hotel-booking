package inbound

// This file is the one internal test in the package: it reads the csp constant
// directly, so that a policy edit and its test cannot drift apart.

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/coreos/go-oidc/v3/oidc"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func Test_SecureHeaders_Should_Send_Every_Header(t *testing.T) {
	// Arrange
	want := map[string]string{
		"Content-Security-Policy":   csp,
		"Strict-Transport-Security": "max-age=31536000",
		"X-Content-Type-Options":    "nosniff",
		"Referrer-Policy":           "same-origin",
	}
	handler := secureHeaders(okHandler())
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	// Assert
	for name, value := range want {
		assert.That(t, "header "+name+" must match", rec.Header().Get(name), value)
	}
}

func Test_CSP_Should_Allow_Data_Urls_For_Mask_Icons(t *testing.T) {
	// Unlike the test above this one is not a tautology. It fails if somebody
	// tightens the policy and takes every mask icon down with it.
	assert.That(t, "csp must allow data: images", strings.Contains(csp, "img-src 'self' data:"), true)
}

func Test_CSP_Should_Not_Allow_Unsafe_Sources(t *testing.T) {
	assert.That(t, "csp must not allow inline code", strings.Contains(csp, "'unsafe-inline'"), false)
	assert.That(t, "csp must not allow eval", strings.Contains(csp, "'unsafe-eval'"), false)
}

func Test_WithSecurity_Cross_Origin_Post_Should_Return_403(t *testing.T) {
	// Arrange
	handler := WithSecurity(okHandler())
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 403", rec.Code, http.StatusForbidden)
}

func Test_WithSecurity_Same_Origin_Post_Should_Return_200(t *testing.T) {
	// Arrange
	handler := WithSecurity(okHandler())
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations", nil)
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_WithSecurity_Post_Without_Browser_Headers_Should_Return_200(t *testing.T) {
	// An MCP client sends neither Sec-Fetch-Site nor Origin, because it is not
	// a browser. The CSRF check must let it through, or /mcp stops answering.

	// Arrange
	handler := WithSecurity(okHandler())
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_WithSecurity_Oversized_Body_Should_Fail_To_Read(t *testing.T) {
	// Arrange
	var readErr error
	handler := WithSecurity(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
	}))
	body := strings.NewReader(strings.Repeat("a", maxRequestBody+1))
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations", body)
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "reading past the cap must fail", readErr != nil, true)
}

func Test_WithSecurity_Body_Under_The_Cap_Should_Be_Readable(t *testing.T) {
	// Arrange
	var readErr error
	handler := WithSecurity(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
	}))
	body := strings.NewReader(strings.Repeat("a", 1024))
	req := httptest.NewRequest(http.MethodPost, "/ui/reservations", body)
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.That(t, "error must be nil", readErr, nil)
}

// ============================================================================
// WithBearerAuth Tests
// ============================================================================

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// workingVerifier is built without touching the network. Every test below
// stops before a signature is checked, so the key set is never used.
func workingVerifier() *oidc.IDTokenVerifier {
	return oidc.NewVerifier("https://issuer.example", nil, &oidc.Config{ClientID: "hotel-booking-mcp"})
}

func Test_WithBearerAuth_Should_Not_Build_The_Verifier_Until_A_Request_Arrives(t *testing.T) {
	// Building it calls the identity provider. If that happened here, it would
	// be happening at startup, which is the whole bug this replaced.

	// Arrange
	calls := 0

	// Act
	_ = WithBearerAuth(func() (*oidc.IDTokenVerifier, error) {
		calls++
		return workingVerifier(), nil
	}, discardLogger(), okHandlerFunc())

	// Assert
	assert.That(t, "the provider must not be called while wiring", calls, 0)
}

func Test_WithBearerAuth_When_The_Provider_Is_Down_Should_Return_503(t *testing.T) {
	// Arrange
	reached := false
	handler := WithBearerAuth(
		func() (*oidc.IDTokenVerifier, error) { return nil, errors.New("connection refused") },
		discardLogger(),
		func(w http.ResponseWriter, r *http.Request) { reached = true },
	)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, httptest.NewRequest(http.MethodPost, "/mcp", nil))

	// Assert
	assert.That(t, "status code must be 503", rec.Code, http.StatusServiceUnavailable)
	assert.That(t, "the handler behind it must not run", reached, false)
}

func Test_WithBearerAuth_Should_Retry_After_The_Provider_Recovers(t *testing.T) {
	// Caching the failure would leave /mcp broken until the next deploy, long
	// after the identity provider came back.

	// Arrange
	calls := 0
	handler := WithBearerAuth(func() (*oidc.IDTokenVerifier, error) {
		calls++
		if calls == 1 {
			return nil, errors.New("connection refused")
		}
		return workingVerifier(), nil
	}, discardLogger(), okHandlerFunc())

	// Act
	first := httptest.NewRecorder()
	handler(first, httptest.NewRequest(http.MethodPost, "/mcp", nil))
	second := httptest.NewRecorder()
	handler(second, httptest.NewRequest(http.MethodPost, "/mcp", nil))

	// Assert
	assert.That(t, "the first request must fail while the provider is down", first.Code, http.StatusServiceUnavailable)
	assert.That(t, "the second must get past the verifier", second.Code, http.StatusUnauthorized)
	assert.That(t, "the provider must have been asked twice", calls, 2)
}

func Test_WithBearerAuth_Should_Build_The_Verifier_Only_Once(t *testing.T) {
	// Arrange
	calls := 0
	handler := WithBearerAuth(func() (*oidc.IDTokenVerifier, error) {
		calls++
		return workingVerifier(), nil
	}, discardLogger(), okHandlerFunc())

	// Act
	for range 3 {
		handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/mcp", nil))
	}

	// Assert
	assert.That(t, "success must be cached", calls, 1)
}

func Test_WithBearerAuth_Without_A_Token_Should_Return_401(t *testing.T) {
	// A missing token is the client's problem (401), a missing provider is
	// ours (503). The two must not look the same.

	// Arrange
	reached := false
	handler := WithBearerAuth(
		func() (*oidc.IDTokenVerifier, error) { return workingVerifier(), nil },
		discardLogger(),
		func(w http.ResponseWriter, r *http.Request) { reached = true },
	)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, httptest.NewRequest(http.MethodPost, "/mcp", nil))

	// Assert
	assert.That(t, "status code must be 401", rec.Code, http.StatusUnauthorized)
	assert.That(t, "the MCP handler must not run", reached, false)
}

func okHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
}
