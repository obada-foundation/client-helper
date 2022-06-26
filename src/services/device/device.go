package device

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"strings"

	svcs "github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/auth"
	"github.com/obada-foundation/client-helper/system/db"
	"github.com/obada-foundation/client-helper/system/encoder"
	"github.com/obada-foundation/client-helper/system/filecrypt"
	ipfssh "github.com/obada-foundation/client-helper/system/ipfs"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/obada-foundation/sdkgo"
)

type DocumentType string

const (
	PhysicalAssetIdentifier DocumentType = "physical_asset_identifier"
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

func parentDocument(docName string, parentDocs []svcs.DeviceDocument) *svcs.DeviceDocument {
	if len(parentDocs) == 0 {
		return nil
	}

	for _, parentDoc := range parentDocs {
		if docName == parentDoc.Name {
			return &parentDoc
		}
	}

	return nil
}

func (ds Service) handleDocuments(ctx context.Context, sd svcs.SaveDevice, parentDocs []svcs.DeviceDocument, saveDocs bool) ([]svcs.DeviceDocument, error) {
	var (
		documents []svcs.DeviceDocument
		secret    []byte
		err       error
	)

	// If we want to save documents to IPFS, we will need encryption for some of them
	if saveDocs {
		obitIDDto, err := makeObitIDDto(sd.SerialNumber, sd.Manufacturer, sd.PartNumber)
		if err != nil {
			return documents, err
		}

		obitID, err := ds.obadasdk.NewObitID(obitIDDto)
		if err != nil {
			return documents, err
		}

		DID := obitID.GetDid()

		secret, err = ds.EncryptionSecret(ctx, DID)
		if err != nil {
			return documents, err
		}
	}

	// Special document type that cover a serial number
	sd.Documents = append(
		sd.Documents,
		svcs.SaveDeviceDocument{
			Name:          string(PhysicalAssetIdentifier),
			ShouldEncrypt: true,
		},
	)

	for _, d := range sd.Documents {
		var documentBytes []byte

		switch d.Name {

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

		// When we have this document in parentDocs and hashes matches
		// we don't need to go over encryption and saving to IPFS
		pd := parentDocument(d.Name, parentDocs)
		if pd != nil && pd.Hash == hash && d.ShouldEncrypt == pd.Encrypted {
			documents = append(documents, *pd)

			continue
		}

		// Encrypt document when true
		if saveDocs && d.ShouldEncrypt {
			documentBytes, err = filecrypt.Encrypt(documentBytes, secret)
			if err != nil {
				return documents, err
			}
		}

		cid, err := ds.ipfs.CreateDocument(documentBytes, saveDocs)
		if err != nil {
			return documents, err
		}

		document := svcs.DeviceDocument{
			Name:      d.Name,
			Hash:      hash,
			URI:       fmt.Sprintf("ipfs://%s", cid),
			Encrypted: d.ShouldEncrypt,
		}

		documents = append(documents, document)
	}

	return documents, nil
}

func (ds Service) Save(ctx context.Context, sd svcs.SaveDevice) (svcs.Device, error) {
	var (
		device       svcs.Device
		parentDevice svcs.Device
	)

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return device, err
	}

	if err := ds.validator.Check(sd); err != nil {
		return device, err
	}

	dto, err := makeObitIDDto(sd.SerialNumber, sd.Manufacturer, sd.PartNumber)
	if err != nil {
		return device, err
	}

	obitID, err := ds.obadasdk.NewObitID(dto)
	if err != nil {
		return device, err
	}

	parentDevice, err = ds.Get(ctx, obitID.GetDid())
	if err != nil && !errors.Is(err, ErrDeviceNotExists) {
		return device, err
	}

	documents, err := ds.handleDocuments(ctx, sd, parentDevice.Documents, true)
	if err != nil {
		return device, err
	}

	device, err = newDevice(ds.obadasdk, sd, documents, &parentDevice)
	if err != nil {
		return device, fmt.Errorf("Cannot make device from given data %+v %w", sd, err)
	}

	batch := ds.db.NewBatch()
	defer batch.Close()

	deviceBytes, err := encoder.DataEncode(device)
	if err != nil {
		return device, err
	}

	DIDkey := makeDIDKey(userID, obitID.GetDid())

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

// Get
func (ds Service) Get(ctx context.Context, key string) (svcs.Device, error) {
	if len(key) == USNLength {
		return ds.GetByUSN(ctx, key)
	}

	return ds.GetByDID(ctx, key)
}

func (ds Service) GetByDID(ctx context.Context, DID string) (svcs.Device, error) {
	var d svcs.Device

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

func (ds Service) GetByUSN(ctx context.Context, USN string) (svcs.Device, error) {
	var d svcs.Device

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
