package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	proto "github.com/gogo/protobuf/proto"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgMintNFT{}, "obit/MsgMintNFT", nil)
	cdc.RegisterConcrete(&MsgUpdateNFT{}, "obit/MsgUpdateNFT", nil)
	cdc.RegisterConcrete(&MsgTransferNFT{}, "obit/MsgTransferNFT", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterInterface("obadafoundation.fullcore.obit.NFTData", (*proto.Message)(nil), &NFTData{})
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMintNFT{},
		&MsgUpdateNFT{},
		&MsgTransferNFT{},
	)

	// this line is used by starport scaffolding # 3
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
