package response

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func JSON(ctx context.Context, w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func Error(ctx context.Context, w http.ResponseWriter, statusCode int, message string, details ...string) {
	log := logrus.WithContext(ctx)

	response := ErrorResponse{
		Error: http.StatusText(statusCode),
	}

	if message != "" {
		response.Message = message
	}

	if statusCode >= http.StatusInternalServerError {
		log.WithFields(logrus.Fields{
			"status_code": statusCode,
			"message":     message,
			"details":     details,
		}).Error("Server error occurred")
	} else {
		log.WithFields(logrus.Fields{
			"status_code": statusCode,
			"message":     message,
		}).Warn("Client error occurred")
	}

	JSON(ctx, w, statusCode, response)
}

func BadRequest(ctx context.Context, w http.ResponseWriter, message string) {
	Error(ctx, w, http.StatusBadRequest, message)
}
