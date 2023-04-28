package ipfs

import (
	"bytes"
	"fmt"

	shell "github.com/ipfs/go-ipfs-api"
)

// IPFS is an interface for IPFS client
type IPFS interface {
	GetDocument(cid string) ([]byte, error)
	CreateDocument(data []byte, saveDocument bool) (string, error)
}

// Client is an implementation of IPFS client
type Client struct {
	sh *shell.Shell
}

// Create is a helper function to create a shell.AddOption for the create flag
func Create(enabled bool) shell.AddOpts {
	return func(rb *shell.RequestBuilder) error {
		rb.Option("create", enabled)
		return nil
	}
}

// NewIPFS creates a new instance of IPFS client
func NewIPFS(rpcURL string) IPFS {
	return &Client{
		sh: shell.NewShell(rpcURL),
	}
}

// GetDocument returns the document from IPFS by CID
func (c Client) GetDocument(cid string) ([]byte, error) {
	data, err := c.sh.Cat(cid)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// CreateDocument creates a new document in IPFS
func (c Client) CreateDocument(data []byte, saveDocument bool) (string, error) {
	bytes.NewBuffer(data)

	cid, err := c.sh.Add(
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
