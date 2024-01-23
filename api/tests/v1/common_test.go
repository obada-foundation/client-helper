package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmostestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/mustafaturan/bus/v3"
	"github.com/obada-foundation/client-helper/api"
	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/events"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/pubkey"
	"github.com/obada-foundation/client-helper/system/obadanode"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/obada-foundation/common/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	db "github.com/tendermint/tm-db"
)

const (
	devToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJFZERTQSIsImtpZCI6Ijg1YmIyMTY1LTkwZTEtNDEzNC1hZjNlLTkwYTRhMGUxZTJjMSJ9.eyJpYXQiOjE2NTU3NjM0OTcsInVpZCI6IjMifQ.zhz_vw4uBLo8QTXqHMWv_yRQhYIR99-mcWMgB_Zn0ylQyc9glyfm9-WfZ_ji15QL5TFkNgqQHTtzyz-F3OBkBQ"
)

//nolint:all // it's ok
func init() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("obada", "obada"+sdk.PrefixPublic)
	config.Seal()
}

// JSON helper for data decode
type JSON map[string]interface{}

// nolint
var c *testutil.Container

func TestMain(m *testing.M) {
	var err error

	c, err = testutil.StartBlockchain("")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer testutil.StopBlockchain(nil, c)

	m.Run()
}

//nolint:all //refactor
func startupT(t *testing.T) (*httptest.Server, func()) {
	ctx := context.Background()

	shutdown := make(chan os.Signal, 1)

	var fn bus.Next = func() string { return "afakeid" }
	b, err := bus.NewBus(fn)
	require.NoError(t, err, "Cannot intialize initalize event bus")

	b.RegisterTopics(events.AccountDeleted, events.AccountCreated)

	logger, lgDefer := testutil.NewTestLoger()

	validator, err := validate.NewValidator()
	require.NoError(t, err)

	database, err := db.NewDB("client-helper-test", db.MemDBBackend, ".")
	require.NoError(t, err)

	nodeClient, err := obadanode.NewClient(
		ctx,
		"obada-testnet",
		fmt.Sprintf("tcp://%s:%d", c.Host, c.Ports["26657"]),
		fmt.Sprintf("%s:%d", c.Host, c.Ports["9090"]),
	)
	require.NoError(t, err, "Cannot initialize OBADA Node client")

	kr := keyring.NewInMemory(cosmostestutil.MakeTestEncodingConfig().Codec)

	accountSvc := account.NewService(validator, database, &nodeClient, kr, b)

	ks, err := pubkey.NewFS("../../../testdata")
	require.NoError(t, err, "reading keys")

	// Auth manager verifies JWT tokens
	a, err := auth.New(auth.Config{
		Log:       logger,
		KeyLookup: ks,
	})
	require.NoError(t, err, "reading keys")

	mux := api.APIMux(api.APIMuxConfig{
		Shutdown: shutdown,
		Log:      logger,
		Auth:     a,

		AccountSvc: accountSvc,
	})

	srv := httptest.NewServer(mux)

	teardown := func() {
		srv.Close()
		lgDefer()
	}

	return srv, teardown

}

func deleteWithAuth(t *testing.T, url string) (*http.Response, error) {
	headers := map[string]string{
		"authorization": fmt.Sprintf("bearer %s", devToken),
	}

	return delete(t, url, headers)
}

//nolint:all //refactor
func delete(t *testing.T, url string, headers map[string]string) (*http.Response, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	defer client.CloseIdleConnections()
	req, err := http.NewRequest("DELETE", url, nil)
	assert.NoError(t, err)
	for header, headerVal := range headers {
		req.Header.Add(header, headerVal)
	}
	return client.Do(req)
}

//nolint:all //refactor
func postWithAuth(t *testing.T, url string, body string) (*http.Response, error) {
	headers := map[string]string{
		"authorization": fmt.Sprintf("bearer %s", devToken),
	}

	return post(t, url, body, headers)
}

func post(t *testing.T, url, body string, headers map[string]string) (*http.Response, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	defer client.CloseIdleConnections()
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	assert.NoError(t, err)
	for header, headerVal := range headers {
		req.Header.Add(header, headerVal)
	}
	return client.Do(req)
}

func getWithAuth(t *testing.T, url string) (*http.Response, error) {
	headers := map[string]string{
		"authorization": fmt.Sprintf("bearer %s", devToken),
	}

	return get(t, url, headers)
}

func get(t *testing.T, url string, headers map[string]string) (*http.Response, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	defer client.CloseIdleConnections()
	//nolint:all //refactor
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)
	for header, headerVal := range headers {
		req.Header.Add(header, headerVal)
	}
	return client.Do(req)
}
