package api

import (
	"os"

	pbacc "github.com/obada-foundation/registry/api/pb/v1/account"
	pbdidoc "github.com/obada-foundation/registry/api/pb/v1/diddoc"
	"github.com/obada-foundation/registry/services/account"
	"github.com/obada-foundation/registry/services/diddoc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer implements server gRPC calls for the DID registry.
type GRPCServer struct {
	pbdidoc.UnimplementedDIDDocServer
	pbacc.UnimplementedAccountServer

	Log *zap.SugaredLogger

	// Services
	DIDDocService  diddoc.DIDDoc
	AccountService account.Account
}

// GRPCConfig defines the dependencies for the gRPC server
type GRPCConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger

	// Services
	DIDDocService  diddoc.DIDDoc
	AccountService account.Account
}

// NewGRPCServer creates a new grpc server
func NewGRPCServer(cfg GRPCConfig) (*grpc.Server, *GRPCServer) {
	srv := &GRPCServer{
		Log: cfg.Log,

		// Services
		DIDDocService:  cfg.DIDDocService,
		AccountService: cfg.AccountService,
	}

	grpcServer := grpc.NewServer()
	pbdidoc.RegisterDIDDocServer(grpcServer, srv)
	pbacc.RegisterAccountServer(grpcServer, srv)
	reflection.Register(grpcServer)

	return grpcServer, srv
}
