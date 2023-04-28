package handlers_test

import (
	"context"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-redis/redismock/v9"
	"github.com/golang/mock/gomock"
	"github.com/mustafaturan/bus/v3"
	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/events"
	"github.com/obada-foundation/client-helper/events/handlers"
	svcs "github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/blockchain"
	"github.com/obada-foundation/client-helper/services/device"
	ipfsclinet "github.com/obada-foundation/client-helper/system/ipfs/mocks"
	"github.com/obada-foundation/client-helper/system/obadanode/mocks"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/obada-foundation/client-helper/testutil"
	obadatypes "github.com/obada-foundation/fullcore/x/obit/types"
	"github.com/obada-foundation/registry/api/pb/v1/account"
	"github.com/obada-foundation/registry/api/pb/v1/diddoc"
	regclient "github.com/obada-foundation/registry/client/mock"
	"github.com/obada-foundation/sdkgo/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tm-db"
)

func GenKeys(t *testing.T) (cryptotypes.PrivKey, cryptotypes.PubKey, string) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr, err := sdk.AccAddressFromHexUnsafe(pubKey.Address().String())
	require.NoError(t, err)

	return privKey, pubKey, addr.String()
}

func Test_EventManager(t *testing.T) {
	t.Parallel()

	t.Run("accountDeletedHandler", accountDeletedHandler)
	t.Run("accountCreatedHandler", accountCreatedHandler)
}

// nolint:gocritic
func accountCreatedHandler(t *testing.T) {
	b, deviceSvc, nodeClientMock, regClient, ipfs, teardown := startupT(t)
	defer teardown()

	ctx := context.Background()
	ctx = auth.SetClaims(ctx, auth.Claims{
		UserID: "1",
	})

	accountAddress := "obada1yxxnd624tgwqm3eyv5smdvjrrydfh9h943qptg"

	data, err := codectypes.NewAnyWithValue(&obadatypes.NFTData{
		Usn: "25rc8AxGbLSr",
	})
	require.NoError(t, err)

	nfts := append(make([]obadatypes.NFT, 0, 1), obadatypes.NFT{
		ClassId: "OBD",
		Id:      "did:obada:64925be84b586363670c1f7e5ada86a37904e590d1f6570d834436331dd3eb88",
		Uri:     "",
		UriHash: "",
		Data:    data,
	})

	nodeClientMock.On("GetNFTByAddress", mock.Anything, accountAddress).
		Return(nfts, nil).
		Once()

	ipfs.On("GetDocument", "bafkreibdklgsqwqv5xci6cmx46j2y3to5y5sqbuoqz7g2qpgjthzgwrz4i").
		Return(
			[]byte(`{"serial_number":"SN123456X", "manufacturer":"Sony", "part_number":"PN123456S"}`),
			nil,
		).Once()

	regClient.EXPECT().GetPublicKey(gomock.Any(), gomock.Eq(&account.GetPublicKeyRequest{Address: accountAddress})).Times(1).
		Return(&account.GetPublicKeyResponse{}, nil)

	regClient.EXPECT().Get(gomock.Any(), gomock.Eq(&diddoc.GetRequest{Did: nfts[0].Id})).Times(1).
		Return(&diddoc.GetResponse{
			Document: &diddoc.DIDDocument{
				Id: nfts[0].Id,
				Metadata: &diddoc.Metadata{
					Objects: append(make([]*diddoc.Object, 0, 1), &diddoc.Object{
						Url: "ipfs://bafkreibdklgsqwqv5xci6cmx46j2y3to5y5sqbuoqz7g2qpgjthzgwrz4i",
						Metadata: map[string]string{
							"type": string(asset.PhysicalAssetIdentifiers),
							"name": string(asset.PhysicalAssetIdentifiers),
						},
					}),
				},
			},
		}, nil)

	err = b.Emit(ctx, events.AccountCreated, accountAddress)
	require.NoError(t, err)

	devices, err := deviceSvc.GetByAddress(ctx, accountAddress)
	require.NoError(t, err)

	d, err := deviceSvc.GetByUSN(ctx, "25rc8AxGbLSr")
	require.NoError(t, err)

	assert.Equal(t, "25rc8AxGbLSr", d.Usn)
	assert.Equal(t, 1, len(d.Documents))
	assert.Equal(t, 1, len(devices))
}

// nolint:gocritic
func accountDeletedHandler(t *testing.T) {
	b, deviceSvc, _, regClient, ipfs, teardown := startupT(t)

	ctx := context.Background()
	ctx = auth.SetClaims(ctx, auth.Claims{
		UserID: "1",
	})

	ipfs.On("CreateDocument", mock.Anything, true).Return("", nil).Twice()

	prKey, _, accountAddress := GenKeys(t)

	regClient.EXPECT().Get(ctx, gomock.Any()).Times(4).
		Return(nil, nil)

	regClient.EXPECT().SaveMetadata(gomock.Any(), gomock.Any()).Times(2).
		Return(nil, nil)

	_, err := deviceSvc.Save(ctx, svcs.SaveDevice{
		SerialNumber: "SN123456",
		Manufacturer: "IBM",
		PartNumber:   "PN123456",
		Address:      accountAddress,
	}, prKey)
	require.NoError(t, err, "Cannot save device")

	_, err = deviceSvc.Save(ctx, svcs.SaveDevice{
		SerialNumber: "SN123457",
		Manufacturer: "IBM",
		PartNumber:   "PN123456",
		Address:      accountAddress,
	}, prKey)
	require.NoError(t, err, "Cannot save device")

	devices, err := deviceSvc.GetByAddress(ctx, accountAddress)
	require.NoError(t, err)

	assert.Equal(t, 2, len(devices))

	err = b.Emit(ctx, events.AccountDeleted, accountAddress)
	require.NoError(t, err)

	devices, err = deviceSvc.GetByAddress(ctx, accountAddress)
	require.NoError(t, err)

	assert.Equal(t, 0, len(devices))

	defer teardown()
}

// nolint:gocritic
func startupT(t *testing.T) (*bus.Bus, *device.Service, *mocks.Client, *regclient.MockClient, *ipfsclinet.IPFS, func()) {
	var fn bus.Next = func() string { return "afakeid" }
	b, err := bus.NewBus(fn)
	require.NoError(t, err, "Cannot initialize event bus")

	logger, lgDefer := testutil.MakeLoger()

	database, err := db.NewDB("client-helper-test", db.MemDBBackend, ".")
	require.NoError(t, err)

	validator, err := validate.NewValidator()
	require.NoError(t, err)

	ipfs := &ipfsclinet.IPFS{}

	ctrl := gomock.NewController(t)
	mockclient := regclient.NewMockClient(ctrl)

	deviceSvc := device.NewService(device.Config{
		Validator: validator,
		DB:        database,
		IPFS:      ipfs,
		Bus:       b,
		Registry:  mockclient,
	})

	nodeClient := &mocks.Client{}

	blockchainSvc := blockchain.NewService(nodeClient, logger)

	dbs, _ := redismock.NewClientMock()

	handlers.Initialize(handlers.Config{
		Bus:         b,
		Logger:      logger,
		RedisClient: dbs,

		// Services
		DeviceSvc:     deviceSvc,
		BlockchainSvc: blockchainSvc,
		Registry:      mockclient,
	})

	teardown := func() {
		lgDefer()
	}

	return b, deviceSvc, nodeClient, mockclient, ipfs, teardown
}
