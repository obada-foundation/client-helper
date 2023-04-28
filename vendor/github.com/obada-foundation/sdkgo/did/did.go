// Package did provides functionality for OBADA DID creation and parsing
package did

import (
	"fmt"
	"log"
	"strings"

	"github.com/obada-foundation/sdkgo/base58"
	"github.com/obada-foundation/sdkgo/hash"
	"github.com/obada-foundation/sdkgo/properties"
)

// DefaultUSNLength the default length of Universal Serial Number
const DefaultUSNLength = 12

// ObadaDIDMethod DID Method for OBADA protocol
// https://www.w3.org/TR/did-core/#dfn-did-methods
const ObadaDIDMethod = "did:obada:"

// DID represent OBADA DID
type DID struct {
	// Universal serial number with limited length
	usn     string
	fullUSN string
	did     string
	hash    hash.Hash
}

// NewDID combines serial number, manufacturer and part number to generate DID
type NewDID struct {
	SerialNumber string
	Manufacturer string
	PartNumber   string
	Logger       *log.Logger
}

// MakeDID creates new OBADA DID
func MakeDID(newDID NewDID) (did *DID, err error) {
	if newDID.Logger != nil {
		newDID.Logger.Printf(
			"MakeDID(%q, %q, %q)",
			newDID.SerialNumber,
			newDID.Manufacturer,
			newDID.PartNumber,
		)
	}

	snProp, err := properties.NewStringProperty(
		"Making serialNumber",
		newDID.SerialNumber,
		newDID.Logger,
	)
	if err != nil {
		return nil, err
	}

	mnProp, err := properties.NewStringProperty(
		"Making manufacturer",
		newDID.Manufacturer,
		newDID.Logger,
	)
	if err != nil {
		return nil, err
	}

	pnProp, err := properties.NewStringProperty(
		"Making partNumber",
		newDID.PartNumber,
		newDID.Logger,
	)
	if err != nil {
		return nil, err
	}

	snHash := snProp.GetHash()
	mh := mnProp.GetHash()
	pnh := pnProp.GetHash()

	sum := hash.SumHashes(newDID.Logger, snHash, mh, pnh)

	h, err := hash.NewHash([]byte(fmt.Sprintf("%x", sum)), newDID.Logger)
	if err != nil {
		return nil, fmt.Errorf("cannot create DID: %w", err)
	}

	hashStr := h.GetHash()

	fullUSN := base58.Encode([]byte(hashStr))

	defer func(logger *log.Logger) {
		if logger != nil {
			logger.Printf("Hash: %s", h.GetHash())
			logger.Printf("DID: %s", did.did)
			logger.Printf("USN: %s", did.usn)
			logger.Printf("Full USN: %s", did.fullUSN)
		}
	}(newDID.Logger)

	return &DID{
		hash:    h,
		did:     ObadaDIDMethod + hashStr,
		fullUSN: fullUSN,
		usn:     fullUSN[:DefaultUSNLength],
	}, nil
}

// GetHash returns DID hash
func (did DID) GetHash() hash.Hash {
	return did.hash
}

// String returns DID string
func (did DID) String() string {
	return did.did
}

// GetUSN returns short universal serial number
func (did DID) GetUSN() string {
	return did.usn
}

// GetFullUSN returns full universal serial number
func (did DID) GetFullUSN() string {
	return did.fullUSN
}

// FromString creates DID struct from DID string
func FromString(str string, logger *log.Logger) (*DID, error) {
	if !strings.HasPrefix(str, ObadaDIDMethod) {
		return nil, ErrNotSupportedDIDMethod
	}

	hashStr := str[len(ObadaDIDMethod):]

	fullUSN := base58.Encode([]byte(hashStr))

	h, err := hash.FromString(hashStr, logger)
	if err != nil {
		return nil, fmt.Errorf("cannot create DID: %w", err)
	}

	return &DID{
		did:     str,
		fullUSN: fullUSN,
		usn:     fullUSN[:DefaultUSNLength],
		hash:    h,
	}, nil
}
