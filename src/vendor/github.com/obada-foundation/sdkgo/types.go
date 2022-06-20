package sdkgo

import (
	"github.com/obada-foundation/sdkgo/properties"
	"log"
)

// ObitIDDto todo add description
type ObitIDDto struct {
	SerialNumberHash string `validate:"required"`
	Manufacturer     string `validate:"required"`
	PartNumber       string `validate:"required"`
}

// ObitDto todo add description
type ObitDto struct {
	ObitIDDto
	TrustAnchorToken string
}

// Obit represent asset data structure
type Obit struct {
	obitID           properties.ObitID
	serialNumberHash properties.StringProperty
	manufacturer     properties.StringProperty
	partNumber       properties.StringProperty
	trustAnchorToken properties.StringProperty
	documents        properties.Documents
	debug            bool
	logger           *log.Logger
}
