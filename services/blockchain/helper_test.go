package blockchain_test

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/obada-foundation/client-helper/services/blockchain"
	"github.com/obada-foundation/client-helper/system/logger"
	"github.com/obada-foundation/client-helper/system/obadanode"
	"github.com/obada-foundation/common/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoinits //needed for the test
func init() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("obada", "obada"+sdk.PrefixPublic)
	config.Seal()
}

func createTestService(t *testing.T) (*blockchain.Service, func()) { //nolint: gocritic
	ctx := context.Background()

	c, err := testutil.StartBlockchain("")
	require.NoError(t, err, "Cannot start blockchain container")

	nodeClient, err := obadanode.NewClient(
		ctx,
		"obada-testnet",
		fmt.Sprintf("tcp://%s:%d", c.Host, c.Ports["26657"]),
		fmt.Sprintf("%s:%d", c.Host, c.Ports["9090"]),
	)
	require.NoError(t, err, "Cannot initialize OBADA Node client")

	lgr, err := logger.New("BLOCKCHAIN-SERVICE-TEST")
	assert.NoError(t, err)

	return blockchain.NewService(&nodeClient, lgr, ""), func() {
		testutil.StopBlockchain(t, c)
	}
}
