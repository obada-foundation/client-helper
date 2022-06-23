package device

type DocumentType string

const (
	PhysicalAssetIdentifier DocumentType = "physical_asset_identifier"
)

type SaveDeviceDocument struct {
	Name string `json:"name" validate:"required"`
	File string `json:"document_file" validate:"required"`
}

type DeviceDocument struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
	Hash string `json:"hash"`
}

type SaveDevice struct {
	SerialNumber     string               `json:"serial_number" validate:"required"`
	Manufacturer     string               `json:"manufacturer"  validate:"required"`
	PartNumber       string               `json:"part_number"  validate:"required"`
	Documents        []SaveDeviceDocument `json:"documents"`
	EncryptDocuments bool                 `json:"encrypt_documents"`
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
	Secret           string           `json:"-"`
}
