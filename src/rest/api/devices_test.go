package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestDevice_save(t *testing.T) {
	ts, _, teardown := startupT(t)
	defer teardown()

	resp, err := postWithAuth(
		t,
		ts.URL+"/api/v1/obits", `{"serial_number":"SN123456", "manufacturer":"IBM", "part_number": "PN123456"}`,
	)
	assert.NoError(t, err)

	b, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode, string(b))
	assert.NoError(t, resp.Body.Close())

	c := JSON{}
	err = json.Unmarshal(b, &c)
	assert.NoError(t, err)

	assert.Equal(t, "did:obada:cc65aa0f47463a5d0488fd758a2524f8bd24981619afc319c8d7b67eeb75c468", c["did"])
	assert.Equal(t, "2zFXFZ3b", c["usn"])
	assert.Equal(t, "bcba7830a0d3b66ed8c5485b60ccc4f09ec50fbe82ff5ede9c3f39819f724af0", c["checksum"])
	assert.Equal(t, "PN123456", c["part_number"])
	assert.Equal(t, "IBM", c["manufacturer"])

}
