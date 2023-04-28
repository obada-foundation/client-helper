package account_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
)

func TestCoinConvert(t *testing.T) {
	if err := types.RegisterDenom("rohi", types.NewDec(1)); err != nil {
		t.Logf("%+v", err)
	}

	if err := types.RegisterDenom("obd", types.NewDec(1000000)); err != nil {
		t.Logf("%+v", err)
	}

	coin1 := types.NewDecCoin("obd", types.NewInt(1))
	coin2, err := types.ConvertDecCoin(coin1, "rohi")
	if err != nil {
		t.Logf("%+v", err)
	}

	val, err := coin2.Amount.Float64()
	if err != nil {
		t.Logf("%+v", err)
	}
	t.Logf("%.2f", val)

	coin3 := types.NewDecCoin("rohi", math.NewInt(3653))
	coin4, err := types.ConvertDecCoin(coin3, "obd")
	if err != nil {
		t.Logf("%+v", err)
	}

	val, err = coin4.Amount.Float64()
	if err != nil {
		t.Logf("%+v", err)
	}

	t.Logf("%.6f", val)

	coin5, err := types.ParseDecCoin("1.1obd")
	if err != nil {
		t.Logf("%+v", err)
	}
	coin6, err := types.ConvertDecCoin(coin5, "rohi")
	if err != nil {
		t.Logf("%+v", err)
	}

	t.Logf("%s", coin6.Amount)
}
