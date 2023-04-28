package client

import (
	"context"

	acc "github.com/obada-foundation/registry/api/pb/v1/account"
	"github.com/obada-foundation/registry/api/pb/v1/diddoc"
	"google.golang.org/grpc"
)

// Client is allows communication with grpc server
type Client interface {
	acc.AccountClient
	diddoc.DIDDocClient

	Close() error
}

type grpcClient struct {
	cc      *grpc.ClientConn
	account acc.AccountClient
	diddoc  diddoc.DIDDocClient
}

// NewClient creates a new instance of Client
func NewClient(conn *grpc.ClientConn) Client {
	return grpcClient{
		cc:      conn,
		account: acc.NewAccountClient(conn),
		diddoc:  diddoc.NewDIDDocClient(conn),
	}
}

// GetPublicKey register a new public key
func (c grpcClient) GetPublicKey(ctx context.Context, msg *acc.GetPublicKeyRequest, opts ...grpc.CallOption) (*acc.GetPublicKeyResponse, error) {
	return c.account.GetPublicKey(ctx, msg, opts...)
}

// RegisterAccount register a new public key
func (c grpcClient) RegisterAccount(ctx context.Context, msg *acc.RegisterAccountRequest, opts ...grpc.CallOption) (*acc.RegisterAccountResponse, error) {
	return c.account.RegisterAccount(ctx, msg, opts...)
}

// Register register a new DID document
func (c grpcClient) Register(ctx context.Context, msg *diddoc.RegisterRequest, opts ...grpc.CallOption) (*diddoc.RegisterResponse, error) {
	return c.diddoc.Register(ctx, msg, opts...)
}

// Get fetches DID document from the registry
func (c grpcClient) Get(ctx context.Context, msg *diddoc.GetRequest, opts ...grpc.CallOption) (*diddoc.GetResponse, error) {
	return c.diddoc.Get(ctx, msg, opts...)
}

// GetMetadataHistory returns metadata history
func (c grpcClient) GetMetadataHistory(ctx context.Context, msg *diddoc.GetMetadataHistoryRequest, opts ...grpc.CallOption) (*diddoc.GetMetadataHistoryResponse, error) {
	return c.diddoc.GetMetadataHistory(ctx, msg, opts...)
}

// SaveMetadata saves DID document metadata
func (c grpcClient) SaveMetadata(ctx context.Context, msg *diddoc.SaveMetadataRequest, opts ...grpc.CallOption) (*diddoc.SaveMetadataResponse, error) {
	return c.diddoc.SaveMetadata(ctx, msg, opts...)
}

// Close close all client connections
func (c grpcClient) Close() error {
	return c.cc.Close()
}
