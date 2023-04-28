package account

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotype "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/obada-foundation/client-helper/auth"
	svcs "github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/encoder"
	"github.com/pkg/errors"
	"github.com/tendermint/tm-db"
)

func (as Service) createNewProfile(profileID string, np svcs.NewProfile) (svcs.Profile, error) { //nolint:gosimple //requires refactoring
	profile := svcs.Profile{ //nolint
		ID:    np.ID,
		Email: np.Email,
	}

	profileBytes, err := encoder.DataEncode(profile)
	if err != nil {
		return profile, err
	}

	// Key that stores imported accounts counter
	b, err := encoder.DataEncode(uint(0))
	if err != nil {
		return profile, err
	}

	batch := as.db.NewBatch()
	defer batch.Close()

	if err := batch.Set(profileKey(profileID), profileBytes); err != nil {
		return profile, err
	}

	if err := batch.Set(accountImportedIdx(profileID), b); err != nil {
		return profile, err
	}

	if err := batch.Write(); err != nil {
		return profile, err
	}

	return profile, nil
}

// RegisterProfile creates a new user profile based on given email
func (as Service) RegisterProfile(ctx context.Context, np svcs.NewProfile) (svcs.Profile, error) {
	var p svcs.Profile

	profileID := auth.GetClaims(ctx).UserID

	if err := as.validator.Check(np); err != nil {
		return p, err
	}

	profileKey := profileKey(profileID)

	hasProfile, err := as.db.Has(profileKey)
	if err != nil {
		return p, err
	}

	if hasProfile {
		return p, ErrProfileExists
	}

	profile, err := as.createNewProfile(profileID, np)
	if err != nil {
		return p, err
	}

	return profile, nil
}

// GetProfile returns the profile of the given user by context value
func (as Service) GetProfile(ctx context.Context) (svcs.Profile, error) {
	var profile svcs.Profile

	profileID := auth.GetClaims(ctx).UserID

	profileKey := profileKey(profileID)

	profileBytes, err := as.db.Get(profileKey)
	if err != nil {
		return profile, err
	}

	b := bytes.NewBuffer(profileBytes)
	dec := gob.NewDecoder(b)

	if er := dec.Decode(&profile); er != nil {
		return profile, er
	}

	return profile, nil
}

// GetAccountPrivateKey returns the private key of the given account
func (as Service) GetAccountPrivateKey(ctx context.Context, address string) (cryptotype.PrivKey, error) {
	key, err := as.keyByAddress(ctx, address)
	if err != nil {
		return nil, err
	}

	armoredKey, err := as.keyring.ExportPrivKeyArmor(key.Name, "")
	if err != nil {
		return nil, err
	}

	privKey, _, err := crypto.UnarmorDecryptPrivKey(armoredKey, "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt private key")
	}

	return privKey, nil
}

func (as Service) keyByAddress(ctx context.Context, address string) (*keyring.Record, error) {
	profileID := auth.GetClaims(ctx).UserID

	addr, err := types.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	keyringAccount, err := as.keyring.KeyByAddress(addr)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(keyringAccount.Name, keyringAccountPrefix(profileID)) {
		return nil, ErrAccountNotExists
	}

	return keyringAccount, nil
}

// GetProfileByAddress returns profileID that owns the given address
func (as Service) GetProfileByAddress(address string) (string, error) {
	profileID := ""

	addr, err := types.AccAddressFromBech32(address)
	if err != nil {
		return profileID, err
	}

	keyringAccount, err := as.keyring.KeyByAddress(addr)
	if err != nil {
		return profileID, err
	}

	keyParts := strings.Split(keyringAccount.Name, "_")

	if len(keyParts) == 0 {
		return profileID, fmt.Errorf("cannot discover account by address: %s and key: %s", address, keyringAccount.Name)
	}

	return keyParts[0], nil
}

// GetProfileAccount returns the account of the given user by context value
func (as Service) GetProfileAccount(ctx context.Context, address string) (svcs.Account, error) {
	keyringAccount, err := as.keyByAddress(ctx, address)
	if err != nil {
		return svcs.Account{}, err
	}

	return as.Keyring2Account(ctx, keyringAccount)
}

// GetProfileAccounts returns all accounts of the given context user
func (as Service) GetProfileAccounts(ctx context.Context) (svcs.ProfileAccounts, error) {
	hdAccounts, err := as.getHDWalletAccounts(ctx)
	if err != nil {
		return svcs.ProfileAccounts{}, fmt.Errorf("cannot get HD accounts: %w", err)
	}

	importedAccounts, err := as.getImportedAccounts(ctx)
	if err != nil {
		return svcs.ProfileAccounts{}, fmt.Errorf("cannot get imported accounts: %w", err)
	}

	return svcs.ProfileAccounts{
		HDAccounts:       hdAccounts,
		ImportedAccounts: importedAccounts,
	}, nil
}

func (as Service) getImportedAccounts(ctx context.Context) ([]svcs.Account, error) {
	accounts := make([]svcs.Account, 0)

	profileID := auth.GetClaims(ctx).UserID

	prefixDB := db.NewPrefixDB(as.db, accountImportedKey(profileID, ""))
	itr, err := prefixDB.Iterator(nil, nil)
	if err != nil {
		return accounts, err
	}

	for ; itr.Valid(); itr.Next() {
		address, err := types.AccAddressFromBech32(string(itr.Key()))
		if err != nil {
			return accounts, err
		}

		key, err := as.keyring.KeyByAddress(address)
		if err != nil {
			return accounts, err
		}

		account, err := as.Keyring2Account(ctx, key)
		if err != nil {
			return accounts, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (as Service) getHDWalletAccounts(ctx context.Context) ([]svcs.Account, error) {
	accounts := make([]svcs.Account, 0)

	profileID := auth.GetClaims(ctx).UserID

	hasWallet, err := as.db.Has(walletKey(profileID))
	if err != nil {
		return accounts, err
	}

	if !hasWallet {
		return accounts, nil
	}

	walletIndex, err := as.GetWalletAccountIndex(ctx)
	if err != nil {
		return accounts, err
	}

	for index := uint(0); index <= walletIndex; index++ {
		key, err := as.keyring.Key(keyringAccountKey(profileID, index))
		if err != nil {
			return accounts, err
		}

		account, err := as.Keyring2Account(ctx, key)
		if err != nil {
			return accounts, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Keyring2Account conventrs a keyring record to client helper account
func (as Service) Keyring2Account(ctx context.Context, key *keyring.Record) (svcs.Account, error) {
	profileID := auth.GetClaims(ctx).UserID

	addr, err := key.GetAddress()
	if err != nil {
		return svcs.Account{}, err
	}

	var accBytes []byte

	if strings.Contains(key.Name, "imported") {
		accBytes, err = as.db.Get(accountImportedKey(profileID, addr.String()))
		if err != nil {
			return svcs.Account{}, err
		}
	} else {
		accBytes, err = as.db.Get(accountHDKey(profileID, addr.String()))
		if err != nil {
			return svcs.Account{}, err
		}
	}

	acc := Account{}

	b := bytes.NewBuffer(accBytes)
	dec := gob.NewDecoder(b)

	if er := dec.Decode(&acc); er != nil {
		return svcs.Account{}, er
	}

	balance, err := as.BalanceByAddress(ctx, addr.String())
	if err != nil {
		return svcs.Account{}, err
	}

	nfts, err := as.nodeClient.GetNFTByAddress(ctx, addr.String())
	if err != nil {
		return svcs.Account{}, err
	}

	pubKey, err := key.GetPubKey()
	if err != nil {
		return svcs.Account{}, err
	}

	return svcs.Account{
		Name:      acc.Name,
		PublicKey: fmt.Sprintf("%X", pubKey.Bytes()),
		Address:   addr.String(),
		Balance:   balance.Balance,
		NFTsCount: uint(len(nfts)),
	}, nil
}
