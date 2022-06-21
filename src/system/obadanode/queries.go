package obadanode

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (c NodeClient) Balance(ctx context.Context, pubKey cryptotypes.PubKey) (*banktypes.QueryBalanceResponse, error) {
	addr, err := types.AccAddressFromHex(pubKey.Address().String())
	if err != nil {
		return nil, err
	}

	req := &banktypes.QueryBalanceRequest{
		Address: addr.String(),
		Denom:   "obd",
	}

	res, err := c.bankClient.Balance(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil

}
