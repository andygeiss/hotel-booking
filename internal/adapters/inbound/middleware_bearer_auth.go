package inbound

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/andygeiss/cloud-native-utils/web"
	"github.com/coreos/go-oidc/v3/oidc"
)

// writeJSONRPCError writes a JSON-RPC 2.0 error response with 401 status.
func writeJSONRPCError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"jsonrpc":"2.0","error":{"code":-32001,"message":"%s"},"id":null}`, message)
}

// WithBearerAuth validates OAuth 2.1 Bearer tokens for MCP endpoints.
// It extracts the token from the Authorization header, verifies it against
// the OIDC provider (Keycloak), and populates the request context with user claims.
func WithBearerAuth(verifier *oidc.IDTokenVerifier, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract Bearer token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSONRPCError(w, "Missing Authorization header")
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeJSONRPCError(w, "Invalid Authorization scheme, expected Bearer")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Check for nil verifier (configuration error)
		if verifier == nil {
			writeJSONRPCError(w, "Invalid or expired token")
			return
		}

		// Verify token with OIDC provider (Keycloak)
		idToken, err := verifier.Verify(r.Context(), tokenString)
		if err != nil {
			writeJSONRPCError(w, "Invalid or expired token")
			return
		}

		// Extract claims from token
		var claims struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
			Name          string `json:"name"`
		}
		if err := idToken.Claims(&claims); err != nil {
			writeJSONRPCError(w, "Failed to parse token claims")
			return
		}

		// Populate context with authenticated user info
		ctx := r.Context()
		ctx = context.WithValue(ctx, web.ContextEmail, claims.Email)
		ctx = context.WithValue(ctx, web.ContextName, claims.Name)
		ctx = context.WithValue(ctx, web.ContextSubject, idToken.Subject)
		ctx = context.WithValue(ctx, web.ContextIssuer, idToken.Issuer)
		ctx = context.WithValue(ctx, web.ContextVerified, claims.EmailVerified)

		next(w, r.WithContext(ctx))
	}
}
