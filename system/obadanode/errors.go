package obadanode

import (
	"errors"
)

var (
	// ErrAccountHasZeroTx is returned when an account has no transactions.
	ErrAccountHasZeroTx = errors.New("account has zero transactions")

	// ErrInsufficientFunds is returned when an account has not enough balance to commit transaction.
	ErrInsufficientFunds = errors.New("insufficient funds")
)
