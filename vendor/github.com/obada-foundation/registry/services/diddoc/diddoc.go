package diddoc

import (
	"bytes"
	"context"
	"encoding/gob"
	"strings"

	immudb "github.com/codenotary/immudb/pkg/client"
	"github.com/obada-foundation/registry/system/encoder"
	"github.com/obada-foundation/registry/types"
	"github.com/obada-foundation/sdkgo/asset"
	sdkdid "github.com/obada-foundation/sdkgo/did"
	"go.uber.org/zap"
)

// DIDDoc defines an API for work with DID documents
type DIDDoc interface {
	// Register registers a new DID document in the registry
	Register(ctx context.Context, did string, vm []types.VerificationMethod, a []string) error

	// Get retrieves a DID document from the registry
	Get(ctx context.Context, did string) (types.DIDDocument, error)

	// GetMetadataHistory returns the history of asset data changes
	GetMetadataHistory(ctx context.Context, did string) (asset.DataArrayVersions, error)

	// SaveMetadata saves metadata to the registry
	SaveMetadata(ctx context.Context, did string, m []asset.Object) error
}

// Service implements DIDDoc
type Service struct {
	db     immudb.ImmuClient
	logger *zap.SugaredLogger
}

// NewService creates a new instance of DIDDoc service
func NewService(db immudb.ImmuClient, logger *zap.SugaredLogger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

// GetMetadataHistory implements DIDDoc GetMetadataHistory
func (s Service) GetMetadataHistory(ctx context.Context, did string) (asset.DataArrayVersions, error) {
	DIDDoc, err := s.Get(ctx, did)
	if err != nil {
		return DIDDoc.MetadataHistory, err
	}

	return DIDDoc.MetadataHistory, nil
}

// SaveMetadata implements DIDDoc SaveMetadata
func (s Service) SaveMetadata(ctx context.Context, did string, m []asset.Object) error {
	DIDDoc, err := s.Get(ctx, did)
	if err != nil {
		return err
	}

	currentVersion := len(DIDDoc.MetadataHistory)
	newVersion := currentVersion + 1

	if newVersion == 1 {
		DIDDoc.MetadataHistory = make(asset.DataArrayVersions, 1)
	}

	dataArray := asset.DataArray{
		Objects: m,
	}

	DIDDoc.MetadataHistory[newVersion] = dataArray
	rootHash, err := asset.RootHash(DIDDoc.MetadataHistory, nil)
	if err != nil {
		return err
	}

	versionHash, err := asset.VersionHash(nil, m)
	if err != nil {
		return err
	}

	dataArray.RootHash = rootHash.GetHash()
	dataArray.VersionHash = versionHash.GetHash()
	DIDDoc.MetadataHistory[newVersion] = dataArray

	gobData, err := encoder.DataEncode(DIDDoc)
	if err != nil {
		return err
	}

	_, err = s.db.Set(ctx, []byte(did), gobData)
	if err != nil {
		return err
	}

	return nil
}

// Register implements DIDDoc Register
func (s Service) Register(ctx context.Context, did string, vm []types.VerificationMethod, a []string) error {
	DID, err := sdkdid.FromString(did, nil)
	if err != nil {
		return err
	}

	if _, err = s.Get(ctx, DID.String()); err == nil {
		return ErrDIDAlreadyRegistered
	}

	DIDDoc := types.NewDIDDoc()
	DIDDoc.ID = DID.String()
	DIDDoc.VerificationMethod = vm
	DIDDoc.Authentication = a

	gobData, err := encoder.DataEncode(DIDDoc)
	if err != nil {
		return err
	}

	_, err = s.db.Set(ctx, []byte(DID.String()), gobData)
	if err != nil {
		return err
	}

	s.logger.Debugf("New DID registered: %q", DID)

	return nil
}

// Get implements DIDDoc Get
func (s Service) Get(ctx context.Context, did string) (types.DIDDocument, error) {
	var DIDDoc types.DIDDocument

	entry, err := s.db.Get(ctx, []byte(did))
	if err != nil {
		if strings.Contains(err.Error(), "key not found") {
			return DIDDoc, ErrDIDNotRegistered
		}

		return DIDDoc, err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(entry.Value))
	if err := dec.Decode(&DIDDoc); err != nil {
		return DIDDoc, err
	}

	lastMetadataVersion := len(DIDDoc.MetadataHistory)

	if lastMetadataVersion > 0 {
		DIDDoc.Metadata = types.Metadata{
			VersionID:   lastMetadataVersion,
			VersionHash: DIDDoc.MetadataHistory[lastMetadataVersion].VersionHash,
			RootHash:    DIDDoc.MetadataHistory[lastMetadataVersion].RootHash,
			Objects:     DIDDoc.MetadataHistory[lastMetadataVersion].Objects,
		}
	}

	return DIDDoc, nil
}
