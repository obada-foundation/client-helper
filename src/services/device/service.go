package device

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"strings"

	cid "github.com/ipfs/go-cid"
	"github.com/obada-foundation/client-helper/system/auth"
	"github.com/obada-foundation/client-helper/system/db"
	"github.com/obada-foundation/client-helper/system/encoder"
	"github.com/obada-foundation/client-helper/system/filecrypt"
	ipfssh "github.com/obada-foundation/client-helper/system/ipfs"
	"github.com/obada-foundation/client-helper/system/validate"
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

func (ds *Service) handleDocuments(ctx context.Context, sd SaveDevice, saveDocs bool) ([]DeviceDocument, error) {
	var (
		documents []DeviceDocument
		builder   cid.V0Builder
	)

	obitIDDto, err := sd.makeObitIDDto(sd.SerialNumber, sd.Manufacturer, sd.PartNumber)
	if err != nil {
		return documents, err
	}

	obitID, err := ds.obadasdk.NewObitID(obitIDDto)
	if err != nil {
		return documents, err
	}

	DID := obitID.GetDid()

	secret, err := ds.EncryptionSecret(ctx, DID)
	if err != nil {
		return documents, err
	}

	sd.Documents = append(
		sd.Documents,
		SaveDeviceDocument{Name: string(PhysicalAssetIdentifier), ShouldEncrypt: true},
	)

	for _, d := range sd.Documents {
		var documentBytes []byte

		switch d.Name {
		// We add this type of document for every device that we create
		case string(PhysicalAssetIdentifier):
			paiDoc := fmt.Sprintf(
				`{"serial_number":"%s","manufacturer":"%s","part_number":"%s"}`,
				sd.SerialNumber,
				sd.Manufacturer,
				sd.PartNumber,
			)

			documentBytes = []byte(paiDoc)
		default:
			documentBytes, err = base64.StdEncoding.DecodeString(d.File)
			if err != nil {
				return documents, err
			}
		}

		// Take a hash of origin content
		hash := fmt.Sprintf("%x", sha256.Sum256(documentBytes))

		// Encrypt document when true
		if d.ShouldEncrypt {
			documentBytes, err = filecrypt.Encrypt(documentBytes, secret)
			if err != nil {
				return documents, err
			}
		}

		c, err := builder.Sum(documentBytes)
		if err != nil {
			return documents, err
		}

		if saveDocs {
			_, err = ds.ipfs.CreateDocument(ctx, DID, d.Name, documentBytes)

			if err != nil {
				return documents, err
			}
		}

		document := DeviceDocument{
			Name:      d.Name,
			Hash:      hash,
			URI:       fmt.Sprintf("ipfs://%s", c.String()),
			Encrypted: d.ShouldEncrypt,
		}

		documents = append(documents, document)
	}

	return documents, nil
}

func (ds *Service) Save(ctx context.Context, sd SaveDevice) (Device, error) {
	var device Device

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return device, err
	}

	if err := ds.validator.Check(sd); err != nil {
		return device, err
	}

	obitIDDto, err := sd.makeObitIDDto(sd.SerialNumber, sd.Manufacturer, sd.PartNumber)
	if err != nil {
		return device, err
	}

	obitID, err := ds.obadasdk.NewObitID(obitIDDto)
	if err != nil {
		return device, err
	}

	device, err = ds.NewDevice(ctx, sd, true)
	if err != nil {
		return device, fmt.Errorf("Cannot make device from given data %+v %w", sd, err)
	}

	batch := ds.db.NewBatch()
	defer batch.Close()

	deviceBytes, err := encoder.DataEncode(device)
	if err != nil {
		return device, err
	}

	DID := obitID.GetDid()

	DIDkey := makeDIDKey(userID, DID)

	if err := batch.Set(DIDkey, deviceBytes); err != nil {
		return device, err
	}

	if err := batch.Set(makeUSNKey(userID, device.Usn), DIDkey); err != nil {
		return device, err
	}

	if err := batch.Write(); err != nil {
		return device, err
	}

	return device, nil
}

func (ds *Service) NewDevice(ctx context.Context, sd SaveDevice, saveDocuments bool) (Device, error) {
	var d Device

	if err := ds.validator.Check(sd); err != nil {
		return d, err
	}

	documents, err := ds.handleDocuments(ctx, sd, saveDocuments)
	if err != nil {
		return d, fmt.Errorf("Cannot handle device documents %+v %w", documents, err)
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
	}, nil
}

// Get
func (ds *Service) Get(ctx context.Context, key string) (Device, error) {
	if len(key) == USNLength {
		return ds.GetByUSN(ctx, key)
	}

	return ds.GetByDID(ctx, key)
}

func (ds *Service) GetByDID(ctx context.Context, DID string) (Device, error) {
	var d Device

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return d, err
	}

	DIDkey := makeDIDKey(userID, DID)

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

func (ds *Service) GetByUSN(ctx context.Context, USN string) (Device, error) {
	var d Device

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return d, err
	}

	USNkey := makeUSNKey(userID, USN)

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

	return ds.GetByDID(ctx, DIDkey[2])

}
