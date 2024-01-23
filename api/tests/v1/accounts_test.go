package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	//nolint:misspell // Seed is corect
	defaultMnemonic string = "radio distance sweet artefact attack liar until video army raccoon green error ceiling size spread burst galaxy bottom cave rubber setup west address must"

	secondMnemonic = "decline maid slush umbrella fame manual prison collect custom usual reform used fatal shock size document smart admit soccer vast desert vivid where must"

	defaultAccount = "obada1yxxnd624tgwqm3eyv5smdvjrrydfh9h943qptg"

	defaultObadaPrivateKey = `
-----BEGIN TENDERMINT PRIVATE KEY-----
type: secp256k1
kdf: bcrypt
salt: 85D3E663D8C4024F67CA5A48C20177F0
        
Bvp92n5ChMmbgutjU1eH17x2fAD2i0Gw4ie2bRwkrJR/zy7UZezgpT0OqwoEJJQk
p1OrbxmEKrKBl16PpWUGWZ5DnD3OkgqzxmVbCf0=
=ekQv
-----END TENDERMINT PRIVATE KEY-----
`
	defaultPublicKey = "020A8D4D4F9920F80F70C1CDE78926929CD1254038B9AF79984CAC65FCE620AB36"
)

func TestAccount(t *testing.T) {
	srv, teardown := startupT(t)
	defer teardown()

	t.Log("Test new profile creation")
	{
		resp, err := postWithAuth(
			t,
			srv.URL+"/api/v1/accounts/register", `{"email":"john.doe@supermail.com"}`,
		)

		assert.NoError(t, err)

		b, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())

		c := JSON{}
		err = json.Unmarshal(b, &c)
		assert.NoError(t, err)

		assert.Equal(t, "3", c["id"])
		assert.Equal(t, "john.doe@supermail.com", c["email"])
	}

	t.Log("Test new profile creation fails when it already exists")
	{
		resp, err := postWithAuth(
			t,
			srv.URL+"/api/v1/accounts/register", `{"email":"john.doe@supermail.com"}`,
		)
		assert.NoError(t, err)

		b, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())

		c := JSON{}
		err = json.Unmarshal(b, &c)
		assert.NoError(t, err)

		assert.Equal(t, "profile already exists", c["error"])
	}

	t.Log("Test new profile HD wallet creation")
	{
		resp, err := postWithAuth(
			t,
			srv.URL+"/api/v1/accounts/new-wallet", `{"mnemonic":"`+defaultMnemonic+`"}`,
		)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	t.Log("Test fetching mnemonic")
	{
		resp, err := getWithAuth(t, srv.URL+"/api/v1/accounts/mnemonic")
		assert.NoError(t, err)

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode, string(b))
		assert.NoError(t, resp.Body.Close())

		c := JSON{}
		err = json.Unmarshal(b, &c)
		assert.NoError(t, err)

		mnemonic := fmt.Sprintf("%s", c["mnemonic"])

		assert.Equal(t, defaultMnemonic, mnemonic)

	}

	t.Log("Test fetching OBADA accounts for newly created profile")
	{
		resp, err := getWithAuth(t, srv.URL+"/api/v1/accounts")
		assert.NoError(t, err)

		b, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, resp.Body.Close())

		var profileAccounts services.ProfileAccounts
		err = json.Unmarshal(b, &profileAccounts)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(profileAccounts.HDAccounts))
		assert.Equal(t, 0, len(profileAccounts.ImportedAccounts))

		acc := profileAccounts.HDAccounts[0]
		assert.Equal(t, sdk.NewDecCoin("obd", sdkmath.NewInt(0)), acc.Balance)
		assert.Equal(t, defaultAccount, acc.Address)
		assert.Equal(t, uint(0), acc.NFTsCount)
		assert.Equal(t, "", acc.Name)
	}

	t.Log("Test that adding a new OBADA account to the profile will fail if last account has zero transactions")
	{
		resp, err := postWithAuth(t, srv.URL+"/api/v1/accounts/new-account", "{}")
		require.NoError(t, err)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		b, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.NoError(t, resp.Body.Close())

		c := JSON{}
		err = json.Unmarshal(b, &c)
		assert.NoError(t, err)

		assert.Equal(t, account.ErrAccountHasZeroTx.Error(), c["error"])
	}
}

func TestAccount_importHDWallet(t *testing.T) {
	srv, teardown := startupT(t)
	defer teardown()

	resp, err := postWithAuth(
		t,
		srv.URL+"/api/v1/accounts/register", `{"email":"john.doe@supermail.com"}`,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	t.Log("Test it returns valid validation message when mnemonic is invalid")
	{
		resp, err = postWithAuth(
			t,
			srv.URL+"/api/v1/accounts/import-wallet", `{"mnemonic":"invalid seed"}`,
		)
		assert.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		b, er := io.ReadAll(resp.Body)
		assert.NoError(t, er)

		require.NoError(t, resp.Body.Close())

		c := JSON{}
		err = json.Unmarshal(b, &c)
		assert.NoError(t, err)

		assert.Equal(t, account.ErrInvalidMnemonic.Error(), c["error"])
	}

	t.Log("Test HD wallet import")
	{ //nolint:gocritic
		resp, err = postWithAuth(
			t,
			srv.URL+"/api/v1/accounts/import-wallet", `{"mnemonic":"`+defaultMnemonic+`"}`,
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.NoError(t, resp.Body.Close())
	}

	t.Log("Test it fails to import a new mnemonic when it already exists")
	{
		resp, err := postWithAuth(
			t,
			srv.URL+"/api/v1/accounts/import-wallet", `{"mnemonic":"`+secondMnemonic+`"}`,
		)

		assert.NoError(t, err)
		require.Equal(t, http.StatusConflict, resp.StatusCode)

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		c := JSON{}
		err = json.Unmarshal(b, &c)
		assert.NoError(t, err)

		assert.Equal(t, account.ErrWalletExists.Error(), c["error"])
	}
}

func TestAccount_importAccount(t *testing.T) {
	srv, teardown := startupT(t)
	defer teardown()

	t.Log("Test account import")
	{

		resp, err := postWithAuth(
			t,
			srv.URL+"/api/v1/accounts/register", `{"email":"john.doe@supermail.com"}`,
		)
		assert.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		payload := map[string]string{
			"private_key":  defaultObadaPrivateKey,
			"account_name": "My test imported account",
		}

		payloadJSONBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		resp, err = postWithAuth(
			t,
			srv.URL+"/api/v1/accounts/import-account", string(payloadJSONBytes),
		)

		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.NoError(t, resp.Body.Close())
	}

	t.Log("Test that imported account was deleted")
	{
		resp, err := deleteWithAuth(
			t,
			srv.URL+"/api/v1/accounts/"+defaultAccount,
		)

		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
		require.NoError(t, resp.Body.Close())
	}
}

func TestAccount_exportAccount(t *testing.T) {
	srv, teardown := startupT(t)
	defer teardown()

	resp, err := postWithAuth(
		t,
		srv.URL+"/api/v1/accounts/register", `{"email":"john.doe@supermail.com"}`,
	)
	assert.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	t.Log("Test account export")
	{
		resp, err := postWithAuth(
			t,
			srv.URL+"/api/v1/accounts/new-wallet", `{"mnemonic":"`+defaultMnemonic+`"}`,
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.NoError(t, resp.Body.Close())

		resp, err = postWithAuth(
			t,
			srv.URL+"/api/v1/accounts/export-account", `{"address":"`+defaultAccount+`", "passphrase": ""}`,
		)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		c := JSON{}
		require.NoError(t, json.Unmarshal(b, &c))

		privateKey := c["private_key"].(string)

		require.True(t, strings.Contains(privateKey, "BEGIN TENDERMINT PRIVATE KEY"))

	}
}

func TestAccount_newMnemonic(t *testing.T) {
	srv, teardown := startupT(t)
	defer teardown()

	t.Log("Mnemonic creation")
	{
		resp, err := getWithAuth(t, srv.URL+"/api/v1/accounts/new-mnemonic")
		assert.NoError(t, err)

		b, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode, string(b))
		assert.NoError(t, resp.Body.Close())

		c := JSON{}
		err = json.Unmarshal(b, &c)
		assert.NoError(t, err)

		mnemonic := fmt.Sprintf("%s", c["mnemonic"])

		assert.Equal(t, 24, len(strings.Split(mnemonic, " ")), "mnemonic should have 24 words")
	}
}

func TestAccount_account(t *testing.T) {
	srv, teardown := startupT(t)
	defer teardown()

	resp, err := postWithAuth(
		t,
		srv.URL+"/api/v1/accounts/new-wallet", `{"mnemonic":"`+defaultMnemonic+`"}`,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	resp, err = postWithAuth(
		t,
		srv.URL+"/api/v1/accounts/export-account", `{"address":"`+defaultAccount+`", "passphrase": ""}`,
	)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	t.Log("Get a single account info")
	{
		resp, err := getWithAuth(t, srv.URL+"/api/v1/accounts/"+defaultAccount)
		assert.NoError(t, err)

		b, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode, string(b))
		assert.NoError(t, resp.Body.Close())

		var acc services.Account
		err = json.Unmarshal(b, &acc)
		assert.NoError(t, err)

		assert.Equal(t, sdk.NewDecCoin("obd", sdkmath.NewInt(0)), acc.Balance)
		assert.Equal(t, defaultAccount, acc.Address)
		assert.Equal(t, uint(0), acc.NFTsCount)
		assert.Equal(t, defaultPublicKey, acc.PublicKey)
	}
}

func TestAccount_updateAccount(t *testing.T) {
	srv, teardown := startupT(t)
	defer teardown()

	resp, err := postWithAuth(
		t,
		srv.URL+"/api/v1/accounts/new-wallet", `{"mnemonic":"`+defaultMnemonic+`"}`,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	t.Log("Update account name")
	{
		resp, err = postWithAuth(
			t,
			srv.URL+"/api/v1/accounts/"+defaultAccount, `{"account_name":"test"}`,
		)

		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
		require.NoError(t, resp.Body.Close())

		resp, err := getWithAuth(t, srv.URL+"/api/v1/accounts/"+defaultAccount)
		assert.NoError(t, err)

		b, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode, string(b))
		assert.NoError(t, resp.Body.Close())

		var acc services.Account
		err = json.Unmarshal(b, &acc)
		assert.NoError(t, err)

		assert.Equal(t, defaultAccount, acc.Address)
		assert.Equal(t, "test", acc.Name)
	}
}
