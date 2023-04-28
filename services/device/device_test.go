package device_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/obada-foundation/client-helper/auth"
	svcs "github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/obada-foundation/registry/api/pb/v1/diddoc"
	"github.com/obada-foundation/registry/types"
	"github.com/obada-foundation/sdkgo/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	service, registryClient, ctx, teardown := createTestService(t)
	defer teardown()

	ctx = auth.SetClaims(ctx, auth.Claims{
		UserID: "1",
	})

	t.Log("Testing Device Service")
	{
		var device svcs.Device

		t.Log("\tTesting Device Save")
		DID := "did:obada:45837cbd255b8008ddc58fff44f7883f32ca3f15b3ac3120670c29f84b0ea020"

		privKey, pubKey, addr := GenKeys(t)

		registryClient.EXPECT().Get(gomock.Any(), gomock.Eq(&diddoc.GetRequest{Did: DID})).Times(2).
			Return(&diddoc.GetResponse{}, nil).
			Return(&diddoc.GetResponse{
				Document: &diddoc.DIDDocument{
					Metadata: &diddoc.Metadata{
						RootHash: "bcba7830a0d3b66ed8c5485b60ccc4f09ec50fbe82ff5ede9c3f39819f724af0",
					},
				},
			}, nil)

		registryClient.EXPECT().Register(gomock.Any(), diddoc.RegisterRequest{
			Did: DID,
			VerificationMethod: append(make([]*diddoc.VerificationMethod, 0, 1), &diddoc.VerificationMethod{
				Id:              DID + "#keys-1",
				Type:            types.Ed25519VerificationKey2018JSONLD,
				PublicKeyBase58: base58.Encode(pubKey.Bytes()),
			}),
			Authentication: []string{
				DID + "#keys-1",
			},
		}).Times(0).Return(&diddoc.RegisterResponse{}, nil)

		registryClient.EXPECT().SaveMetadata(gomock.Any(), gomock.Any()).Times(1).Return(nil, nil)

		device, err := service.Save(ctx, svcs.SaveDevice{
			SerialNumber: "SN123456",
			Manufacturer: "IBM",
			PartNumber:   "PN123456",
			Address:      addr,
		}, privKey)
		require.NoError(t, err, "Cannot save device")

		assert.Equal(t, DID, device.DID)
		assert.Equal(t, addr, device.Address)
		assert.Equal(t, "bcba7830a0d3b66ed8c5485b60ccc4f09ec50fbe82ff5ede9c3f39819f724af0", device.Checksum)

		t.Log("\tTesting Device Get")
		{
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

		}
	}
}

func TestService_SaveValidation(t *testing.T) {
	service, _, ctx, teardown := createTestService(t)
	defer teardown()

	ctx = auth.SetClaims(ctx, auth.Claims{
		UserID: "1",
	})

	t.Log("Testing Device Save validation")
	{
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
					{
						Field: "address",
						Error: "address is a required field",
					},
				},
			},
		}

		for _, tc := range validationTestCases {
			if _, err := service.Save(ctx, tc.given, nil); err != nil {
				if !validate.IsFieldErrors(err) {
					t.Fatalf(err.Error())
				}

				assert.Equal(t, validate.FieldErrors(tc.want), validate.GetFieldErrors(err))
			}
		}

	}
}

func TestService_DeleteByAddress(t *testing.T) {
	service, registryClient, ctx, teardown := createTestService(t)
	defer teardown()

	privKey, _, addr1 := GenKeys(t)
	privKey2, _, addr2 := GenKeys(t)

	ctx = auth.SetClaims(ctx, auth.Claims{
		UserID: "1",
	})

	registryClient.EXPECT().Get(gomock.Any(), gomock.Any()).Times(6).
		Return(&diddoc.GetResponse{}, nil).
		Return(&diddoc.GetResponse{}, nil)

	registryClient.EXPECT().SaveMetadata(gomock.Any(), gomock.Any()).Times(3).Return(nil, nil)

	_, err := service.Save(ctx, svcs.SaveDevice{
		SerialNumber: "SN123456",
		Manufacturer: "IBM",
		PartNumber:   "PN123456",
		Address:      addr1,
	}, privKey)
	require.NoError(t, err, "Cannot save device")

	_, err = service.Save(ctx, svcs.SaveDevice{
		SerialNumber: "SN123457",
		Manufacturer: "IBM",
		PartNumber:   "PN123456",
		Address:      addr1,
	}, privKey)
	require.NoError(t, err, "Cannot save device")

	_, err = service.Save(ctx, svcs.SaveDevice{
		SerialNumber: "SN123458",
		Manufacturer: "IBM",
		PartNumber:   "PN123456",
		Address:      addr2,
	}, privKey2)
	require.NoError(t, err, "Cannot save device")

	deletedRecords, err := service.DeleteByAddress(ctx, addr1)
	require.NoError(t, err, "Cannot delete devices by address")

	assert.Equal(t, uint(2), deletedRecords)

}

func GenKeys(t *testing.T) (cryptotypes.PrivKey, cryptotypes.PubKey, string) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr, err := sdk.AccAddressFromHexUnsafe(pubKey.Address().String())
	require.NoError(t, err)

	return privKey, pubKey, addr.String()
}
