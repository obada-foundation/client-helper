package services

import "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

type NewAccount struct {
	ID    string `json:"-" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type Account struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type Wallet struct {
	PrivateKey secp256k1.PrivKey `json:"-"`
}

type Balance struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

type SaveDeviceDocument struct {
	Name          string `json:"name" validate:"required"`
	File          string `json:"document_file" validate:"required"`
	ShouldEncrypt bool   `json:"should_encrypt"`
}

type DeviceDocument struct {
	Name      string `json:"name"`
	URI       string `json:"uri"`
	Hash      string `json:"hash"`
	Encrypted bool   `json:"encrypted"`
}

type SaveDevice struct {
	SerialNumber string               `json:"serial_number" validate:"required"`
	Manufacturer string               `json:"manufacturer"  validate:"required"`
	PartNumber   string               `json:"part_number"  validate:"required"`
	Documents    []SaveDeviceDocument `json:"documents"`
}

type Device struct {
	Usn              string           `json:"usn"`
	DID              string           `json:"did"`
	Checksum         string           `json:"checksum"`
	SerialNumberHash string           `json:"serial_number_hash"`
	Manufacturer     string           `json:"manufacturer"`
	PartNumber       string           `json:"part_number"`
	TrustAnchorToken string           `json:"trust_anchor_token"`
	Documents        []DeviceDocument `json:"documents"`
}

type SendNFT struct {
	ReceiverArr string `json:"receiver"`
}
