package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	proto "github.com/gogo/protobuf/proto"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgMintObit{}, "obit/MintObit", nil)
	cdc.RegisterConcrete(&MsgSend{}, "obit/MsgSend", nil)
	cdc.RegisterConcrete(&MsgCreateTa{}, "obit/CreateTa", nil)
	cdc.RegisterConcrete(&MsgUpdateTa{}, "obit/UpdateTa", nil)
	cdc.RegisterConcrete(&MsgDeleteTa{}, "obit/DeleteTa", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterInterface("obadafoundation.fullcore.obit.NFTData", (*proto.Message)(nil), &NFTData{})
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMintObit{},
		&MsgSend{},
		&MsgCreateTa{},
		&MsgUpdateTa{},
		&MsgDeleteTa{},
	)

	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
