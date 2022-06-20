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
	"github.com/obada-foundation/client-helper/system/auth"
	"github.com/obada-foundation/sdkgo"
	"go.uber.org/zap"
)

const hardBodyLimit = 1024 * 64 // limit size of body

// Rest is a rest access server
type Rest struct {
	Version string

	ClientHelperURL string
	Auth            *auth.Auth
	AccountService  *account.Service
	DB              *badger.DB
	Logger          *zap.SugaredLogger
	SSLConfig       SSLConfig
	httpServer      *http.Server
	httpsServer     *http.Server
	lock            sync.Mutex

	pubRest      public
	accountsRest accounts
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

	s.pubRest, s.accountsRest = s.controllerGroups() // assign controllers for groups

	// api routes
	router.Route("/api/v1", func(rapi chi.Router) {
		// protected routes, require auth
		rapi.Group(func(rauth chi.Router) {
			rauth.Use(mid.Authenticate(s.Auth))

			rauth.Route("/accounts", func(account chi.Router) {
				account.Post("/", s.accountsRest.createAccount)
				account.Get("/me", s.accountsRest.myAccount)
			})
		})

		rapi.Route("/obit", func(obit chi.Router) {
			obit.Post("/did", s.pubRest.generateObit)
			obit.Post("/checksum", s.pubRest.generateChecksum)
		})

		rapi.Route("/obits", func(obits chi.Router) {
			obits.Get("/{key}", s.pubRest.getObit)
			obits.Post("/", s.pubRest.saveObit)
			obits.Get("/", s.pubRest.search)
			obits.Get("/{key}/to-chain", s.pubRest.uploadToChain)
			obits.Get("/{key}/from-chain", s.pubRest.downloadFromChain)
		})
	})

	return router
}

func (s *Rest) controllerGroups() (public, accounts) {
	sdk, _ := sdkgo.NewSdk(zap.NewStdLog(s.Logger.Desugar()), false)

	obitSvc := services.NewObitService(s.Logger, s.DB, sdk)
	obadaChain, _ := client.NewClient("obada-testnet", "tcp://52.206.218.105:26657", "52.206.218.105:9090")
	walletSvc := wallet.NewWalletService()

	pubGrp := public{
		logger:        s.Logger,
		obitService:   obitSvc,
		chainService:  &obadaChain,
		walletService: walletSvc,
	}

	accountGrp := accounts{
		logger:     s.Logger,
		accountSvc: s.AccountService,
	}

	return pubGrp, accountGrp
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
