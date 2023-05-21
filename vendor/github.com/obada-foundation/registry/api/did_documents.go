package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	pb "github.com/obada-foundation/registry/api/pb/v1/diddoc"
	"github.com/obada-foundation/registry/services/diddoc"
	"github.com/obada-foundation/registry/types"
	"github.com/obada-foundation/sdkgo/asset"
	"github.com/obada-foundation/sdkgo/base58"
	sdkdid "github.com/obada-foundation/sdkgo/did"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Get DID document from the registry
func (s GRPCServer) Get(ctx context.Context, msg *pb.GetRequest) (*pb.GetResponse, error) {
	doc, err := s.checkIfRegistered(ctx, msg.GetDid())
	if err != nil {
		return nil, err
	}

	pbdoc := &pb.DIDDocument{
		Context:              doc.Context,
		Id:                   doc.ID,
		Controller:           doc.Controller,
		Authentication:       doc.Authentication,
		AssertionMethod:      doc.AssertionMethod,
		CapabilityInvocation: doc.CapabilityInvocation,
		CapabilityDelegation: doc.CapabilityDelegation,
		KeyAgreement:         doc.KeyAgreement,
		AlsoKnownAs:          doc.AlsoKnownAs,
	}

	pbdoc.Metadata = &pb.Metadata{
		VersionId:   int32(doc.Metadata.VersionID),
		VersionHash: doc.Metadata.VersionHash,
		RootHash:    doc.Metadata.RootHash,
	}

	for _, obj := range doc.Metadata.Objects {
		pbdoc.Metadata.Objects = append(pbdoc.Metadata.Objects, &pb.Object{
			Url:                     obj.URL,
			HashEncryptedDataObject: obj.HashEncryptedDataObject,
			HashUnencryptedObject:   obj.HashUnencryptedObject,
			HashUnencryptedMetadata: obj.HashUnencryptedMetadata,
			HashEncryptedMetadata:   obj.HashEncryptedMetadata,
			DataObjectHash:          obj.DataObjectHash,
			Metadata:                obj.Metadata,
		})
	}

	for _, service := range doc.Service {
		pbdoc.Service = append(pbdoc.Service, &pb.Service{
			Id:              service.ID,
			Type:            service.Type,
			ServiceEndpoint: service.ServiceEndpoint,
		})
	}

	for _, verifyMethod := range doc.VerificationMethod {
		pbdoc.VerificationMethod = append(pbdoc.VerificationMethod, &pb.VerificationMethod{
			Id:         verifyMethod.ID,
			Type:       verifyMethod.Type,
			Controller: verifyMethod.Controller,
			//	PublicKeyJwk: verifyMethod.PublicKeyJwk,
			PublicKeyMultibase: verifyMethod.PublicKeyMultibase,
			PublicKeyBase58:    verifyMethod.PublicKeyBase58,
		})
	}

	return &pb.GetResponse{
		Document: pbdoc,
	}, nil
}

// Register DID in the registry
func (s GRPCServer) Register(ctx context.Context, msg *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	resp := &pb.RegisterResponse{}

	verificationMethods := make([]types.VerificationMethod, len(msg.VerificationMethod))
	for _, verifyMethod := range msg.GetVerificationMethod() {
		verificationMethods = append(verificationMethods, types.VerificationMethod{
			ID:         verifyMethod.GetId(),
			Type:       verifyMethod.GetType(),
			Controller: verifyMethod.Controller,
			//	PublicKeyJwk: verifyMethod.PublicKeyJwk,
			PublicKeyMultibase: verifyMethod.GetPublicKeyMultibase(),
			PublicKeyBase58:    verifyMethod.GetPublicKeyBase58(),
		})
	}

	if err := s.DIDDocService.Register(ctx, msg.GetDid(), verificationMethods, msg.GetAuthentication()); err != nil {
		if errors.Is(err, diddoc.ErrDIDAlreadyRegistered) {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}

		if err != sdkdid.ErrNotSupportedDIDMethod {
			return resp, fmt.Errorf("cannot create DID from string: %w", err)
		}

		return resp, err
	}

	return resp, nil
}

// SaveMetadata updates metadata for DID
func (s GRPCServer) SaveMetadata(ctx context.Context, msg *pb.SaveMetadataRequest) (*pb.SaveMetadataResponse, error) {
	resp := &pb.SaveMetadataResponse{}

	if len(msg.GetSignature()) == 0 {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
	}

	data := msg.GetData()

	DIDDoc, err := s.checkIfRegistered(ctx, data.GetDid())
	if err != nil {
		return resp, err
	}

	hash, err := MetadataDeterministicChecksum(data)
	if err != nil {
		return resp, err
	}

	for _, authKey := range DIDDoc.Authentication {
		for _, method := range DIDDoc.VerificationMethod {
			if method.ID == authKey && method.PublicKeyBase58 != "" {
				pubKey := secp256k1.PubKey{
					Key: base58.Decode(method.PublicKeyBase58),
				}

				if pubKey.VerifySignature(hash[:], msg.GetSignature()) {
					objects := make([]asset.Object, 0, len(data.GetObjects()))

					for _, obj := range data.GetObjects() {
						objects = append(objects, asset.Object{
							URL:                     obj.GetUrl(),
							HashEncryptedDataObject: obj.GetHashEncryptedDataObject(),
							HashUnencryptedObject:   obj.GetHashUnencryptedObject(),
							Metadata:                obj.GetMetadata(),
							HashUnencryptedMetadata: obj.GetHashUnencryptedMetadata(),
							HashEncryptedMetadata:   obj.GetHashEncryptedMetadata(),
						})
					}

					if err := s.DIDDocService.SaveMetadata(ctx, data.GetDid(), objects); err != nil {
						return resp, err
					}

					return resp, nil
				}
			}
		}
	}

	return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
}

// GetMetadataHistory returns historical records of metadata changes
func (s GRPCServer) GetMetadataHistory(ctx context.Context, msg *pb.GetMetadataHistoryRequest) (*pb.GetMetadataHistoryResponse, error) {
	if _, err := s.checkIfRegistered(ctx, msg.GetDid()); err != nil {
		return nil, err
	}

	history, err := s.DIDDocService.GetMetadataHistory(ctx, msg.GetDid())
	if err != nil {
		return nil, err
	}

	pbhistory := make(map[int32]*pb.DataArray)

	for version, data := range history {

		pbhistoryVerData := &pb.DataArray{
			VersionHash: data.VersionHash,
			RootHash:    data.RootHash,
		}

		for _, obj := range data.Objects {
			pbhistoryVerData.Objects = append(pbhistoryVerData.Objects, &pb.Object{
				Url:                     obj.URL,
				HashEncryptedDataObject: obj.HashEncryptedDataObject,
				HashUnencryptedObject:   obj.HashUnencryptedObject,
				HashUnencryptedMetadata: obj.HashUnencryptedMetadata,
				HashEncryptedMetadata:   obj.HashEncryptedMetadata,
				DataObjectHash:          obj.DataObjectHash,
				Metadata:                obj.Metadata,
			})
		}

		pbhistory[int32(version)] = pbhistoryVerData
	}

	return &pb.GetMetadataHistoryResponse{
		MetadataHistory: pbhistory,
	}, nil
}

func (s GRPCServer) checkIfRegistered(ctx context.Context, did string) (*types.DIDDocument, error) {
	didDoc, err := s.DIDDocService.Get(ctx, did)
	if err != nil {
		if errors.Is(err, diddoc.ErrDIDNotRegistered) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}

		return nil, err
	}

	return &didDoc, nil
}
