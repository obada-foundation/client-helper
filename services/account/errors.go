package account

import (
	"errors"
)

var (
	// ErrProfileExists is returned when a profile already exists.
	ErrProfileExists = errors.New("profile already exists")

	// ErrWalletExists wallet already exists
	ErrWalletExists = errors.New("profile wallet already exists")

	// ErrWalletNotExists no wallet
	ErrWalletNotExists = errors.New("profile wallet doesn't exists")

	// ErrProfileNotExists no profile
	ErrProfileNotExists = errors.New("profile doesn't exists")

	// ErrAccountNotExists is returned when an account doesn't exist.
	ErrAccountNotExists = errors.New("account doesn't exists")

	// ErrAccountExists account already exists
	ErrAccountExists = errors.New("account already exists")

	// ErrAccountHasZeroTx no tx for account
	ErrAccountHasZeroTx = errors.New("please use the prior address before creating a new one")

	// ErrInvalidMnemonic invalid mnemonic
	ErrInvalidMnemonic = errors.New("invalid mnemonic")

	// ErrHDAccountDelete cannot delete hd account
	ErrHDAccountDelete = errors.New("cannot delete hd account")
)

// IsAccountError errors that can send back to the client
func IsAccountError(err error) bool {
	return errors.Is(err, ErrProfileExists) ||
		errors.Is(err, ErrAccountHasZeroTx) ||
		errors.Is(err, ErrAccountExists) ||
		errors.Is(err, ErrWalletExists) ||
		errors.Is(err, ErrInvalidMnemonic) ||
		errors.Is(err, ErrWalletNotExists)
}
