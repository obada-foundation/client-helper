// Package encryption provides encrypt/descrypt functionality for OBADA needs
package encryption

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	ecies "github.com/ecies/go/v2"
)

// Encrypt given data
func Encrypt(pubKey cryptotypes.PubKey, data []byte) ([]byte, error) {
	encryptionKey, err := ecies.NewPublicKeyFromBytes(pubKey.Bytes())
	if err != nil {
		return nil, err
	}

	encryptedBytes, err := ecies.Encrypt(encryptionKey, data)
	if err != nil {
		return nil, err
	}

	return encryptedBytes, nil
}

// Decrypt given encrypted data
func Decrypt(privKey cryptotypes.PrivKey, encData []byte) ([]byte, error) {
	decryptionKey := ecies.NewPrivateKeyFromBytes(privKey.Bytes())

	decBytes, err := ecies.Decrypt(decryptionKey, encData)
	if err != nil {
		return nil, err
	}

	return decBytes, nil
}
