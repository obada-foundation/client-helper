package encoder

import (
	"bytes"
	"encoding/gob"
)

// DataEncode encodes data to bytes
func DataEncode(data interface{}) ([]byte, error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)

	if err := e.Encode(data); err != nil {
		return b.Bytes(), err
	}

	return b.Bytes(), nil
}
