package account_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmostestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/mustafaturan/bus/v3"
	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/events"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/system/obadanode"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/stretchr/testify/require"
	db "github.com/tendermint/tm-db"
)

//nolint:gochecknoinits //requred for test
func init() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("obada", "obada"+sdk.PrefixPublic)
	config.Seal()
}

func createTestService(t *testing.T) (*bus.Bus, *account.Service, context.Context, func()) {
	v, err := validate.NewValidator()
	require.NoError(t, err, "Cannot initialize validation")

	database, err := db.NewDB("accounts", db.MemDBBackend, "./testdb")
	require.NoError(t, err, "Cannot initialize database")

	nodeClient, err := obadanode.NewClient(
		context.Background(),
		"obada-testnet",
		fmt.Sprintf("tcp://%s:%d", c.Host, c.Ports["26657"]),
		fmt.Sprintf("%s:%d", c.Host, c.Ports["9090"]),
	)
	require.NoError(t, err, "Cannot initialize OBADA Node client")

	defaultConfig := cosmostestutil.MakeTestEncodingConfig()

	kr := keyring.NewInMemory(defaultConfig.Codec)

	var fn bus.Next = func() string { return "afakeid" }
	b, err := bus.NewBus(fn)

	b.RegisterTopics(events.AccountDeleted, events.AccountCreated)

	require.NoError(t, err, "Cannot initialize event bus")

	service := account.NewService(v, database, &nodeClient, kr, b)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	ctx = auth.SetClaims(ctx, auth.Claims{
		UserID: "3",
	})

	_, err = makeProfile(t, service, ctx)
	require.NoError(t, err)

	deferFn := func() {
		cancel()
	}

	return b, service, ctx, deferFn
}
