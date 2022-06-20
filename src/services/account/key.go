package account

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

type PrivateKeyGenerator interface {
	GeneratePrivateKey() secp256k1.PrivKey
}

type PrivateKeyConcreteGenerator struct {
}
