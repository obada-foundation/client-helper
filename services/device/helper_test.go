package device_test

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/mustafaturan/bus/v3"
	"github.com/obada-foundation/client-helper/events"
	"github.com/obada-foundation/client-helper/services/device"
	"github.com/obada-foundation/client-helper/system/validate"
	registryclient "github.com/obada-foundation/registry/client/mock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tm-db"
)

// nolint
func init() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("obada", "obada"+sdk.PrefixPublic)
	config.Seal()
}

// nolint
type IPFSTestClient struct {
}

// nolint
func (c IPFSTestClient) CreateDocument(data []byte, saveDocument bool) (string, error) {
	return "", nil
}

// nolint
func (c IPFSTestClient) GetDocument(cid string) ([]byte, error) {
	return nil, nil
}

func createTestService(t *testing.T) (*device.Service, *registryclient.MockClient, context.Context, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	v, err := validate.NewValidator()
	require.NoError(t, err, "Cannot initialize validation")

	d, err := db.NewDB("devices", db.MemDBBackend, "./testdb")
	require.NoError(t, err, "Cannot initialize database")

	ipfs := &IPFSTestClient{}

	var fn bus.Next = func() string { return "afakeid" }
	b, err := bus.NewBus(fn)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	mockclient := registryclient.NewMockClient(ctrl)

	b.RegisterTopics(events.DeviceSaved)

	tearDown := func() {
		d.Close()
		cancel()
	}

	return device.NewService(device.Config{
		Validator: v,
		DB:        d,
		IPFS:      ipfs,
		Bus:       b,
		Registry:  mockclient,
	}), mockclient, ctx, tearDown
}
