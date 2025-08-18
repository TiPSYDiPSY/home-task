package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserData(ctx context.Context, userID uint64) (User, error)
	UpdateUserBalance(ctx context.Context, transaction Transaction) error
}

const (
	ReadTimeoutSeconds  = 5
	WriteTimeoutSeconds = 10
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrDuplicateTransaction = errors.New("duplicate transaction")
	ErrInsufficientFunds    = errors.New("insufficient funds")
)

func (r *PostgresDBDataStore) GetUserData(ctx context.Context, userID uint64) (user User, err error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, ReadTimeoutSeconds*time.Second)
	defer cancel()

	return user, r.db.WithContext(ctxWithTimeout).First(&user, userID).Error
}

func (r *PostgresDBDataStore) UpdateUserBalance(ctx context.Context, transaction Transaction) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, WriteTimeoutSeconds*time.Second)
	defer cancel()

	if err := r.db.WithContext(ctxWithTimeout).Transaction(func(tx *gorm.DB) error {
		if err := r.checkTransactionExists(tx, transaction.UserID, transaction.TransactionID); err != nil {
			return err
		}

		if err := r.updateUserBalanceAtomic(tx, transaction.UserID, transaction.Amount); err != nil {
			return err
		}

		return r.createTransactionRecord(tx, transaction)
	}); err != nil {
		return fmt.Errorf("failed to execute balance update transaction: %w", err)
	}

	return nil
}

func (*PostgresDBDataStore) checkTransactionExists(tx *gorm.DB, userID uint64, transactionID string) error {
	var count int64
	if err := tx.Model(&Transaction{}).
		Where("user_id = ? AND transaction_id = ?", userID, transactionID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check transaction existence: %w", err)
	}

	if count > 0 {
		return ErrDuplicateTransaction
	}

	return nil
}

func (r *PostgresDBDataStore) updateUserBalanceAtomic(tx *gorm.DB, userID uint64, amount int64) error {
	result := tx.Model(&User{}).
		Where("id = ? AND balance + ? >= 0", userID, amount).
		Update("balance", gorm.Expr("balance + ?", amount))

	if result.Error != nil {
		return fmt.Errorf("failed to update user balance: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return r.determineUpdateFailureReason(tx, userID)
	}

	return nil
}

func (*PostgresDBDataStore) determineUpdateFailureReason(tx *gorm.DB, userID uint64) error {
	var userExists bool
	if err := tx.Model(&User{}).
		Select("1").
		Where("id = ?", userID).
		Limit(1).
		Find(&userExists).Error; err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if !userExists {
		return ErrUserNotFound
	}

	return ErrInsufficientFunds
}

func (*PostgresDBDataStore) createTransactionRecord(tx *gorm.DB, transaction Transaction) error {
	if err := tx.Create(&transaction).Error; err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	return nil
}
