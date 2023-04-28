package blockchain

import (
	"context"
	"errors"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/obadanode"
)

// Send sends coins from one account to another.
func (bs Service) Send(ctx context.Context, account services.Account, toAddress, amount string, privKey cryptotypes.PrivKey) error {
	fromAddress, err := sdk.AccAddressFromBech32(account.Address)
	if err != nil {
		return err
	}

	recepientAddress, err := sdk.AccAddressFromBech32(toAddress)
	if err != nil {
		return err
	}

	ok, err := bs.nodeClient.HasAccount(ctx, account.Address)
	if err != nil {
		return err
	}

	if !ok {
		return ErrInsufficientFunds
	}

	coins, err := sdk.ParseCoinsNormalized(amount)
	if err != nil {
		return err
	}

	msg := types.NewMsgSend(fromAddress, recepientAddress, coins)

	resp, err := bs.nodeClient.SendTx(ctx, msg, privKey)
	if err != nil {
		if errors.Is(err, obadanode.ErrInsufficientFunds) {
			return ErrInsufficientFunds
		}

		return err
	}

	bs.logger.Info("Coins were transferred", msg, resp)

	return nil
}
