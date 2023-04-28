package blockchain_test

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/blockchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const receiverAddress string = "obada12fhqkqnednplwcvdz5hstzzrzn9348n3eaxj38"

type tests struct {
	service *blockchain.Service
	ctx     context.Context
}

func TestService(t *testing.T) {
	t.Parallel()

	service, deferFn := createTestService(t)
	defer deferFn()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tests := tests{
		service: service,
		ctx:     ctx,
	}

	t.Run("testSend", tests.testSend)
	t.Run("testMintNFT", tests.testMintNFT)
	t.Run("testTransferNFT", tests.testTransferNFT)
	t.Run("testGetNFTByAddress", tests.testGetNFTByAddress)
}

func (ts tests) testSend(t *testing.T) {
	privKey := secp256k1.GenPrivKey()

	accAddress := sdk.AccAddress(privKey.PubKey().Address().Bytes()).String()

	t.Log("Test coin send from account with zero tx")
	account := services.Account{
		Address: accAddress,
	}

	err := ts.service.Send(ts.ctx, account, receiverAddress, "1obd", privKey)
	require.ErrorIs(t, err, blockchain.ErrInsufficientFunds)
}

func (ts tests) testMintNFT(t *testing.T) {
	privKey := secp256k1.GenPrivKey()

	t.Log("Test miniting NFT from account with zero tx")
	account := services.Device{
		Address: receiverAddress,
	}

	err := ts.service.MintNFT(ts.ctx, account, privKey)
	require.ErrorIs(t, err, blockchain.ErrInsufficientFunds)
}

func (ts tests) testTransferNFT(t *testing.T) {
	privKey := secp256k1.GenPrivKey()

	t.Log("Test transferring NFT to account with zero tx")

	err := ts.service.TransferNFT(ts.ctx, "did:obada:12345", receiverAddress, privKey)
	require.ErrorIs(t, err, blockchain.ErrInsufficientFunds)
}

func (ts tests) testGetNFTByAddress(t *testing.T) {
	nfts, err := ts.service.GetNFTByAddress(ts.ctx, "obada1wup66kj5gq0nv0u4ttkn9gfneq9wx3475044p8")
	require.NoError(t, err)

	assert.Equal(t, 0, len(nfts))

}
