package types

import (
	"errors"
)

var (
	// ErrUnauthorized throws when lack of permissions
	ErrUnauthorized = errors.New("unauthorized")

	// ErrUnauthorizedNoSignature throws when request has not signature field
	ErrUnauthorizedNoSignature = errors.New("unauthorized, no signature")
)
