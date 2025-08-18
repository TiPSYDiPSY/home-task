package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	errs "github.com/TiPSYDiPSY/home-task/internal/errors"

	"github.com/TiPSYDiPSY/home-task/internal/db"
	"github.com/TiPSYDiPSY/home-task/internal/model/api"
)

type UserService interface {
	GetBalance(ctx context.Context, userID uint64) (api.BalanceResponse, error)
	UpdateBalance(ctx context.Context, req api.TransactionRequest, UserID uint64, SourceType string) error
}

type userService struct {
	repo                  db.UserRepository
	centsToDollarsDecimal decimal.Decimal // Move to struct field to avoid global variable
}

const (
	CentsToDollarsMultiplier = 100
	DecimalPlaces            = 2
)

func newUserService(repo db.UserRepository) UserService {
	return &userService{
		repo:                  repo,
		centsToDollarsDecimal: decimal.NewFromInt(CentsToDollarsMultiplier),
	}
}

func (s *userService) GetBalance(ctx context.Context, userID uint64) (api.BalanceResponse, error) {
	user, err := s.repo.GetUserData(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.BalanceResponse{}, errs.ErrUserNotFound
		}

		return api.BalanceResponse{}, fmt.Errorf("GetUserData error: %w", err)
	}

	centsDecimal := decimal.NewFromInt(user.Balance)
	dollarsDecimal := centsDecimal.Div(s.centsToDollarsDecimal)

	return api.BalanceResponse{
		UserID:  user.ID,
		Balance: dollarsDecimal.StringFixed(DecimalPlaces),
	}, nil
}

func (s *userService) UpdateBalance(ctx context.Context, req api.TransactionRequest, userID uint64, sourceType string) error {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return errs.ErrInvalidAmountFormat
	}

	if req.State == "lose" {
		amount = amount.Neg()
	}

	amountInCents := amount.Mul(s.centsToDollarsDecimal).IntPart()

	if err := s.repo.UpdateUserBalance(ctx, db.Transaction{
		UserID:        userID,
		State:         req.State,
		SourceType:    sourceType,
		TransactionID: req.TransactionID,
		Amount:        amountInCents,
	}); err != nil {
		switch {
		case errors.Is(err, db.ErrUserNotFound):
			return errs.ErrUserNotFound
		case errors.Is(err, db.ErrDuplicateTransaction):
			return errs.ErrTransactionExists
		case errors.Is(err, db.ErrInsufficientFunds):
			return errs.ErrInsufficientFunds
		default:
			return fmt.Errorf("UpdateUserBalance error: %w", err)
		}
	}

	return nil
}
