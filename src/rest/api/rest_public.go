package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/obada-foundation/client-helper/blockchain/client"
	"github.com/obada-foundation/client-helper/blockchain/wallet"
	"github.com/obada-foundation/client-helper/rest"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/utils"
	"go.uber.org/zap"
)

type public struct {
	obitService   *services.ObitService
	chainService  *client.ObadaChainClient
	walletService *wallet.WalletService
	logger        *zap.SugaredLogger
}

func (rpub *public) getObit(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	localNFT, err := rpub.obitService.Get(key)

	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, rpub.logger)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &localNFT)
}

// RequestObitDID request for generate Obit DID
type GenerateObitDIDReq struct {
	SerialNumber string `json:"serial_number"`
	Manufacturer string `json:"manufacturer"`
	PartNumber   string `json:"part_number"`
}

type GenerateObitDIDResp struct {
	SerialNumberHash string `json:"serial_number_hash"`
	USN              string `json:"usn"`
	DID              string `json:"did"`
	USNBase58        string `json:"usn_base58"`
}

func (rpub *public) generateObit(w http.ResponseWriter, r *http.Request) {
	var requestData GenerateObitDIDReq
	var resp GenerateObitDIDResp

	if err := render.DecodeJSON(http.MaxBytesReader(w, r.Body, hardBodyLimit), &requestData); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't decode request data", rest.ErrDecode, rpub.logger)
		return
	}

	obit, err := rpub.obitService.GenerateObit(
		requestData.SerialNumber,
		requestData.Manufacturer,
		requestData.PartNumber,
	)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, rpub.logger)
		return
	}

	resp.DID = obit.GetDid()
	resp.USN = obit.GetUsn()
	resp.USNBase58 = obit.GetFullUsn()
	resp.SerialNumberHash, _ = utils.HashStr(requestData.SerialNumber)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &resp)
}

func (rpub *public) saveObit(w http.ResponseWriter, r *http.Request) {
	var requestData services.NFTPayload

	if err := render.DecodeJSON(http.MaxBytesReader(w, r.Body, hardBodyLimit), &requestData); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't decode request data", rest.ErrDecode, rpub.logger)
		return
	}

	wallet := rpub.walletService.GetWallet("")

	localNFT, err := rpub.obitService.Save(requestData, wallet.GetObadaAddress())
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't save local NFT", rest.ErrDecode, rpub.logger)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &localNFT)
}

type GenerateObitChecksumResp struct {
	Checksum   string `json:"checksum"`
	ComputeLog string `json:"compute_log"`
}

func (rpub *public) generateChecksum(w http.ResponseWriter, r *http.Request) {
	var requestData services.NFTPayload
	var resp GenerateObitChecksumResp

	if err := render.DecodeJSON(http.MaxBytesReader(w, r.Body, hardBodyLimit), &requestData); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't decode request data", rest.ErrDecode, rpub.logger)
		return
	}

	localNFT, capturedLog, err := rpub.obitService.MakeLocalNFT(requestData, true)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, rpub.logger)
		return
	}

	resp.Checksum = localNFT.Checksum
	resp.ComputeLog = capturedLog.String()

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &resp)
}

type SearchResponse struct {
	Data []services.LocalNFT `json:"data"`
	Meta interface{}         `json:"meta"`
}

func (rpub *public) search(w http.ResponseWriter, r *http.Request) {
	var resp SearchResponse

	localNFTs, err := rpub.obitService.Search("")
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, rpub.logger)
		return
	}

	resp.Data = localNFTs

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &resp)
}

func (rpub *public) uploadToChain(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	localNFT, err := rpub.obitService.Get(key)

	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, rpub.logger)
		return
	}

	ctx := context.Background()

	wallet := rpub.walletService.CreateWallet()

	res, err := rpub.chainService.Mint(ctx, wallet.PrivateKey, localNFT)

	if err != nil {
		rpub.logger.Error(err)
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, rpub.logger)
		return
	}

	rpub.logger.Debug(res)

	render.Status(r, http.StatusNoContent)
}

func (rpub *public) downloadFromChain(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusNoContent)
}
