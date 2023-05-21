package testutil

import (
	"fmt"
	"testing"
	"time"
)

const dockerImage = "obada/fullcore"

// StartBlockchain starts OBADA blockchain node instance
func StartBlockchain(tag string) (*Container, error) {
	if tag == "" {
		tag = "develop-testnet"
	}

	args := []string{}
	image := fmt.Sprintf("%s:%s", dockerImage, tag)

	c, err := StartContainer(image, []string{"26657", "9090"}, args...)
	if err != nil {
		return nil, err
	}

	time.Sleep(7 * time.Second)

	return c, nil
}

// StopBlockchain stops a running OBADA node instance.
func StopBlockchain(t *testing.T, c *Container) {
	if err := StopContainer(c.ID); err != nil {
		t.Logf("ERROR: cannot stop blockchain node container: %s\n %+v", err, c)
	}

	fmt.Println("Stopped:", c.ID)
}
