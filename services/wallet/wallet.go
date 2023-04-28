package wallet

import (
	"bytes"
	"encoding/gob"
	"strings"

	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/encoder"
	"github.com/tendermint/tm-db"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"go.uber.org/zap"
)

// Service is a wallet service
type Service struct {
	db     db.DB
	logger *zap.SugaredLogger
}

// NewService creates a new wallet service
func NewService(dbs db.DB, logger *zap.SugaredLogger) *Service {
	return &Service{
		db:     dbs,
		logger: logger,
	}
}

// NewMasterKey creates a new master key
func (ws Service) NewMasterKey(accountID, kid, mnemonic string) (*bip32.Key, error) {
	mKey, err := masterKeyFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	masterKeyBytes, err := encoder.DataEncode(services.MasterKey{
		ID:  kid,
		Key: mKey.String(),
	})
	if err != nil {
		return nil, err
	}

	dbKey := masterKey(accountID, kid)

	if err := ws.db.Set(dbKey, masterKeyBytes); err != nil {
		return nil, err
	}

	if _, err := ws.NewPrivateKey(accountID, kid, mKey); err != nil {
		return nil, err
	}

	return mKey, nil
}

// GetMasterKey returns master key
func (ws Service) GetMasterKey(accountID, kid string) (*bip32.Key, error) {
	dbKey := masterKey(accountID, kid)

	masterKeyBytes, err := ws.db.Get(dbKey)
	if err != nil {
		return nil, err
	}

	masterKey := services.MasterKey{}
	b := bytes.NewBuffer(masterKeyBytes)
	dec := gob.NewDecoder(b)

	if er := dec.Decode(&masterKey); er != nil {
		return nil, er
	}

	mKey, err := bip32.B58Deserialize(masterKey.Key)
	if err != nil {
		return nil, err
	}

	return mKey, nil
}

// NewPrivateKey creates a new private key
func (ws Service) NewPrivateKey(accountID, kid string, mKey *bip32.Key) (*bip32.Key, error) {

	pPkeys, err := ws.GetAllPrivateKeys(accountID, kid)
	if err != nil {
		return nil, err
	}

	pKey, err := mKey.NewChildKey(uint32(len(pPkeys)) + 1)
	if err != nil {
		return nil, err
	}

	dbKey := privateKey(accountID, kid, pKey.PublicKey().String())

	if err := ws.db.Set(dbKey, []byte(pKey.String())); err != nil {
		return nil, err
	}

	return pKey, nil
}

// GetAllMasterKeys returns all master keys
func (ws Service) GetAllMasterKeys(accountID string) ([]services.MasterKey, error) {
	var keys []services.MasterKey

	dbKey := masterKey(accountID, "")

	ws.logger.Error("test:", string(dbKey))

	iterator, err := ws.db.Iterator(nil, nil)
	if err != nil {
		return keys, err
	}

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		//nolint: gocritic // for future refactoring
		if strings.Contains(string(iterator.Key()), string(dbKey)) && !strings.Contains(string(iterator.Key()), "private-keys") {
			mKey := services.MasterKey{}
			masterKeyBytes, err := ws.db.Get(iterator.Key())
			if err != nil {
				return keys, err
			}

			b := bytes.NewBuffer(masterKeyBytes)
			dec := gob.NewDecoder(b)

			if err := dec.Decode(&mKey); err != nil {
				return keys, err
			}

			keys = append(keys, mKey)
		}
	}

	return keys, nil
}

// GetAllPrivateKeys returns all private keys
func (ws Service) GetAllPrivateKeys(accountID, kid string) ([]*bip32.Key, error) {
	privKeys := []*bip32.Key{}

	dbKey := privateKey(accountID, kid, "")

	iterator, err := ws.db.Iterator(nil, nil)
	if err != nil {
		return privKeys, err
	}

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Key())

		if strings.Contains(key, string(dbKey)) {
			privateKeyBytes, err := ws.db.Get(iterator.Key())
			if err != nil {
				return privKeys, err
			}

			pKey, err := bip32.B58Deserialize(string(privateKeyBytes))
			if err != nil {
				return nil, err
			}

			privKeys = append(privKeys, pKey)
		}
	}

	return privKeys, nil
}

// GetPrivateKey returns private key
func (ws Service) GetPrivateKey(accountID, kid, publicKey string) (*bip32.Key, error) {
	dbKey := privateKey(accountID, kid, publicKey)

	privateKeyBytes, err := ws.db.Get(dbKey)
	if err != nil {
		return nil, err
	}

	pKey, err := bip32.B58Deserialize(string(privateKeyBytes))
	if err != nil {
		return nil, err
	}

	return pKey, nil
}

func masterKeyFromMnemonic(mnemonic string) (*bip32.Key, error) {
	seed := bip39.NewSeed(mnemonic, "")

	return bip32.NewMasterKey(seed)
}
