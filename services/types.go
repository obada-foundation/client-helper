package services

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewProfile request data for creating a new profile
type NewProfile struct {
	ID    string `json:"-" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// Profile is a user profile
type Profile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// MasterKey stores master key, needs to be removed once we stop storing aster key
type MasterKey struct {
	ID  string `json:"id"`
	Key string `json:"master_key"`
}

// Wallet user wallet
type Wallet struct {
	Mnemonic     string `json:"-"`
	AccountIndex uint   `json:"-"`
}

// ProfileAccounts stores all accounts separated on accout types
type ProfileAccounts struct {
	HDAccounts       []Account `json:"hd_accounts"`
	ImportedAccounts []Account `json:"imported_accounts"`
}

// Account client helper account
type Account struct {
	Name      string      `json:"name"`
	PublicKey string      `json:"pub_key"`
	Address   string      `json:"address"`
	Balance   sdk.DecCoin `json:"balance"`
	NFTsCount uint        `json:"nft_count"`
}

// Balance account balance
type Balance struct {
	Address string      `json:"address"`
	Balance sdk.DecCoin `json:"balance"`
}

// SaveDeviceDocument request data for saving device documents
type SaveDeviceDocument struct {
	Name          string `json:"name" validate:"required"`
	Description   string `json:"description"`
	File          string `json:"document_file" validate:"required"`
	Type          string `json:"type" validate:"required"`
	DeviceAddress string `json:"device_address"`
	ShouldEncrypt bool   `json:"should_encrypt"`
}

// DeviceDocument device document (asset)
type DeviceDocument struct {
	Name        string `json:"name"`
	URI         string `json:"uri"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Hash        string `json:"hash"`
	Encrypted   bool   `json:"encrypted"`
}

// SaveDevice request data for saving device information
type BatchSaveDevice struct {
	ShouldMint bool         `json:"should_mint"`
	Obits      []SaveDevice `json:"obits"`
	Address    string       `json:"address" validate:"required"`
}

// SaveDevice request data for saving device information
type SaveDevice struct {
	SerialNumber string               `json:"serial_number" validate:"required"`
	Manufacturer string               `json:"manufacturer"  validate:"required"`
	PartNumber   string               `json:"part_number"  validate:"required"`
	Documents    []SaveDeviceDocument `json:"documents"`
	Address      string               `json:"address"  validate:"required"`
}

// Device is ClientHelper device (asset)
type Device struct {
	Usn          string           `json:"usn"`
	DID          string           `json:"did"`
	Checksum     string           `json:"checksum"`
	SerialNumber string           `json:"serial_number"`
	Manufacturer string           `json:"manufacturer"`
	PartNumber   string           `json:"part_number"`
	Documents    []DeviceDocument `json:"documents"`
	Address      string           `json:"address"`
}

// SendNFT request data for sending NFT
type SendNFT struct {
	ReceiverArr string `json:"receiver"`
}

// MintBatchNFT request data for minting batch NFTs
type MintBatchNFT struct {
	Nfts []string `json:"nfts"`
}
