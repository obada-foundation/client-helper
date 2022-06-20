package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/obada-foundation/client-helper/system/auth"
)

var (
	ErrUnauthorized = errors.New("token is unauthorized")
	ErrExpired      = errors.New("token is expired")
	ErrNBFInvalid   = errors.New("token nbf validation failed")
	ErrIATInvalid   = errors.New("token iat validation failed")
	ErrNoTokenFound = errors.New("no token found")
	ErrAlgoInvalid  = errors.New("algorithm mismatch")
)

func VerifyRequest(a *auth.Auth, r *http.Request) (auth.Claims, error) {
	authStr := r.Header.Get("authorization")

	// Parse the authorization header.
	parts := strings.Split(authStr, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return auth.Claims{}, ErrNoTokenFound
	}

	// Validate the token is signed by us.
	claims, err := a.ValidateToken(parts[1])
	if err != nil {
		log.Println(err)
		return auth.Claims{}, err
	}

	return claims, nil
}

func Authenticate(a *auth.Auth) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			claims, err := VerifyRequest(a, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := auth.SetClaims(r.Context(), claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(hfn)
	}
}
