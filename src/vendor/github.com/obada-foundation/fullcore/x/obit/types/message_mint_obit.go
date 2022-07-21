package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgMintObit     = "mint_obit"
	TypeMsgSend         = "send_obit"
	TypeMsgEditMetadata = "edit_metadata"
)

var _ sdk.Msg = &MsgMintObit{}

func NewMsgMintObit(creator, snHash, manufacturer, pn, taToken, uri, uriHash string) *MsgMintObit {
	return &MsgMintObit{
		Creator:          creator,
		SerialNumberHash: snHash,
		Manufacturer:     manufacturer,
		PartNumber:       pn,
		TrustAnchorToken: taToken,
		Uri:              uri,
		UriHash:          uriHash,
	}
}

func (msg *MsgMintObit) Route() string {
	return RouterKey
}

func (msg *MsgMintObit) Type() string {
	return TypeMsgMintObit
}

func (msg *MsgMintObit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgMintObit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMintObit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgEditMetadata{}

func NewMsgEditMetadata(DID, editor string, metadata NFTData) *MsgEditMetadata {
	return &MsgEditMetadata{
		Did:     DID,
		Editor:  editor,
		NFTData: &metadata,
	}
}

func (msg *MsgEditMetadata) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Editor)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid editor address (%s)", err)
	}

	return nil
}

func (msg *MsgEditMetadata) GetSigners() []sdk.AccAddress {
	editor, err := sdk.AccAddressFromBech32(msg.Editor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{editor}
}

func (msg *MsgEditMetadata) Route() string {
	return RouterKey
}

func (msg *MsgEditMetadata) Type() string {
	return TypeMsgEditMetadata
}

var _ sdk.Msg = &MsgSend{}

func NewMsgSend(DID, sender, receiver string) *MsgSend {
	return &MsgSend{
		Did:      DID,
		Sender:   sender,
		Receiver: receiver,
	}
}

func (msg *MsgSend) Route() string {
	return RouterKey
}

func (msg *MsgSend) Type() string {
	return TypeMsgSend
}

func (msg *MsgSend) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *MsgSend) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSend) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver address (%s)", err)
	}
	return nil
}
