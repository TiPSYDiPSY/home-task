package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/TiPSYDiPSY/home-task/internal/util/response"
)

type ContextKey string

const (
	SourceTypeKey ContextKey = "source_type"
)

func getValidSourceTypes() map[string]bool {
	return map[string]bool{
		"game":    true,
		"server":  true,
		"payment": true,
	}
}

func SourceTypeValidator(next http.Handler) http.Handler {
	validSourceTypes := getValidSourceTypes()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sourceType := strings.TrimSpace(r.Header.Get("Source-Type"))
		if sourceType == "" {
			response.BadRequest(ctx, w, "Source-Type header is required")

			return
		}

		sourceType = strings.ToLower(sourceType)

		if !validSourceTypes[sourceType] {
			response.BadRequest(ctx, w, "Source-Type must be one of: game, server, payment")

			return
		}

		ctx = context.WithValue(ctx, SourceTypeKey, sourceType)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetSourceType(ctx context.Context) string {
	if sourceType, ok := ctx.Value(SourceTypeKey).(string); ok {
		return sourceType
	}

	return ""
}
