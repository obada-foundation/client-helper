package api

import (
	"crypto/sha256"
	"sort"

	"github.com/golang/protobuf/jsonpb"
	pbdiddoc "github.com/obada-foundation/registry/api/pb/v1/diddoc"
)

func MetadataDeterministicChecksum(data *pbdiddoc.SaveMetadataRequest_Data) ([32]byte, error) {
	marshaler := jsonpb.Marshaler{OrigName: true}
	jsonString, err := marshaler.MarshalToString(data)

	if err != nil {
		return [32]byte{}, err
	}

	// Sort the JSON bytes
	sortedBytes := []byte(jsonString)
	sort.Slice(sortedBytes, func(i, j int) bool {
		return sortedBytes[i] < sortedBytes[j]
	})

	// Hash the serialized message
	hash := sha256.Sum256(sortedBytes)

	return hash, nil
}
