package device

import (
	"fmt"

	svcs "github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/utils"
	"github.com/obada-foundation/sdkgo"
	"github.com/obada-foundation/sdkgo/hash"
)

func newDevice(sdk *sdkgo.Sdk, sd svcs.SaveDevice, docs []svcs.DeviceDocument, pd *svcs.Device) (svcs.Device, error) {
	var (
		d        svcs.Device
		checksum hash.Hash
	)

	obit, err := makeSdkObit(sdk, sd)
	if err != nil {
		return d, fmt.Errorf("Cannot create Obit from given data %+v %w", sd, err)
	}

	if pd != nil {
		// here we put parent checksum
		checksum, err = obit.GetChecksum(nil)
	} else {
		checksum, err = obit.GetChecksum(nil)
	}

	if err != nil {
		return d, fmt.Errorf("Cannot get Obit checksum from given data %+v %w", sd, err)
	}

	did := obit.GetObitID()

	return svcs.Device{
		Usn:              did.GetUsn(),
		DID:              did.GetDid(),
		Checksum:         checksum.GetHash(),
		SerialNumberHash: obit.GetSerialNumberHash().GetValue(),
		Manufacturer:     obit.GetManufacturer().GetValue(),
		PartNumber:       obit.GetPartNumber().GetValue(),
		TrustAnchorToken: obit.GetTrustAnchorToken().GetValue(),
		Documents:        docs,
	}, nil
}

func makeObitIDDto(sn, man, pn string) (sdkgo.ObitIDDto, error) {
	serialNumberHash, err := utils.HashStr(sn)

	if err != nil {
		return sdkgo.ObitIDDto{}, err
	}

	return sdkgo.ObitIDDto{
		SerialNumberHash: serialNumberHash,
		Manufacturer:     man,
		PartNumber:       pn,
	}, nil
}

func makeSdkObit(sdk *sdkgo.Sdk, sd svcs.SaveDevice) (sdkgo.Obit, error) {
	var obit sdkgo.Obit

	IDDto, err := makeObitIDDto(sd.SerialNumber, sd.Manufacturer, sd.PartNumber)

	obit, err = sdk.NewObit(sdkgo.ObitDto{
		ObitIDDto: IDDto,
	})

	return obit, err
}
