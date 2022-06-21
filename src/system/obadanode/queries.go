package obadanode

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (c NodeClient) Balance(ctx context.Context, pubKey cryptotypes.PubKey) (*banktypes.QueryBalanceResponse, error) {
	req := &banktypes.QueryBalanceRequest{
		Address: types.AccAddress(pubKey.Address().Bytes()).String(),
		Denom:   "OBD",
	}

	res, err := c.bankClient.Balance(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil

}
