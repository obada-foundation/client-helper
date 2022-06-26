package obadanode

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	txtypes "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	obadatypes "github.com/obada-foundation/fullcore/x/obit/types"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"google.golang.org/grpc"
)

type NodeClient struct {
	conn *grpc.ClientConn

	clientHTTP  *rpchttp.HTTP
	authClient  authtypes.QueryClient
	bankClient  banktypes.QueryClient
	obadaClient obadatypes.QueryClient

	cdc      *codec.ProtoCodec
	txConfig client.TxConfig
	chainID  string
}

func NewClient(chainID, rpcURI, grpcURI string) (NodeClient, error) {
	var (
		c = NodeClient{
			chainID: chainID,
		}
		encCfg = simapp.MakeTestEncodingConfig()
		err    error
	)

	if c.clientHTTP, err = rpchttp.New(rpcURI, "/websocket"); err != nil {
		return c, err
	}

	if c.conn, err = grpc.Dial(grpcURI, grpc.WithInsecure()); err != nil {
		return c, err
	}

	c.authClient = authtypes.NewQueryClient(c.conn)
	c.bankClient = banktypes.NewQueryClient(c.conn)
	c.obadaClient = obadatypes.NewQueryClient(c.conn)

	c.cdc = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	c.txConfig = txtypes.NewTxConfig(c.cdc, txtypes.DefaultSignModes)

	return c, nil
}

func (c NodeClient) Close() {
	c.conn.Close()
}
