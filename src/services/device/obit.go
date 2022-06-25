package device

import (
	"github.com/obada-foundation/client-helper/utils"
	"github.com/obada-foundation/sdkgo"
)

func (ds *SaveDevice) makeObitIDDto(sn, man, pn string) (sdkgo.ObitIDDto, error) {
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

func (ds *Service) makeSdkObit(sdk *sdkgo.Sdk, sd SaveDevice) (sdkgo.Obit, error) {
	var obit sdkgo.Obit

	IDDto, err := sd.makeObitIDDto(sd.SerialNumber, sd.Manufacturer, sd.PartNumber)

	obit, err = ds.obadasdk.NewObit(sdkgo.ObitDto{
		ObitIDDto: IDDto,
	})

	return obit, err
}
