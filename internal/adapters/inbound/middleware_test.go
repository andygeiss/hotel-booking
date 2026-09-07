package inbound

// This file is the one internal test in the package: it reads the csp constant
// directly, so that a policy edit and its test cannot drift apart.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
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
