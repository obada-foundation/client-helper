package nft

import (
	"context"

	codestypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/obada-foundation/client-helper/services"
	node "github.com/obada-foundation/client-helper/system/obadanode"
	"github.com/obada-foundation/fullcore/x/obit/types"
	"go.uber.org/zap"
)

type Service struct {
	nodeClient *node.NodeClient
	logger     *zap.SugaredLogger
}

func NewService(client *node.NodeClient, logger *zap.SugaredLogger) *Service {
	return &Service{
		nodeClient: client,
		logger:     logger,
	}
}

type nftAnyResolver struct{}

func (m *nftAnyResolver) Resolve(typeURL string) (proto.Message, error) {
	return new(types.NFTData), nil
}

func NFTtoJSON(nft *types.NFT) (string, error) {
	m := jsonpb.Marshaler{
		AnyResolver: &nftAnyResolver{},
	}

	return m.MarshalToString(nft)
}

func AnyToNFTData(any *codestypes.Any) {

}

func (ns Service) NFT(ctx context.Context, DID string) (*types.NFT, error) {
	nft, err := ns.nodeClient.GetNFT(ctx, DID)
	if err != nil {
		return nil, err
	}

	return nft, nil
}

func (ns Service) MintGasEstimate(ctx context.Context, d services.Device, addreess string) error {
	msg := ns.buildMintMsg(d, addreess)

	resp, _, err := ns.nodeClient.CalculateGas(ctx, msg)
	ns.logger.Info("NFT was minted", resp)

	return err
}

func (ns Service) buildMintMsg(d services.Device, address string) *types.MsgMintObit {
	var docs []types.NFTDocument

	for _, d := range d.Documents {
		docs = append(docs, types.NFTDocument{
			Name: d.Name,
			Uri:  d.URI,
			Hash: d.Hash,
		})
	}

	return &types.MsgMintObit{
		Creator:          address,
		SerialNumberHash: d.SerialNumberHash,
		Manufacturer:     d.Manufacturer,
		PartNumber:       d.PartNumber,
		Documents:        docs,
	}
}

func (ns Service) EditMetadata(ctx context.Context, d services.Device, privKey secp256k1.PrivKey) error {
	accAddress := sdk.AccAddress(privKey.PubKey().Address().Bytes()).String()

	nft, err := ns.NFT(ctx, d.DID)
	if err != nil {
		return err
	}

	nftData := &types.NFTData{}

	proto.Unmarshal(nft.Data.GetValue(), nftData)

	var docs []types.NFTDocument

	for _, d := range d.Documents {
		docs = append(docs, types.NFTDocument{
			Name: d.Name,
			Uri:  d.URI,
			Hash: d.Hash,
		})
	}

	nftData.Documents = docs

	msg := &types.MsgEditMetadata{
		Did:     nft.Id,
		Editor:  accAddress,
		NFTData: nftData,
	}

	resp, err := ns.nodeClient.SendTx(ctx, msg, &privKey)
	ns.logger.Info("NFT metadata was updated", resp)

	return err
}

func (ns Service) Mint(ctx context.Context, d services.Device, privKey secp256k1.PrivKey) error {
	var docs []types.NFTDocument

	accAddress := sdk.AccAddress(privKey.PubKey().Address().Bytes()).String()

	for _, d := range d.Documents {
		docs = append(docs, types.NFTDocument{
			Name: d.Name,
			Uri:  d.URI,
			Hash: d.Hash,
		})
	}

	msg := ns.buildMintMsg(d, accAddress)

	resp, err := ns.nodeClient.SendTx(ctx, msg, &privKey)
	ns.logger.Info("NFT was minted", resp)

	return err
}

func (ns Service) Send(ctx context.Context, DID, receiverAddr string, privKey secp256k1.PrivKey) error {
	accAddr := sdk.AccAddress(privKey.PubKey().Address().Bytes()).String()

	msg := &types.MsgSend{
		Did:      DID,
		Sender:   accAddr,
		Receiver: receiverAddr,
	}

	resp, err := ns.nodeClient.SendTx(ctx, msg, &privKey)
	ns.logger.Info("NFT was transfered", msg, resp)

	return err
}
