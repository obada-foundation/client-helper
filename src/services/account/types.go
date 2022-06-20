package account

import "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

type NewAccount struct {
	ID    string `json:"id" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type Account struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type Wallet struct {
	Address    string            `json:"address"`
	Balance    int               `json:"balance"`
	PrivateKey secp256k1.PrivKey `json:"-"`
}
