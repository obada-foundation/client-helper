package account

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/auth"
	"github.com/obada-foundation/client-helper/system/db"
	"github.com/obada-foundation/client-helper/system/obadanode"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("obada", "obada"+sdk.PrefixPublic)
	config.Seal()

	t.Log("Testing Account Service")

	v, err := validate.NewValidator()
	require.NoError(t, err, "Cannot intialize validation")

	db, err := db.NewDB("accounts", db.MemDBBackend, "./testdb")
	defer db.Close()
	require.NoError(t, err, "Cannot intialize database")

	nodeClient, err := obadanode.NewClient(
		"obada-testnet",
		"tcp://52.206.218.105:26657",
		"52.206.218.105:9090",
	)
	defer nodeClient.Close()
	require.NoError(t, err, "Cannot OBADA Node client")

	service := NewService(v, db, &nodeClient)

	t.Log("Testing Account creation")

	na := services.NewAccount{
		ID:    uuid.New().String(),
		Email: "jon.doe@supermail.com",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ctx = auth.SetClaims(ctx, auth.Claims{
		UserID: "1",
	})

	na.ID = "1"

	a, err := service.Create(ctx, na)
	require.NoError(t, err, "Cannot create account")

	assert.Equal(t, na.ID, a.ID)
	assert.Equal(t, na.Email, a.Email)

	t.Log("Testing Account find by ID")

	fa, err := service.Show(ctx)
	require.NoError(t, err, "Cannot find account that was previostly created")

	assert.Equal(t, fa, a)

	t.Log("Testing Account wallet fetch")
	_, err = service.Wallet(ctx)
	require.NoError(t, err, "Cannot fetch the wallet")

	t.Log("Testing Account balance fetch")
	balance, err := service.Balance(ctx)
	require.NoError(t, err, "Cannot fetch the balance")

	assert.Equal(t, 0, balance.Balance)
	assert.True(t, strings.HasPrefix(balance.Address, "obada1"))

	t.Log("Testing Account won't be created if already exists")

	a, err = service.Create(ctx, na)
	if err != nil {
		if !errors.Is(err, ErrAccountExists) {
			t.Fatalf("Cannot create account: %s", err.Error())
		}
	}

	t.Log("Testing Account creation validation")

	type validationTest struct {
		given services.NewAccount
		want  []validate.FieldError
	}

	validationTestCases := []validationTest{
		{
			given: services.NewAccount{},
			want: []validate.FieldError{
				{
					Field: "ID",
					Error: "ID is a required field",
				},
				{
					Field: "email",
					Error: "email is a required field",
				},
			},
		},
		{
			given: services.NewAccount{
				Email: "brokenemail",
			},
			want: []validate.FieldError{
				{
					Field: "ID",
					Error: "ID is a required field",
				},
				{
					Field: "email",
					Error: "email must be a valid email address",
				},
			},
		},
	}

	for _, tc := range validationTestCases {
		_, err = service.Create(ctx, tc.given)
		if err != nil {
			if !validate.IsFieldErrors(err) {
				t.Fatalf(err.Error())
			}

			assert.Equal(t, validate.FieldErrors(tc.want), validate.GetFieldErrors(err))
		}
	}
}
