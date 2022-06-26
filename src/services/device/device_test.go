package device

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	svcs "github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/auth"
	"github.com/obada-foundation/client-helper/system/db"
	"github.com/obada-foundation/client-helper/system/ipfs"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/obada-foundation/sdkgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("obada", "obada"+sdk.PrefixPublic)
	config.Seal()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ctx = auth.SetClaims(ctx, auth.Claims{
		UserID: "1",
	})

	t.Log("Testing Device Service")

	v, err := validate.NewValidator()
	require.NoError(t, err, "Cannot intialize validation")

	db, err := db.NewDB("accounts", db.MemDBBackend, "./testdb")
	defer db.Close()
	require.NoError(t, err, "Cannot intialize database")

	sdk, err := sdkgo.NewSdk(nil, false)
	require.NoError(t, err, "Cannot intialize OBADA SDK")

	ipfsShell := ipfs.NewIPFS("http://localhost:5001")

	service := NewService(v, db, sdk, ipfsShell)

	t.Log("\tTesting Device Save")

	device, err := service.Save(ctx, svcs.SaveDevice{
		SerialNumber: "SN123456",
		Manufacturer: "IBM",
		PartNumber:   "PN123456",
	})
	require.NoError(t, err, "Cannot save device")

	assert.Equal(t, "did:obada:cc65aa0f47463a5d0488fd758a2524f8bd24981619afc319c8d7b67eeb75c468", device.DID)
	assert.Equal(t, "bcba7830a0d3b66ed8c5485b60ccc4f09ec50fbe82ff5ede9c3f39819f724af0", device.Checksum)

	t.Log("\tTesting Device Get")

	type deviceGetTest struct {
		given string
		want  svcs.Device
	}

	deviceGetCases := []deviceGetTest{
		{
			given: device.DID,
			want:  device,
		},
		{
			given: device.Usn,
			want:  device,
		},
	}

	for _, tc := range deviceGetCases {
		getDevice, err := service.Get(ctx, tc.given)
		require.NoError(t, err, "Cannot get device", tc.given)

		assert.Equal(t, tc.want, getDevice)
	}

	t.Log("\tTesting Device Save validation")

	type validationTest struct {
		given svcs.SaveDevice
		want  []validate.FieldError
	}

	validationTestCases := []validationTest{
		{
			given: svcs.SaveDevice{},
			want: []validate.FieldError{
				{
					Field: "serial_number",
					Error: "serial_number is a required field",
				},
				{
					Field: "manufacturer",
					Error: "manufacturer is a required field",
				},
				{
					Field: "part_number",
					Error: "part_number is a required field",
				},
			},
		},
	}

	for _, tc := range validationTestCases {
		_, err := service.Save(ctx, tc.given)
		if err != nil {
			if !validate.IsFieldErrors(err) {
				t.Fatalf(err.Error())
			}

			assert.Equal(t, validate.FieldErrors(tc.want), validate.GetFieldErrors(err))
		}
	}
}
