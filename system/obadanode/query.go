package obadanode

import (
	"context"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	obadatypes "github.com/obada-foundation/fullcore/x/obit/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const baseDenom = "rohi"

// BaseDenomMetadata returns the metadata for the base denom
func (c NodeClient) BaseDenomMetadata(ctx context.Context) (banktypes.Metadata, error) {
	var md banktypes.Metadata

	denoms, err := c.bankClient.DenomsMetadata(ctx, &banktypes.QueryDenomsMetadataRequest{})
	if err != nil {
		return md, fmt.Errorf("cannot get denom metadata: %w", err)
	}

	if len(denoms.Metadatas) == 0 {
		return md, fmt.Errorf("no denom metadata registered")
	}

	for _, md = range denoms.Metadatas {
		if md.Base == baseDenom {
			return md, nil
		}
	}

	return md, fmt.Errorf("%q base denom metadata are not registered on chain", baseDenom)
}

// Balance implements the Balance method of the Client interface
func (c NodeClient) Balance(ctx context.Context, pubKey cryptotypes.PubKey) (*banktypes.QueryBalanceResponse, error) {
	addr := types.AccAddress(pubKey.Address())

	req := &banktypes.QueryBalanceRequest{
		Address: addr.String(),
		Denom:   "rohi",
	}
	res, err := c.bankClient.Balance(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil

}

// BalanceByAddress implements the BalanceByAddress method of the Client interface
func (c NodeClient) BalanceByAddress(ctx context.Context, address string) (*banktypes.QueryBalanceResponse, error) {
	req := &banktypes.QueryBalanceRequest{
		Address: address,
		Denom:   "rohi",
	}

	res, err := c.bankClient.Balance(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil

}

// GetNFTByAddress implements the GetNFTByAddress method of the Client interface
func (c NodeClient) GetNFTByAddress(ctx context.Context, address string) ([]obadatypes.NFT, error) {
	resp, err := c.obadaClient.GetNFTByAddress(ctx, &obadatypes.QueryGetNFTByAddressRequest{
		Address: address,
	})
	if err != nil {
		return nil, err
	}

	return resp.NFT, nil

}

// GetNFT implements the GetNFT method of the Client interface
func (c NodeClient) GetNFT(ctx context.Context, did string) (*obadatypes.NFT, error) {
	resp, err := c.obadaClient.GetNFT(ctx, &obadatypes.QueryGetNFTRequest{
		Id: did,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil

}

// HasAccount returns true if the account exists on blockchain (has transactions)
func (c NodeClient) HasAccount(ctx context.Context, address string) (bool, error) {
	if _, err := c.Account(ctx, address); err != nil {
		if err == ErrAccountHasZeroTx {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// Account returns the account details for a gived address
func (c NodeClient) Account(ctx context.Context, address string) (acc authtypes.AccountI, err error) {
	req := &authtypes.QueryAccountRequest{Address: address}

	res, err := c.authClient.Account(ctx, req)
	if err != nil {
		statusError, ok := status.FromError(err)
		if !ok || statusError.Code() != codes.NotFound {
			return
		}

		err = ErrAccountHasZeroTx

		return
	}

	if err = c.cdc.UnpackAny(res.GetAccount(), &acc); err != nil {
		return
	}

	return
}
