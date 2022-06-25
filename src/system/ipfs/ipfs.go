package ipfs

import (
	"context"
	"fmt"
	"log"

	shell "github.com/ipfs/go-ipfs-api"
	files "github.com/ipfs/go-ipfs-files"
)

type IPFS struct {
	sh *shell.Shell
}

func NewIPFS(rpcURL string) *IPFS {
	return &IPFS{
		sh: shell.NewShell(rpcURL),
	}
}

func (s *IPFS) CreateDocument(ctx context.Context, DID, fileName string, data []byte) (string, error) {
	dir := files.NewSliceDirectory([]files.DirEntry{
		files.FileEntry(fileName, files.NewBytesFile(data)),
	})

	reader := files.NewMultiFileReader(dir, true)

	path := fmt.Sprintf("/%s/%s", DID, fileName)

	// See info: https://docs.ipfs.io/reference/http/api/#api-v0-files-write
	err := s.sh.Request("files/write").
		Arguments(path).
		Body(reader).
		// Make parent directories as needed
		Option("parents", true).
		// Use raw blocks for newly created leaf nodes
		Option("raw-leaves", true).
		// Create the file if it does not exist
		Option("create", true).
		// CID version
		Option("cid-version", 0).
		Exec(ctx, nil)

	if err != nil {
		return "", err
	}

	stat, err := s.sh.FilesStat(ctx, path)
	if err != nil {
		return "", err
	}

	log.Printf("File %+v", stat)

	return stat.Hash, err
}
