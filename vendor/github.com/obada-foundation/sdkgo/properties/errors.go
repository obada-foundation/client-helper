package properties

import (
	"errors"
)

var (
	// ErrEmptyMetadataKey throws when record has an empty key
	ErrEmptyMetadataKey = errors.New("metadata key cannot be empty")
)
