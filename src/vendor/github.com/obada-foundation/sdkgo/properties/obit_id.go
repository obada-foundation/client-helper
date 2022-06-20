package properties

import (
	"fmt"
	"github.com/obada-foundation/sdkgo/base58"
	"github.com/obada-foundation/sdkgo/hash"
	"log"
)

// ObitID represent obit identifier
type ObitID struct {
	usn     string
	fullUsn string
	did     string
	hash    hash.Hash
}

// NewObitIDProperty creates new ObitID from given arguments
func NewObitIDProperty(serialNumberHash, manufacturer, partNumber StringProperty, logger *log.Logger, debug bool) (ObitID, error) {
	var id ObitID

	snhHash := serialNumberHash.GetHash()
	mh := manufacturer.GetHash()
	pnh := partNumber.GetHash()

	sum := serialNumberHash.GetHash().GetDec() + manufacturer.GetHash().GetDec() + partNumber.GetHash().GetDec()

	if debug {
		logger.Printf(
			"\n<|%s|> => NewObitIDProperty(%v, %v, %v) -> (%d + %d + %d) -> %d",
			"Making ObitID",
			serialNumberHash,
			manufacturer,
			partNumber,
			snhHash.GetDec(),
			mh.GetDec(),
			pnh.GetDec(),
			sum,
		)
	}

	h, err := hash.NewHash([]byte(fmt.Sprintf("%x", sum)), logger, debug)

	if err != nil {
		return id, fmt.Errorf("cannot create obit id: %w", err)
	}

	hashStr := h.GetHash()

	id.hash = h
	id.did = "did:obada:" + hashStr

	fullUsn := base58.Encode([]byte(hashStr))

	id.usn = fullUsn[:8]
	id.fullUsn = fullUsn

	if debug {
		logger.Printf("Hash: %s", h.GetHash())
		logger.Printf("Did: %s", id.did)
		logger.Printf("Usn: %s", id.usn)
		logger.Printf("Full Usn: %s", id.fullUsn)
	}

	return id, nil
}

// GetHash returns ObitID hash
func (id *ObitID) GetHash() hash.Hash {
	return id.hash
}

// GetDid returns obit DID
func (id *ObitID) GetDid() string {
	return id.did
}

// GetUsn returns short universal serial number
func (id *ObitID) GetUsn() string {
	return id.usn
}

// GetFullUsn returns full universal serial number
func (id *ObitID) GetFullUsn() string {
	return id.fullUsn
}
