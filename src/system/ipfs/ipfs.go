package ipfs

import (
	"bytes"
	"fmt"

	shell "github.com/ipfs/go-ipfs-api"
)

type IPFS struct {
	sh *shell.Shell
}

func Create(enabled bool) shell.AddOpts {
	return func(rb *shell.RequestBuilder) error {
		rb.Option("create", enabled)
		return nil
	}
}

func NewIPFS(rpcURL string) *IPFS {
	return &IPFS{
		sh: shell.NewShell(rpcURL),
	}
}

func (s *IPFS) CreateDocument(data []byte, saveDocument bool) (string, error) {
	bytes.NewBuffer(data)

	cid, err := s.sh.Add(
		bytes.NewBuffer(data),
		shell.OnlyHash(!saveDocument),
		Create(saveDocument),
		shell.Pin(false),
		shell.RawLeaves(true),
	)
	if err != nil {
		return "", fmt.Errorf("cannot submit document to IPFS %w", err)
	}

	return cid, nil
}
