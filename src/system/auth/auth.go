package auth

import (
	eddsa "crypto/ed25519"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type KeyLookup interface {
	PublicKey(kid string) (eddsa.PublicKey, error)
}

type Auth struct {
	activeKID string
	keyLookup KeyLookup
	method    jwt.SigningMethod
	keyFunc   func(t *jwt.Token) (interface{}, error)
	parser    jwt.Parser
}

func New(activeKID string, keyLookup KeyLookup) (*Auth, error) {
	_, err := keyLookup.PublicKey(activeKID)
	if err != nil {
		return nil, err
		return nil, errors.New("active KID doesn't exists in store")
	}

	method := jwt.GetSigningMethod("EdDSA")
	if method == nil {
		return nil, errors.New("configuring algorithm EdDSA")
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}

		kidID, ok := kid.(string)
		if !ok {
			return nil, errors.New("user token key id (kid) must be a string")
		}

		return keyLookup.PublicKey(kidID)
	}

	parser := jwt.Parser{
		ValidMethods: []string{"EdDSA"},
	}

	a := Auth{
		activeKID: activeKID,
		keyLookup: keyLookup,
		method:    method,
		keyFunc:   keyFunc,
		parser:    parser,
	}

	return &a, nil
}

func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {
	var claims Claims

	token, err := a.parser.ParseWithClaims(tokenStr, &claims, a.keyFunc)
	if err != nil {
		return Claims{}, fmt.Errorf("parsing token: %w", err)
	}

	if !token.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}

	return claims, nil
}
