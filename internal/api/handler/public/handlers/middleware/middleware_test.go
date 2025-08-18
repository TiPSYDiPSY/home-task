package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceTypeValidator(t *testing.T) {
	tests := []struct {
		name           string
		sourceType     string
		wantHTTPCode   int
		wantBody       string
		shouldCallNext bool
	}{
		{
			name:           "valid source type - game",
			sourceType:     "game",
			wantHTTPCode:   http.StatusOK,
			wantBody:       "success",
			shouldCallNext: true,
		},
		{
			name:           "valid source type - server",
			sourceType:     "server",
			wantHTTPCode:   http.StatusOK,
			wantBody:       "success",
			shouldCallNext: true,
		},
		{
			name:           "valid source type - payment",
			sourceType:     "payment",
			wantHTTPCode:   http.StatusOK,
			wantBody:       "success",
			shouldCallNext: true,
		},
		{
			name:           "valid source type with uppercase - GAME",
			sourceType:     "GAME",
			wantHTTPCode:   http.StatusOK,
			wantBody:       "success",
			shouldCallNext: true,
		},
		{
			name:           "valid source type with mixed case - Server",
			sourceType:     "Server",
			wantHTTPCode:   http.StatusOK,
			wantBody:       "success",
			shouldCallNext: true,
		},
		{
			name:           "valid source type with spaces - ' game '",
			sourceType:     " game ",
			wantHTTPCode:   http.StatusOK,
			wantBody:       "success",
			shouldCallNext: true,
		},
		{
			name:           "invalid source type - invalid",
			sourceType:     "invalid",
			wantHTTPCode:   http.StatusBadRequest,
			wantBody:       `{"error":"Bad Request","message":"Source-Type must be one of: game, server, payment"}`,
			shouldCallNext: false,
		},
		{
			name:           "invalid source type - empty string",
			sourceType:     "",
			wantHTTPCode:   http.StatusBadRequest,
			wantBody:       `{"error":"Bad Request","message":"Source-Type header is required"}`,
			shouldCallNext: false,
		},
		{
			name:           "invalid source type - spaces only",
			sourceType:     "   ",
			wantHTTPCode:   http.StatusBadRequest,
			wantBody:       `{"error":"Bad Request","message":"Source-Type header is required"}`,
			shouldCallNext: false,
		},
		{
			name:           "invalid source type - random value",
			sourceType:     "random",
			wantHTTPCode:   http.StatusBadRequest,
			wantBody:       `{"error":"Bad Request","message":"Source-Type must be one of: game, server, payment"}`,
			shouldCallNext: false,
		},
		{
			name:           "invalid source type - partial match",
			sourceType:     "gam",
			wantHTTPCode:   http.StatusBadRequest,
			wantBody:       `{"error":"Bad Request","message":"Source-Type must be one of: game, server, payment"}`,
			shouldCallNext: false,
		},
		{
			name:           "invalid source type - numbers",
			sourceType:     "123",
			wantHTTPCode:   http.StatusBadRequest,
			wantBody:       `{"error":"Bad Request","message":"Source-Type must be one of: game, server, payment"}`,
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextHandlerCalled = true

				if tt.shouldCallNext {
					sourceType := GetSourceType(r.Context())
					expectedSourceType := strings.ToLower(strings.TrimSpace(tt.sourceType))
					assert.Equal(t, expectedSourceType, sourceType, "Source type should be set in context")
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			middleware := SourceTypeValidator(nextHandler)

			req := httptest.NewRequest(http.MethodPost, "/test", nil)

			if tt.name != "invalid source type - empty string" {
				req.Header.Set("Source-Type", tt.sourceType)
			}

			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantHTTPCode, rr.Code)
			assert.Equal(t, tt.shouldCallNext, nextHandlerCalled)

			if tt.shouldCallNext {
				assert.Equal(t, tt.wantBody, rr.Body.String())
			} else {
				assert.JSONEq(t, tt.wantBody, rr.Body.String())
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			}
		})
	}
}

func TestSourceTypeValidator_ContextPropagation(t *testing.T) {
	var capturedContext context.Context

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContext = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	middleware := SourceTypeValidator(nextHandler)

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Source-Type", "game")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.NotNil(t, capturedContext)
	sourceType := GetSourceType(capturedContext)
	assert.Equal(t, "game", sourceType)
}

func TestSourceTypeValidator_NoHeader(t *testing.T) {
	nextHandlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := SourceTypeValidator(nextHandler)

	req := httptest.NewRequest(http.MethodPost, "/test", nil)

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.False(t, nextHandlerCalled)
	assert.JSONEq(t, `{"error":"Bad Request","message":"Source-Type header is required"}`, rr.Body.String())
}

func TestSourceTypeValidator_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		sourceType   string
		wantHTTPCode int
		description  string
	}{
		{
			name:         "source type with tabs",
			sourceType:   "\tgame\t",
			wantHTTPCode: http.StatusOK,
			description:  "Should trim tabs and accept valid source type",
		},
		{
			name:         "source type with newlines",
			sourceType:   "\ngame\n",
			wantHTTPCode: http.StatusOK,
			description:  "Should trim newlines and accept valid source type",
		},
		{
			name:         "source type with mixed whitespace",
			sourceType:   " \t\ngame\n\t ",
			wantHTTPCode: http.StatusOK,
			description:  "Should trim all whitespace and accept valid source type",
		},
		{
			name:         "source type case insensitive - PayMeNt",
			sourceType:   "PayMeNt",
			wantHTTPCode: http.StatusOK,
			description:  "Should be case insensitive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := SourceTypeValidator(nextHandler)

			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			req.Header.Set("Source-Type", tt.sourceType)

			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantHTTPCode, rr.Code, tt.description)
		})
	}
}
