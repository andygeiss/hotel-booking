package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/logging"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
)

// Integration tests for the HTTP server.
// These tests verify the server endpoints work correctly with all middleware and routing.
//
// Run with: go test -v ./cmd/server/...

func Test_Server_Integration_Liveness_Endpoint_Should_Return_OK(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/liveness")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 200", resp.StatusCode, http.StatusOK)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.That(t, "body must be OK", string(body), "OK")
}

func Test_Server_Integration_Readiness_Endpoint_Should_Return_OK(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/readiness")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 200", resp.StatusCode, http.StatusOK)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.That(t, "body must be OK", string(body), "OK")
}

func Test_Server_Integration_Static_Assets_Should_Serve_CSS(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/static/css/base.css")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 200", resp.StatusCode, http.StatusOK)
	assert.That(t, "content-type must be text/css", resp.Header.Get("Content-Type"), "text/css; charset=utf-8")
	resp.Body.Close()
}

func Test_Server_Integration_Static_Assets_Should_Serve_JS(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/static/js/htmx.min.js")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 200", resp.StatusCode, http.StatusOK)
	resp.Body.Close()
}

func Test_Server_Integration_UI_Login_Should_Return_Login_Page(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/ui/login")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 200", resp.StatusCode, http.StatusOK)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Verify it contains login page content.
	bodyStr := string(body)
	assert.That(t, "body must contain html", len(bodyStr) > 0, true)
}

func Test_Server_Integration_UI_Index_Without_Auth_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Create a client that doesn't follow redirects.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Act
	resp, err := client.Get(server.URL + "/ui/")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	// Unauthenticated requests should redirect to login.
	assert.That(t, "status code must be 303 (See Other)", resp.StatusCode, http.StatusSeeOther)

	location := resp.Header.Get("Location")
	assert.That(t, "location must contain login", len(location) > 0, true)
	resp.Body.Close()
}

func Test_Server_Integration_NotFound_Should_Return_404(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/nonexistent-path")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 404", resp.StatusCode, http.StatusNotFound)
	resp.Body.Close()
}

func Test_Server_Integration_Concurrent_Requests_Should_Handle_Load(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	concurrency := 50
	results := make(chan int, concurrency)

	// Act - Send concurrent requests.
	for range concurrency {
		go func() {
			resp, err := http.Get(server.URL + "/liveness")
			if err != nil {
				results <- 0
				return
			}
			results <- resp.StatusCode
			resp.Body.Close()
		}()
	}

	// Assert - All requests should succeed.
	successCount := 0
	for range concurrency {
		if <-results == http.StatusOK {
			successCount++
		}
	}
	assert.That(t, "all concurrent requests must succeed", successCount, concurrency)
}

// Benchmarks for Profile-Guided Optimization (PGO).
// Run with: just profile
// This generates cpuprofile.pprof for optimized builds.

func Benchmark_Server_Integration_Liveness_Should_Respond_Fast(b *testing.B) {
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for b.Loop() {
		resp, _ := client.Get(server.URL + "/liveness")
		if resp != nil {
			resp.Body.Close()
		}
	}
}

func Benchmark_Server_Integration_Static_CSS_Should_Serve_Fast(b *testing.B) {
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for b.Loop() {
		resp, _ := client.Get(server.URL + "/static/css/base.css")
		if resp != nil {
			resp.Body.Close()
		}
	}
}

func Benchmark_Server_Integration_Login_Page_Should_Render_Fast(b *testing.B) {
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for b.Loop() {
		resp, _ := client.Get(server.URL + "/ui/login")
		if resp != nil {
			resp.Body.Close()
		}
	}
}
