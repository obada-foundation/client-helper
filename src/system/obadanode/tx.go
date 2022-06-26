package obadanode

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	obadatypes "github.com/obada-foundation/fullcore/x/obit/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func (c *NodeClient) Mint(ctx context.Context, priv cryptotypes.PrivKey, msg *obadatypes.MsgMintObit) (*ctypes.ResultBroadcastTx, error) {
	accAddress := sdk.AccAddress(priv.PubKey().Address().Bytes()).String()
	nonce, err := c.Nonce(ctx, accAddress)

	if err != nil {
		return nil, err
	}

	res, err := c.SendTx(ctx, msg, priv, nonce)
	if err != nil {
		return nil, err
	}

	return res, nil
}
