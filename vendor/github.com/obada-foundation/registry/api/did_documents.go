package api

import (
	"context"
	"errors"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	pb "github.com/obada-foundation/registry/api/pb/v1/diddoc"
	"github.com/obada-foundation/registry/services/diddoc"
	"github.com/obada-foundation/registry/types"
	"github.com/obada-foundation/sdkgo/asset"
	sdkdid "github.com/obada-foundation/sdkgo/did"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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

func (s GRPCServer) checkSignature(pubKey cryptotypes.PubKey, sig []byte, msg proto.Message) error {
	if len(sig) == 0 {
		return status.Errorf(codes.InvalidArgument, "empty signature")
	}

	hash, err := ProtoDeterministicChecksum(msg)
	if err != nil {
		return err
	}

	if pubKey.VerifySignature(hash[:], sig) {
		return nil
	}

	return status.Errorf(codes.Unauthenticated, "unauthorized")
}

// SaveMetadata updates metadata for DID
func (s GRPCServer) SaveMetadata(ctx context.Context, msg *pb.SaveMetadataRequest) (*pb.SaveMetadataResponse, error) {
	resp := &pb.SaveMetadataResponse{}

	data := msg.GetData()

	if len(msg.GetSignature()) == 0 {
		return resp, status.Errorf(codes.InvalidArgument, "empty signature")
	}

	if data.GetAuthenticationKeyId() == "" {
		return resp, status.Errorf(codes.InvalidArgument, "empty autherntication id")
	}

	pubKey, err := s.DIDDocService.GetVerificationKeyByAuthID(ctx, data.GetDid(), data.GetAuthenticationKeyId())
	if err != nil {
		if errors.Is(err, diddoc.ErrDIDNotRegistered) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}

		return resp, err
	}

	if err := s.checkSignature(pubKey, msg.GetSignature(), data); err != nil {
		return resp, err
	}

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

// SaveVerificationMethods saves verification methods for DID
func (s GRPCServer) SaveVerificationMethods(ctx context.Context, msg *pb.MsgSaveVerificationMethods) (*pb.SaveVerificationMethodsResponse, error) {
	resp := &pb.SaveVerificationMethodsResponse{}

	data := msg.GetData()

	if len(msg.GetSignature()) == 0 {
		return resp, status.Errorf(codes.InvalidArgument, "empty signature")
	}

	if data.GetAuthenticationKeyId() == "" {
		return resp, status.Errorf(codes.InvalidArgument, "empty autherntication id")
	}

	pubKey, err := s.DIDDocService.GetVerificationKeyByAuthID(ctx, data.GetDid(), data.GetAuthenticationKeyId())
	if err != nil {
		if errors.Is(err, diddoc.ErrDIDNotRegistered) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}

		return resp, err
	}

	if err := s.checkSignature(pubKey, msg.GetSignature(), data); err != nil {
		return resp, err
	}

	vms := make([]types.VerificationMethod, 0, len(data.GetVerificationMethods()))

	for _, vm := range data.GetVerificationMethods() {
		vms = append(vms, types.VerificationMethod{
			Context:         vm.GetContext(),
			ID:              vm.GetId(),
			Type:            vm.GetType(),
			Controller:      vm.GetController(),
			PublicKeyBase58: vm.GetPublicKeyBase58(),
		})
	}

	if err := s.DIDDocService.SaveVerificationMethods(ctx, data.GetDid(), vms, data.GetAuthentication()); err != nil {
		return resp, err
	}

	return resp, nil
}
