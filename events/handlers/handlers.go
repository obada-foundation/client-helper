package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/mustafaturan/bus/v3"
	"github.com/obada-foundation/client-helper/events"
	"github.com/obada-foundation/client-helper/services/blockchain"
	"github.com/obada-foundation/client-helper/services/device"
	pbacc "github.com/obada-foundation/registry/api/pb/v1/account"
	registry "github.com/obada-foundation/registry/client"
	"github.com/obada-foundation/sdkgo/base58"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config handlers config
type Config struct {
	Bus           *bus.Bus
	Logger        *zap.SugaredLogger
	RedisClient   *redis.Client
	DeviceSvc     *device.Service
	BlockchainSvc *blockchain.Service
	Registry      registry.Client
	Keyring       keyring.Keyring
}

// EventManager manages ClientHelper events
type EventManager struct {
	b             *bus.Bus
	logger        *zap.SugaredLogger
	redis         *redis.Client
	deviceSvc     *device.Service
	blockchainSvc *blockchain.Service
	registry      registry.Client
	kr            keyring.Keyring
}

// Initialize initializes handlers
func Initialize(cfg Config) {
	manager := &EventManager{
		b:             cfg.Bus,
		logger:        cfg.Logger,
		redis:         cfg.RedisClient,
		deviceSvc:     cfg.DeviceSvc,
		blockchainSvc: cfg.BlockchainSvc,
		registry:      cfg.Registry,
		kr:            cfg.Keyring,
	}

	manager.RegisterEvents()
	manager.RegisterHandlers()
}

// RegisterEvents registers events
func (em EventManager) RegisterEvents() {
	em.b.RegisterTopics(
		// Account events
		events.AccountDeleted,
		events.AccountCreated,

		// Device events
		events.DeviceSaved,

		// NFT events
		events.NftMinted,
		events.NftTransfered,
		events.NftMetadataUpdated,
	)
}

func (em EventManager) accountDeletedHandler(ctx context.Context, e bus.Event) {
	accAddress := fmt.Sprintf("%v", e.Data)

	if _, err := em.deviceSvc.DeleteByAddress(ctx, accAddress); err != nil {
		em.logger.Errorw("failed to delete devices", "EVENT", events.AccountDeleted, "address", accAddress, "error", err)
		return
	}

	em.redis.Publish(ctx, events.AccountDeleted, accAddress)

	em.logger.Infow("account deleted", "EVENT", events.AccountDeleted, "account address", accAddress)
}

func (em EventManager) accountCreatedHandler(ctx context.Context, e bus.Event) {
	accAddress := fmt.Sprintf("%v", e.Data)

	if _, err := em.registry.GetPublicKey(ctx, &pbacc.GetPublicKeyRequest{Address: accAddress}); err != nil {
		er, ok := status.FromError(err)
		if !ok || er.Code() != codes.NotFound {
			em.logger.Errorw("failed to check if account key is in the registry", "EVENT", events.AccountCreated, "address", accAddress, "error", err)
		}

		addr, _ := types.AccAddressFromBech32(accAddress)

		key, err := em.kr.KeyByAddress(addr)
		if err != nil {
			em.logger.Errorw("cannot find key in keyring", "EVENT", events.AccountCreated, "address", accAddress, "error", err)
		}

		if err == nil {
			pubKey, _ := key.GetPubKey()

			msg := &pbacc.RegisterAccountRequest{
				Pubkey: base58.Encode(pubKey.Bytes()),
			}

			if _, err := em.registry.RegisterAccount(ctx, msg); err != nil {
				em.logger.Errorw("failed to register account in the registry", "EVENT", events.AccountCreated, "address", accAddress, "error", err)
			}

			if err == nil {
				em.logger.Infow("added to the registry", "EVENT", events.AccountCreated, "account address", accAddress)
			}
		}
	}

	nfts, err := em.blockchainSvc.GetNFTByAddress(ctx, accAddress)
	if err != nil {
		em.logger.Errorw("failed to fetch nfts from blockchain", "EVENT", events.AccountCreated, "address", accAddress, "error", err)
		return
	}

	for _, NFT := range nfts {
		if err := em.deviceSvc.ImportDevice(ctx, NFT, accAddress); err != nil {
			em.logger.Errorw("failed to import nft", "EVENT", events.AccountCreated, "address", accAddress, "nft", NFT.Id, "error", err)
			continue
		}
	}

	em.logger.Infow("assets imported", "EVENT", events.AccountCreated, "account address", accAddress)
}

// RegisterHandlers registers handlers
func (em EventManager) RegisterHandlers() {

	em.b.RegisterHandler(events.AccountDeletedHandler, bus.Handler{
		Handle:  em.accountDeletedHandler,
		Matcher: events.AccountDeleted,
	})

	em.b.RegisterHandler(events.AccountCreatedHandler, bus.Handler{
		Handle:  em.accountCreatedHandler,
		Matcher: events.AccountCreated,
	})

	{
		h := bus.Handler{
			Handle: func(ctx context.Context, e bus.Event) {
				jsonData, err := json.Marshal(e.Data)
				if err != nil {
					em.logger.Errorw("failed to marshal json data", "EVENT", events.DeviceSaved, "error", err)
					return
				}

				em.redis.Publish(ctx, events.DeviceSaved, string(jsonData))

				em.logger.Infow("device saved", "EVENT", events.DeviceSaved, "device", e.Data)
			},
			Matcher: events.DeviceSaved,
		}

		em.b.RegisterHandler(events.DeviceSavedHandler, h)
	}

	// NFT event handlers
	{
		h := bus.Handler{
			Handle: func(ctx context.Context, e bus.Event) {
				DID := fmt.Sprintf("%v", e.Data)

				em.redis.Publish(ctx, events.NftMinted, DID)

				em.logger.Infow("nft minted", "EVENT", events.NftMinted, "DID", DID)
			},
			Matcher: events.NftMinted,
		}

		em.b.RegisterHandler(events.NftMintedHandler, h)
	}
	{
		h := bus.Handler{
			Handle: func(ctx context.Context, e bus.Event) {
				DID := fmt.Sprintf("%v", e.Data)

				em.redis.Publish(ctx, events.NftMetadataUpdated, DID)
			},
			Matcher: events.NftMetadataUpdated,
		}

		em.b.RegisterHandler(events.NftMetadataUpdatedHandler, h)
	}
	{
		h := bus.Handler{
			Handle: func(ctx context.Context, e bus.Event) {
				DID := fmt.Sprintf("%v", e.Data)

				em.redis.Publish(ctx, events.NftTransfered, DID)

				em.logger.Infow("nft received", "EVENT", events.NftTransfered, "DID", DID)
			},
			Matcher: events.NftTransfered,
		}

		em.b.RegisterHandler(events.NftTransferedHandler, h)
	}
}
