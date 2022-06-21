package account

import (
	"errors"
)

var (
	ErrAccountExists    = errors.New("account already exists")
	ErrAccountNotExists = errors.New("account doesn't exists")
)
