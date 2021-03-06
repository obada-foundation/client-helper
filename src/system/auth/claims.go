package auth

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

const (
	RoleUser = "USER"
)

type Claims struct {
	jwt.StandardClaims
	Roles  []string `json:"roles"`
	UserID string   `json:"uid"`
}

func (c Claims) Authorized(roles ...string) bool {
	for _, has := range c.Roles {
		for _, want := range roles {
			if has == want {
				return true
			}
		}
	}

	return false
}

type ctxKey int

const key ctxKey = 1

func SetClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, key, claims)
}

func GetClaims(ctx context.Context) (Claims, error) {
	v, ok := ctx.Value(key).(Claims)
	if !ok {
		return Claims{}, errors.New("claim value missing from the context")
	}

	return v, nil
}

func GetUserID(ctx context.Context) (string, error) {
	claims, err := GetClaims(ctx)
	if err != nil {
		return "", err
	}

	return claims.UserID, nil
}
