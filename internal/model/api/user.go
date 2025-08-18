package api

type BalanceResponse struct {
	UserID  uint64 `json:"userId"` //nolint: tagliatelle // Per API spec
	Balance string `json:"balance"`
}

type TransactionRequest struct {
	State         string `json:"state"         validate:"required,oneof=win lose"`
	Amount        string `json:"amount"        validate:"required,decimal2"`
	TransactionID string `json:"transactionId" validate:"required"` //nolint: tagliatelle // Per API spec
}
