package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
	"github.com/obada-foundation/client-helper/blockchain/client"
	"github.com/obada-foundation/client-helper/blockchain/wallet"
	mid "github.com/obada-foundation/client-helper/rest/api/middleware"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/device"
	"github.com/obada-foundation/client-helper/services/nft"
	"github.com/obada-foundation/client-helper/system/auth"
	"github.com/obada-foundation/sdkgo"
	"go.uber.org/zap"
)

const hardBodyLimit = 1024 * 64 // limit size of body

// Rest is a rest access server
type Rest struct {
	Version string

	ClientHelperURL string

	// Services
	AccountService *account.Service
	DeviceService  *device.Service
	NFTService     *nft.Service

	// System
	Auth   *auth.Auth
	DB     *badger.DB
	Logger *zap.SugaredLogger

	// Server
	SSLConfig   SSLConfig
	httpServer  *http.Server
	httpsServer *http.Server
	lock        sync.Mutex

	// Route groups
	pubRest     public
	accountRest accountGroup
	deviceRest  deviceGroup
	nftRest     nftGroup
}

// Run the lister and request's router, activate rest server
func (s *Rest) Run(address string, port int) {
	if address == "*" {
		address = ""
	}

	switch s.SSLConfig.Type {
	case None:
		s.Logger.Infof("activate http rest server on %s:%d", address, port)

		s.lock.Lock()
		s.httpServer = s.makeHTTPServer(address, port, s.routes())
		s.httpServer.ErrorLog = zap.NewStdLog(s.Logger.Desugar())
		s.lock.Unlock()

		err := s.httpServer.ListenAndServe()
		s.Logger.Warnf("http server terminated, %s", err)
	case Static:
		s.Logger.Warnf("activate https server mode on %s:%d", address, port)

		s.lock.Lock()
		s.httpsServer = s.makeHTTPSServer(address, port, s.routes())
		s.httpsServer.ErrorLog = zap.NewStdLog(s.Logger.Desugar())
		s.lock.Unlock()

		err := s.httpsServer.ListenAndServeTLS(s.SSLConfig.Cert, s.SSLConfig.Key)
		s.Logger.Warnf("https server terminated, %s", err)
	}
}

func (s *Rest) makeHTTPServer(address string, port int, router http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", address, port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		// WriteTimeout:      120 * time.Second, // TODO: such a long timeout needed for blocking export (backup) request
		IdleTimeout: 30 * time.Second,
	}
}

func (s *Rest) routes() chi.Router {
	router := chi.NewRouter()

	middlewareLogger := zap.NewStdLog(s.Logger.Desugar())
	mdLogger := chimid.RequestLogger(&chimid.DefaultLogFormatter{Logger: middlewareLogger, NoColor: false})

	router.Use(mdLogger)
	router.Use(chimid.Throttle(1000), chimid.RealIP, mid.Recoverer(s.Logger))
	router.Use(mid.AppInfo("Client Helper", "OBADA Foundation", s.Version), mid.Ping)

	s.pubRest, s.accountRest, s.deviceRest, s.nftRest = s.controllerGroups() // assign controllers for groups

	// api routes
	router.Route("/api/v1", func(rapi chi.Router) {
		// protected routes, require auth
		rapi.Group(func(rauth chi.Router) {
			rauth.Use(mid.Authenticate(s.Auth))

			rauth.Route("/accounts", func(account chi.Router) {
				account.Post("/", s.accountRest.create)
				account.Get("/my-balance", s.accountRest.balance)
			})

			rauth.Route("/obits", func(obits chi.Router) {
				obits.Get("/{key}", s.deviceRest.get)
				obits.Post("/", s.deviceRest.save)
				obits.Get("/", s.pubRest.search)
			})

			rauth.Route("/nft", func(nfts chi.Router) {
				nfts.Post("/{key}/mint", s.nftRest.mint)
				nfts.Post("/{key}/send", s.nftRest.transfer)
				nfts.Get("/{key}", s.nftRest.nft)
			})
		})

		rapi.Route("/obit", func(obit chi.Router) {
			obit.Post("/did", s.pubRest.generateObit)
			obit.Post("/checksum", s.pubRest.generateChecksum)
		})
	})

	return router
}

func (s *Rest) controllerGroups() (public, accountGroup, deviceGroup, nftGroup) {
	deviceGrp := deviceGroup{
		logger:     s.Logger,
		deviceSvc:  s.DeviceService,
		accountSvc: s.AccountService,
	}

	accountGrp := accountGroup{
		logger:     s.Logger,
		accountSvc: s.AccountService,
	}

	nftGrp := nftGroup{
		logger:     s.Logger,
		accountSvc: s.AccountService,
		deviceSvc:  s.DeviceService,
		nftSvc:     s.NFTService,
	}

	sdk, _ := sdkgo.NewSdk(nil, false)

	obitSvc := services.NewObitService(s.Logger, s.DB, sdk)
	obadaChain, _ := client.NewClient("obada-testnet", "tcp://52.206.218.105:26657", "52.206.218.105:9090")
	walletSvc := wallet.NewWalletService()

	publicGrp := public{
		logger:        s.Logger,
		obitService:   obitSvc,
		chainService:  &obadaChain,
		walletService: walletSvc,
	}

	return publicGrp, accountGrp, deviceGrp, nftGrp
}

// Shutdown rest http server
func (s *Rest) Shutdown() {
	s.Logger.Warnf("shutdown rest server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.lock.Lock()

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.Logger.Debugf("http shutdown error, %s", err)
		}
		s.Logger.Debug("shutdown http server completed")
	}

	if s.httpsServer != nil {
		s.Logger.Warn("shutdown https server")
		if err := s.httpsServer.Shutdown(ctx); err != nil {
			s.Logger.Debug("https shutdown error, %s", err)
		}
		s.Logger.Debug("shutdown https server completed")
	}
	s.lock.Unlock()
}
