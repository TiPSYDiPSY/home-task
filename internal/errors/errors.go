package errors

import "errors"

type ValidationError struct {
	Field   string
	Message string
}

type DatabaseError struct {
	Operation string
	Err       error
}

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrInvalidAmountFormat = errors.New("invalid amount format")
	ErrTransactionExists   = errors.New("transaction already exists")
)

func (e ValidationError) Error() string {
	return e.Message
}

func (e DatabaseError) Error() string {
	return "database operation failed: " + e.Operation
}

func (e DatabaseError) Unwrap() error {
	return e.Err
}
