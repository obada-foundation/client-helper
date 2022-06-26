package obadanode

import (
	"context"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	obadatypes "github.com/obada-foundation/fullcore/x/obit/types"
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

func (c NodeClient) GetNFT(ctx context.Context, DID string) (*obadatypes.NFT, error) {
	resp, err := c.obadaClient.GetNft(ctx, &obadatypes.QueryGetNftRequest{
		Did: DID,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil

}

func (c NodeClient) Account(ctx context.Context, address string) (acc authtypes.AccountI, err error) {
	req := &authtypes.QueryAccountRequest{Address: address}

	res, err := c.authClient.Account(ctx, req)
	if err != nil {

		fmt.Printf("cannot get account information: %s", err.Error())
		return
	}

	if err = c.cdc.UnpackAny(res.GetAccount(), &acc); err != nil {
		return
	}

	return
}
