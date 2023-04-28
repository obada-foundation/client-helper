package account

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mustafaturan/bus/v3"
	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/events"
	svcs "github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/encoder"
	"github.com/obada-foundation/client-helper/system/obadanode"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/tendermint/tm-db"
)

// Service is the account service
type Service struct {
	validator  *validate.Validator
	db         db.DB
	nodeClient obadanode.Client
	keyring    keyring.Keyring
	eventBus   *bus.Bus
}

// Account contain fields that represent account
type Account struct {
	Name string `json:"name"`
}

// NewService creates new account service
func NewService(v *validate.Validator, database db.DB, c obadanode.Client, k keyring.Keyring, eb *bus.Bus) *Service {
	return &Service{
		validator:  v,
		db:         database,
		nodeClient: c,
		keyring:    k,
		eventBus:   eb,
	}
}

// GetImportedAccountIndex returns the imported account index
func (as Service) GetImportedAccountIndex(ctx context.Context) (uint, error) {
	var index uint

	profileID := auth.GetClaims(ctx).UserID

	indexBytes, err := as.db.Get(accountImportedIdx(profileID))
	if err != nil {
		return index, err
	}

	b := bytes.NewBuffer(indexBytes)
	dec := gob.NewDecoder(b)

	if er := dec.Decode(&index); er != nil {
		return index, er
	}

	return index, nil
}

//nolint:all // potentially can be removed
func (as Service) incrementImportedAccountIndex(ctx context.Context) error {
	profileID := auth.GetClaims(ctx).UserID

	index, err := as.GetImportedAccountIndex(ctx)
	if err != nil {
		return err
	}

	// Increment the index
	b, err := encoder.DataEncode(index + 1)
	if err != nil {
		return err
	}

	if err := as.db.SetSync(accountImportedIdx(profileID), b); err != nil {
		return err
	}

	return nil
}

// ImportAccount imports an account to the keying
func (as Service) ImportAccount(ctx context.Context, privateKey, passphrase string, acc Account) error {
	profileID := auth.GetClaims(ctx).UserID

	idx, err := as.GetImportedAccountIndex(ctx)
	if err != nil {
		return err
	}

	newIdx := idx + 1

	importKey := keyringAccountImportedKey(profileID, newIdx)

	if er := as.keyring.ImportPrivKey(importKey, privateKey, passphrase); er != nil {
		return fmt.Errorf("cannot import private key to keyring: %w", er)
	}

	krAccount, err := as.keyring.Key(importKey)
	if err != nil {
		return fmt.Errorf("cannot fetch account by key %q: %w", importKey, err)
	}

	accAddress, err := krAccount.GetAddress()
	if err != nil {
		return err
	}

	batch := as.db.NewBatch()
	defer batch.Close()

	accountBytes, err := encoder.DataEncode(acc)
	if err != nil {
		return err
	}

	if er := batch.Set(accountImportedKey(profileID, accAddress.String()), accountBytes); er != nil {
		return er
	}

	// Increment the index
	idxBytes, err := encoder.DataEncode(newIdx)
	if err != nil {
		return err
	}

	if err := batch.Set(accountImportedIdx(profileID), idxBytes); err != nil {
		return err
	}

	if err := batch.WriteSync(); err != nil {
		if er := as.keyring.Delete(importKey); er != nil {
			return fmt.Errorf("%s : %w", er.Error(), err)
		}

		return err
	}

	return as.eventBus.Emit(ctx, events.AccountCreated, accAddress.String())
}

// HasAccount idenfify that account can be used by given context
func (as Service) HasAccount(ctx context.Context, address string) bool {
	profileID := auth.GetUserID(ctx)

	accountOwnnerID, err := as.GetProfileByAddress(address)
	if err != nil {
		return false
	}

	return accountOwnnerID == profileID
}

// DeleteAccount deletes an imported account
func (as Service) DeleteAccount(ctx context.Context, address string) error {
	profileID := auth.GetUserID(ctx)

	accountAddress, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	key, err := as.keyring.KeyByAddress(accountAddress)
	if err != nil {
		return err
	}

	if !strings.Contains(key.Name, "imported") {
		return ErrHDAccountDelete
	}

	if err := as.db.Delete(accountImportedKey(profileID, address)); err != nil {
		return err
	}

	if err := as.keyring.DeleteByAddress(accountAddress); err != nil {
		return err
	}

	return as.eventBus.Emit(ctx, events.AccountDeleted, address)
}

// ExportAccount expots an account
func (as Service) ExportAccount(ctx context.Context, address, passphrase string) (string, error) {
	profileID := auth.GetClaims(ctx).UserID

	accountAddress, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return "", err
	}

	keyInfo, err := as.keyring.KeyByAddress(accountAddress)
	if err != nil {
		return "", err
	}

	if !strings.Contains(keyInfo.Name, keyringAccountPrefix(profileID)) {
		return "", ErrAccountNotExists
	}

	return as.keyring.ExportPrivKeyArmor(keyInfo.Name, passphrase)
}

// UpdateAccountName updates the account name
func (as Service) UpdateAccountName(ctx context.Context, address, newAccountName string) error {
	profileID := auth.GetUserID(ctx)

	keys := [][]byte{
		accountHDKey(profileID, ""),
		accountImportedKey(profileID, ""),
	}

	for _, key := range keys {
		prefixDB := db.NewPrefixDB(as.db, key)
		itr, err := prefixDB.Iterator(nil, nil)
		if err != nil {
			return err
		}

		for ; itr.Valid(); itr.Next() {
			key := itr.Key()

			if strings.Contains(string(key), address) {
				account := Account{}

				b := bytes.NewBuffer(itr.Value())
				dec := gob.NewDecoder(b)

				if err := dec.Decode(&account); err != nil {
					return err
				}

				account.Name = newAccountName

				accountBytes, err := encoder.DataEncode(account)
				if err != nil {
					return err
				}

				return prefixDB.Set(key, accountBytes)
			}
		}

		_ = itr.Close()
	}

	return ErrAccountNotExists

}

// NewAccount creates a new OBADA account from HD wallet
func (as Service) NewAccount(ctx context.Context, acc Account) (svcs.Account, error) {
	var account svcs.Account

	profileID := auth.GetClaims(ctx).UserID

	wallet, err := as.GetWallet(ctx)
	if err != nil {
		return account, err
	}

	hasAccounts := false

	prefixDB := db.NewPrefixDB(as.db, accountHDKey(profileID, ""))
	itr, err := prefixDB.Iterator(nil, nil)
	if err != nil {
		return account, err
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		hasAccounts = true
		break
	}

	// When we have existing accounts we first check that last account has a zero transactions
	if hasAccounts {
		accounts, er := as.getHDWalletAccounts(ctx)
		if er != nil {
			return account, er
		}

		lastAccount := accounts[wallet.AccountIndex]

		ok, er := as.nodeClient.HasAccount(ctx, lastAccount.Address)
		if er != nil {
			return account, er
		}

		if !ok {
			return account, ErrAccountHasZeroTx
		}

		wallet.AccountIndex++
	}

	obadaAccount := keyringAccountKey(profileID, wallet.AccountIndex)

	hdPath := hd.CreateHDPath(118, uint32(wallet.AccountIndex), 0).String()

	keyringAccount, err := as.keyring.NewAccount(obadaAccount, wallet.Mnemonic, "", hdPath, hd.Secp256k1)
	if err != nil {
		if strings.Contains(err.Error(), "duplicated address created") {
			return account, ErrAccountExists
		}
		return account, fmt.Errorf("cannot create keyring account: %w", err)
	}

	batch := as.db.NewBatch()
	defer batch.Close()

	accountBytes, err := encoder.DataEncode(acc)
	if err != nil {
		return account, err
	}

	addr, err := keyringAccount.GetAddress()
	if err != nil {
		return account, err
	}

	if er := batch.Set(accountHDKey(profileID, addr.String()), accountBytes); er != nil {
		return account, er
	}

	walletBytes, err := encoder.DataEncode(wallet)
	if err != nil {
		return account, err
	}

	if er := batch.Set(walletKey(profileID), walletBytes); er != nil {
		return account, er
	}

	if e := batch.WriteSync(); e != nil {
		if er := as.keyring.Delete(obadaAccount); er != nil {
			return account, fmt.Errorf("%s : %w", e.Error(), er)
		}

		return account, e
	}

	account, err = as.Keyring2Account(ctx, keyringAccount)
	if err != nil {
		return account, err
	}

	if err := as.eventBus.Emit(ctx, events.AccountCreated, addr.String()); err != nil {
		return account, err
	}

	return account, nil
}

// BalanceByAddress returns the balance of an account
func (as Service) BalanceByAddress(ctx context.Context, address string) (svcs.Balance, error) {
	var balance svcs.Balance

	nodeBalance, err := as.nodeClient.BalanceByAddress(ctx, address)
	if err != nil {
		return balance, err
	}

	decCoin := sdk.NewDecCoinFromCoin(*nodeBalance.Balance)

	obdBalance, err := sdk.ConvertDecCoin(decCoin, "obd")
	if err != nil {
		return balance, err
	}

	return svcs.Balance{
		Address: address,
		Balance: obdBalance,
	}, nil
}

// Wallet rreturns profile wallet
func (as Service) Wallet(ctx context.Context) (svcs.Wallet, error) {
	var wallet svcs.Wallet

	profileID := auth.GetClaims(ctx).UserID

	waKey := walletKey(profileID)

	hasWallet, err := as.db.Has(waKey)
	if err != nil {
		return wallet, err
	}

	if !hasWallet {
		return wallet, ErrProfileNotExists
	}

	walletBytes, err := as.db.Get(waKey)
	if err != nil {
		return wallet, err
	}

	b := bytes.NewBuffer(walletBytes)
	dec := gob.NewDecoder(b)

	if err := dec.Decode(&wallet); err != nil {
		return wallet, err
	}

	return wallet, nil
}
