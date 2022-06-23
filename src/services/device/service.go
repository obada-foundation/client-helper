package device

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/obada-foundation/client-helper/system/db"
	"github.com/obada-foundation/client-helper/system/encoder"
	"github.com/obada-foundation/client-helper/system/filecrypt"
	ipfssh "github.com/obada-foundation/client-helper/system/ipfs"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/obada-foundation/client-helper/utils"
	"github.com/obada-foundation/sdkgo"
)

const USNLength = 8

type Service struct {
	validator *validate.Validator
	db        db.DB
	obadasdk  *sdkgo.Sdk
	ipfs      *ipfssh.IPFS
}

func NewService(v *validate.Validator, db db.DB, sdk *sdkgo.Sdk, ipfs *ipfssh.IPFS) *Service {

	return &Service{
		validator: v,
		db:        db,
		obadasdk:  sdk,
		ipfs:      ipfs,
	}
}

func (ds *Service) makeSdkObit(sdk *sdkgo.Sdk, payload SaveDevice) (sdkgo.Obit, error) {
	var obit sdkgo.Obit

	serialNumberHash, err := utils.HashStr(payload.SerialNumber)
	if err != nil {
		return obit, err
	}

	obit, err = sdk.NewObit(sdkgo.ObitDto{
		ObitIDDto: sdkgo.ObitIDDto{
			SerialNumberHash: serialNumberHash,
			Manufacturer:     payload.Manufacturer,
			PartNumber:       payload.PartNumber,
		},
	})

	return obit, err
}

func makeUSNKey(accountID, usn string) []byte {
	return []byte(fmt.Sprintf("devices:%s:usn:%s", accountID, usn))
}

func makeDIDKey(accountID, DID string) []byte {
	return []byte(fmt.Sprintf("devices:%s:%s", accountID, DID))
}

func (ds *Service) handleDocuments(sd SaveDevice) ([]DeviceDocument, string, error) {
	var documents []DeviceDocument
	var secret string

	if sd.EncryptDocuments {
		secret = uuid.New().String()[:32]
	}

	for _, d := range sd.Documents {
		documentBytes, err := base64.StdEncoding.DecodeString(d.File)
		if err != nil {
			return documents, secret, err
		}

		if sd.EncryptDocuments {
			documentBytes, err = filecrypt.Encrypt(documentBytes, secret)
			if err != nil {
				return documents, secret, err
			}
		}

		cid, err := ds.ipfs.CreateDocument(documentBytes)
		if err != nil {
			return documents, secret, err
		}

		document := DeviceDocument{
			Name: d.Name,
		}

		documents = append(documents, document)
	}

	return documents, secret, nil
}

func (ds *Service) Save(ctx context.Context, accountID string, sd SaveDevice) (Device, error) {
	device, err := ds.NewDevice(ctx, sd)
	if err != nil {
		return device, fmt.Errorf("Cannot make device from given data %+v %w", sd, err)
	}

	batch := ds.db.NewBatch()
	defer batch.Close()

	deviceBytes, err := encoder.DataEncode(device)
	if err != nil {
		return device, err
	}

	DIDkey := makeDIDKey(accountID, device.DID)

	if err := batch.Set(DIDkey, deviceBytes); err != nil {
		return device, err
	}

	if err := batch.Set(makeUSNKey(accountID, device.Usn), DIDkey); err != nil {
		return device, err
	}

	if err := batch.Write(); err != nil {
		return device, err
	}

	return device, nil
}

func (ds *Service) NewDevice(ctx context.Context, sd SaveDevice) (Device, error) {
	var d Device

	documents, secret, err := ds.handleDocuments(sd)
	if err != nil {
		return d, fmt.Errorf("Cannot handle device documents %+v %w", documents, err)
	}

	if err := ds.validator.Check(sd); err != nil {
		return d, err
	}

	obit, err := ds.makeSdkObit(ds.obadasdk, sd)
	if err != nil {
		return d, fmt.Errorf("Cannot create Obit from given data %+v %w", sd, err)
	}

	checksum, err := obit.GetChecksum(nil)
	if err != nil {
		return d, fmt.Errorf("Cannot get Obit checksum from given data %+v %w", sd, err)
	}

	did := obit.GetObitID()

	return Device{
		Usn:              did.GetUsn(),
		DID:              did.GetDid(),
		Checksum:         checksum.GetHash(),
		SerialNumberHash: obit.GetSerialNumberHash().GetValue(),
		Manufacturer:     obit.GetManufacturer().GetValue(),
		PartNumber:       obit.GetPartNumber().GetValue(),
		TrustAnchorToken: obit.GetTrustAnchorToken().GetValue(),
		Documents:        documents,
		Secret:           secret,
	}, nil
}

// Get
func (ds *Service) Get(ctx context.Context, accountID, key string) (Device, error) {
	if len(key) == USNLength {
		return ds.GetByUSN(ctx, accountID, key)
	}

	return ds.GetByDID(ctx, accountID, key)
}

func (ds *Service) GetByDID(ctx context.Context, accountID, DID string) (Device, error) {
	var d Device

	DIDkey := makeDIDKey(accountID, DID)

	ok, err := ds.db.Has(DIDkey)
	if err != nil {
		return d, err
	}

	if !ok {
		return d, ErrDeviceNotExists
	}

	deviceBytes, err := ds.db.Get(DIDkey)
	if err != nil {
		return d, err
	}

	buf := bytes.NewBuffer(deviceBytes)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(&d); err != nil {
		return d, err
	}

	return d, nil
}

func (ds *Service) GetByUSN(ctx context.Context, accountID, USN string) (Device, error) {
	var d Device

	USNkey := makeUSNKey(accountID, USN)

	ok, err := ds.db.Has(USNkey)
	if err != nil {
		return d, err
	}

	if !ok {
		return d, ErrDeviceNotExists
	}

	DIDbytes, err := ds.db.Get(USNkey)
	if err != nil {
		return d, err
	}

	DIDkey := strings.SplitAfterN(string(DIDbytes), ":", 3)

	return ds.GetByDID(ctx, accountID, DIDkey[2])

}
