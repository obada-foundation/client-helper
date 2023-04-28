package device

import (
	"errors"
)

var (
	// ErrDeviceNotExists device not exists
	ErrDeviceNotExists = errors.New("device doesn't exists")
)
