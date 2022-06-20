package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgCreateTa = "create_ta"
	TypeMsgUpdateTa = "update_ta"
	TypeMsgDeleteTa = "delete_ta"
)

var _ sdk.Msg = &MsgCreateTa{}

func NewMsgCreateTa(creator string, title string, pubkey string) *MsgCreateTa {
	return &MsgCreateTa{
		Creator: creator,
		Title:   title,
		Pubkey:  pubkey,
	}
}

func (msg *MsgCreateTa) Route() string {
	return RouterKey
}

func (msg *MsgCreateTa) Type() string {
	return TypeMsgCreateTa
}

func (msg *MsgCreateTa) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateTa) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateTa) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateTa{}

func NewMsgUpdateTa(creator string, id uint64, title string, pubkey string) *MsgUpdateTa {
	return &MsgUpdateTa{
		Id:      id,
		Creator: creator,
		Title:   title,
		Pubkey:  pubkey,
	}
}

func (msg *MsgUpdateTa) Route() string {
	return RouterKey
}

func (msg *MsgUpdateTa) Type() string {
	return TypeMsgUpdateTa
}

func (msg *MsgUpdateTa) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateTa) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateTa) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgDeleteTa{}

func NewMsgDeleteTa(creator string, id uint64) *MsgDeleteTa {
	return &MsgDeleteTa{
		Id:      id,
		Creator: creator,
	}
}
func (msg *MsgDeleteTa) Route() string {
	return RouterKey
}

func (msg *MsgDeleteTa) Type() string {
	return TypeMsgDeleteTa
}

func (msg *MsgDeleteTa) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteTa) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteTa) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
