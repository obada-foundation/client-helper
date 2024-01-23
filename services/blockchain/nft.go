package blockchain

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck // wait for refactoring
	"github.com/golang/protobuf/proto"  //nolint:staticcheck // wait for refactoring
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/obadanode"
	"github.com/obada-foundation/fullcore/x/obit/types"
)

type nftAnyResolver struct{}

// Resolve implements jsonpb.AnyResolver interface.
func (m *nftAnyResolver) Resolve(_ string) (proto.Message, error) {
	return new(types.NFTData), nil
}

// NFTtoJSON converts NFT to JSON.
func NFTtoJSON(nft *types.NFT) (string, error) {
	m := jsonpb.Marshaler{
		AnyResolver: &nftAnyResolver{},
	}

	return m.MarshalToString(nft)
}

// JSONtoNFT converts JSON to NFT.
func JSONtoNFT(json string) (*types.NFT, error) {
	nft := new(types.NFT)

	um := jsonpb.Unmarshaler{
		AnyResolver: &nftAnyResolver{},
	}

	if err := um.Unmarshal(strings.NewReader(json), nft); err != nil {
		return nil, err
	}

	return nft, nil
}

// GetNFT returns NFT by given DID.
func (bs Service) GetNFT(ctx context.Context, did string) (*types.NFT, error) {
	nft, err := bs.nodeClient.GetNFT(ctx, did)
	if err != nil {
		return nil, err
	}

	return nft, nil
}

// GetNFTByAddress returns NFTs by given address.
func (bs Service) GetNFTByAddress(ctx context.Context, address string) ([]types.NFT, error) {
	nfts, err := bs.nodeClient.GetNFTByAddress(ctx, address)
	if err != nil {
		return nfts, err
	}

	return nfts, nil
}

// MintGasEstimate returns gas estimate for minting NFT.
func (bs Service) MintGasEstimate(ctx context.Context, d services.Device, addreess string) error {
	msg := bs.buildMintMsg(d, addreess)

	resp, _, err := bs.nodeClient.CalculateGas(ctx, msg)
	bs.logger.Info("Gas estimation for minting", resp)

	return err
}

func (bs Service) buildMintMsg(d services.Device, address string) *types.MsgMintNFT {

	URI := fmt.Sprintf("%s/api/v1.0/diddoc/%s", bs.registryURL, d.DID)

	return &types.MsgMintNFT{
		Creator: address,
		Id:      d.DID,
		Uri:     URI,
		Usn:     d.Usn,
		UriHash: d.Checksum,
	}
}

// EditNFTMetadata edits NFT metadata.
func (bs Service) EditNFTMetadata(ctx context.Context, d services.Device, privKey cryptotypes.PrivKey) error {
	accAddress := sdk.AccAddress(privKey.PubKey().Address().Bytes()).String()

	nft, err := bs.GetNFT(ctx, d.DID)
	if err != nil {
		return err
	}

	nftData := &types.NFTData{}

	if er := proto.Unmarshal(nft.Data.GetValue(), nftData); er != nil {
		return er
	}

	msg := &types.MsgUpdateUriHash{
		Id:      nft.Id,
		Editor:  accAddress,
		UriHash: d.Checksum,
	}

	resp, err := bs.nodeClient.SendTx(ctx, msg, privKey)
	if err != nil {
		if errors.Is(err, obadanode.ErrInsufficientFunds) {
			return ErrInsufficientFunds
		}

		return err
	}
	bs.logger.Info("NFT metadata was updated", resp)

	return err
}

// MintNFT creates new NFT.
func (bs Service) MintNFT(ctx context.Context, d services.Device, privKey cryptotypes.PrivKey) error {
	accAddress := sdk.AccAddress(privKey.PubKey().Address().Bytes()).String()

	ok, err := bs.nodeClient.HasAccount(ctx, accAddress)
	if err != nil {
		return err
	}

	if !ok {
		return ErrInsufficientFunds
	}

	msg := bs.buildMintMsg(d, accAddress)

	resp, err := bs.nodeClient.SendTx(ctx, msg, privKey)
	if err != nil {
		if errors.Is(err, obadanode.ErrInsufficientFunds) {
			return ErrInsufficientFunds
		}

		return err
	}
	bs.logger.Info("NFT was minted", resp)

	return nil
}

// TransferNFT transfers NFT to another address.
func (bs Service) TransferNFT(ctx context.Context, did, receiverAddr string, privKey cryptotypes.PrivKey) error {
	accAddress := sdk.AccAddress(privKey.PubKey().Address().Bytes()).String()

	ok, err := bs.nodeClient.HasAccount(ctx, accAddress)
	if err != nil {
		return err
	}

	if !ok {
		return ErrInsufficientFunds
	}

	msg := &types.MsgTransferNFT{
		Id:       did,
		Sender:   accAddress,
		Receiver: receiverAddr,
	}

	resp, err := bs.nodeClient.SendTx(ctx, msg, privKey)
	if err != nil {
		if errors.Is(err, obadanode.ErrInsufficientFunds) {
			return ErrInsufficientFunds
		}

		return err
	}

	bs.logger.Info("NFT transfer request was sent", msg, resp)

	return nil
}
