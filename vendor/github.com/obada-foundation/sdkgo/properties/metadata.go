package properties

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/obada-foundation/sdkgo/hash"
)

// Metadata stores metadata records
type Metadata struct {
	logger  *log.Logger
	mu      sync.RWMutex
	records map[string]string
}

// NewMetadata creates new metadata records store
func NewMetadata(m map[string]string, logger *log.Logger) (*Metadata, error) {
	metadata := &Metadata{
		logger:  logger,
		records: make(map[string]string, len(m)),
	}

	for key, value := range m {
		if err := metadata.AddRecord(key, value); err != nil {
			return nil, err
		}
	}

	return metadata, nil
}

// AddRecord sets new metadata key/value pair
func (m *Metadata) AddRecord(key, value string) error {
	if key == "" {
		return ErrEmptyMetadataKey
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.records[key] = value

	return nil
}

// ToJSON returns JSON bytes
func (m *Metadata) ToJSON() ([]byte, error) {
	b, err := json.Marshal(m.records)
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}

// String returns JSON as a string
func (m *Metadata) String() (string, error) {
	jsonBytes, err := m.ToJSON()
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// Hash sum of metada records hashes
func (m *Metadata) Hash() (*hash.Hash, error) {
	recordHashes := make([]hash.Hash, 0)

	for key, value := range m.records {
		metaKeyHash, err := NewStringProperty("Metadata key", key, m.logger)
		if err != nil {
			return nil, err
		}

		metaKeyValue, err := NewStringProperty("Metadata value", value, m.logger)
		if err != nil {
			return nil, err
		}

		recordSum := hash.SumHashes(m.logger, metaKeyHash.GetHash(), metaKeyValue.GetHash())

		h, err := hash.NewHash([]byte(fmt.Sprintf("%x", recordSum)), m.logger)
		if err != nil {
			return nil, err
		}

		recordHashes = append(recordHashes, h)
	}

	h, err := hash.NewHash([]byte(fmt.Sprintf("%x", hash.SumHashes(m.logger, recordHashes...))), m.logger)
	if err != nil {
		return nil, err
	}

	return &h, nil
}
