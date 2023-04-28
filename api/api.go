package api

import (
	"net/http"
	"os"

	middleware "github.com/obada-foundation/client-helper/api/middleware/v1"
	"github.com/obada-foundation/client-helper/api/v1"
	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/blockchain"
	"github.com/obada-foundation/client-helper/services/device"
	"github.com/obada-foundation/client-helper/system/web"
	"go.uber.org/zap"
)

// APIMuxConfig defines the dependencies for the API mux.
type APIMuxConfig struct { //nolint:revive //for future refactoring
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	Auth     *auth.Auth

	// Services
	AccountSvc    *account.Service
	BlockchainSvc *blockchain.Service
	DeviceSvc     *device.Service
	ObitSvc       *services.ObitService
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) http.Handler { //nolint:revive //for future refactoring

	app := web.NewApp(
		cfg.Shutdown,
		middleware.Logger(cfg.Log),
		middleware.Errors(cfg.Log),
		middleware.Metrics(),
		middleware.Panics(),
	)

	v1.Routes(app, v1.Config{
		Log:  cfg.Log,
		Auth: cfg.Auth,

		// Services
		AccountSvc:    cfg.AccountSvc,
		BlockchainSvc: cfg.BlockchainSvc,
		DeviceSvc:     cfg.DeviceSvc,
		ObitSvc:       cfg.ObitSvc,
	})

	return app
}
