package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TiPSYDiPSY/home-task/internal/util/validation"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/TiPSYDiPSY/home-task/internal/api/handler/public/handlers/middleware"
	"github.com/TiPSYDiPSY/home-task/internal/model/api"
	"github.com/TiPSYDiPSY/home-task/internal/service"
)

func TestParseUserID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		expectedID    uint64
		expectedError string
	}{
		{
			name:          "valid user ID",
			userID:        "123",
			expectedID:    123,
			expectedError: "",
		},
		{
			name:          "user ID 1",
			userID:        "1",
			expectedID:    1,
			expectedError: "",
		},
		{
			name:          "large user ID",
			userID:        "18446744073709551615",
			expectedID:    18446744073709551615,
			expectedError: "",
		},
		{
			name:          "zero user ID",
			userID:        "0",
			expectedID:    0,
			expectedError: "user ID must be positive",
		},
		{
			name:          "negative user ID",
			userID:        "-1",
			expectedID:    0,
			expectedError: "invalid user ID format",
		},
		{
			name:          "invalid format - letters",
			userID:        "abc",
			expectedID:    0,
			expectedError: "invalid user ID format",
		},
		{
			name:          "invalid format - float",
			userID:        "12.34",
			expectedID:    0,
			expectedError: "invalid user ID format",
		},
		{
			name:          "empty user ID",
			userID:        "",
			expectedID:    0,
			expectedError: "user ID is required",
		},
		{
			name:          "user ID with spaces",
			userID:        "  123  ",
			expectedID:    0,
			expectedError: "invalid user ID format",
		},
		{
			name:          "overflow uint64",
			userID:        "18446744073709551616",
			expectedID:    0,
			expectedError: "invalid user ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/user/placeholder/balance", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("userID", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			userID, err := parseUserID(req)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, uint64(0), userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	type prepareMocks func(*service.MockUserService)
	type args struct {
		userID string
	}

	tests := []struct {
		name         string
		args         args
		prepareMocks prepareMocks
		wantHTTPCode int
		wantBody     string
	}{
		{
			name: "successful balance retrieval",
			args: args{userID: "1"},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().GetBalance(mock.Anything, uint64(1)).Return(
					api.BalanceResponse{
						UserID:  1,
						Balance: "15.50",
					}, nil)
			},
			wantHTTPCode: http.StatusOK,
			wantBody: `{
				"data": {
					"userId": 1,
					"balance": "15.50"
				}
			}`,
		},
		{
			name: "zero balance",
			args: args{userID: "2"},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().GetBalance(mock.Anything, uint64(2)).Return(
					api.BalanceResponse{
						UserID:  2,
						Balance: "0.00",
					}, nil)
			},
			wantHTTPCode: http.StatusOK,
			wantBody: `{
				"data": {
					"userId": 2,
					"balance": "0.00"
				}
			}`,
		},
		{
			name: "large balance",
			args: args{userID: "3"},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().GetBalance(mock.Anything, uint64(3)).Return(
					api.BalanceResponse{
						UserID:  3,
						Balance: "12345.67",
					}, nil)
			},
			wantHTTPCode: http.StatusOK,
			wantBody: `{
				"data": {
					"userId": 3,
					"balance": "12345.67"
				}
			}`,
		},
		{
			name: "user not found",
			args: args{userID: "999"},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().GetBalance(mock.Anything, uint64(999)).Return(
					api.BalanceResponse{}, service.ErrUserNotFound)
			},
			wantHTTPCode: http.StatusNotFound,
			wantBody: `{
				"error": "Not Found",
				"message": "user not found"
			}`,
		},
		{
			name: "internal server error",
			args: args{userID: "5"},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().GetBalance(mock.Anything, uint64(5)).Return(
					api.BalanceResponse{}, errors.New("database connection failed"))
			},
			wantHTTPCode: http.StatusInternalServerError,
			wantBody: `{
				"error": "Internal Server Error",
				"message": "internal server Error"
			}`,
		},
		{
			name:         "invalid user ID format",
			args:         args{userID: "invalid"},
			prepareMocks: func(mockService *service.MockUserService) {},
			wantHTTPCode: http.StatusBadRequest,
			wantBody: `{
				"error": "Bad Request",
				"message": "invalid user ID format"
			}`,
		},
		{
			name:         "zero user ID",
			args:         args{userID: "0"},
			prepareMocks: func(mockService *service.MockUserService) {},
			wantHTTPCode: http.StatusBadRequest,
			wantBody: `{
				"error": "Bad Request",
				"message": "user ID must be positive"
			}`,
		},
		{
			name:         "empty user ID",
			args:         args{userID: ""},
			prepareMocks: func(mockService *service.MockUserService) {},
			wantHTTPCode: http.StatusBadRequest,
			wantBody: `{
				"error": "Bad Request",
				"message": "user ID is required"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service.NewMockUserService(t)

			tt.prepareMocks(mockService)

			req := httptest.NewRequest(http.MethodGet, "/user/placeholder/balance", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("userID", tt.args.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler := GetBalance(mockService)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantHTTPCode, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, rr.Body.String())
		})
	}
}

func TestUpdateBalance(t *testing.T) {
	type prepareMocks func(*service.MockUserService)
	type args struct {
		userID     string
		sourceType string
		body       interface{}
	}

	tests := []struct {
		name         string
		args         args
		prepareMocks prepareMocks
		wantHTTPCode int
		wantBody     string
	}{
		{
			name: "successful win transaction",
			args: args{
				userID:     "1",
				sourceType: "game",
				body: api.TransactionRequest{
					State:         "win",
					Amount:        "10.50",
					TransactionID: "txn-123",
				},
			},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().UpdateBalance(mock.Anything, api.TransactionRequest{
					State:         "win",
					Amount:        "10.50",
					TransactionID: "txn-123",
				}, uint64(1), "game").Return(nil)
			},
			wantHTTPCode: http.StatusOK,
			wantBody: `{
				"data": {
					"status": "success",
					"message": "Transaction processed successfully"
				}
			}`,
		},
		{
			name: "successful lose transaction",
			args: args{
				userID:     "2",
				sourceType: "game",
				body: api.TransactionRequest{
					State:         "lose",
					Amount:        "5.25",
					TransactionID: "txn-456",
				},
			},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().UpdateBalance(mock.Anything, api.TransactionRequest{
					State:         "lose",
					Amount:        "5.25",
					TransactionID: "txn-456",
				}, uint64(2), "game").Return(nil)
			},
			wantHTTPCode: http.StatusOK,
			wantBody: `{
				"data": {
					"status": "success",
					"message": "Transaction processed successfully"
				}
			}`,
		},
		{
			name: "user not found",
			args: args{
				userID:     "999",
				sourceType: "game",
				body: api.TransactionRequest{
					State:         "win",
					Amount:        "10.00",
					TransactionID: "txn-notfound",
				},
			},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().UpdateBalance(mock.Anything, api.TransactionRequest{
					State:         "win",
					Amount:        "10.00",
					TransactionID: "txn-notfound",
				}, uint64(999), "game").Return(service.ErrUserNotFound)
			},
			wantHTTPCode: http.StatusNotFound,
			wantBody: `{
				"error": "Not Found",
				"message": "user not found"
			}`,
		},
		{
			name: "duplicate transaction",
			args: args{
				userID:     "3",
				sourceType: "game",
				body: api.TransactionRequest{
					State:         "win",
					Amount:        "10.00",
					TransactionID: "txn-duplicate",
				},
			},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().UpdateBalance(mock.Anything, api.TransactionRequest{
					State:         "win",
					Amount:        "10.00",
					TransactionID: "txn-duplicate",
				}, uint64(3), "game").Return(errors.New("transaction already exists: duplicate"))
			},
			wantHTTPCode: http.StatusConflict,
			wantBody: `{
				"error": "Conflict",
				"message": "transaction with this ID already exists"
			}`,
		},
		{
			name: "insufficient funds",
			args: args{
				userID:     "4",
				sourceType: "game",
				body: api.TransactionRequest{
					State:         "lose",
					Amount:        "1000.00",
					TransactionID: "txn-insufficient",
				},
			},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().UpdateBalance(mock.Anything, api.TransactionRequest{
					State:         "lose",
					Amount:        "1000.00",
					TransactionID: "txn-insufficient",
				}, uint64(4), "game").Return(errors.New("insufficient funds: balance too low"))
			},
			wantHTTPCode: http.StatusBadRequest,
			wantBody: `{
				"error": "Bad Request",
				"message": "insufficient funds for this transaction"
			}`,
		},
		{
			name: "invalid amount format",
			args: args{
				userID:     "5",
				sourceType: "game",
				body: api.TransactionRequest{
					State:         "win",
					Amount:        "invalid",
					TransactionID: "txn-invalid-amount",
				},
			},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().UpdateBalance(mock.Anything, api.TransactionRequest{
					State:         "win",
					Amount:        "invalid",
					TransactionID: "txn-invalid-amount",
				}, uint64(5), "game").Return(errors.New("invalid amount format: not a number"))
			},
			wantHTTPCode: http.StatusBadRequest,
			wantBody: `{
				"error": "Bad Request",
				"message": "invalid amount format"
			}`,
		},
		{
			name: "internal server error",
			args: args{
				userID:     "6",
				sourceType: "game",
				body: api.TransactionRequest{
					State:         "win",
					Amount:        "10.00",
					TransactionID: "txn-server-error",
				},
			},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().UpdateBalance(mock.Anything, api.TransactionRequest{
					State:         "win",
					Amount:        "10.00",
					TransactionID: "txn-server-error",
				}, uint64(6), "game").Return(errors.New("database connection failed"))
			},
			wantHTTPCode: http.StatusInternalServerError,
			wantBody: `{
				"error": "Internal Server Error",
				"message": "failed to process transaction"
			}`,
		},
		{
			name: "invalid user ID format",
			args: args{
				userID:     "invalid",
				sourceType: "game",
				body: api.TransactionRequest{
					State:         "win",
					Amount:        "10.00",
					TransactionID: "txn-789",
				},
			},
			prepareMocks: func(mockService *service.MockUserService) {},
			wantHTTPCode: http.StatusBadRequest,
			wantBody: `{
				"error": "Bad Request",
				"message": "invalid user ID format"
			}`,
		},
		{
			name: "zero user ID",
			args: args{
				userID:     "0",
				sourceType: "game",
				body: api.TransactionRequest{
					State:         "win",
					Amount:        "10.00",
					TransactionID: "txn-zero",
				},
			},
			prepareMocks: func(mockService *service.MockUserService) {},
			wantHTTPCode: http.StatusBadRequest,
			wantBody: `{
				"error": "Bad Request",
				"message": "user ID must be positive"
			}`,
		},
		{
			name: "invalid JSON format",
			args: args{
				userID:     "7",
				sourceType: "game",
				body:       `{"state": "win", "amount": "10.00"`,
			},
			prepareMocks: func(mockService *service.MockUserService) {},
			wantHTTPCode: http.StatusBadRequest,
			wantBody: `{
				"error": "Bad Request",
				"message": "invalid JSON format"
			}`,
		},
		{
			name: "empty request body",
			args: args{
				userID:     "8",
				sourceType: "game",
				body:       "",
			},
			prepareMocks: func(mockService *service.MockUserService) {},
			wantHTTPCode: http.StatusBadRequest,
			wantBody: `{
				"error": "Bad Request",
				"message": "invalid JSON format"
			}`,
		},
		{
			name: "missing source type",
			args: args{
				userID:     "9",
				sourceType: "",
				body: api.TransactionRequest{
					State:         "win",
					Amount:        "10.00",
					TransactionID: "txn-no-source",
				},
			},
			prepareMocks: func(mockService *service.MockUserService) {
				mockService.EXPECT().UpdateBalance(mock.Anything, api.TransactionRequest{
					State:         "win",
					Amount:        "10.00",
					TransactionID: "txn-no-source",
				}, uint64(9), "").Return(nil)
			},
			wantHTTPCode: http.StatusOK,
			wantBody: `{
				"data": {
					"status": "success",
					"message": "Transaction processed successfully"
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service.NewMockUserService(t)

			tt.prepareMocks(mockService)

			var bodyBytes []byte
			if str, ok := tt.args.body.(string); ok {
				bodyBytes = []byte(str)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tt.args.body)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/user/placeholder/transaction", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("userID", tt.args.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			if tt.args.sourceType != "" {
				ctx := context.WithValue(req.Context(), middleware.SourceTypeKey, tt.args.sourceType)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler := UpdateBalance(mockService, validation.NewValidator())

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantHTTPCode, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, rr.Body.String())
		})
	}
}
