package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HashStr hash string to SHA256
func HashStr(str string) (string, error) {
	h := sha256.New()

	if _, err := h.Write([]byte(str)); err != nil {
		return "", fmt.Errorf("cannot wite bytes %v to hasher: %w", []byte(str), err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
