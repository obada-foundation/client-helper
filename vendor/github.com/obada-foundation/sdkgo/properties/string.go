package properties

import (
	"fmt"
	"log"

	"github.com/obada-foundation/sdkgo/hash"
)

// StringProperty represent a string property and string hash
type StringProperty struct {
	value string
	hash  hash.Hash
}

// NewStringProperty creates a new obit property from the string
func NewStringProperty(description, value string, logger *log.Logger) (StringProperty, error) {
	var sp StringProperty

	if logger != nil {
		logger.Printf("\n <|%s|> => NewStringProperty(%q)", description, value)
	}

	h, err := hash.NewHash([]byte(value), logger)
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
