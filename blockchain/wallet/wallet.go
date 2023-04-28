package wallet

import (
	digitalSignature "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Wallet is a struct that contains the private and public keys of a wallet
type Wallet struct {
	PrivateKey *digitalSignature.PrivKey
	PublicKey  types.PubKey
}

// GetObadaAddress returns the address of the wallet
func (w *Wallet) GetObadaAddress() string {
	return sdk.AccAddress(w.PrivateKey.PubKey().Address().Bytes()).String()
}

// WalletService is a struct that contains the wallets
type WalletService struct { //nolint:revive //for refactoring
}

// NewWalletService returns a new WalletService
func NewWalletService() *WalletService {
	return &WalletService{}
}

// GetWallet returns the wallet
func (ws *WalletService) GetWallet(_ string) *Wallet {
	return ws.CreateWallet()
}

// CreateWallet creates a new wallet
func (ws *WalletService) CreateWallet() *Wallet {
	secret := "fukusima"

	privKey := digitalSignature.GenPrivKeyFromSecret([]byte(secret))
	pubKey := privKey.PubKey()

	return &Wallet{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}
}
