package account_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
)

func TestCoinConvert(t *testing.T) {
	coin := types.NewCoin("rohi", sdkmath.NewInt(1000000000000000000))

	t.Log(coin)
}
