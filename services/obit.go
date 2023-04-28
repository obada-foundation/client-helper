package services

import (
	"bytes"
	"log"

	didsdk "github.com/obada-foundation/sdkgo/did"
	"go.uber.org/zap"
)

// ObitService holds service dependencies
type ObitService struct {
	logger *zap.SugaredLogger
}

// NewObitService creates new service
func NewObitService(logger *zap.SugaredLogger) *ObitService {
	return &ObitService{
		logger: logger,
	}
}

//nolint:all //needs to be refactored
type NFTPayload struct {
	SerialNumber     string `json:"serial_number"`
	Manufacturer     string `json:"manufacturer"`
	PartNumber       string `json:"part_number"`
	TrustAnchorToken string `json:"trust_anchor_token"`
}

//nolint:all //needs to be refactored
type LocalNFT struct {
	Usn              string `json:"usn"`
	DID              string `json:"did"`
	Owner            string `json:"owner"`
	Checksum         string `json:"checksum"`
	SerialNumber     string `json:"serial_number"`
	Manufacturer     string `json:"manufacturer"`
	PartNumber       string `json:"part_number"`
	TrustAnchorToken string `json:"trust_anchor_token"`
}

// GenerateObit generates Obit DID
func (svc *ObitService) GenerateObit(serialNumber, manufacturer, partNumber string, logger *log.Logger) (*didsdk.DID, error) {
	obit, err := didsdk.MakeDID(didsdk.NewDID{
		SerialNumber: serialNumber,
		Manufacturer: manufacturer,
		PartNumber:   partNumber,
		Logger:       logger,
	})

	return obit, err
}

//nolint:all //needs to be refactored
func (svc *ObitService) MakeLocalNFT(payload NFTPayload, catchLogs bool) (LocalNFT, bytes.Buffer, error) {
	var capturedLogs bytes.Buffer
	var err error
	var localNFT LocalNFT

	did, err := svc.GenerateObit(payload.SerialNumber, payload.Manufacturer, payload.PartNumber, nil)
	if err != nil {
		return localNFT, capturedLogs, err
	}

	localNFT = LocalNFT{
		Usn:          did.GetFullUSN(),
		DID:          did.String(),
		Checksum:     "",
		SerialNumber: payload.SerialNumber,
		Manufacturer: payload.Manufacturer,
		PartNumber:   payload.PartNumber,
	}

	return localNFT, capturedLogs, nil
}
