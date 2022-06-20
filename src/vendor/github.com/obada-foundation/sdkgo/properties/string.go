package properties

import (
	"fmt"
	"github.com/obada-foundation/sdkgo/hash"
	"log"
)

// StringProperty represent a string property and string hash
type StringProperty struct {
	value string
	hash  hash.Hash
}

// NewStringProperty creates a new obit property from the string
func NewStringProperty(description, value string, logger *log.Logger, debug bool) (StringProperty, error) {
	var sp StringProperty

	if debug {
		logger.Printf("\n <|%s|> => NewStringProperty(%v)", description, sp)
	}

	h, err := hash.NewHash([]byte(value), logger, debug)

	if err != nil {
		return sp, fmt.Errorf("cannot hash %q: %w", value, err)
	}

	sp.hash = h
	sp.value = value

	return sp, nil
}

// GetValue returns a string value
func (sp StringProperty) GetValue() string {
	return sp.value
}

// GetHash returns string hash
func (sp StringProperty) GetHash() hash.Hash {
	return sp.hash
}
