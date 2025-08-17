package api

type BalanceResponse struct {
	UserID  uint64 `json:"user_id"`
	Balance string `json:"balance"`
}
