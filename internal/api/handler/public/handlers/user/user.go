package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/TiPSYDiPSY/home-task/internal/api/handler/public/handlers/middleware"
	"github.com/TiPSYDiPSY/home-task/internal/model/api"

	"github.com/go-chi/chi/v5"

	"github.com/TiPSYDiPSY/home-task/internal/util/response"
	"github.com/TiPSYDiPSY/home-task/internal/util/validation"

	"github.com/TiPSYDiPSY/home-task/internal/service"
)

const (
	DecimalBase = 10
	BitSize     = 64
)

func UpdateBalance(userService service.UserService, validation *validation.Validator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logrus.WithContext(ctx)

		userID, err := parseUserID(r)
		if err != nil {
			response.BadRequest(ctx, w, err.Error())

			return
		}

		sourceType := middleware.GetSourceType(ctx)

		var request api.TransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.WithError(err).Error("Failed to decode request body")
			response.BadRequest(ctx, w, "invalid JSON format")

			return
		}

		if err := validation.ValidateStruct(&request); err != nil {
			log.WithError(err).Error("Request validation failed")
			response.BadRequest(ctx, w, err.Error())

			return
		}

		if err := userService.UpdateBalance(ctx, request, userID, sourceType); err != nil {
			log.WithError(err).Error("Failed to update user balance")

			switch {
			case errors.Is(err, service.ErrUserNotFound):
				response.Error(ctx, w, http.StatusNotFound, "user not found")
			case strings.Contains(err.Error(), "transaction already exists"):
				response.Error(ctx, w, http.StatusConflict, "transaction with this ID already exists")
			case strings.Contains(err.Error(), "insufficient funds"):
				response.Error(ctx, w, http.StatusBadRequest, "insufficient funds for this transaction")
			case strings.Contains(err.Error(), "invalid amount format"):
				response.Error(ctx, w, http.StatusBadRequest, "invalid amount format")
			default:
				response.Error(ctx, w, http.StatusInternalServerError, "failed to process transaction")
			}

			return
		}

		response.OK(ctx, w, map[string]string{
			"status":  "success",
			"message": "Transaction processed successfully",
		})
	}
}

func GetBalance(userService service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, err := parseUserID(r)
		if err != nil {
			response.BadRequest(ctx, w, err.Error())

			return
		}

		balanceResponse, err := userService.GetBalance(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				response.Error(ctx, w, http.StatusNotFound, "user not found")
			default:
				response.Error(ctx, w, http.StatusInternalServerError, "internal server Error")
			}

			return
		}

		response.OK(ctx, w, balanceResponse)
	}
}

func parseUserID(r *http.Request) (uint64, error) {
	userIDStr := chi.URLParam(r, "userID")
	if userIDStr == "" {
		return 0, errors.New("user ID is required")
	}

	userID, err := strconv.ParseUint(userIDStr, DecimalBase, BitSize)
	if err != nil {
		return 0, errors.New("invalid user ID format")
	}

	if userID == 0 {
		return 0, errors.New("user ID must be positive")
	}

	return userID, nil
}
