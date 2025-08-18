package middleware

import (
	"bytes"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type LoggingConfig struct {
	BodyLoggingEnabled bool
	ServiceName        string
}

type LoggingMiddleware struct {
	Config LoggingConfig
	Tracer trace.Tracer
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func NewLoggingMiddleware(config LoggingConfig) *LoggingMiddleware {
	serviceName := config.ServiceName
	if serviceName == "" {
		serviceName = "home-task"
	}

	return &LoggingMiddleware{
		Config: config,
		Tracer: otel.Tracer(serviceName),
	}
}

// Middleware
// nolint: funlen // 62 lines is acceptable for this middleware instead of 60 lines
func (m *LoggingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.Tracer.Start(r.Context(), "http_request",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.RequestURI),
			),
		)
		defer span.End()

		spanContext := span.SpanContext()

		var traceID, spanID string

		if spanContext.IsValid() {
			traceID = spanContext.TraceID().String()
			spanID = spanContext.SpanID().String()
		} else {
			traceID = "invalid_trace"
			spanID = "invalid_span"
		}

		lrw := newLoggingResponseWriter(w)
		start := time.Now()

		var body string

		if m.Config.BodyLoggingEnabled && slices.Contains([]string{http.MethodPost, http.MethodPut, http.MethodPatch}, r.Method) {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				body = string(bodyBytes)
			}
		}

		initialLogFields := logrus.Fields{
			"http_method": r.Method,
			"request_uri": r.RequestURI,
			"trace_id":    traceID,
			"span_id":     spanID,
		}

		if body != "" {
			initialLogFields["request_body"] = body
		}

		logrus.WithContext(ctx).WithFields(initialLogFields).Info("HTTP Request Started")

		next.ServeHTTP(lrw, r.WithContext(ctx))

		duration := time.Since(start)

		completionLogFields := logrus.Fields{
			"http_method": r.Method,
			"request_uri": r.RequestURI,
			"status_code": lrw.statusCode,
			"duration_ms": duration.Milliseconds(),
			"trace_id":    traceID,
			"span_id":     spanID,
		}

		logrus.WithContext(ctx).WithFields(completionLogFields).Info("HTTP Request Completed")
	})
}
