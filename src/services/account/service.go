package account

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/obada-foundation/client-helper/system/auth"
	"github.com/obada-foundation/client-helper/system/db"
	"github.com/obada-foundation/client-helper/system/encoder"
	"github.com/obada-foundation/client-helper/system/obadanode"
	"github.com/obada-foundation/client-helper/system/validate"
)

type Service struct {
	validator  *validate.Validator
	db         db.DB
	nodeClient obadanode.NodeClient
}

func NewService(v *validate.Validator, db db.DB, client obadanode.NodeClient) *Service {
	return &Service{
		validator:  v,
		db:         db,
		nodeClient: client,
	}
}

func accountKey(ID string) []byte {
	return []byte(fmt.Sprintf("accounts:%s", ID))
}

func walletKey(ID string) []byte {
	return []byte(fmt.Sprintf("accounts:%s:wallet", ID))
}

func (as *Service) createAccount(accountKey []byte, na NewAccount, batch db.Batch) (Account, error) {
	account := Account{
		ID:    na.ID,
		Email: na.Email,
	}

	accountBytes, err := encoder.DataEncode(account)
	if err != nil {
		return account, err
	}

	if err := batch.Set(accountKey, accountBytes); err != nil {
		return account, err
	}

	return account, nil
}

func (as *Service) createWallet(walletKey []byte, batch db.Batch) (Wallet, error) {
	privKey := secp256k1.GenPrivKey()

	wallet := Wallet{
		PrivateKey: *privKey,
	}

	walletBytes, err := encoder.DataEncode(wallet)
	if err != nil {
		return wallet, err
	}

	if err := batch.Set(walletKey, walletBytes); err != nil {
		return wallet, err
	}

	return wallet, nil
}

// Create creates a new account based on given email, returns an access token for helper API
func (as *Service) Create(ctx context.Context, na NewAccount) (Account, error) {
	var acc Account

	if err := as.validator.Check(na); err != nil {
		return acc, err
	}

	accKey := accountKey(na.ID)

	hasAcc, err := as.db.Has(accKey)
	if err != nil {
		return acc, err
	}

	if hasAcc {
		return acc, ErrAccountExists
	}

	batch := as.db.NewBatch()
	defer batch.Close()

	account, err := as.createAccount(accKey, na, batch)
	if err != nil {
		return acc, err
	}

	if _, err := as.createWallet(walletKey(na.ID), batch); err != nil {
		return acc, err
	}

	if err := batch.Write(); err != nil {
		return acc, err
	}

	return account, nil
}

func (as *Service) Balance(ctx context.Context) (Balance, error) {
	var balance Balance

	wallet, err := as.Wallet(ctx)
	if err != nil {
		return balance, err
	}

	pubKey := wallet.PrivateKey.PubKey()

	nodeBalance, err := as.nodeClient.Balance(ctx, pubKey)
	if err != nil {
		return balance, err
	}

	addr, err := types.AccAddressFromHex(pubKey.Address().String())
	if err != nil {
		return balance, err
	}

	return Balance{
		Address: addr.String(),
		Balance: int(nodeBalance.Balance.Amount.Uint64()),
	}, nil
}

func (as *Service) Wallet(ctx context.Context) (Wallet, error) {
	var wallet Wallet

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return wallet, err
	}

	waKey := walletKey(userID)

	hasWallet, err := as.db.Has(waKey)
	if err != nil {
		return wallet, err
	}

	if !hasWallet {
		return wallet, ErrAccountNotExists
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

func (as *Service) Show(ctx context.Context) (Account, error) {
	var account Account

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return account, err
	}

	accountBytes, err := as.db.Get(accountKey(userID))
	if err != nil {
		return account, err
	}

	b := bytes.NewBuffer(accountBytes)
	dec := gob.NewDecoder(b)

	if err = dec.Decode(&account); err != nil {
		return account, err
	}

	return account, nil
}
