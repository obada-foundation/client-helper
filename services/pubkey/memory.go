package pubkey

import (
	eddsa "crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v4"
)

// MemStore in memoty key storage
type MemStore struct {
	mu    sync.RWMutex
	store map[string]*eddsa.PublicKey
}

// New creates in memory new key store.
func New() *MemStore {
	return &MemStore{
		store: make(map[string]*eddsa.PublicKey),
	}
}

// NewMap creates key store from map
func NewMap(store map[string]*eddsa.PublicKey) *MemStore {
	return &MemStore{
		store: store,
	}
}

// NewFS creates a new key store from the directory
func NewFS(path string) (*MemStore, error) {
	ks := New()

	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walkdir failure: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".pem" {
			return nil
		}

		// nolint:gosec // for future refactoring
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("opening key file: %w", err)
		}
		// nolint:gosec // for future refactoring
		defer file.Close()

		// limit PEM file size to 1 megabyte. This should be reasonable for
		// almost any PEM file and prevents shenanigans like linking the file
		// to /dev/random or something like that.
		publicPEM, err := io.ReadAll(io.LimitReader(file, 1024*1024))
		if err != nil {
			return fmt.Errorf("reading public key: %w", err)
		}

		publicKey, err := jwt.ParseEdPublicKeyFromPEM(publicPEM)
		if err != nil {
			return fmt.Errorf("parsing auth public key: %w", err)
		}

		pkey, ok := publicKey.(eddsa.PublicKey)
		if !ok {
			return fmt.Errorf("token integrity issue")
		}

		ks.store[strings.TrimSuffix(filepath.Base(file.Name()), ".pem")] = &pkey
		return nil
	}

	if err := filepath.Walk(path, fn); err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	return ks, nil
}

// Add adds a public key and combination kid to the store.
func (ks *MemStore) Add(publicKey *eddsa.PublicKey, kid string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.store[kid] = publicKey
}

// Remove public a private key and combination kid to the store.
func (ks *MemStore) Remove(kid string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	delete(ks.store, kid)
}

// PublicKey searches the key store for a given kid and returns the public key.
func (ks *MemStore) PublicKey(kid string) (eddsa.PublicKey, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	publicKey, found := ks.store[kid]
	if !found {
		return nil, errors.New("kid lookup failed")
	}

	return *publicKey, nil
}
