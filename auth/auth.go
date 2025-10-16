package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// ExtractBearerToken extracts the bearer token from an Authorization header
func ExtractBearerToken(header string) (string, bool) {
	const bearerPrefix = "Bearer "
	if header == "" {
		return "", false
	}
	if len(header) < len(bearerPrefix) || !strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(bearerPrefix):])
	if token == "" {
		return "", false
	}
	return token, true
}

// tokensMatch compares two tokens using constant-time comparison
func tokensMatch(expected, candidate string) bool {
	if expected == "" {
		return true
	}
	if candidate == "" || len(candidate) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(candidate)) == 1
}

// AuthorizeHTTPBearer authorizes HTTP requests with Bearer token authentication
func AuthorizeHTTPBearer(expected string, r *http.Request) bool {
	if expected == "" {
		return true
	}
	candidate, ok := ExtractBearerToken(r.Header.Get("Authorization"))
	if !ok {
		return false
	}
	return tokensMatch(expected, candidate)
}

// AuthorizeSSEToken authorizes SSE requests with token authentication
func AuthorizeSSEToken(expected string, r *http.Request) bool {
	if expected == "" {
		return true
	}
	if tokensMatch(expected, r.URL.Query().Get("token")) {
		return true
	}
	if candidate, ok := ExtractBearerToken(r.Header.Get("Authorization")); ok {
		return tokensMatch(expected, candidate)
	}
	return false
}

// WriteUnauthorized writes an HTTP 401 Unauthorized response
func WriteUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer realm=\"gojq-mcp\"")
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
