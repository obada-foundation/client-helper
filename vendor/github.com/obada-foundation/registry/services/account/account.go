package account

import (
	"context"
	"fmt"
	"strings"

	immudb "github.com/codenotary/immudb/pkg/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/obada-foundation/sdkgo/base58"
	"go.uber.org/zap"
)

// Account defines an interface API for registerening account addresses and their public keys
type Account interface {
	// RegisterAccount registers an account public key
	Register(ctx context.Context, pubKey string) error

	// GetPublicKey returns the base58 public key for the given address
	GetPublicKey(ctx context.Context, address string) (string, error)
}

// Service is impementation of the Account interface
type Service struct {
	db     immudb.ImmuClient
	logger *zap.SugaredLogger
}

// NewService creates a new account service
func NewService(db immudb.ImmuClient, logger *zap.SugaredLogger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

// Register implemnts method of the Account interface, it takes a base58 public key and registers it
func (s Service) Register(ctx context.Context, pubKeyB58 string) error {
	if pubKeyB58 == "" {
		return ErrPublicKeyIsEmpty
	}

	pubKey := &secp256k1.PubKey{
		Key: base58.Decode(pubKeyB58),
	}

	if len(pubKey.Key) != secp256k1.PubKeySize {
		return ErrInvalidPublicKey
	}

	addr := types.AccAddress(pubKey.Address()).String()

	if _, err := s.db.Set(ctx, []byte(addr), []byte(pubKeyB58)); err != nil {
		return err
	}

	s.logger.Debugw("new address registered", "address", addr)

	return nil
}

// GetPublicKey implements method of the Account interface, it takes blockchain address and returns the base58 encoded public key
func (s Service) GetPublicKey(ctx context.Context, address string) (string, error) {
	addr, err := types.AccAddressFromBech32(address)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidAddress, err)
	}

	entry, err := s.db.Get(ctx, []byte(addr.String()))
	if err != nil {
		if strings.Contains(err.Error(), "key not found") {
			return "", ErrPubKeyNotRegistered
		}
		return "", err
	}

	return string(entry.GetValue()), nil
}
