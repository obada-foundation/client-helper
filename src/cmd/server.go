package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/obada-foundation/client-helper/rest/api"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/device"
	"github.com/obada-foundation/client-helper/services/pubkey"
	"github.com/obada-foundation/client-helper/system/auth"
	"github.com/obada-foundation/client-helper/system/ipfs"
	"github.com/obada-foundation/client-helper/system/obadanode"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/obada-foundation/sdkgo"
)

// ServerCommand with command line flags and env
type ServerCommand struct {
	Port    int       `long:"port" env:"SERVER_PORT" default:"9090" description:"port"`
	Address string    `long:"address" env:"SERVER_ADDRESS" default:"" description:"listening address"`
	SSL     SSLGroup  `group:"ssl" namespace:"ssl" env-namespace:"SSL"`
	Auth    AuthGroup `group:"auth" namespace:"auth" env-namespace:"AUTH"`
	Node    NodeGroup `group:"node" namespace:"node" env-namespace:"NODE"`
	IPFS    IPFSGroup `group:"ipfs" namespace:"ipfs" env-namespace:"IPFS"`

	CommonOpts
}

type AuthGroup struct {
	KeysFolder string `long:"keys-folder" env:"KEYS_FOLDER" default:"/home/obada/keys" description:"Folder where public keys for verification are stored"`
	ActiveKID  string `long:"active-kid" env:"ACTIVE_KID" default:"85bb2165-90e1-4134-af3e-90a4a0e1e2c1" description:"Active public key that should be used by default"`
}

// SSLGroup defines options group for server ssl params
type SSLGroup struct {
	Type string `long:"type" env:"TYPE" description:"ssl support" choice:"none" choice:"static" default:"none"` // nolint
	Cert string `long:"cert" env:"CERT" description:"path to cert.pem file"`
	Key  string `long:"key" env:"KEY" description:"path to key.pem file"`
}

// NodeGroup defines options for connection to the blockchain node
type NodeGroup struct {
	ChainID string `long:"chain-id" env:"CHAIN_ID" description:"" default:"obada-testnet"`
	RpcURL  string `long:"rpc-url" env:"RPC_URL" description:"" default:"tcp://52.206.218.105:26657"`
	GrpcURL string `long:"grpc-url" env:"GRPC_URL" description:"" default:"52.206.218.105:9090"`
}

type IPFSGroup struct {
	RPC_URL string `long:"url" env:"RPC_URL" description:"IPFS RPC url to connect"`
}

// serverApp holds all active objects
type serverApp struct {
	*ServerCommand
	restSrv    *api.Rest
	terminated chan struct{}
}

// Execute is the entry point for "server" command, called by flag parser
func (s *ServerCommand) Execute(_ []string) error {
	s.Logger.Infof("start server on port %s:%d", s.Address, s.Port)

	ctx, cancel := context.WithCancel(context.Background())

	go func() { // catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		s.Logger.Warnf("interrupt signal")
		cancel()
	}()

	app, err := s.newServerApp()
	if err != nil {
		return err
	}

	if err := app.run(ctx); err != nil {
		s.Logger.Errorf("server terminated with error %+v", err)
		return err
	}

	return nil
}

// newServerApp prepares application and return it with all active parts
// doesn't start anything
func (s *ServerCommand) newServerApp() (*serverApp, error) {
	nodeClient, err := obadanode.NewClient(
		s.Node.ChainID,
		s.Node.RpcURL,
		s.Node.GrpcURL,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize node connection: %w", err)
	}

	sslConfig := s.makeSSLConfig()

	ks, err := pubkey.NewFS(s.Auth.KeysFolder)
	if err != nil {
		return nil, fmt.Errorf("reading keys: %w", err)
	}

	// Auth manager verifies JWT tokens
	a, err := auth.New(s.Auth.ActiveKID, ks)
	if err != nil {
		return nil, err
	}

	// Validator package
	validator, err := validate.NewValidator()
	if err != nil {
		return nil, err
	}

	// Account service manage OBADA wallets
	accountSvc := account.NewService(validator, s.DB, nodeClient)

	// Device manager initialization
	sdk, err := sdkgo.NewSdk(nil, false)
	if err != nil {
		return nil, err
	}

	// IPFS shell intialization
	ipfsShell := ipfs.NewIPFS(s.IPFS.RPC_URL)

	deviceSvc := device.NewService(validator, s.DB, sdk, ipfsShell)

	srv := &api.Rest{
		AccountService: accountSvc,
		DeviceService:  deviceSvc,
		Logger:         s.Logger,
		SSLConfig:      sslConfig,
		Auth:           a,
	}

	return &serverApp{
		ServerCommand: s,
		restSrv:       srv,
		terminated:    make(chan struct{}),
	}, nil
}

func (s *ServerCommand) makeSSLConfig() (config api.SSLConfig) {
	switch s.SSL.Type {
	case "none":
		config.Type = api.None

	}

	config.Cert = s.SSL.Cert
	config.Key = s.SSL.Key

	return config
}

// Run all application objects
func (a *serverApp) run(ctx context.Context) error {
	go func() {
		// shutdown on context cancellation
		<-ctx.Done()
		a.Logger.Info("shutdown initiated")
		a.restSrv.Shutdown()
	}()

	a.restSrv.Run(a.Address, a.Port)

	close(a.terminated)
	return nil
}

// Wait for application completion (termination)
func (a *serverApp) Wait() {
	<-a.terminated
}
