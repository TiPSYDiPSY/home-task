package service

import (
	"context"
	"errors"
	"testing"

	errs "github.com/TiPSYDiPSY/home-task/internal/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/TiPSYDiPSY/home-task/internal/db"
	"github.com/TiPSYDiPSY/home-task/internal/model/api"
)

func TestNewUserService(t *testing.T) {
	mockRepo := db.NewMockUserRepository(t)
	service := newUserService(mockRepo)

	assert.NotNil(t, service)
	assert.Implements(t, (*UserService)(nil), service)
}

func TestGetBalance(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		userID         uint64
		mockSetup      func(*db.MockUserRepository)
		expectedResult api.BalanceResponse
		expectedError  error
	}{
		{
			name:   "successful balance retrieval",
			userID: 1,
			mockSetup: func(mockRepo *db.MockUserRepository) {
				mockRepo.EXPECT().GetUserData(ctx, uint64(1)).Return(db.User{
					ID:      1,
					Balance: 1500,
				}, nil)
			},
			expectedResult: api.BalanceResponse{
				UserID:  1,
				Balance: "15.00",
			},
			expectedError: nil,
		},
		{
			name:   "user not found",
			userID: 999,
			mockSetup: func(mockRepo *db.MockUserRepository) {
				mockRepo.EXPECT().GetUserData(ctx, uint64(999)).Return(db.User{}, gorm.ErrRecordNotFound)
			},
			expectedResult: api.BalanceResponse{},
			expectedError:  errs.ErrUserNotFound,
		},
		{
			name:   "database error",
			userID: 1,
			mockSetup: func(mockRepo *db.MockUserRepository) {
				mockRepo.EXPECT().GetUserData(ctx, uint64(1)).Return(db.User{}, errors.New("database connection error"))
			},
			expectedResult: api.BalanceResponse{},
			expectedError:  errors.New("GetUserData error: database connection error"),
		},
		{
			name:   "zero balance",
			userID: 2,
			mockSetup: func(mockRepo *db.MockUserRepository) {
				mockRepo.EXPECT().GetUserData(ctx, uint64(2)).Return(db.User{
					ID:      2,
					Balance: 0,
				}, nil)
			},
			expectedResult: api.BalanceResponse{
				UserID:  2,
				Balance: "0.00",
			},
			expectedError: nil,
		},
		{
			name:   "large balance",
			userID: 3,
			mockSetup: func(mockRepo *db.MockUserRepository) {
				mockRepo.EXPECT().GetUserData(ctx, uint64(3)).Return(db.User{
					ID:      3,
					Balance: 123456789,
				}, nil)
			},
			expectedResult: api.BalanceResponse{
				UserID:  3,
				Balance: "1234567.89",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := db.NewMockUserRepository(t)
			tt.mockSetup(mockRepo)

			service := newUserService(mockRepo)
			result, err := service.GetBalance(ctx, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestUpdateBalance(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       api.TransactionRequest
		userID        uint64
		sourceType    string
		mockSetup     func(*db.MockUserRepository)
		expectedError error
	}{
		{
			name: "successful win transaction",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "10.50",
				TransactionID: "txn-123",
			},
			userID:     1,
			sourceType: "game",
			mockSetup: func(mockRepo *db.MockUserRepository) {
				expectedTransaction := db.Transaction{
					UserID:        1,
					State:         "win",
					SourceType:    "game",
					TransactionID: "txn-123",
					Amount:        1050,
				}
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "successful lose transaction",
			request: api.TransactionRequest{
				State:         "lose",
				Amount:        "5.25",
				TransactionID: "txn-456",
			},
			userID:     2,
			sourceType: "game",
			mockSetup: func(mockRepo *db.MockUserRepository) {
				expectedTransaction := db.Transaction{
					UserID:        2,
					State:         "lose",
					SourceType:    "game",
					TransactionID: "txn-456",
					Amount:        -525,
				}
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "invalid amount format",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "invalid",
				TransactionID: "txn-789",
			},
			userID:        1,
			sourceType:    "game",
			mockSetup:     func(mockRepo *db.MockUserRepository) {},
			expectedError: errors.New("invalid amount format"),
		},
		{
			name: "user not found",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "10.00",
				TransactionID: "txn-999",
			},
			userID:     999,
			sourceType: "game",
			mockSetup: func(mockRepo *db.MockUserRepository) {
				expectedTransaction := db.Transaction{
					UserID:        999,
					State:         "win",
					SourceType:    "game",
					TransactionID: "txn-999",
					Amount:        1000,
				}
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(db.ErrUserNotFound)
			},
			expectedError: errs.ErrUserNotFound,
		},
		{
			name: "duplicate transaction",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "10.00",
				TransactionID: "txn-duplicate",
			},
			userID:     1,
			sourceType: "game",
			mockSetup: func(mockRepo *db.MockUserRepository) {
				expectedTransaction := db.Transaction{
					UserID:        1,
					State:         "win",
					SourceType:    "game",
					TransactionID: "txn-duplicate",
					Amount:        1000,
				}
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(db.ErrDuplicateTransaction)
			},
			expectedError: errors.New("transaction already exists"),
		},
		{
			name: "insufficient funds",
			request: api.TransactionRequest{
				State:         "lose",
				Amount:        "100.00",
				TransactionID: "txn-insufficient",
			},
			userID:     1,
			sourceType: "game",
			mockSetup: func(mockRepo *db.MockUserRepository) {
				expectedTransaction := db.Transaction{
					UserID:        1,
					State:         "lose",
					SourceType:    "game",
					TransactionID: "txn-insufficient",
					Amount:        -10000,
				}
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(db.ErrInsufficientFunds)
			},
			expectedError: errors.New("insufficient funds"),
		},
		{
			name: "database error",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "10.00",
				TransactionID: "txn-db-error",
			},
			userID:     1,
			sourceType: "game",
			mockSetup: func(mockRepo *db.MockUserRepository) {
				expectedTransaction := db.Transaction{
					UserID:        1,
					State:         "win",
					SourceType:    "game",
					TransactionID: "txn-db-error",
					Amount:        1000,
				}
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(errors.New("database connection error"))
			},
			expectedError: errors.New("UpdateUserBalance error"),
		},
		{
			name: "zero amount win",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "0.00",
				TransactionID: "txn-zero",
			},
			userID:     1,
			sourceType: "game",
			mockSetup: func(mockRepo *db.MockUserRepository) {
				expectedTransaction := db.Transaction{
					UserID:        1,
					State:         "win",
					SourceType:    "game",
					TransactionID: "txn-zero",
					Amount:        0,
				}
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "decimal amount with many places",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "10.999",
				TransactionID: "txn-decimal",
			},
			userID:     1,
			sourceType: "game",
			mockSetup: func(mockRepo *db.MockUserRepository) {
				expectedTransaction := db.Transaction{
					UserID:        1,
					State:         "win",
					SourceType:    "game",
					TransactionID: "txn-decimal",
					Amount:        1099,
				}
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := db.NewMockUserRepository(t)
			tt.mockSetup(mockRepo)

			service := newUserService(mockRepo)
			err := service.UpdateBalance(ctx, tt.request, tt.userID, tt.sourceType)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_EdgeCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                  string
		userID                uint64
		sourceType            string
		request               api.TransactionRequest
		expectedAmountInCents int64
		mockSetup             func(*db.MockUserRepository, db.Transaction)
		expectedError         error
	}{
		{
			name:       "very large amount conversion",
			userID:     1,
			sourceType: "game",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "999999.99",
				TransactionID: "txn-large",
			},
			expectedAmountInCents: 99999999,
			mockSetup: func(mockRepo *db.MockUserRepository, expectedTransaction db.Transaction) {
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "negative amount string for win",
			userID:     1,
			sourceType: "game",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "-10.00",
				TransactionID: "txn-negative-win",
			},
			expectedAmountInCents: -1000,
			mockSetup: func(mockRepo *db.MockUserRepository, expectedTransaction db.Transaction) {
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "amount with single decimal place",
			userID:     1,
			sourceType: "game",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "10.5",
				TransactionID: "txn-single-decimal",
			},
			expectedAmountInCents: 1050,
			mockSetup: func(mockRepo *db.MockUserRepository, expectedTransaction db.Transaction) {
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "whole number amount",
			userID:     1,
			sourceType: "game",
			request: api.TransactionRequest{
				State:         "lose",
				Amount:        "50",
				TransactionID: "txn-whole-number",
			},
			expectedAmountInCents: -5000,
			mockSetup: func(mockRepo *db.MockUserRepository, expectedTransaction db.Transaction) {
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "very small amount",
			userID:     1,
			sourceType: "game",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "0.01",
				TransactionID: "txn-small",
			},
			expectedAmountInCents: 1,
			mockSetup: func(mockRepo *db.MockUserRepository, expectedTransaction db.Transaction) {
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "amount with three decimal places",
			userID:     1,
			sourceType: "game",
			request: api.TransactionRequest{
				State:         "win",
				Amount:        "12.345",
				TransactionID: "txn-three-decimals",
			},
			expectedAmountInCents: 1234,
			mockSetup: func(mockRepo *db.MockUserRepository, expectedTransaction db.Transaction) {
				mockRepo.EXPECT().UpdateUserBalance(ctx, expectedTransaction).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := db.NewMockUserRepository(t)
			service := newUserService(mockRepo)

			expectedTransaction := db.Transaction{
				UserID:        tt.userID,
				State:         tt.request.State,
				SourceType:    tt.sourceType,
				TransactionID: tt.request.TransactionID,
				Amount:        tt.expectedAmountInCents,
			}

			tt.mockSetup(mockRepo, expectedTransaction)

			err := service.UpdateBalance(ctx, tt.request, tt.userID, tt.sourceType)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_BalanceConversion(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		balanceInCents  int64
		expectedBalance string
	}{
		{"zero balance", 0, "0.00"},
		{"one cent", 1, "0.01"},
		{"ten cents", 10, "0.10"},
		{"one dollar", 100, "1.00"},
		{"ten dollars fifty cents", 1050, "10.50"},
		{"large amount", 123456789, "1234567.89"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := db.NewMockUserRepository(t)
			service := newUserService(mockRepo)

			mockRepo.EXPECT().GetUserData(ctx, uint64(1)).Return(db.User{
				ID:      1,
				Balance: tt.balanceInCents,
			}, nil)

			result, err := service.GetBalance(ctx, 1)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBalance, result.Balance)
			assert.Equal(t, uint64(1), result.UserID)
		})
	}
}
