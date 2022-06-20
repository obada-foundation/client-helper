package properties

import (
	"fmt"
	"github.com/obada-foundation/sdkgo/hash"
	"log"
	"strconv"
)

// KvCollection representing slice of records and their hash
type KvCollection struct {
	records []Record
	hash    hash.Hash
}

// Record of key/value
type Record struct {
	key   StringProperty
	value StringProperty
	hash  hash.Hash
}

// KV represents key/value mapping to Obit KV collection record.
type KV struct {
	Key   string
	Value string
}

// NewKVCollection Holds KV records and compute hash of those records
func NewKVCollection(description string, kvs []KV, logger *log.Logger, debug bool) (KvCollection, error) {
	var p KvCollection
	var kvDec uint64

	if debug {
		logger.Printf("\n <|%s|> => NewKVCollection(%v)", description, kvs)
	}

	for _, kv := range kvs {
		description = "\t" + description + " :: creating key/value record"

		r, err := NewRecord(description, kv.Key, kv.Value, logger, debug)

		if err != nil {
			return p, err
		}

		rh := r.GetHash()
		kvDec += rh.GetDec()

		p.records = append(p.records, r)
	}

	h, err := hash.NewHash([]byte(strconv.FormatUint(kvDec, 10)), logger, debug)

	if err != nil {
		return p, fmt.Errorf("cannot hash %q: %w", kvDec, err)
	}

	p.hash = h

	return p, nil
}

// GetAll returns slice of records
func (kvcp *KvCollection) GetAll() []Record {
	return kvcp.records
}

// GetHash returns hash of all records
func (kvcp *KvCollection) GetHash() hash.Hash {
	return kvcp.hash
}

// NewRecord creates a new key/value record
func NewRecord(description, key, value string, logger *log.Logger, debug bool) (Record, error) {
	var r Record

	if debug {
		logger.Printf("\n |%s| => NewRecord(%q, %q)", description, key, value)
	}

	k, err := NewStringProperty("New record key", key, logger, debug)

	if err != nil {
		return r, err
	}

	v, err := NewStringProperty("New record value", value, logger, debug)

	if err != nil {
		return r, err
	}

	kh := k.GetHash()
	vh := v.GetHash()
	kvDec := kh.GetDec() + vh.GetDec()

	if debug {
		logger.Printf("(%d + %d) -> %d", kh.GetDec(), vh.GetDec(), kvDec)
	}

	h, err := hash.NewHash([]byte(strconv.FormatUint(kvDec, 10)), logger, debug)

	if err != nil {
		return r, err
	}

	r.key = k
	r.value = v
	r.hash = h

	return r, nil
}

// GetKey returns a record key
func (r *Record) GetKey() StringProperty {
	return r.key
}

// GetValue returns a record value
func (r *Record) GetValue() StringProperty {
	return r.value
}

// GetHash returns a record hash
func (r *Record) GetHash() hash.Hash {
	return r.hash
}
