package api

import (
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/obada-foundation/client-helper/rest"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/device"
	nftSvc "github.com/obada-foundation/client-helper/services/nft"
	"go.uber.org/zap"
)

type nftGroup struct {
	logger *zap.SugaredLogger

	deviceSvc  *device.Service
	accountSvc *account.Service
	nftSvc     *nftSvc.Service
}

func (ng *nftGroup) nft(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	ctx := r.Context()

	device, err := ng.deviceSvc.Get(ctx, key)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, ng.logger)
		return
	}

	nft, err := ng.nftSvc.NFT(ctx, device.DID)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, ng.logger)
		return
	}

	data, err := nftSvc.NFTtoJSON(nft)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, ng.logger)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, data)
}

func (ng *nftGroup) transfer(w http.ResponseWriter, r *http.Request) {
	var req services.SendNFT

	key := chi.URLParam(r, "key")

	ctx := r.Context()

	if err := render.DecodeJSON(http.MaxBytesReader(w, r.Body, hardBodyLimit), &req); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't decode request data", rest.ErrDecode, ng.logger)
		return
	}

	wallet, err := ng.accountSvc.Wallet(ctx)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, ng.logger)
		return
	}

	privKey := wallet.PrivateKey

	device, err := ng.deviceSvc.Get(ctx, key)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, ng.logger)
		return
	}

	if err := ng.nftSvc.Send(ctx, device.DID, req.ReceiverArr, privKey); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, ng.logger)
		return
	}

	render.Status(r, http.StatusCreated)
}

func (ng *nftGroup) mintGasEstimate(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	ctx := r.Context()

	wallet, err := ng.accountSvc.Wallet(ctx)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, ng.logger)
		return
	}

	privKey := wallet.PrivateKey

	device, err := ng.deviceSvc.Get(ctx, key)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, ng.logger)
		return
	}

	address := sdk.AccAddress(privKey.PubKey().Address().Bytes()).String()

	if err := ng.nftSvc.MintGasEstimate(ctx, device, address); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, ng.logger)
		return
	}

	render.Status(r, http.StatusCreated)
}

func (ng *nftGroup) updateMetadata(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	ctx := r.Context()

	wallet, err := ng.accountSvc.Wallet(ctx)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, ng.logger)
		return
	}

	privKey := wallet.PrivateKey

	device, err := ng.deviceSvc.Get(ctx, key)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, ng.logger)
		return
	}

	ng.logger.Debug("found device", device)

	nft, err := ng.nftSvc.NFT(ctx, device.DID)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, ng.logger)
		return
	}

	ng.logger.Debug("found nft", nft)

	if err := ng.nftSvc.EditMetadata(ctx, device, privKey); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, ng.logger)
		return
	}

	render.Status(r, http.StatusOK)
}

func (ng *nftGroup) mint(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	ctx := r.Context()

	wallet, err := ng.accountSvc.Wallet(ctx)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, ng.logger)
		return
	}

	privKey := wallet.PrivateKey

	device, err := ng.deviceSvc.Get(ctx, key)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, ng.logger)
		return
	}

	if err := ng.nftSvc.Mint(ctx, device, privKey); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, ng.logger)
		return
	}

	render.Status(r, http.StatusCreated)
}
