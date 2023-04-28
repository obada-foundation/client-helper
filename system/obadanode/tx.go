package obadanode

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txs "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// BuildUnsignedTx builds a transaction to be signed given a set of messages.
// Once created, the fee, memo, and messages are set.
func (c NodeClient) BuildUnsignedTx(msgs ...sdk.Msg) (client.TxBuilder, error) {
	tsn := c.txConfig.NewTxBuilder()

	if err := tsn.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	return tsn, nil
}

// getSimPK gets the public key to use for building a simulation tx.
// Note, we should only check for keys in the keybase if we are in simulate and execute mode,
// e.g. when using --gas=auto.
// When using --dry-run, we are is simulation mode only and should not check the keybase.
// Ref: https://github.com/cosmos/cosmos-sdk/issues/11283
func (c NodeClient) getSimPK() cryptotypes.PubKey {
	return &secp256k1.PubKey{}
}

// BuildSimTx creates an unsigned tx with an empty single signature and returns
// the encoded transaction or an error if the unsigned transaction cannot be
// built.
func (c NodeClient) BuildSimTx(msgs ...sdk.Msg) ([]byte, error) {
	tsn, err := c.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	pk := c.getSimPK()

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := signing.SignatureV2{
		PubKey: pk,
		Data: &signing.SingleSignatureData{
			SignMode: c.txConfig.SignModeHandler().DefaultMode(),
		},
		Sequence: 0,
	}
	if err := tsn.SetSignatures(sig); err != nil {
		return nil, err
	}

	return c.txConfig.TxEncoder()(tsn.GetTx())
}

// CalculateGas simulates the execution of a transaction and returns the
// simulation response obtained by the query and the adjusted gas amount.
func (c NodeClient) CalculateGas(ctx context.Context, msgs ...sdk.Msg,
) (*txs.SimulateResponse, uint64, error) {
	txBytes, err := c.BuildSimTx(msgs...)
	if err != nil {
		return nil, 0, err
	}

	simRes, err := c.serviceClient.Simulate(ctx, &txs.SimulateRequest{
		TxBytes: txBytes,
	})
	if err != nil {
		return nil, 0, err
	}

	return simRes, uint64(1 * float64(simRes.GasInfo.GasUsed)), nil
}

// SendTx sends a transaction to the node.
func (c NodeClient) SendTx(ctx context.Context, msg sdk.Msg, priv cryptotypes.PrivKey) (*ctypes.ResultBroadcastTx, error) {
	accAddress := sdk.AccAddress(priv.PubKey().Address().Bytes()).String()
	nonce, err := c.Nonce(ctx, accAddress)
	if err != nil {
		return nil, err
	}

	tsn, err := c.BuildTx(ctx, msg, priv, nonce)
	if err != nil {
		return nil, err
	}

	txBytes, err := c.txConfig.TxEncoder()(tsn)
	if err != nil {
		return nil, err
	}

	res, err := c.clientHTTP.BroadcastTxSync(ctx, txBytes)

	if err != nil {
		return nil, err
	}
	// Note: In async case, response is returned before TxCheck
	// res, err := c.clientHTTP.BroadcastTxAsync(ctx, txBytes)
	if errRes := client.CheckTendermintError(err, txBytes); errRes != nil {
		return nil, fmt.Errorf("code: %d, log: %s, codespace: %s", errRes.Code, errRes.Logs, res.Codespace)
	}

	if res.Code != 0 {
		if strings.Contains(res.Log, "insufficient funds") {
			return nil, ErrInsufficientFunds
		}
		return nil, fmt.Errorf("code: %d, log: %s, codespace: %s", res.Code, res.Log, res.Codespace)
	}

	return res, nil
}

// BuildTx builds a transaction given a set of messages and a private key.
func (c NodeClient) BuildTx(ctx context.Context, msg sdk.Msg, priv cryptotypes.PrivKey, accSeq uint64) (authsigning.Tx, error) {
	txBuilder := c.txConfig.NewTxBuilder()
	txBuilder.GetTx().GetFee()

	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}
	txBuilder.SetGasLimit(uint64(200000))
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("obd", sdk.NewInt(1))))

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
	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	return txBuilder.GetTx(), nil
}

// Nonce returns the nonce for a given address.
func (c NodeClient) Nonce(ctx context.Context, address string) (uint64, error) {
	acc, err := c.Account(ctx, address)
	if err != nil {
		return 0, err
	}

	return acc.GetSequence(), nil
}
