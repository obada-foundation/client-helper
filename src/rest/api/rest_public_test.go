package api

import (
	"log"
	"testing"
)

// Test owns state for running and shutting down tests.
type Test struct {
	Logger   *log.Logger
	Teardown func()

	t *testing.T
}

func NewPublicAPI(t *testing.T) *Test {
	return nil
}
