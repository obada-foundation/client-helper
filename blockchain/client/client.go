package client

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	txtypes "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/obada-foundation/client-helper/services"
	obadatypes "github.com/obada-foundation/fullcore/x/obit/types"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc"
)

// ObadaChainClient stores dependencies needed for making requests to the node
type ObadaChainClient struct {
	conn *grpc.ClientConn

	clientHTTP *rpchttp.HTTP
	authClient authtypes.QueryClient

	cdc      *codec.ProtoCodec
	txConfig client.TxConfig
	chainID  string
}

// NewClient creates a new client
func NewClient(chainID, rpcURI, grpcURI string) (ObadaChainClient, error) {
	var (
		c = ObadaChainClient{
			chainID: chainID,
		}
		encCfg = simapp.MakeTestEncodingConfig()
		err    error
	)

	if c.clientHTTP, err = rpchttp.New(rpcURI, "/websocket"); err != nil {
		return c, err
	}

	if c.conn, err = grpc.Dial(grpcURI, grpc.WithInsecure()); err != nil { //nolint:staticcheck //for future refactoring
		return c, err
	}

	c.authClient = authtypes.NewQueryClient(c.conn)

	c.cdc = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	c.txConfig = txtypes.NewTxConfig(c.cdc, txtypes.DefaultSignModes)

	return c, nil
}

// Close closes the connection to the node
func (c ObadaChainClient) Close() {
	_ = c.conn.Close()
}

// BuildTx builds a transaction
func (c *ObadaChainClient) BuildTx(ctx context.Context, msg sdk.Msg, priv cryptotypes.PrivKey, accSeq uint64) (authsigning.Tx, error) {
	txBuilder := c.txConfig.NewTxBuilder()

	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}
	txBuilder.SetGasLimit(uint64(200000))

	// First round: we gather all the signer infos. We use the "set empty signature" hack to do that.
	if er := txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  c.txConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: accSeq,
	}); er != nil {
		return nil, er
	}

	accAddress := sdk.AccAddress(priv.PubKey().Address().Bytes()).String()

	acc, err := c.Account(ctx, accAddress)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	signerData := authsigning.SignerData{
		ChainID:       c.chainID,
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      accSeq,
	}
	sigV2, err := tx.SignWithPrivKey(
		c.txConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, c.txConfig, accSeq)
	if err != nil {
		return nil, err
	}
	if er := txBuilder.SetSignatures(sigV2); er != nil {
		return nil, er
	}

	return txBuilder.GetTx(), nil
}

// Account returns the account info
func (c ObadaChainClient) Account(ctx context.Context, address string) (acc authtypes.AccountI, err error) {
	req := &authtypes.QueryAccountRequest{Address: address}

	res, err := c.authClient.Account(ctx, req)
	if err != nil {

		fmt.Printf("cannot get account information: %s", err.Error())
		return
	}

	if err = c.cdc.UnpackAny(res.GetAccount(), &acc); err != nil {
		return
	}

	return
}

// SendTx sends a transaction to the chain
func (c *ObadaChainClient) SendTx(ctx context.Context, msg sdk.Msg, priv cryptotypes.PrivKey, seq uint64) (*ctypes.ResultBroadcastTx, error) {
	tsn, err := c.BuildTx(ctx, msg, priv, seq)
	if err != nil {
		return nil, err
	}

	txBytes, err := c.txConfig.TxEncoder()(tsn)
	if err != nil {
		return nil, err
	}

	res, err := c.clientHTTP.BroadcastTxSync(ctx, txBytes)
	// Note: In async case, response is returned before TxCheck
	// res, err := c.clientHTTP.BroadcastTxAsync(ctx, txBytes)
	if errRes := client.CheckTendermintError(err, txBytes); errRes != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("code: %d, log: %s, codespace: %s", res.Code, res.Log, res.Codespace)
	}

	return res, nil
}

// Nonce returns nonce
func (c ObadaChainClient) Nonce(ctx context.Context, address string) (uint64, error) {
	acc, err := c.Account(ctx, address)
	if err != nil {
		return 0, err
	}

	return acc.GetSequence(), nil
}

// Mint is creating new NFT
func (c *ObadaChainClient) Mint(ctx context.Context, priv cryptotypes.PrivKey, localNFT services.LocalNFT) (*ctypes.ResultBroadcastTx, error) {
	accAddress := sdk.AccAddress(priv.PubKey().Address().Bytes()).String()
	nonce, err := c.Nonce(ctx, accAddress)

	if err != nil {
		return nil, err
	}

	msg := obadatypes.MsgMintNFT{
		Creator: accAddress,
		Id:      localNFT.SerialNumber,
		Uri:     "",
		UriHash: "",
		Usn:     "",
	}

	res, err := c.SendTx(ctx, &msg, priv, nonce)
	if err != nil {
		return nil, err
	}

	return res, nil
}
