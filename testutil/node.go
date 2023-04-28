package testutil

import (
	"fmt"
	"testing"
	"time"

	commonts "github.com/obada-foundation/common/testutil"
)

const dockerImage = "obada/fullcore"

// StartBlockchain starts OBADA blockchain node instance
func StartBlockchain(tag string) (*commonts.Container, error) {
	if tag == "" {
		tag = "develop-testnet"
	}

	args := []string{}
	image := fmt.Sprintf("%s:%s", dockerImage, tag)

	c, err := commonts.StartContainer(image, []string{"26657"}, args...)
	if err != nil {
		return nil, err
	}

	time.Sleep(7 * time.Second)

	return c, nil
}

// StopBlockchain stops a running OBADA node instance.
func StopBlockchain(t *testing.T, c *commonts.Container) {
	if err := commonts.StopContainer(c.ID); err != nil {
		t.Logf("ERROR: cannot stop blockchain node container: %s\n %+v", err, c)
	}

	fmt.Println("Stopped:", c.ID)
}
