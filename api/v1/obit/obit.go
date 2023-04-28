package obit

import (
	"context"
	"fmt"
	"net/http"

	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/web"
	"github.com/obada-foundation/client-helper/utils"
)

// Handlers holds dependencies
type Handlers struct {
	ObitSvc *services.ObitService
}

// GenerateObitDIDReq request data for GenerateObitDID
type GenerateObitDIDReq struct {
	SerialNumber string `json:"serial_number"`
	Manufacturer string `json:"manufacturer"`
	PartNumber   string `json:"part_number"`
}

// GenerateObitDIDResp response for generate Obit DID
type GenerateObitDIDResp struct {
	SerialNumberHash string `json:"serial_number_hash"`
	USN              string `json:"usn"`
	DID              string `json:"did"`
	USNBase58        string `json:"usn_base58"`
}

//nolint:all needs to be refactored
func (h Handlers) GenerateObit(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var requestData GenerateObitDIDReq
	var resp GenerateObitDIDResp

	if err := web.Decode(r, &requestData); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	obit, err := h.ObitSvc.GenerateObit(
		requestData.SerialNumber,
		requestData.Manufacturer,
		requestData.PartNumber,
		nil,
	)

	if err != nil {
		return err
	}

	resp.DID = obit.String()
	resp.USN = obit.GetUSN()
	resp.USNBase58 = obit.GetFullUSN()
	resp.SerialNumberHash, _ = utils.HashStr(requestData.SerialNumber)

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// GenerateObitChecksumResp response for generate Obit DID
type GenerateObitChecksumResp struct {
	Checksum   string `json:"checksum"`
	ComputeLog string `json:"compute_log"`
}

// GenerateChecksum generate checksum for Obit DID
func (h Handlers) GenerateChecksum(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var requestData services.NFTPayload
	var resp GenerateObitChecksumResp

	if err := web.Decode(r, &requestData); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	localNFT, capturedLog, err := h.ObitSvc.MakeLocalNFT(requestData, true)
	if err != nil {
		return err
	}

	resp.Checksum = localNFT.Checksum
	resp.ComputeLog = capturedLog.String()

	return web.Respond(ctx, w, resp, http.StatusOK)
}
