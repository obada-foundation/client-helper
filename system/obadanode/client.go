package obadanode

import (
	"context"
	"math"

	sdkmath "cosmossdk.io/math"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx"
	txtypes "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	proto "github.com/gogo/protobuf/proto"
	obadatypes "github.com/obada-foundation/fullcore/x/obit/types"
	"google.golang.org/grpc"
)

// Client describe OBADA node methods needed for interaction with client helper
type Client interface {
	// Query methods

	// Balance returns the balance of account
	Balance(ctx context.Context, pubKey cryptotypes.PubKey) (*banktypes.QueryBalanceResponse, error)

	// BalanceByAddress returns the balance of specified address
	BalanceByAddress(ctx context.Context, address string) (*banktypes.QueryBalanceResponse, error)

	// BaseDenomMetadata returns the metadata of base denom
	BaseDenomMetadata(ctx context.Context) (banktypes.Metadata, error)

	// GetNFTByAddress returns the NFTs of specified address
	GetNFTByAddress(ctx context.Context, address string) ([]obadatypes.NFT, error)

	// GetNFT returns the NFT with given NFT
	GetNFT(ctx context.Context, DID string) (*obadatypes.NFT, error)

	// HasAccount returns true if there at least one tx recordred in blockchain
	HasAccount(ctx context.Context, address string) (bool, error)

	// Account returns the account of specified address
	Account(ctx context.Context, address string) (acc sdk.AccountI, err error)

	// Tx methods
	SendTx(ctx context.Context, cnf TxCustomConfig) (*ctypes.ResultBroadcastTx, error)

	// CalculateGas returns the gas needed to execute the given message
	CalculateGas(ctx context.Context, msgs ...sdk.Msg) (*tx.SimulateResponse, uint64, error)

	// DecodeTx decodes the given tx bytes
	DecodeTx(b []byte) (Tx, error)
}

// NodeClient stores dependencies for OBADA client
type NodeClient struct {
	conn *grpc.ClientConn

	clientHTTP    *rpchttp.HTTP
	authClient    authtypes.QueryClient
	bankClient    banktypes.QueryClient
	obadaClient   obadatypes.QueryClient
	serviceClient tx.ServiceClient

	cdc      *codec.ProtoCodec
	txConfig client.TxConfig
	chainID  string
}

// Tx blockchain transaction
type Tx struct {
	sdk.Tx

	codec codec.ProtoCodecMarshaler
}

// NewClient creates a new OBADA node client
func NewClient(ctx context.Context, chainID, rpcURI, grpcURI string) (NodeClient, error) {
	var (
		c = NodeClient{
			chainID: chainID,
		}
		encCfg = testutil.MakeTestEncodingConfig()
		err    error
	)

	if c.clientHTTP, err = rpchttp.New(rpcURI, "/websocket"); err != nil {
		return c, err
	}

	if c.conn, err = grpc.Dial(grpcURI, grpc.WithInsecure()); err != nil { // nolint:staticcheck // for further refactoring
		return c, err
	}

	c.serviceClient = tx.NewServiceClient(c.conn)
	c.authClient = authtypes.NewQueryClient(c.conn)
	c.bankClient = banktypes.NewQueryClient(c.conn)
	c.obadaClient = obadatypes.NewQueryClient(c.conn)

	encCfg.InterfaceRegistry.RegisterInterface("AccountI", (*sdk.AccountI)(nil), &authtypes.BaseAccount{})
	encCfg.InterfaceRegistry.RegisterInterface("obadafoundation.fullcore.obit.NFTData", (*proto.Message)(nil), &obadatypes.NFTData{})
	encCfg.InterfaceRegistry.RegisterImplementations((*sdk.Msg)(nil),
		&obadatypes.MsgMintNFT{},
		&obadatypes.MsgUpdateNFT{},
		&obadatypes.MsgTransferNFT{},
		&obadatypes.MsgUpdateUriHash{},
		&obadatypes.MsgBatchTransferNFT{},
		&obadatypes.MsgBatchMintNFT{},
	)

	c.cdc = codec.NewProtoCodec(encCfg.InterfaceRegistry)

	c.txConfig = txtypes.NewTxConfig(c.cdc, txtypes.DefaultSignModes)

	baseDenomMetdata, err := c.BaseDenomMetadata(ctx)
	if err != nil {
		return c, err
	}

	if _, ok := sdk.GetDenomUnit(baseDenom); !ok {
		if er := sdk.RegisterDenom(baseDenom, sdkmath.LegacyNewDec(1)); er != nil {
			return c, er
		}
	}

	for _, denomUnit := range baseDenomMetdata.DenomUnits {
		if denomUnit.Denom != baseDenom {
			if _, ok := sdk.GetDenomUnit(denomUnit.Denom); !ok {
				exp := int64(1 * math.Pow10(int(denomUnit.Exponent)))

				if er := sdk.RegisterDenom(denomUnit.Denom, sdkmath.LegacyNewDec(exp)); er != nil {
					return c, er
				}
			}
		}
	}

	return c, nil
}

// Close close the connection to the node
func (c NodeClient) Close() {
	_ = c.conn.Close()
}

// DecodeTx decodes transaction from bytes
func (c NodeClient) DecodeTx(b []byte) (Tx, error) {
	transaction, err := txtypes.DefaultTxDecoder(c.cdc)(b)
	if err != nil {
		return Tx{}, err
	}

	return Tx{
		transaction,
		c.cdc,
	}, nil

}
