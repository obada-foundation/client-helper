package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tmjson "github.com/cometbft/cometbft/libs/json"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	jsonrpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/getsentry/sentry-go"
	eb "github.com/mustafaturan/bus/v3"
	"github.com/obada-foundation/client-helper/api"
	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/bus"
	"github.com/obada-foundation/client-helper/events"
	"github.com/obada-foundation/client-helper/events/handlers"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/blockchain"
	"github.com/obada-foundation/client-helper/services/device"
	"github.com/obada-foundation/client-helper/services/pubkey"
	"github.com/obada-foundation/client-helper/system/ipfs"
	"github.com/obada-foundation/client-helper/system/obadanode"
	"github.com/obada-foundation/client-helper/system/validate"
	obadatypes "github.com/obada-foundation/fullcore/x/obit/types"
	registry "github.com/obada-foundation/registry/client"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServerCommand with command line flags and env
type ServerCommand struct {
	Port            int           `long:"port" env:"SERVER_PORT" default:"9090" description:"port"`
	Address         string        `long:"address" env:"SERVER_ADDRESS" default:"" description:"listening address"`
	ReadTimeout     time.Duration `long:"read-timeout" env:"READ_TIMEOUT" default:"5s" description:"read timeout"`
	WriteTimeout    time.Duration `long:"write-timeout" env:"WRITE_TIMEOUT" default:"10s" description:"write timeout"`
	IdleTimeout     time.Duration `long:"idle-timeout" env:"IDLE_TIMEOUT" default:"120s" description:"idle timeout"`
	ShutdownTimeout time.Duration `long:"shutdown-timeout" env:"SHUTDOWN_TIMEOUT" default:"20s" description:"shutdown timeout"`
	SentryDSN       string        `long:"sentry-dsn" env:"SENTRY_DSN" default:"" description:"sentry dsn"`
	Redis           RedisGroup    `group:"redis" namespace:"redis" env-namespace:"REDIS"`
	Registry        RegistryGroup `group:"registry" namespace:"registry" env-namespace:"REGISTRY"`
	SSL             SSLGroup      `group:"ssl" namespace:"ssl" env-namespace:"SSL"`
	Auth            AuthGroup     `group:"auth" namespace:"auth" env-namespace:"AUTH"`
	Node            NodeGroup     `group:"node" namespace:"node" env-namespace:"NODE"`
	IPFS            IPFSGroup     `group:"ipfs" namespace:"ipfs" env-namespace:"IPFS"`
	Keyring         KeyringGroup  `group:"keyring" namespace:"keyring" env-namespace:"KEYRING"`

	CommonOpts
}

// RedisGroup redis config options
type RedisGroup struct {
	Addr     string `long:"addr" env:"ADDR" default:"redis:6379" description:"redis address"`
	Password string `long:"password" env:"PASSWORD" default:"" description:"redis password"`
	DB       int    `long:"db" env:"DB" default:"0" description:"redis db"`
}

// KeyringGroup keyring config options
type KeyringGroup struct {
	Dir string `long:"dir" env:"DIR" default:"/home/obada/keyring" description:"keyring directory"`
}

// AuthGroup auth config options
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
	RPCURL  string `long:"rpc-url" env:"RPC_URL" description:"" default:"tcp://52.206.218.105:26657"`
	GrpcURL string `long:"grpc-url" env:"GRPC_URL" description:"" default:"52.206.218.105:9090"`
}

// IPFSGroup defines options for connection to the IPFS node
type IPFSGroup struct {
	RPCURL string `long:"url" env:"RPC_URL" description:"IPFS RPC url to connect"`
}

// RegistryGroup defines options for connection to the OBADA DID registry
type RegistryGroup struct {
	GrpcURL string `long:"url" env:"URL" description:"Registry HTTP URL"`
	HTTPUrl string `long:"http-url" env:"HTTP_URL" description:"Registry HTTP API URL"`
}

// Execute is the entry point for "server" command, called by flag parser
func (s *ServerCommand) Execute(_ []string) error {
	s.Logger.Infow("startup", "status", "initializing API support")

	ctx := context.Background()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Initialize event bus
	eventBus, err := bus.NewBus()
	if err != nil {
		return fmt.Errorf("initialize event bus: %w", err)
	}

	nodeClient, err := obadanode.NewClient(
		ctx,
		s.Node.ChainID,
		s.Node.RPCURL,
		s.Node.GrpcURL,
	)
	if err != nil {
		return fmt.Errorf("initialize node connection: %w", err)
	}

	// Validator package
	validator, err := validate.NewValidator()
	if err != nil {
		return err
	}

	kr, err := keyring.New("client-helper", keyring.BackendTest, s.Keyring.Dir, nil, testutil.MakeTestEncodingConfig().Codec)
	if err != nil {
		return fmt.Errorf("creating keyring error: %w", err)
	}

	accountSvc := account.NewService(validator, s.DB, nodeClient, kr, eventBus)
	blockchainSvc := blockchain.NewService(nodeClient, s.Logger, s.Registry.HTTPUrl)

	// IPFS client init
	ipfsShell := ipfs.NewIPFS(s.IPFS.RPCURL)

	conn, err := grpc.Dial(s.Registry.GrpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("creating registry client: %w", err)
	}

	// Registry init
	regClient := registry.NewClient(conn)

	deviceSvc := device.NewService(device.Config{
		Validator: validator,
		DB:        s.DB,
		IPFS:      ipfsShell,
		Bus:       eventBus,
		Registry:  regClient,
	})

	obitSvc := services.NewObitService(s.Logger)

	ks, err := pubkey.NewFS(s.Auth.KeysFolder)
	if err != nil {
		return fmt.Errorf("reading keys: %w", err)
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn: s.SentryDSN,
	})
	if err != nil {
		return fmt.Errorf("sentry.Init: %w", err)
	}

	// Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     s.Redis.Addr,
		Password: s.Redis.Password,
		DB:       s.Redis.DB,
	})

	// Initialize system events
	handlers.Initialize(handlers.Config{
		Bus:         eventBus,
		Logger:      s.Logger,
		RedisClient: rdb,
		Registry:    regClient,
		Keyring:     kr,

		DeviceSvc:     deviceSvc,
		BlockchainSvc: blockchainSvc,
	})

	// Auth manager verifies JWT tokens
	a, err := auth.New(auth.Config{
		Log:       s.Logger,
		KeyLookup: ks,
	})
	if err != nil {
		return err
	}

	wsClient, err := s.makeWsClient(ctx, wsClientConfig{
		nodeClient: nodeClient,
		deviceSvc:  deviceSvc,
		accountSvc: accountSvc,
		obitSvc:    obitSvc,
		bus:        eventBus,
	})
	if err != nil {
		return fmt.Errorf("cannot initialize node ws client: %w", err)
	}

	apiServer := s.makeAPIServer(api.APIMuxConfig{
		Shutdown: shutdown,
		Log:      s.Logger,
		Auth:     a,

		AccountSvc:    accountSvc,
		BlockchainSvc: blockchainSvc,
		DeviceSvc:     deviceSvc,
		ObitSvc:       obitSvc,
		Registry:      regClient,
	})

	serverErrors := make(chan error, 1)

	go func() {
		s.Logger.Infow("startup", "status", "api router started", "host", apiServer.Addr)
		serverErrors <- apiServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("api error: %w", err)

	case sig := <-shutdown:
		s.Logger.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer s.Logger.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		shutdownCtx, cancel := context.WithTimeout(ctx, s.ShutdownTimeout)
		defer cancel()

		if err := apiServer.Shutdown(shutdownCtx); err != nil {
			_ = apiServer.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}

		if err := wsClient.UnsubscribeAll(shutdownCtx); err != nil {
			return fmt.Errorf("cannot unsubscribe : %w", err)
		}

		if err := wsClient.Stop(); err != nil {
			return fmt.Errorf("could not stop node ws client gracefully: %w", err)
		}

		if err := s.DB.Close(); err != nil {
			return fmt.Errorf("could not close database: %w", err)
		}
	}

	return nil
}

func (s *ServerCommand) makeAPIServer(cfg api.APIMuxConfig) *http.Server {
	apiMux := api.APIMux(cfg)

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.Address, s.Port),
		Handler:      apiMux,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
		IdleTimeout:  s.IdleTimeout,
		ErrorLog:     zap.NewStdLog(s.Logger.Desugar()),
	}
}

type wsClientConfig struct {
	nodeClient obadanode.Client
	deviceSvc  *device.Service
	accountSvc *account.Service
	obitSvc    *services.ObitService
	bus        *eb.Bus
}

func (s *ServerCommand) makeWsClient(ctx context.Context, cfg wsClientConfig) (*jsonrpcclient.WSClient, error) {
	client, err := jsonrpcclient.NewWS(s.Node.RPCURL, "/websocket")
	if err != nil {
		return nil, err
	}

	if er := client.Start(); er != nil {
		return nil, er
	}

	if er := client.Subscribe(ctx, "tm.event = 'Tx'"); er != nil {
		return nil, er
	}

	go func() {
		for { //nolint:gosimple // ignore for now requires further refactoring
			select {
			case resp, ok := <-client.ResponsesCh:
				if ok {
					result := ctypes.ResultEvent{}

					if err = tmjson.Unmarshal(resp.Result, &result); err != nil {
						s.Logger.Errorw("unmarshal tx", "error", err)
					}

					for event, val := range result.Events {
						if event == "message.action" { //nolint:gocritic

							dataTx, ok := result.Data.(tmtypes.EventDataTx)
							if !ok {
								s.Logger.Error("decoding tx")
								break
							}

							tx, err := cfg.nodeClient.DecodeTx(dataTx.GetTx())
							if err != nil {
								s.Logger.Errorw("ecoding tx", "error", err)
								break
							}

							switch val[0] {
							case "mint_nft":
								for _, msg := range tx.GetMsgs() {
									msg, ok := msg.(*obadatypes.MsgMintNFT)
									if ok {
										s.Logger.Infow("obit metadata were updated", "data", result)

										// for future refactoring
										_ = cfg.bus.Emit(ctx, events.NftMinted, msg.Id)

										s.Logger.Infow("obit metadata were updated", "data", msg)
									}
								}
								s.Logger.Infow("obit was minted", "data", result.Data)
							case "update_uri_hash":
								for _, msg := range tx.GetMsgs() {
									msg, ok := msg.(*obadatypes.MsgUpdateUriHash)
									if ok {
										// for future refactoring
										_ = cfg.bus.Emit(ctx, events.NftMetadataUpdated, msg.Id)

										nft, err := cfg.nodeClient.GetNFT(ctx, msg.Id)
										if err != nil {
											s.Logger.Errorw("cannot get NFT by DID", "did", msg.Id, "error", err)
											break
										}

										profileID, err := cfg.accountSvc.GetProfileByAddress(msg.Editor)
										if err != nil {
											s.Logger.Errorw("caanot find profile", "address", msg.Editor, "error", err)
											break
										}

										authCtx := auth.SetClaims(ctx, auth.Claims{UserID: profileID})

										if err := cfg.deviceSvc.ImportDevice(authCtx, *nft, msg.Editor); err != nil {
											s.Logger.Errorw("cannot import device updates", "msg", msg, "error", err)
										}

										s.Logger.Infow("obit metadata were updated", "data", msg)
									}
								}

							case "transfer_nft":
								for _, msg := range tx.GetMsgs() {
									msg, ok := msg.(*obadatypes.MsgTransferNFT)
									if ok {
										// for future refactoring
										_ = cfg.bus.Emit(ctx, events.NftTransfered, msg.Id)

										nft, err := cfg.nodeClient.GetNFT(ctx, msg.Id)
										if err != nil {
											s.Logger.Errorw("cannot get NFT by DID", "did", msg.Id, "error", err)
											break
										}

										profileID, err := cfg.accountSvc.GetProfileByAddress(msg.Receiver)
										if err != nil {
											s.Logger.Errorw("cannot find profile", "address", msg.Receiver, "error", err)
											break
										}

										authCtx := auth.SetClaims(ctx, auth.Claims{UserID: profileID})

										if err := cfg.deviceSvc.ImportDevice(authCtx, *nft, msg.Receiver); err != nil {
											s.Logger.Errorw("cannot import device updates", "msg", msg, "error", err)
											break
										}

										s.Logger.Infow("nft was received", "nft", nft.Id)
									}
								}
							}
						}
					}
				}
			}
		}
	}()

	return client, nil
}
