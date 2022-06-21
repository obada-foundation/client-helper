package api

import (
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/pubkey"
	"github.com/obada-foundation/client-helper/system/auth"
	"github.com/obada-foundation/client-helper/system/db"
	"github.com/obada-foundation/client-helper/system/logger"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/stretchr/testify/assert"

	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var devToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJFZERTQSIsImtpZCI6Ijg1YmIyMTY1LTkwZTEtNDEzNC1hZjNlLTkwYTRhMGUxZTJjMSJ9.eyJpYXQiOjE2NTU3NjM0OTcsInVpZCI6IjMifQ.zhz_vw4uBLo8QTXqHMWv_yRQhYIR99-mcWMgB_Zn0ylQyc9glyfm9-WfZ_ji15QL5TFkNgqQHTtzyz-F3OBkBQ"

// startupT runs fully configured testing server
// srvHook is an optional func to set some Rest param after the creation but prior to Run
func startupT(t *testing.T, srvHook ...func(srv *Rest)) (ts *httptest.Server, srv *Rest, teardown func()) {
	logger, err := logger.New("CLIENT-HELPER-TEST")
	assert.NoError(t, err)

	validator, err := validate.NewValidator()
	assert.NoError(t, err)

	database, err := db.NewDB("client-helper-test", db.MemDBBackend, ".")
	assert.NoError(t, err)

	accountSvc := account.NewService(validator, database)

	ks, err := pubkey.NewFS("./testdata")
	assert.NoError(t, err, "reading keys")

	// Auth manager verifies JWT tokens
	auth, err := auth.New("85bb2165-90e1-4134-af3e-90a4a0e1e2c1", ks)
	assert.NoError(t, err, "reading keys")

	srv = &Rest{
		Auth:           auth,
		Logger:         logger,
		AccountService: accountSvc,
	}

	ts = httptest.NewServer(srv.routes())

	teardown = func() {
		ts.Close()
	}

	return ts, srv, teardown
}

func postWithAuth(t *testing.T, url string, body string) (*http.Response, error) {
	headers := map[string]string{
		"authorization": fmt.Sprintf("bearer %s", devToken),
	}

	return post(t, url, body, headers)
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
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	for header, headerVal := range headers {
		req.Header.Add(header, headerVal)
	}
	return client.Do(req)
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
