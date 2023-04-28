package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
)

// Claims represents the claims in a JWT.
type Claims struct {
	jwt.StandardClaims
	Roles  []string `json:"roles"`
	UserID string   `json:"uid"`
}

type ctxKey int

const key ctxKey = 1

// SetClaims sets the claims in the context.
func SetClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, key, claims)
}

// GetClaims returns the claims from the context.
func GetClaims(ctx context.Context) Claims {
	v, ok := ctx.Value(key).(Claims)
	if !ok {
		return Claims{}
	}
	return v
}

// GetUserID returns the user ID from the context.
func GetUserID(ctx context.Context) string {
	return GetClaims(ctx).UserID
}
