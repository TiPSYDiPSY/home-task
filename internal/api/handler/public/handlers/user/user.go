package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/TiPSYDiPSY/home-task/internal/util/response"

	"github.com/TiPSYDiPSY/home-task/internal/service"
)

func UpdateBalance(userService service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

func GetBalance(userService service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logrus.WithContext(ctx)

		userID, err := parseUserID(r)
		if err != nil {
			response.HandleError(ctx, w, http.StatusBadRequest, err.Error())

			return
		}

		balanceResponse, err := userService.GetBalance(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				response.HandleError(ctx, w, http.StatusNotFound, "user not found")
			default:
				log.WithError(err).Error("Error getting user balance")
				response.HandleError(ctx, w, http.StatusInternalServerError, "internal server error")
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

	return strconv.ParseUint(userIDStr, 10, 64)
}
