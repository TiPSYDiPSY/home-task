package api

type BalanceResponse struct {
	UserID  uint64 `json:"user_id"`
	Balance string `json:"balance"`
}

type TransactionRequest struct {
	State         string `json:"state"`
	Amount        string `json:"amount"`
	TransactionID string `json:"transactionId"` //nolint: tagliatelle // Per API spec
}
