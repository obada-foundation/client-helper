package device

import (
	"context"
	"crypto/rand"

	"github.com/obada-foundation/client-helper/system/auth"
)

func (ds Service) EncryptionSecret(ctx context.Context, DID string) ([]byte, error) {
	secretBytes := make([]byte, 32)

	userID, err := auth.GetUserID(ctx)
	secretKey := makeSecretKey(userID, DID)

	ok, err := ds.db.Has(secretKey)
	if err != nil {
		return secretBytes, err
	}

	if !ok {
		if _, err := rand.Read(secretBytes); err != nil {
			return secretBytes, nil
		}

		if err := ds.db.Set(secretKey, secretBytes); err != nil {
			return secretBytes, nil
		}

		return secretBytes, nil
	}

	return ds.db.Get(secretKey)
}
