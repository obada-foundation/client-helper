package account_test

import (
	"fmt"
	"testing"

	"github.com/obada-foundation/common/testutil"
)

// nolint
var c *testutil.Container

func TestMain(m *testing.M) {
	var err error

	c, err = testutil.StartBlockchain("")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer testutil.StopBlockchain(nil, c)

	m.Run()
}
