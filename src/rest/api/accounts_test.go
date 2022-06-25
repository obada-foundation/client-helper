package api

import (
	"context"
	"strings"

	"github.com/obada-foundation/client-helper/services/account"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type JSON map[string]interface{}

func TestAccount_create(t *testing.T) {
	ts, _, teardown := startupT(t)
	defer teardown()

	resp, err := postWithAuth(
		t,
		ts.URL+"/api/v1/accounts", `{"email":"john.doe@supermail.com"}`,
	)
	assert.NoError(t, err)

	b, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode, string(b))
	assert.NoError(t, resp.Body.Close())

	c := JSON{}
	err = json.Unmarshal(b, &c)
	assert.NoError(t, err)

	assert.Equal(t, "3", c["id"])
	assert.Equal(t, "john.doe@supermail.com", c["email"])

	resp, err = postWithAuth(
		t,
		ts.URL+"/api/v1/accounts", `{"email":"john.doe@supermail.com"}`,
	)

	b, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode, string(b))
	assert.NoError(t, resp.Body.Close())

	err = json.Unmarshal(b, &c)
	assert.NoError(t, err)

	assert.Equal(t, "account already exists", c["error"])
}

func TestAccount_balance(t *testing.T) {
	ts, srv, teardown := startupT(t)
	defer teardown()

	resp, err := getWithAuth(t, ts.URL+"/api/v1/accounts/my-balance")
	assert.NoError(t, err)

	b, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	require.Equal(t, http.StatusNotFound, resp.StatusCode, string(b))
	assert.NoError(t, resp.Body.Close())

	ctx := context.Background()

	_, err = srv.AccountService.Create(ctx, account.NewAccount{
		ID:    "3",
		Email: "foo@bar.com",
	})

	assert.NoError(t, err)

	resp, err = getWithAuth(t, ts.URL+"/api/v1/accounts/my-balance")
	assert.NoError(t, err)

	b, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode, string(b))
	assert.NoError(t, resp.Body.Close())

	c := JSON{}
	err = json.Unmarshal(b, &c)
	assert.NoError(t, err)

	assert.Equal(t, 0, int(c["balance"].(float64)))
	assert.True(t, strings.HasPrefix(c["address"].(string), "obada"))
}
