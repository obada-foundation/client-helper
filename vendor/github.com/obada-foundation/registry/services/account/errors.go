package account

import (
	"errors"
)

var (
	// ErrPublicKeyIsEmpty thrown when public key is empty
	ErrPublicKeyIsEmpty = errors.New("public key is empty")

	// ErrInvalidPublicKey thrown when public key is invalid
	ErrInvalidPublicKey = errors.New("invalid public key")

	// ErrInvalidAddress thrown when address is invalid or not registered
	ErrInvalidAddress = errors.New("invalid address")

	// ErrPubKeyNotRegistered thrown when public key is not registered
	ErrPubKeyNotRegistered = errors.New("public key is not registered")
)
