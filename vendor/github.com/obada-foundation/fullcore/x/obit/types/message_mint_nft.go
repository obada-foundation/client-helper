package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgMintNFT     = "mint_nft"
	TypeMsgTransferNFT = "transfer_nft"
	TypeMsgUpdateNFT   = "update_nft"
)

var _ sdk.Msg = &MsgMintNFT{}

func NewMsgMintNFT(creator, uri, uriHash, usn string) *MsgMintNFT {
	return &MsgMintNFT{
		Creator: creator,
		Uri:     uri,
		UriHash: uriHash,
		Usn:     usn,
	}
}

func (msg *MsgMintNFT) Route() string {
	return RouterKey
}

func (msg *MsgMintNFT) Type() string {
	return TypeMsgMintNFT
}

func (msg *MsgMintNFT) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgMintNFT) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMintNFT) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateNFT{}

func NewMsgUpdateNFT(DID, editor string, metadata NFTData) *MsgUpdateNFT {
	return &MsgUpdateNFT{
		Id:      DID,
		Editor:  editor,
		NFTData: &metadata,
	}
}

func (msg *MsgUpdateNFT) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Editor)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid editor address (%s)", err)
	}

	return nil
}

func (msg *MsgUpdateNFT) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateNFT) GetSigners() []sdk.AccAddress {
	editor, err := sdk.AccAddressFromBech32(msg.Editor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{editor}
}

func (msg *MsgUpdateNFT) Route() string {
	return RouterKey
}

func (msg *MsgUpdateNFT) Type() string {
	return TypeMsgUpdateNFT
}

var _ sdk.Msg = &MsgTransferNFT{}

func NewMsgTransferNFT(DID, sender, receiver string) *MsgTransferNFT {
	return &MsgTransferNFT{
		Id:       DID,
		Sender:   sender,
		Receiver: receiver,
	}
}

func (msg *MsgTransferNFT) Route() string {
	return RouterKey
}

func (msg *MsgTransferNFT) Type() string {
	return TypeMsgTransferNFT
}

func (msg *MsgTransferNFT) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *MsgTransferNFT) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgTransferNFT) ValidateBasic() error {
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
