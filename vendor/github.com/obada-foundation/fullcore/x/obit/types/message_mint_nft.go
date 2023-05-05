package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typeerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// TypeMsgMintNFT defines the type string of MsgMintNFT
	TypeMsgMintNFT = "mint_nft"

	// TypeMsgTransferNFT defines the type string of MsgTransferNFT
	TypeMsgTransferNFT = "transfer_nft"

	// TypeMsgUpdateNFT defines the type string of MsgUpdateNFT
	TypeMsgUpdateNFT = "update_nft"
)

var _ sdk.Msg = &MsgMintNFT{}

// NewMsgMintNFT is a constructor function for MsgMintNFT
func NewMsgMintNFT(creator, uri, uriHash, usn string) *MsgMintNFT {
	return &MsgMintNFT{
		Creator: creator,
		Uri:     uri,
		UriHash: uriHash,
		Usn:     usn,
	}
}

// Route implements Msg.
func (msg *MsgMintNFT) Route() string {
	return RouterKey
}

// Type implements Msg.
func (msg *MsgMintNFT) Type() string {
	return TypeMsgMintNFT
}

// GetSigners implements Msg.
func (msg *MsgMintNFT) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes implements Msg.
func (msg *MsgMintNFT) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements Msg.
func (msg *MsgMintNFT) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(typeerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateNFT{}

// NewMsgUpdateNFT is a constructor function for MsgUpdateNFT
func NewMsgUpdateNFT(did, editor string, metadata NFTData) *MsgUpdateNFT {
	return &MsgUpdateNFT{
		Id:      did,
		Editor:  editor,
		NFTData: &metadata,
	}
}

// ValidateBasic implements Msg.
func (msg *MsgUpdateNFT) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Editor)
	if err != nil {
		return sdkerrors.Wrapf(typeerrors.ErrInvalidAddress, "invalid editor address (%s)", err)
	}

	return nil
}

// GetSignBytes implements Msg.
func (msg *MsgUpdateNFT) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg.
func (msg *MsgUpdateNFT) GetSigners() []sdk.AccAddress {
	editor, err := sdk.AccAddressFromBech32(msg.Editor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{editor}
}

// Route implements Msg.
func (msg *MsgUpdateNFT) Route() string {
	return RouterKey
}

// Type implements Msg.
func (msg *MsgUpdateNFT) Type() string {
	return TypeMsgUpdateNFT
}

var _ sdk.Msg = &MsgTransferNFT{}

// NewMsgTransferNFT is a constructor function for MsgTransferNFT
func NewMsgTransferNFT(did, sender, receiver string) *MsgTransferNFT {
	return &MsgTransferNFT{
		Id:       did,
		Sender:   sender,
		Receiver: receiver,
	}
}

// Route implements Msg.
func (msg *MsgTransferNFT) Route() string {
	return RouterKey
}

// Type implements Msg.
func (msg *MsgTransferNFT) Type() string {
	return TypeMsgTransferNFT
}

// GetSigners implements Msg.
func (msg *MsgTransferNFT) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// GetSignBytes implements Msg.
func (msg *MsgTransferNFT) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements Msg.
func (msg *MsgTransferNFT) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(typeerrors.ErrInvalidAddress, "invalid sender address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrapf(typeerrors.ErrInvalidAddress, "invalid receiver address (%s)", err)
	}
	return nil
}
