package account

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/obada-foundation/client-helper/system/db"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	t.Log("Testing Account Service")

	v, err := validate.NewValidator()
	if err != nil {
		t.Fatalf("Cannot intialize validation: %s", err.Error())
	}

	db, err := db.NewDB("accounts", db.MemDBBackend, "./testdb")
	if err != nil {
		t.Fatalf("Cannot intialize database: %s", err.Error())
	}

	service := NewService(v, db)

	t.Log("Testing Account creation")

	na := NewAccount{
		ID:    uuid.New().String(),
		Email: "jon.doe@supermail.com",
	}

	a, err := service.Create(na)
	if err != nil {
		t.Fatalf("Cannot create account: %s", err.Error())
	}

	assert.Equal(t, na.ID, a.ID)
	assert.Equal(t, na.Email, a.Email)

	t.Log("Testing Account find by ID")

	fa, err := service.Find(a.ID)
	if err != nil {
		t.Fatalf("Cannot find account that was previostly created: %s", err.Error())
	}

	assert.Equal(t, fa, a)

	t.Log("Testing Account wallet fetch")
	balance, err := service.Wallet(a.ID)
	if err != nil {
		t.Fatalf("Cannot get account: %s", err.Error())
	}

	assert.Equal(t, 0, balance.Balance)
	assert.True(t, strings.HasPrefix(balance.Address, "obada1"))

	t.Log("Testing Account won't create if already exists")

	a, err = service.Create(na)
	if err != nil {
		if !errors.Is(err, ErrAccountExists) {
			t.Fatalf("Cannot create account: %s", err.Error())
		}
	}

	t.Log("Testing Account creation validation")

	type validationTest struct {
		given NewAccount
		want  []validate.FieldError
	}

	validationTestCases := []validationTest{
		{
			given: NewAccount{},
			want: []validate.FieldError{
				{
					Field: "id",
					Error: "id is a required field",
				},
				{
					Field: "email",
					Error: "email is a required field",
				},
			},
		},
		{
			given: NewAccount{
				Email: "brokenemail",
			},
			want: []validate.FieldError{
				{
					Field: "id",
					Error: "id is a required field",
				},
				{
					Field: "email",
					Error: "email must be a valid email address",
				},
			},
		},
	}

	for _, tc := range validationTestCases {
		_, err = service.Create(tc.given)
		if err != nil {
			if !validate.IsFieldErrors(err) {
				t.Fatalf(err.Error())
			}

			assert.Equal(t, validate.FieldErrors(tc.want), validate.GetFieldErrors(err))
		}
	}
}
