package keyring

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/go-bip39"
)

const (
	// bits of entropy to draw when creating a mnemonic
	defaultEntropySize = 256
	addressSuffix      = "address"
	infoSuffix         = "info"
)

// NewKeying creates a new keyring with a single key
func NewKeying(_, keyringPath string) (keyring.Keyring, error) {
	return keyring.New("client-helper", keyring.BackendFile, keyringPath, nil, network.DefaultConfig().Codec)
}

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
