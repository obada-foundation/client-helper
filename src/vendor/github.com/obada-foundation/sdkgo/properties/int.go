package properties

import (
	"github.com/obada-foundation/sdkgo/hash"
	"log"
	"strconv"
)

// IntProperty hosts hash and value of int property
type IntProperty struct {
	value int64
	hash  hash.Hash
}

// NewIntProperty creates a new obit property from integer value
func NewIntProperty(description string, i int64, logger *log.Logger, debug bool) (IntProperty, error) {
	var ip IntProperty

	if debug {
		logger.Printf("\n <|%s|> => NewIntProperty(%v)", description, i)
	}

	h, err := hash.NewHash([]byte(strconv.FormatInt(i, 10)), logger, debug)

	if err != nil {
		return ip, err
	}

	ip.hash = h
	ip.value = i

	return ip, nil
}

// GetValue returns property value in int64
func (sp IntProperty) GetValue() int64 {
	return sp.value
}

// GetHash returns a hash of int64 value
func (sp IntProperty) GetHash() hash.Hash {
	return sp.hash
}
