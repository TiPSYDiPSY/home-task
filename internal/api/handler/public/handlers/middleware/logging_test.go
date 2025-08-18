package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

func TestLoggingMiddleware_Middleware(t *testing.T) {
	var logBuffer bytes.Buffer
	logrus.SetOutput(&logBuffer)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer func() {
		logrus.SetOutput(io.Discard)
	}()

	tests := []struct {
		name               string
		config             LoggingConfig
		method             string
		path               string
		body               string
		expectedStatusCode int
		shouldLogBody      bool
	}{
		{
			name: "GET request without body logging",
			config: LoggingConfig{
				BodyLoggingEnabled: false,
				ServiceName:        "test-service",
			},
			method:             http.MethodGet,
			path:               "/test",
			body:               "",
			expectedStatusCode: http.StatusOK,
			shouldLogBody:      false,
		},
		{
			name: "POST request with body logging enabled",
			config: LoggingConfig{
				BodyLoggingEnabled: true,
				ServiceName:        "test-service",
			},
			method:             http.MethodPost,
			path:               "/users",
			body:               `{"name":"test"}`,
			expectedStatusCode: http.StatusCreated,
			shouldLogBody:      true,
		},
		{
			name: "POST request with body logging disabled",
			config: LoggingConfig{
				BodyLoggingEnabled: false,
				ServiceName:        "test-service",
			},
			method:             http.MethodPost,
			path:               "/users",
			body:               `{"name":"test"}`,
			expectedStatusCode: http.StatusCreated,
			shouldLogBody:      false,
		},
		{
			name: "PUT request with body logging enabled",
			config: LoggingConfig{
				BodyLoggingEnabled: true,
				ServiceName:        "test-service",
			},
			method:             http.MethodPut,
			path:               "/users/1",
			body:               `{"name":"updated"}`,
			expectedStatusCode: http.StatusOK,
			shouldLogBody:      true,
		},
		{
			name: "PATCH request with body logging enabled",
			config: LoggingConfig{
				BodyLoggingEnabled: true,
				ServiceName:        "test-service",
			},
			method:             http.MethodPatch,
			path:               "/users/1",
			body:               `{"name":"patched"}`,
			expectedStatusCode: http.StatusOK,
			shouldLogBody:      true,
		},
		{
			name: "DELETE request - no body logging even when enabled",
			config: LoggingConfig{
				BodyLoggingEnabled: true,
				ServiceName:        "test-service",
			},
			method:             http.MethodDelete,
			path:               "/users/1",
			body:               "",
			expectedStatusCode: http.StatusNoContent,
			shouldLogBody:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuffer.Reset()

			middleware := NewLoggingMiddleware(tt.config)

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.body != "" {
					bodyBytes, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Equal(t, tt.body, string(bodyBytes))
				}

				w.WriteHeader(tt.expectedStatusCode)
				w.Write([]byte("test response"))
			})

			wrappedHandler := middleware.Middleware(testHandler)

			var reqBody io.Reader
			if tt.body != "" {
				reqBody = strings.NewReader(tt.body)
			}
			req := httptest.NewRequest(tt.method, tt.path, reqBody)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatusCode, recorder.Code)
			assert.Equal(t, "test response", recorder.Body.String())

			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, "HTTP Request Started")
			assert.Contains(t, logOutput, "HTTP Request Completed")
			assert.Contains(t, logOutput, tt.method)
			assert.Contains(t, logOutput, tt.path)

			if tt.shouldLogBody && tt.body != "" {
				escapedBody := strings.ReplaceAll(tt.body, `"`, `\"`)
				assert.Contains(t, logOutput, escapedBody)
			} else if tt.body != "" {
				assert.NotContains(t, logOutput, tt.body)
			}

			assert.Contains(t, logOutput, "trace_id")
			assert.Contains(t, logOutput, "span_id")

			assert.Contains(t, logOutput, "duration_ms")
			assert.Contains(t, logOutput, "status_code")
		})
	}
}

func TestLoggingMiddleware_WithTracing(t *testing.T) {
	var logBuffer bytes.Buffer
	logrus.SetOutput(&logBuffer)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer func() {
		logrus.SetOutput(io.Discard)
	}()

	config := LoggingConfig{
		BodyLoggingEnabled: false,
		ServiceName:        "test-service",
	}
	middleware := NewLoggingMiddleware(config)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware.Middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "trace_id")
	assert.Contains(t, logOutput, "span_id")
}

func TestLoggingMiddleware_WithInvalidSpan(t *testing.T) {
	var logBuffer bytes.Buffer
	logrus.SetOutput(&logBuffer)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer func() {
		logrus.SetOutput(io.Discard)
	}()

	config := LoggingConfig{
		BodyLoggingEnabled: false,
		ServiceName:        "test-service",
	}

	middleware := &LoggingMiddleware{
		Config: config,
		Tracer: otel.Tracer("test"),
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware.Middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestLoggingMiddleware_ErrorReadingBody(t *testing.T) {
	var logBuffer bytes.Buffer
	logrus.SetOutput(&logBuffer)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer func() {
		logrus.SetOutput(io.Discard)
	}()

	config := LoggingConfig{
		BodyLoggingEnabled: true,
		ServiceName:        "test-service",
	}
	middleware := NewLoggingMiddleware(config)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware.Middleware(testHandler)

	req := httptest.NewRequest(http.MethodPost, "/test", &errorReader{})
	recorder := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "HTTP Request Started")
	assert.Contains(t, logOutput, "HTTP Request Completed")
	assert.NotContains(t, logOutput, "request_body")
}

type errorReader struct{}

func (e *errorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestLoggingMiddleware_Integration(t *testing.T) {
	var logBuffer bytes.Buffer
	logrus.SetOutput(&logBuffer)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	defer func() {
		logrus.SetOutput(io.Discard)
	}()

	config := LoggingConfig{
		BodyLoggingEnabled: true,
		ServiceName:        "integration-test",
	}
	middleware := NewLoggingMiddleware(config)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write(bodyBytes)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})

	wrappedHandler := middleware.Middleware(testHandler)

	requestBody := `{"name":"John","email":"john@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, requestBody, recorder.Body.String())
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "HTTP Request Started")
	assert.Contains(t, logOutput, "HTTP Request Completed")
	assert.Contains(t, logOutput, `\"name\":\"John\"`)
	assert.Contains(t, logOutput, `\"email\":\"john@example.com\"`)
	assert.Contains(t, logOutput, "POST")
	assert.Contains(t, logOutput, "/api/users")
	assert.Contains(t, logOutput, "201") // status code
	assert.Contains(t, logOutput, "duration_ms")
}
