package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/TiPSYDiPSY/home-task/internal/db"
	"github.com/TiPSYDiPSY/home-task/internal/model/api"
)

type UserService interface {
	GetBalance(ctx context.Context, userID uint64) (api.BalanceResponse, error)
	UpdateBalance(ctx context.Context, req api.TransactionRequest, UserID uint64, SourceType string) error
}

type userService struct {
	repo db.UserRepository
}

const (
	CentsToDollarsMultiplier = 100
	DecimalPlaces            = 2
)

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

	centsDecimal := decimal.NewFromInt(user.Balance)
	dollarsDecimal := centsDecimal.Div(decimal.NewFromInt(CentsToDollarsMultiplier))

	return api.BalanceResponse{
		UserID:  user.ID,
		Balance: dollarsDecimal.StringFixed(DecimalPlaces),
	}, nil
}

func (s *userService) UpdateBalance(ctx context.Context, req api.TransactionRequest, userID uint64, sourceType string) error {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount format: %w", err)
	}

	if req.State == "lose" {
		amount = amount.Neg()
	}

	amountInCents := amount.Mul(decimal.NewFromInt(CentsToDollarsMultiplier)).IntPart()

	if err := s.repo.UpdateUserBalance(ctx, db.Transaction{
		UserID:        userID,
		State:         req.State,
		SourceType:    sourceType,
		TransactionID: req.TransactionID,
		Amount:        amountInCents,
	}); err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			return ErrUserNotFound
		}

		if errors.Is(err, db.ErrDuplicateTransaction) {
			return fmt.Errorf("transaction already exists: %w", err)
		}

		if errors.Is(err, db.ErrInsufficientFunds) {
			return fmt.Errorf("insufficient funds: %w", err)
		}

		return fmt.Errorf("UpdateUserBalance error: %w", err)
	}

	return nil
}
