package account

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/go-bip39"
	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/events"
	svcs "github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/encoder"
)

// ImportWallet imports HD wallet and fetch existing accounts from the blockchain
func (as Service) ImportWallet(ctx context.Context, mnemonic string, force bool) error {
	if !bip39.IsMnemonicValid(mnemonic) {
		return ErrInvalidMnemonic
	}

	if _, err := as.NewWallet(ctx, mnemonic, force); err != nil {
		if er := as.deleteWallet(ctx); er != nil {
			return fmt.Errorf("%s : %w", er.Error(), err)
		}

		return err
	}

	// Create accounts until not receive ErrAccountHasZeroTx
	for {
		if _, err := as.NewAccount(ctx, Account{}); err != nil {
			if !errors.Is(err, ErrAccountHasZeroTx) {
				if er := as.deleteWallet(ctx); er != nil {
					return fmt.Errorf("%s : %w", er.Error(), err)
				}

				return err
			}

			break
		}
	}

	return nil

}

// NewWallet created new HD wallet attached to the user account
func (as Service) NewWallet(ctx context.Context, mnemonic string, force bool) (*svcs.Wallet, error) {
	profileID := auth.GetClaims(ctx).UserID

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, ErrInvalidMnemonic
	}

	wk := walletKey(profileID)

	hasWallet, err := as.db.Has(wk)
	if err != nil {
		return nil, err
	}

	if hasWallet && !force {
		return nil, ErrWalletExists
	}

	if hasWallet {
		if er := as.deleteWallet(ctx); er != nil {
			return nil, fmt.Errorf("cannot delete a wallet: %w", er)
		}
	}

	wallet := svcs.Wallet{
		Mnemonic:     mnemonic,
		AccountIndex: 0,
	}

	walletBytes, err := encoder.DataEncode(wallet)
	if err != nil {
		return nil, err
	}

	if err := as.db.Set(wk, walletBytes); err != nil {
		return nil, err
	}

	if _, err := as.NewAccount(ctx, Account{}); err != nil {
		return nil, err
	}

	return &wallet, nil
}

// GetWallet returns a wallet attached to the profile
func (as Service) GetWallet(ctx context.Context) (svcs.Wallet, error) {
	var wallet svcs.Wallet

	profileID := auth.GetClaims(ctx).UserID

	wk := walletKey(profileID)

	hasWallet, err := as.db.Has(wk)
	if err != nil {
		return wallet, err
	}

	if !hasWallet {
		return wallet, ErrWalletNotExists
	}

	walletBytes, err := as.db.Get(wk)
	if err != nil {
		return wallet, err
	}

	b := bytes.NewBuffer(walletBytes)
	dec := gob.NewDecoder(b)

	if er := dec.Decode(&wallet); er != nil {
		return wallet, er
	}

	return wallet, nil
}

func (as Service) deleteWallet(ctx context.Context) error {
	profileID := auth.GetUserID(ctx)
	batch := as.db.NewBatch()
	defer batch.Close()

	wallet, err := as.GetWallet(ctx)
	if err != nil {
		return err
	}

	for {
		obadaAccount := keyringAccountKey(profileID, wallet.AccountIndex)

		account, err := as.keyring.Key(obadaAccount)
		if err != nil {
			// When we import wallet and then on account discovery we find that such account already exists
			// key with 0 index be not created so error not found should happen
			if !errors.Is(err, sdkerrors.ErrKeyNotFound) {
				return err
			}
		}

		if account != nil {
			if err := as.keyring.Delete(obadaAccount); err != nil {
				return err
			}

			addr, err := account.GetAddress()
			if err != nil {
				return err
			}

			if err := batch.Delete(accountHDKey(profileID, addr.String())); err != nil {
				return err
			}

			if err := as.eventBus.Emit(ctx, events.AccountDeleted, addr.String()); err != nil {
				return err
			}
		}

		if wallet.AccountIndex > 0 {
			wallet.AccountIndex--
			continue
		}

		break
	}

	if er := batch.Delete(walletKey(profileID)); er != nil {
		return er
	}

	return batch.WriteSync()
}

// GetWalletAccountIndex returns the wallet account index
func (as Service) GetWalletAccountIndex(ctx context.Context) (uint, error) {
	wallet, err := as.GetWallet(ctx)
	if err != nil {
		return 0, err
	}

	return wallet.AccountIndex, nil
}
