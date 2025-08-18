package user

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/TiPSYDiPSY/home-task/internal/api/handler/public/handlers/middleware"
	customErrors "github.com/TiPSYDiPSY/home-task/internal/errors"
	"github.com/TiPSYDiPSY/home-task/internal/model/api"

	"github.com/go-chi/chi/v5"

	"github.com/TiPSYDiPSY/home-task/internal/util/response"
	"github.com/TiPSYDiPSY/home-task/internal/util/validation"

	"github.com/TiPSYDiPSY/home-task/internal/service"
)

const (
	DecimalBase        = 10
	BitSize            = 64
	MaxRequestBodySize = 1024
)

func UpdateBalance(userService service.UserService, valid *validation.Validator) http.HandlerFunc {
	logger := logrus.StandardLogger()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, err := parseUserID(r)
		if err != nil {
			response.BadRequest(ctx, w, err.Error())

			return
		}

		sourceType := middleware.GetSourceType(ctx)

		body, err := io.ReadAll(io.LimitReader(r.Body, MaxRequestBodySize))
		if err != nil {
			logger.WithError(err).Error("Failed to read request body")
			response.BadRequest(ctx, w, "invalid request body")

			return
		}

		var request api.TransactionRequest
		if err := json.Unmarshal(body, &request); err != nil {
			logger.WithError(err).Error("Failed to decode request body")
			response.BadRequest(ctx, w, "invalid JSON format")

			return
		}

		if err := valid.ValidateStruct(&request); err != nil {
			logger.WithError(err).Error("Request valid failed")
			response.BadRequest(ctx, w, err.Error())

			return
		}

		if err := userService.UpdateBalance(ctx, request, userID, sourceType); err != nil {
			logger.WithError(err).Error("Failed to update user balance")

			switch {
			case errors.Is(err, customErrors.ErrUserNotFound):
				response.Error(ctx, w, http.StatusNotFound, "user not found")
			case errors.Is(err, customErrors.ErrTransactionExists):
				response.Error(ctx, w, http.StatusConflict, "transaction with this ID already exists")
			case errors.Is(err, customErrors.ErrInsufficientFunds):
				response.Error(ctx, w, http.StatusBadRequest, "insufficient funds for this transaction")
			case errors.Is(err, customErrors.ErrInvalidAmountFormat):
				response.Error(ctx, w, http.StatusBadRequest, "invalid amount format")
			default:
				response.Error(ctx, w, http.StatusInternalServerError, "failed to process transaction")
			}

			return
		}

		response.JSON(ctx, w, http.StatusOK, map[string]string{
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
			case errors.Is(err, customErrors.ErrUserNotFound):
				response.Error(ctx, w, http.StatusNotFound, "user not found")
			default:
				response.Error(ctx, w, http.StatusInternalServerError, "internal server error")
			}

			return
		}

		response.JSON(ctx, w, http.StatusOK, balanceResponse)
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
