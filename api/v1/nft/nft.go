package nft

import (
	"context"
	"fmt"
	"net/http"

	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/blockchain"
	"github.com/obada-foundation/client-helper/services/device"
	"github.com/obada-foundation/client-helper/system/web"
)

// Handlers holds dependencies
type Handlers struct {
	AccountSvc    *account.Service
	DeviceSvc     *device.Service
	BlockchainSvc *blockchain.Service
}

// NFT reponds NFT
func (h Handlers) NFT(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	key := web.Param(r, "key")

	d, err := h.DeviceSvc.Get(ctx, key)
	if err != nil {
		return err
	}

	nft, err := h.BlockchainSvc.GetNFT(ctx, d.DID)
	if err != nil {
		return err
	}

	data, err := blockchain.NFTtoJSON(nft)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, data, http.StatusOK)
}

// Mint creates a new NFT
func (h Handlers) Mint(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	key := web.Param(r, "key")

	d, err := h.DeviceSvc.Get(ctx, key)
	if err != nil {
		return err
	}

	privKey, err := h.AccountSvc.GetAccountPrivateKey(ctx, d.Address)
	if err != nil {
		return err
	}

	if err := h.BlockchainSvc.MintNFT(ctx, d, privKey); err != nil {
		return err
	}

	return web.RespondWithNoContent(ctx, w, http.StatusCreated)
}

// UpdateMetadata updates NFT metadata
func (h Handlers) UpdateMetadata(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	key := web.Param(r, "key")

	d, err := h.DeviceSvc.Get(ctx, key)
	if err != nil {
		return err
	}

	if _, er := h.BlockchainSvc.GetNFT(ctx, d.DID); er != nil {
		return er
	}

	privKey, err := h.AccountSvc.GetAccountPrivateKey(ctx, d.Address)
	if err != nil {
		return err
	}

	if er := h.BlockchainSvc.EditNFTMetadata(ctx, d, privKey); er != nil {
		return er
	}

	return web.RespondWithNoContent(ctx, w, http.StatusOK)
}

// Transfer transfers NFT
func (h Handlers) Transfer(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req services.SendNFT

	key := web.Param(r, "key")

	if err := web.Decode(r, &req); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	d, err := h.DeviceSvc.Get(ctx, key)
	if err != nil {
		return err
	}

	privKey, err := h.AccountSvc.GetAccountPrivateKey(ctx, d.Address)
	if err != nil {
		return err
	}

	if err := h.BlockchainSvc.TransferNFT(ctx, d.DID, req.ReceiverArr, privKey); err != nil {
		return err
	}

	if err := h.DeviceSvc.Delete(ctx, d.DID); err != nil {
		return err
	}

	return web.RespondWithNoContent(ctx, w, http.StatusCreated)
}
