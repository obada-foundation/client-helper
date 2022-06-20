package wallet

import (
	digitalSignature "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Wallet struct {
	PrivateKey *digitalSignature.PrivKey
	PublicKey  types.PubKey
}

func (w *Wallet) GetObadaAddress() string {
	return sdk.AccAddress(w.PrivateKey.PubKey().Address().Bytes()).String()
}

type WalletService struct {
}

func NewWalletService() *WalletService {
	return &WalletService{}
}

func (ws *WalletService) GetWallet(walletID string) *Wallet {
	return ws.CreateWallet()
}

func (ws *WalletService) CreateWallet() *Wallet {
	secret := "fukusima"

	privKey := digitalSignature.GenPrivKeyFromSecret([]byte(secret))
	pubKey := privKey.PubKey()

	return &Wallet{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}
}
