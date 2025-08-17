package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/TiPSYDiPSY/home-task/internal/db"
	"github.com/TiPSYDiPSY/home-task/internal/model/api"
)

type UserService interface {
	GetBalance(ctx context.Context, userID uint64) (api.BalanceResponse, error)
}

type userService struct {
	repo db.UserRepository
}

var ErrUserNotFound = errors.New("user not found")

func newUserService(repo db.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) GetBalance(ctx context.Context, userID uint64) (api.BalanceResponse, error) {
	user, err := s.repo.GetUserData(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.BalanceResponse{}, ErrUserNotFound
		}

		return api.BalanceResponse{}, fmt.Errorf("GetUserData error: %w", err)
	}

	return mapToBalanceResponse(user), nil
}

func mapToBalanceResponse(user db.User) api.BalanceResponse {
	major := user.Balance / 100
	cents := user.Balance % 100

	return api.BalanceResponse{
		UserID:  user.ID,
		Balance: fmt.Sprintf("%d.%02d", major, cents),
	}
}
