package v1

import (
	"context"
	"net/http"

	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/system/web"
)

// Account checkes if profile has an access for the given account
func Account(accountSvc *account.Service) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			address := web.Param(r, "address")

			if ok := accountSvc.HasAccount(ctx, address); !ok {
				return auth.NewAuthError("permission denied")
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
