package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	appErrors "github.com/obada-foundation/client-helper/api/errors"
	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/blockchain"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/obada-foundation/client-helper/system/web"
	"go.uber.org/zap"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *zap.SugaredLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if err := handler(ctx, w, r); err != nil {
				log.Errorw("ERROR", "trace_id", web.GetTraceID(ctx), "message", err)

				var er appErrors.ErrorResponse
				var status int
				switch {
				case blockchain.IsAcceptableError(err):
					er = appErrors.ErrorResponse{
						Error: err.Error(),
					}

					status = http.StatusBadRequest
				case account.IsAccountError(err):
					er = appErrors.ErrorResponse{
						Error: err.Error(),
					}

					status = http.StatusBadRequest

					if errors.Is(err, account.ErrWalletNotExists) {
						status = http.StatusNotFound
					}

					if errors.Is(err, account.ErrWalletExists) {
						status = http.StatusConflict
					}

				case validate.IsFieldErrors(err):
					fieldErrors := validate.GetFieldErrors(err)
					er = appErrors.ErrorResponse{
						Error:  "data validation error",
						Fields: fieldErrors.Fields(),
					}
					status = http.StatusBadRequest

				case appErrors.IsRequestError(err):
					reqErr := appErrors.GetRequestError(err)
					er = appErrors.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				case auth.IsAuthError(err):
					er = appErrors.ErrorResponse{
						Error: http.StatusText(http.StatusUnauthorized),
					}
					status = http.StatusUnauthorized

				default:
					sentry.CaptureException(err)

					// Temporary show up internal error to the user (issue 45-346 requested by Rohi)
					er = appErrors.ErrorResponse{
						Error: err.Error(),
					}

					//nolint:gocritic //temporary commented because of Rohi request
					// er = appErrors.ErrorResponse{
					// 	Error: http.StatusText(http.StatusInternalServerError),
					// }
					status = http.StatusInternalServerError
				}

				if er := web.Respond(ctx, w, er, status); er != nil {
					return er
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down the service.
				if web.IsShutdown(err) {
					return err
				}
			}

			return nil
		}

		return h
	}

	return m
}
