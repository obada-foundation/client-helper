package diddoc

import (
	"errors"
)

var (
	// ErrDIDNotRegistered thows when requested DID is not present in registry
	ErrDIDNotRegistered = errors.New("DID is not registered")

	// ErrDIDAlreadyRegistered thows when system tries to register DID that is already registered
	ErrDIDAlreadyRegistered = errors.New("DID was already registered")
)
