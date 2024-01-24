package device

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/golang/protobuf/proto" // nolint:staticcheck //need check
	"github.com/mustafaturan/bus/v3"
	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/events"
	svcs "github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/system/encoder"
	ipfssh "github.com/obada-foundation/client-helper/system/ipfs"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/obada-foundation/fullcore/x/obit/types"
	regapi "github.com/obada-foundation/registry/api"
	"github.com/obada-foundation/registry/api/pb/v1/diddoc"
	"github.com/obada-foundation/registry/client"
	regtypes "github.com/obada-foundation/registry/types"
	"github.com/obada-foundation/sdkgo/asset"
	"github.com/obada-foundation/sdkgo/base58"
	sdkdid "github.com/obada-foundation/sdkgo/did"
	"github.com/obada-foundation/sdkgo/encryption"
	db "github.com/tendermint/tm-db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config is the device service configuration
type Config struct {
	Validator *validate.Validator
	DB        db.DB
	IPFS      ipfssh.IPFS
	Bus       *bus.Bus
	Registry  client.Client
}

// Service holds dependencies
type Service struct {
	validator *validate.Validator
	db        db.DB
	ipfs      ipfssh.IPFS
	eventBus  *bus.Bus
	registry  client.Client
}

// NewService creates a new device service
func NewService(cfg Config) *Service {
	return &Service{
		registry:  cfg.Registry,
		validator: cfg.Validator,
		db:        cfg.DB,
		ipfs:      cfg.IPFS,
		eventBus:  cfg.Bus,
	}
}

// nolint:unused // need refactoring
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

func (ds Service) handleDocuments(_ context.Context, sd svcs.SaveDevice, pk cryptotypes.PubKey, saveDocs bool) ([]svcs.DeviceDocument, error) {
	//nolint:prealloc //need refactoring
	var (
		documents []svcs.DeviceDocument
		err       error
	)

	hasIdentifier := false

	for _, d := range sd.Documents {
		if d.Type == string(asset.PhysicalAssetIdentifiers) {
			hasIdentifier = true
		}
	}

	if !hasIdentifier {
		// Special document type that covers stores serial number
		sd.Documents = append(
			sd.Documents,
			svcs.SaveDeviceDocument{
				Name:          string(asset.PhysicalAssetIdentifiers),
				Type:          string(asset.PhysicalAssetIdentifiers),
				ShouldEncrypt: false,
			},
		)
	}

	for _, d := range sd.Documents {
		var documentBytes []byte

		switch d.Type {

		case string(asset.PhysicalAssetIdentifiers):
			documentBytes, err = json.Marshal(asset.PhysicalAssetIdentifiersScheme{
				SerialNumber: sd.SerialNumber,
				Manufacturer: sd.Manufacturer,
				PartNumber:   sd.PartNumber,
			})

			if err != nil {
				return documents, err
			}
		default:
			documentBytes, err = base64.StdEncoding.DecodeString(d.File)
			if err != nil {
				return documents, err
			}
		}

		// Take a hash of origin content
		hash := fmt.Sprintf("%x", sha256.Sum256(documentBytes))

		// Encrypt document when true
		if saveDocs && d.ShouldEncrypt {
			documentBytes, err = encryption.Encrypt(pk, documentBytes)
			if err != nil {
				return documents, err
			}
		}

		cid, err := ds.ipfs.CreateDocument(documentBytes, saveDocs)
		if err != nil {
			return documents, err
		}

		document := svcs.DeviceDocument{
			Name:        d.Name,
			Hash:        hash,
			URI:         fmt.Sprintf("ipfs://%s", cid),
			Encrypted:   d.ShouldEncrypt,
			Type:        d.Type,
			Description: d.Description,
		}

		documents = append(documents, document)
	}

	return documents, nil
}

// Save a device and register it in DID registry
func (ds Service) Save(ctx context.Context, sd svcs.SaveDevice, pk cryptotypes.PrivKey) (svcs.Device, error) {
	var device svcs.Device

	userID := auth.GetClaims(ctx).UserID

	if err := ds.validator.Check(sd); err != nil {
		return device, err
	}

	DID, err := sdkdid.MakeDID(sdkdid.NewDID{
		SerialNumber: sd.SerialNumber,
		Manufacturer: sd.Manufacturer,
		PartNumber:   sd.PartNumber,
	})
	if err != nil {
		return device, err
	}

	verifyMethodID := fmt.Sprintf("%s#keys-1", DID.String())

	_, err = ds.registry.Get(ctx, &diddoc.GetRequest{
		Did: DID.String(),
	})

	if err != nil {
		er, ok := status.FromError(err)
		if !ok {
			return device, err
		}

		if er.Code() != codes.NotFound {
			return device, err
		}

		vm := append(make([]*diddoc.VerificationMethod, 0, 1), &diddoc.VerificationMethod{
			Id:              verifyMethodID,
			Type:            regtypes.Ed25519VerificationKey2018JSONLD,
			Controller:      DID.String(),
			PublicKeyBase58: base58.Encode(pk.PubKey().Bytes()),
		})

		// Register DID in OBADA registry
		_, erReg := ds.registry.Register(ctx, &diddoc.RegisterRequest{
			Did:                DID.String(),
			VerificationMethod: vm,
			Authentication: []string{
				verifyMethodID,
			},
		})
		if erReg != nil {
			return device, fmt.Errorf("cannot register DID in the registry: %w", erReg)
		}
	}

	documents, err := ds.handleDocuments(ctx, sd, pk.PubKey(), true)
	if err != nil {
		return device, err
	}

	objs := make([]*diddoc.Object, 0, len(documents))
	for _, d := range documents {
		encHash := ""
		if d.Encrypted {
			encHash = "xxx"
		}

		objs = append(objs, &diddoc.Object{
			Url: d.URI,
			Metadata: map[string]string{
				"type":        d.Type,
				"name":        d.Name,
				"description": d.Description,
			},
			HashUnencryptedObject:   d.Hash,
			HashEncryptedDataObject: encHash,
		})
	}

	data := &diddoc.SaveMetadataRequest_Data{
		Did:                 DID.String(),
		AuthenticationKeyId: verifyMethodID,
		Objects:             objs,
	}

	hash, err := regapi.ProtoDeterministicChecksum(data)
	if err != nil {
		return device, err
	}

	signature, err := pk.Sign(hash[:])
	if err != nil {
		return device, err
	}

	_, err = ds.registry.SaveMetadata(ctx, &diddoc.SaveMetadataRequest{
		Signature: signature,
		Data:      data,
	})
	if err != nil {
		return device, fmt.Errorf("cannot save metadata to registry: %w", err)
	}

	resp, err := ds.registry.Get(ctx, &diddoc.GetRequest{
		Did: DID.String(),
	})
	if err != nil {
		return device, err
	}

	DIDDoc := resp.GetDocument()

	device = svcs.Device{
		Usn:          DID.GetUSN(),
		DID:          DID.String(),
		Checksum:     DIDDoc.GetMetadata().GetRootHash(),
		SerialNumber: sd.SerialNumber,
		Manufacturer: sd.Manufacturer,
		PartNumber:   sd.PartNumber,
		Documents:    documents,
		Address:      sd.Address,
	}

	batch := ds.db.NewBatch()
	defer batch.Close()

	deviceBytes, err := encoder.DataEncode(device)
	if err != nil {
		return device, err
	}

	DIDkey := makeDIDKey(userID, DID.String())
	if err := batch.Set(DIDkey, deviceBytes); err != nil {
		return device, err
	}

	if err := batch.Set(makeUSNKey(userID, device.Usn), DIDkey); err != nil {
		return device, err
	}

	if err := batch.Set(makeAddressKey(userID, device.Address, DID.String()), []byte(DID.String())); err != nil {
		return device, err
	}

	if err := batch.Write(); err != nil {
		return device, err
	}

	evt := DeviceSaved{
		Device:    device,
		ProfileID: auth.GetUserID(ctx),
	}

	if err := ds.eventBus.Emit(ctx, events.DeviceSaved, evt); err != nil {
		return device, err
	}

	return device, nil
}

// ImportDevice imports a device from a given DID
func (ds Service) ImportDevice(ctx context.Context, nft types.NFT, address string) error {
	userID := auth.GetUserID(ctx)
	nftData := &types.NFTData{}
	deviceDocuments := make([]svcs.DeviceDocument, 0)
	cid := ""

	if err := proto.Unmarshal(nft.Data.GetValue(), nftData); err != nil {
		return fmt.Errorf("cannot unmarshall nft data: %w", err)
	}

	resp, err := ds.registry.Get(ctx, &diddoc.GetRequest{
		Did: nft.GetId(),
	})
	if err != nil {
		return fmt.Errorf("cannot obtain info about DID %q from registry while importing:%w", nft.Id, err)
	}

	for _, doc := range resp.GetDocument().GetMetadata().GetObjects() {
		md := doc.GetMetadata()

		docType := md["type"]
		docName := md["name"]
		docDescription := md["description"]
		if docType == string(asset.PhysicalAssetIdentifiers) {
			uriParts := strings.Split(doc.Url, "://")
			cid = uriParts[1]
		}

		deviceDocuments = append(deviceDocuments, svcs.DeviceDocument{
			Name:        docName,
			URI:         doc.Url,
			Description: docDescription,
			Type:        docType,
			Hash:        doc.HashUnencryptedObject,
			Encrypted:   len(doc.HashEncryptedDataObject) > 0,
		})
	}

	if cid == "" {
		return fmt.Errorf("missing physical asset identifier: %+v", nft)
	}

	didDataBytes, err := ds.ipfs.GetDocument(cid)
	if err != nil {
		return err
	}

	physicalAssetIdentifier := asset.PhysicalAssetIdentifiersScheme{}
	if er := json.Unmarshal(didDataBytes, &physicalAssetIdentifier); er != nil {
		return er
	}

	device := svcs.Device{
		DID:          nft.Id,
		Usn:          nftData.Usn,
		Checksum:     nft.UriHash,
		Documents:    deviceDocuments,
		SerialNumber: physicalAssetIdentifier.SerialNumber,
		Manufacturer: physicalAssetIdentifier.Manufacturer,
		PartNumber:   physicalAssetIdentifier.PartNumber,
		Address:      address,
	}

	batch := ds.db.NewBatch()
	defer batch.Close()

	deviceBytes, err := encoder.DataEncode(device)
	if err != nil {
		return err
	}

	DIDkey := makeDIDKey(userID, nft.Id)

	if err := batch.Set(DIDkey, deviceBytes); err != nil {
		return err
	}

	if err := batch.Set(makeUSNKey(userID, device.Usn), DIDkey); err != nil {
		return err
	}

	if err := batch.Set(makeAddressKey(userID, device.Address, nft.Id), []byte(nft.Id)); err != nil {
		return err
	}

	if err := batch.Write(); err != nil {
		return err
	}

	evt := DeviceSaved{
		Device:    device,
		ProfileID: auth.GetUserID(ctx),
	}

	return ds.eventBus.Emit(ctx, events.DeviceSaved, evt)
}

// GetByAddress fetch all devices owned by the given address
func (ds Service) GetByAddress(ctx context.Context, address string) ([]svcs.Device, error) {
	profileID := auth.GetUserID(ctx)

	devices := make([]svcs.Device, 0)

	prefixDB := db.NewPrefixDB(ds.db, makeAddressKey(profileID, address, ""))

	itr, err := prefixDB.Iterator(nil, nil)
	if err != nil {
		return devices, err
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		DID := string(itr.Value())

		device, err := ds.GetByDID(ctx, DID)
		if err != nil {
			return devices, err
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// GetByUser fetch all devices owned by user
func (ds Service) GetByUser(ctx context.Context) ([]svcs.Device, error) {
	profileID := auth.GetUserID(ctx)

	devices := make([]svcs.Device, 0)

	prefixDB := db.NewPrefixDB(ds.db, makeUSNKey(profileID, ""))

	itr, err := prefixDB.Iterator(nil, nil)
	if err != nil {
		return devices, err
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		DIDbytes, err := ds.db.Get(itr.Value())
		if err != nil {
			return devices, err
		}

		DIDkey := strings.SplitAfterN(string(DIDbytes), ":", 3)

		device, err := ds.GetByDID(ctx, DIDkey[2])
		if err != nil {
			return devices, err
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// DeleteByAddress deletes all devices and all associated keys for device
func (ds Service) DeleteByAddress(ctx context.Context, address string) (uint, error) {
	profileID := auth.GetUserID(ctx)
	deletedRecords := uint(0)

	prefixDB := db.NewPrefixDB(ds.db, makeAddressKey(profileID, address, ""))

	batch := ds.db.NewBatch()
	defer batch.Close()

	itr, err := prefixDB.Iterator(nil, nil)
	if err != nil {
		return deletedRecords, err
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		DID := string(itr.Value())

		device, err := ds.GetByDID(ctx, DID)
		if err != nil {
			return deletedRecords, err
		}

		if device.Address != address {
			return deletedRecords, fmt.Errorf("data integrity error for address %s and DID %s", address, DID)
		}

		if err := batch.Delete(makeUSNKey(profileID, device.Usn)); err != nil {
			return deletedRecords, err
		}

		if err := batch.Delete(makeAddressKey(profileID, address, DID)); err != nil {
			return deletedRecords, err
		}

		if err := batch.Delete(makeDIDKey(profileID, DID)); err != nil {
			return deletedRecords, err
		}

		deletedRecords++
	}

	if err := batch.WriteSync(); err != nil {
		return deletedRecords, err
	}

	return deletedRecords, nil
}

// Delete deletes device by key
func (ds Service) Delete(ctx context.Context, key string) error {
	userID := auth.GetClaims(ctx).UserID

	device, err := ds.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("cannot delete device %s: %w", key, err)
	}

	if err := ds.db.DeleteSync(makeDIDKey(userID, device.DID)); err != nil {
		return fmt.Errorf("cannot delete device %s: %w", key, err)
	}

	return nil
}

// Get fetches device by DID or USN
func (ds Service) Get(ctx context.Context, key string) (svcs.Device, error) {
	if len(key) == sdkdid.DefaultUSNLength {
		return ds.GetByUSN(ctx, key)
	}

	return ds.GetByDID(ctx, key)
}

// GetByDIDs fetches many devices by DIDs
func (ds Service) GetByDIDs(ctx context.Context, dids []string) ([]svcs.Device, error) {
	devices := make([]svcs.Device, 0, len(dids))

	for _, did := range dids {
		device, err := ds.GetByDID(ctx, did)
		if err != nil {
			return devices, err
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// GetByDIDUnsafe fetches device by DID without checking if the device belongs to the user
func (ds Service) GetByDIDUnsafe(_ context.Context, did, userID string) (svcs.Device, error) {
	var d svcs.Device

	DIDkey := makeDIDKey(userID, did)

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

// GetByDID fetches device by DID
func (ds Service) GetByDID(ctx context.Context, did string) (svcs.Device, error) {
	var d svcs.Device

	userID := auth.GetClaims(ctx).UserID

	DIDkey := makeDIDKey(userID, did)

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

// GetByUSN fetches device by USN
func (ds Service) GetByUSN(ctx context.Context, usn string) (svcs.Device, error) {
	var d svcs.Device

	userID := auth.GetClaims(ctx).UserID

	USNkey := makeUSNKey(userID, usn)

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
