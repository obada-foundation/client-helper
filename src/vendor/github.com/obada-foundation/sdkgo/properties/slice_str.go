package properties

import (
	"fmt"
	"github.com/obada-foundation/sdkgo/hash"
	"log"
	"strconv"
	"strings"
)

// SliceStrProperty represent slice of strings property and their hash
type SliceStrProperty struct {
	value []string
	hash  hash.Hash
}

// NewSliceStrProperty creates a new obit property from slice strings
func NewSliceStrProperty(description string, value []string, logger *log.Logger, debug bool) (SliceStrProperty, error) {
	var stp SliceStrProperty

	if debug {
		logger.Printf("\n <|%s|> => NewSliceStrProperty(%v)", description, value)
	}

	dec := make([]uint64, 0, len(value))
	var decTotal uint64

	for _, str := range value {
		h, err := hash.NewHash([]byte(str), logger, debug)

		if err != nil {
			return stp, fmt.Errorf("cannot hash %q: %w", value, err)
		}

		decTotal += h.GetDec()

		dec = append(dec, h.GetDec())
	}

	if debug {
		var logStr []string

		for _, decValue := range dec {
			logStr = append(logStr, strconv.FormatUint(decValue, 10))
		}

		logger.Printf("(%s) => %d", strings.Join(logStr, " + "), decTotal)
	}

	h, err := hash.NewHash([]byte(strconv.FormatUint(decTotal, 10)), logger, debug)

	if err != nil {
		return stp, fmt.Errorf("cannot hash %q: %w", value, err)
	}

	stp.hash = h
	stp.value = value

	return stp, nil
}

// GetValue returns a slice of strings
func (stp SliceStrProperty) GetValue() []string {
	return stp.value
}

// GetHash returns a slice of strings hash
func (stp SliceStrProperty) GetHash() hash.Hash {
	return stp.hash
}
