package api

import (
	"crypto/sha256"
	"sort"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtoDeterministicChecksum returns a deterministic checksum of the given proto message.
func ProtoDeterministicChecksum(m proto.Message) ([32]byte, error) {
	marshaler := protojson.MarshalOptions{
		Indent:          "  ",
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	sortedBytes, err := marshaler.Marshal(m)
	if err != nil {
		return [32]byte{}, err
	}

	// Sort the JSON bytes
	sort.Slice(sortedBytes, func(i, j int) bool {
		return sortedBytes[i] < sortedBytes[j]
	})

	// Hash the serialized message
	hash := sha256.Sum256(sortedBytes)

	return hash, nil
}
