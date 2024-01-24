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
	regapi "github.com/obada-foundation/registry/api"
	pbacc "github.com/obada-foundation/registry/api/pb/v1/account"
	"github.com/obada-foundation/registry/api/pb/v1/diddoc"
	registry "github.com/obada-foundation/registry/client"
)

// Handlers holds dependencies
type Handlers struct {
	AccountSvc    *account.Service
	DeviceSvc     *device.Service
	BlockchainSvc *blockchain.Service
	Registry      registry.Client
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

// BatchMint mints a batch of NFTs
func (h Handlers) BatchMint(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req services.MintBatchNFT

	if err := web.Decode(r, &req); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	devices, err := h.DeviceSvc.GetByDIDs(ctx, req.Nfts)
	if err != nil {
		return err
	}

	privKey, err := h.AccountSvc.GetAccountPrivateKey(ctx, devices[0].Address)
	if err != nil {
		return err
	}

	if err := h.BlockchainSvc.BatchMintNFT(ctx, devices, privKey); err != nil {
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

	resp, err := h.Registry.GetPublicKey(ctx, &pbacc.GetPublicKeyRequest{
		Address: req.ReceiverArr,
	})
	if err != nil {
		return err
	}

	DIDDoc, err := h.Registry.Get(ctx, &diddoc.GetRequest{Did: d.DID})
	if err != nil {
		return err
	}

	vms := make([]*diddoc.VerificationMethod, 0)
	authId := fmt.Sprintf("%s#keys-1", d.DID)

	for _, doc := range DIDDoc.GetDocument().GetVerificationMethod() {
		if doc.GetId() == authId {
			doc.PublicKeyBase58 = resp.GetPubkey()
		}

		vms = append(vms, doc)
	}

	data := &diddoc.MsgSaveVerificationMethods_Data{
		Did:                 d.DID,
		AuthenticationKeyId: authId,
		Authentication:      DIDDoc.Document.Authentication,
		VerificationMethods: vms,
	}

	hash, err := regapi.ProtoDeterministicChecksum(data)
	if err != nil {
		return err
	}

	signature, err := privKey.Sign(hash[:])
	if err != nil {
		return err
	}

	_, err = h.Registry.SaveVerificationMethods(ctx, &diddoc.MsgSaveVerificationMethods{
		Data:      data,
		Signature: signature,
	})
	if err != nil {
		return err
	}

	if err := h.DeviceSvc.Delete(ctx, d.DID); err != nil {
		return fmt.Errorf("cannot delete device after transfer: %w", err)
	}

	return web.RespondWithNoContent(ctx, w, http.StatusCreated)
}
