package account_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

func TestCoinConvert(t *testing.T) {
	coin := types.NewCoin("rohi", types.NewInt(1000000000000000000))

	t.Log(coin)
}
