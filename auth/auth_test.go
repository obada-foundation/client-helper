package auth_test

import (
	"bufio"
	"bytes"
	eddsa "crypto/ed25519"
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/services/pubkey"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	success = "\u2713"
	failed  = "\u2717"

	token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJFZERTQSIsImtpZCI6Ijg1YmIyMTY1LTkwZTEtNDEzNC1hZjNlLTkwYTRhMGUxZTJjMSJ9.eyJpYXQiOjE2NTU3NjM0OTcsInVpZCI6IjMifQ.zhz_vw4uBLo8QTXqHMWv_yRQhYIR99-mcWMgB_Zn0ylQyc9glyfm9-WfZ_ji15QL5TFkNgqQHTtzyz-F3OBkBQ"
)

func Test_Auth(t *testing.T) {
	log, teardown := newUnit(t)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		teardown()
	}()

	t.Log("Given the need to be able to authenticate and authorize access.")
	{
		testProfileID := 0
		t.Logf("\tTest %d:\tWhen handling a single user.", testProfileID)
		{
			cfg := auth.Config{
				Log:       log,
				KeyLookup: &keyStore{},
			}

			a, err := auth.New(cfg)
			assert.NoError(t, err)

			parsedClaims, err := a.ValidateToken(token)
			assert.NoError(t, err, fmt.Sprintf("\t%s\tTest %d:\tShould be able to authenticate the claims: %v", failed, testProfileID, err))

			assert.Equal(t, "3", parsedClaims.UserID)
		}
	}
}

//nolint:gocritic // ok for unnamed result
func newUnit(t *testing.T) (*zap.SugaredLogger, func()) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writer := bufio.NewWriter(&buf)
	log := zap.New(
		zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel),
		zap.WithCaller(true),
	).Sugar()

	teardown := func() {
		t.Helper()

		log.Sync()

		writer.Flush()
		fmt.Println("******************** LOGS ********************")
		fmt.Print(buf.String())
		fmt.Println("******************** LOGS ********************")
	}

	return log, teardown
}

type keyStore struct{}

func (ks *keyStore) PublicKey(kid string) (eddsa.PublicKey, error) {
	store, err := pubkey.NewFS("../testdata")
	if err != nil {
		return nil, err
	}

	pk, err := store.PublicKey(kid)
	if err != nil {
		return nil, err
	}

	return pk, nil
}
