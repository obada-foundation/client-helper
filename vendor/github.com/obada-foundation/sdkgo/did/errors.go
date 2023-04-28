package did

import (
	"errors"
)

var (
	// ErrNotSupportedDIDMethod thows when not supported methods are passed for DID creation
	ErrNotSupportedDIDMethod = errors.New("given DID method is not supported")
)
