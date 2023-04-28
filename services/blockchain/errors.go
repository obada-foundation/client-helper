package blockchain

import (
	"errors"
)

var (
	// ErrInsufficientFunds is returned when the account balance has insufficient funds to complete the transaction.
	ErrInsufficientFunds = errors.New("out of funds")
)

// IsAcceptableError returns true if the error is acceptable to return to the client.
func IsAcceptableError(err error) bool {
	return errors.Is(err, ErrInsufficientFunds)
}
