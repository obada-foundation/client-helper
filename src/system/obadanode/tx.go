package obadanode

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func (c *NodeClient) SendTx(ctx context.Context, msg sdk.Msg, priv cryptotypes.PrivKey) (*ctypes.ResultBroadcastTx, error) {
	accAddress := sdk.AccAddress(priv.PubKey().Address().Bytes()).String()
	nonce, err := c.Nonce(ctx, accAddress)
	if err != nil {
		return nil, err
	}

	tx, err := c.BuildTx(ctx, msg, priv, nonce)
	if err != nil {
		return nil, err
	}

	txBytes, err := c.txConfig.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}

	res, err := c.clientHTTP.BroadcastTxSync(ctx, txBytes)
	// Note: In async case, response is returnd before TxCheck
	// res, err := c.clientHTTP.BroadcastTxAsync(ctx, txBytes)
	if errRes := client.CheckTendermintError(err, txBytes); errRes != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("code: %d, log: %s, codespace: %s\n", res.Code, res.Log, res.Codespace)
	}

	return res, nil
}

func (c *NodeClient) BuildTx(ctx context.Context, msg sdk.Msg, priv cryptotypes.PrivKey, accSeq uint64) (authsigning.Tx, error) {
	txBuilder := c.txConfig.NewTxBuilder()

	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}
	txBuilder.SetGasLimit(uint64(200000))

	// First round: we gather all the signer infos. We use the "set empty signature" hack to do that.
	if err = txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  c.txConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: accSeq,
	}); err != nil {
		return nil, err
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
	if err = txBuilder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	return txBuilder.GetTx(), nil
}

func (c NodeClient) Nonce(ctx context.Context, address string) (uint64, error) {
	acc, err := c.Account(ctx, address)
	if err != nil {
		return 0, err
	}

	return acc.GetSequence(), nil
}
