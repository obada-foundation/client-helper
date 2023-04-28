package wallet_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/obada-foundation/client-helper/services/wallet"
	"github.com/obada-foundation/client-helper/system/logger"
	"github.com/obada-foundation/client-helper/system/obadanode"
	"github.com/obada-foundation/common/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tm-db"
)

func TestService(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("obada", "obada"+sdk.PrefixPublic)
	config.Seal()

	c, err := testutil.StartBlockchain("")
	require.NoError(t, err, "Cannot start blockchain container")
	defer testutil.StopBlockchain(t, c)

	t.Log("Testing Wallet Service")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log, err := logger.New("TEST Keys Service")
	require.NoError(t, err, "Cannot initialize logger")

	database, err := db.NewDB("wallets", db.MemDBBackend, "./testdb")
	require.NoError(t, err, "Cannot initialize database")
	defer database.Close()

	nodeClient, err := obadanode.NewClient(
		ctx,
		"obada-testnet",
		fmt.Sprintf("tcp://%s:%d", c.Host, c.Ports["26657"]),
		fmt.Sprintf("%s:%d", c.Host, c.Ports["9090"]),
	)

	require.NoError(t, err, "Cannot OBADA Node client")
	defer nodeClient.Close()

	service := wallet.NewService(database, log)

	t.Log("Testing master key creation")

	accountID := "1"
	mnemonic := "raw tube meadow accident giraffe chase rotate desert tribe dish they chuckle focus harsh cattle net actual bulb virtual blanket supply staff split ribbon"
	kid := "my new master kid"

	masterKey, err := service.NewMasterKey(accountID, kid, mnemonic)
	require.NoError(t, err, "Cannot create master key")

	assert.Equal(t, "xprv9s21ZrQH143K3MgoVTvruPwMq9fsfZM2S8L5HwTNmazrpz6hKe9CuGQfimSJuVcE8CkLqk18gr2cyGYg95W3Fv9aGGriRN3b5dgzSpRUAME", masterKey.String())

	t.Log("Testing master that private key was created by default with private key")

	pKeys, err := service.GetAllPrivateKeys(accountID, kid)
	require.NoError(t, err, "Cannot get a private keys")

	for _, pKey := range pKeys {
		privateKey, er := service.GetPrivateKey(accountID, kid, pKey.PublicKey().String())
		require.NoError(t, er, "Cannot get a private key")
		assert.Equal(t, pKey, privateKey)
	}

	t.Log("Testing creation of private key")

	for idx := 1; idx <= 9; idx++ {
		newpKey, er := service.NewPrivateKey(accountID, kid, masterKey)
		require.NoError(t, er, "Cannot create a new private key")

		newpKeys, er := service.GetAllPrivateKeys(accountID, kid)
		require.NoError(t, er, "Cannot get private keys")

		assert.Equal(t, newpKey, newpKeys[idx])
	}

	t.Log("Testing getting all master keys")
	allMasterKeys, err := service.GetAllMasterKeys(accountID)
	require.NoError(t, err, "Cannot get all master keys")

	assert.Equal(t, 1, len(allMasterKeys))
}
