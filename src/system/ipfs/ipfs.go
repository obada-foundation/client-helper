package ipfs

import (
	"bytes"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
)

type IPFS struct {
	sh *shell.Shell
}

func NewIPFS(URL string) *IPFS {
	return &IPFS{
		shell.NewShell(URL),
	}
}

func (s *IPFS) CreateDocument(data []byte) (string, error) {
	bytes.NewBuffer(data)

	cid, err := s.sh.Add(bytes.NewBuffer(data))
	if err != nil {
		return "", errors.Wrap(err, "cannot submit document to IPFS")
	}

	return "http://localhost:8084/ipfs/" + cid, nil
}
