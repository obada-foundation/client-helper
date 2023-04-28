package account

import (
	"github.com/cosmos/go-bip39"
)

const (
	// bits of entropy to draw when creating a mnemonic
	defaultEntropySize = 256
	addressSuffix      = "address"
	infoSuffix         = "info"
)

// NewMnemonic generates a new mnemonic, derives a hierarchical deterministic
func NewMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(defaultEntropySize)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}
