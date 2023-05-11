package v1

import (
	"net/http"

	middleware "github.com/obada-foundation/client-helper/api/middleware/v1"
	"github.com/obada-foundation/client-helper/api/v1/accounts"
	"github.com/obada-foundation/client-helper/api/v1/nft"
	"github.com/obada-foundation/client-helper/api/v1/obit"
	"github.com/obada-foundation/client-helper/api/v1/obits"
	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/blockchain"
	"github.com/obada-foundation/client-helper/services/device"
	"github.com/obada-foundation/client-helper/system/web"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by route handlers.
type Config struct {
	Log  *zap.SugaredLogger
	Auth *auth.Auth

	// Services
	AccountSvc    *account.Service
	BlockchainSvc *blockchain.Service
	DeviceSvc     *device.Service
	ObitSvc       *services.ObitService
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "api/v1"

	authenticate := middleware.Authenticate(cfg.Auth)
	accountMw := middleware.Account(cfg.AccountSvc)

	accountsGrp := accounts.Handlers{
		AccountSvc:    cfg.AccountSvc,
		BlockchainSvc: cfg.BlockchainSvc,
	}

	app.Handle(http.MethodGet, version, "/accounts", accountsGrp.Accounts, authenticate)
	app.Handle(http.MethodPost, version, "/accounts/register", accountsGrp.Register, authenticate)
	app.Handle(http.MethodPost, version, "/accounts/new-wallet", accountsGrp.NewWallet, authenticate)
	app.Handle(http.MethodPost, version, "/accounts/import-wallet", accountsGrp.ImportWallet, authenticate)
	app.Handle(http.MethodPost, version, "/accounts/new-account", accountsGrp.NewAccount, authenticate)
	app.Handle(http.MethodPost, version, "/accounts/import-account", accountsGrp.ImportAccount, authenticate)
	app.Handle(http.MethodPost, version, "/accounts/export-account", accountsGrp.ExportAccount, authenticate)
	app.Handle(http.MethodGet, version, "/accounts/new-mnemonic", accountsGrp.NewMnemonic, authenticate)
	app.Handle(http.MethodGet, version, "/accounts/mnemonic", accountsGrp.Mnemonic, authenticate)
	app.Handle(http.MethodGet, version, "/accounts/:address", accountsGrp.Account, authenticate, accountMw)
	app.Handle(http.MethodPost, version, "/accounts/:address", accountsGrp.UpdateAccount, authenticate, accountMw)
	app.Handle(http.MethodDelete, version, "/accounts/:address", accountsGrp.DeleteAccount, authenticate, accountMw)
	app.Handle(http.MethodPost, version, "/accounts/:address/send-coins", accountsGrp.SendCoins, authenticate, accountMw)

	obitsGrp := obits.Handlers{
		AccountSvc: cfg.AccountSvc,
		DeviceSvc:  cfg.DeviceSvc,
	}

	app.Handle(http.MethodGet, version, "/obits/:key", obitsGrp.Obit, authenticate)
	app.Handle(http.MethodGet, version, "/obits", obitsGrp.Search, authenticate)
	app.Handle(http.MethodPost, version, "/obits", obitsGrp.Save, authenticate)

	obitGrp := obit.Handlers{
		ObitSvc: cfg.ObitSvc,
	}

	app.Handle(http.MethodPost, version, "/obit/did", obitGrp.GenerateObit)
	app.Handle(http.MethodPost, version, "/obit/checksum", obitGrp.GenerateChecksum)

	nftGrp := nft.Handlers{
		AccountSvc:    cfg.AccountSvc,
		DeviceSvc:     cfg.DeviceSvc,
		BlockchainSvc: cfg.BlockchainSvc,
	}

	app.Handle(http.MethodGet, version, "/nft/:key", nftGrp.NFT, authenticate)
	app.Handle(http.MethodPost, version, "/nft/:key/mint", nftGrp.Mint, authenticate)
	app.Handle(http.MethodPost, version, "/nft/:key/metadata", nftGrp.UpdateMetadata, authenticate)
	app.Handle(http.MethodPost, version, "/nft/:key/send", nftGrp.Transfer, authenticate)
}
