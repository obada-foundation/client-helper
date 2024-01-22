// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	context "context"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"

	cosmos_sdktypes "github.com/cosmos/cosmos-sdk/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	mock "github.com/stretchr/testify/mock"

	obadanode "github.com/obada-foundation/client-helper/system/obadanode"

	obittypes "github.com/obada-foundation/fullcore/x/obit/types"

	tx "github.com/cosmos/cosmos-sdk/types/tx"

	types "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

func (_m *Client) BaseDenomMetadata(ctx context.Context) (banktypes.Metadata, error) {
	return banktypes.Metadata{}, nil
}

// Account provides a mock function with given fields: ctx, address
func (_m *Client) Account(ctx context.Context, address string) (types.AccountI, error) {
	ret := _m.Called(ctx, address)

	var r0 types.AccountI
	if rf, ok := ret.Get(0).(func(context.Context, string) types.AccountI); ok {
		r0 = rf(ctx, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.AccountI)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Balance provides a mock function with given fields: ctx, pubKey
func (_m *Client) Balance(ctx context.Context, pubKey cryptotypes.PubKey) (*banktypes.QueryBalanceResponse, error) {
	ret := _m.Called(ctx, pubKey)

	var r0 *banktypes.QueryBalanceResponse
	if rf, ok := ret.Get(0).(func(context.Context, cryptotypes.PubKey) *banktypes.QueryBalanceResponse); ok {
		r0 = rf(ctx, pubKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*banktypes.QueryBalanceResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, cryptotypes.PubKey) error); ok {
		r1 = rf(ctx, pubKey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BalanceByAddress provides a mock function with given fields: ctx, address
func (_m *Client) BalanceByAddress(ctx context.Context, address string) (*banktypes.QueryBalanceResponse, error) {
	ret := _m.Called(ctx, address)

	var r0 *banktypes.QueryBalanceResponse
	if rf, ok := ret.Get(0).(func(context.Context, string) *banktypes.QueryBalanceResponse); ok {
		r0 = rf(ctx, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*banktypes.QueryBalanceResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CalculateGas provides a mock function with given fields: ctx, msgs
func (_m *Client) CalculateGas(ctx context.Context, msgs ...cosmos_sdktypes.Msg) (*tx.SimulateResponse, uint64, error) {
	_va := make([]interface{}, len(msgs))
	for _i := range msgs {
		_va[_i] = msgs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *tx.SimulateResponse
	if rf, ok := ret.Get(0).(func(context.Context, ...cosmos_sdktypes.Msg) *tx.SimulateResponse); ok {
		r0 = rf(ctx, msgs...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tx.SimulateResponse)
		}
	}

	var r1 uint64
	if rf, ok := ret.Get(1).(func(context.Context, ...cosmos_sdktypes.Msg) uint64); ok {
		r1 = rf(ctx, msgs...)
	} else {
		r1 = ret.Get(1).(uint64)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, ...cosmos_sdktypes.Msg) error); ok {
		r2 = rf(ctx, msgs...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DecodeTx provides a mock function with given fields: b
func (_m *Client) DecodeTx(b []byte) (obadanode.Tx, error) {
	ret := _m.Called(b)

	var r0 obadanode.Tx
	if rf, ok := ret.Get(0).(func([]byte) obadanode.Tx); ok {
		r0 = rf(b)
	} else {
		r0 = ret.Get(0).(obadanode.Tx)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(b)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNFT provides a mock function with given fields: ctx, DID
func (_m *Client) GetNFT(ctx context.Context, DID string) (*obittypes.NFT, error) {
	ret := _m.Called(ctx, DID)

	var r0 *obittypes.NFT
	if rf, ok := ret.Get(0).(func(context.Context, string) *obittypes.NFT); ok {
		r0 = rf(ctx, DID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*obittypes.NFT)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, DID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNFTByAddress provides a mock function with given fields: ctx, address
func (_m *Client) GetNFTByAddress(ctx context.Context, address string) ([]obittypes.NFT, error) {
	ret := _m.Called(ctx, address)

	var r0 []obittypes.NFT
	if rf, ok := ret.Get(0).(func(context.Context, string) []obittypes.NFT); ok {
		r0 = rf(ctx, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]obittypes.NFT)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HasAccount provides a mock function with given fields: ctx, address
func (_m *Client) HasAccount(ctx context.Context, address string) (bool, error) {
	ret := _m.Called(ctx, address)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, address)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendTx provides a mock function with given fields: ctx, msg, priv
func (_m *Client) SendTx(ctx context.Context, msg cosmos_sdktypes.Msg, priv cryptotypes.PrivKey) (*coretypes.ResultBroadcastTx, error) {
	ret := _m.Called(ctx, msg, priv)

	var r0 *coretypes.ResultBroadcastTx
	if rf, ok := ret.Get(0).(func(context.Context, cosmos_sdktypes.Msg, cryptotypes.PrivKey) *coretypes.ResultBroadcastTx); ok {
		r0 = rf(ctx, msg, priv)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBroadcastTx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, cosmos_sdktypes.Msg, cryptotypes.PrivKey) error); ok {
		r1 = rf(ctx, msg, priv)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewClient(t mockConstructorTestingTNewClient) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
