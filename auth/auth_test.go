package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		expected  string
		expectOK  bool
	}{
		{
			name:     "valid bearer token",
			header:   "Bearer abc123",
			expected: "abc123",
			expectOK: true,
		},
		{
			name:     "bearer token with extra spaces",
			header:   "Bearer   token123   ",
			expected: "token123",
			expectOK: true,
		},
		{
			name:     "empty header",
			header:   "",
			expected: "",
			expectOK: false,
		},
		{
			name:     "missing bearer prefix",
			header:   "abc123",
			expected: "",
			expectOK: false,
		},
		{
			name:     "wrong case bearer",
			header:   "bearer abc123",
			expected: "abc123",
			expectOK: true,
		},
		{
			name:     "only bearer prefix",
			header:   "Bearer ",
			expected: "",
			expectOK: false,
		},
		{
			name:     "bearer with extra text",
			header:   "Bearer token extra",
			expected: "token extra",
			expectOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, ok := extractBearerToken(tt.header)
			assert.Equal(t, tt.expectOK, ok)
			assert.Equal(t, tt.expected, token)
		})
	}
}

func TestTokensMatch(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		candidate string
		expectedMatch bool
	}{
		{
			name:     "matching tokens",
			expected: "secret123",
			candidate: "secret123",
			expectedMatch: true,
		},
		{
			name:     "non-matching tokens",
			expected: "secret123",
			candidate: "wrong123",
			expectedMatch: false,
		},
		{
			name:     "empty expected token",
			expected: "",
			candidate: "any",
			expectedMatch: true,
		},
		{
			name:     "empty candidate token",
			expected: "secret",
			candidate: "",
			expectedMatch: false,
		},
		{
			name:     "different lengths",
			expected: "short",
			candidate: "verylongtoken",
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokensMatch(tt.expected, tt.candidate)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}

func TestAuthorizeHTTPBearer(t *testing.T) {
	tests := []struct {
		name       string
		expected   string
		authHeader string
		expectedAuth bool
	}{
		{
			name:       "valid authorization",
			expected:   "secret123",
			authHeader: "Bearer secret123",
			expectedAuth: true,
		},
		{
			name:       "invalid token",
			expected:   "secret123",
			authHeader: "Bearer wrong123",
			expectedAuth: false,
		},
		{
			name:       "no auth header",
			expected:   "secret123",
			authHeader: "",
			expectedAuth: false,
		},
		{
			name:       "no expected token",
			expected:   "",
			authHeader: "Bearer anytoken",
			expectedAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			result := AuthorizeHTTPBearer(tt.expected, req)
			assert.Equal(t, tt.expectedAuth, result)
		})
	}
}

func TestAuthorizeSSEToken(t *testing.T) {
	tests := []struct {
		name       string
		expected   string
		queryToken string
		authHeader string
		expectedAuth bool
	}{
		{
			name:       "valid query token",
			expected:   "secret123",
			queryToken: "secret123",
			authHeader: "",
			expectedAuth: true,
		},
		{
			name:       "valid bearer token",
			expected:   "secret123",
			queryToken: "",
			authHeader: "Bearer secret123",
			expectedAuth: true,
		},
		{
			name:       "invalid tokens",
			expected:   "secret123",
			queryToken: "wrong",
			authHeader: "Bearer also_wrong",
			expectedAuth: false,
		},
		{
			name:       "no expected token",
			expected:   "",
			queryToken: "any",
			authHeader: "",
			expectedAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.queryToken != "" {
				q := req.URL.Query()
				q.Set("token", tt.queryToken)
				req.URL.RawQuery = q.Encode()
			}
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			result := AuthorizeSSEToken(tt.expected, req)
			assert.Equal(t, tt.expectedAuth, result)
		})
	}
}

func TestWriteUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	WriteUnauthorized(w)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "Unauthorized\n", w.Body.String())
	assert.Equal(t, "Bearer realm=\"gojq-mcp\"", w.Header().Get("WWW-Authenticate"))
}
