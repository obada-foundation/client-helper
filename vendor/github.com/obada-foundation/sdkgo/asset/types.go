package asset

// ObjectType is meta type for all objects
type ObjectType string

const (
	// PhysicalAssetIdentifiers https://github.com/obada-foundation/standard/blob/main/1.0-information-architecture/1.2-asset-data-model.md#1251--required-data-objects
	PhysicalAssetIdentifiers ObjectType = "physicalAssetIdentifiers"

	// Image https://github.com/obada-foundation/standard/blob/main/1.0-information-architecture/1.2-asset-data-model.md#1252-optional-data-object-types
	Image ObjectType = "image"

	// MainImage https://github.com/obada-foundation/standard/blob/main/1.0-information-architecture/1.2-asset-data-model.md#1252-optional-data-object-types
	MainImage ObjectType = "mainImage"

	// FunctionalityReport https://github.com/obada-foundation/standard/blob/main/1.0-information-architecture/1.2-asset-data-model.md#1252-optional-data-object-types
	FunctionalityReport ObjectType = "functionalityReport"

	// DataSanitizationReport https://github.com/obada-foundation/standard/blob/main/1.0-information-architecture/1.2-asset-data-model.md#1252-optional-data-object-types
	DataSanitizationReport ObjectType = "dataSanitizationReport"

	// DispositionReport https://github.com/obada-foundation/standard/blob/main/1.0-information-architecture/1.2-asset-data-model.md#1252-optional-data-object-types
	DispositionReport ObjectType = "dispositionReport"
)

// PhysicalAssetIdentifiersScheme represent special document that help to identify asset
type PhysicalAssetIdentifiersScheme struct {
	SerialNumber string `json:"serial_number"`
	Manufacturer string `json:"manufacturer"`
	PartNumber   string `json:"part_number"`
}

// Object is the temp name
type Object struct {
	URL                     string            `json:"url"`
	HashEncryptedDataObject string            `json:"hashEncryptedDataObject"`
	HashUnencryptedObject   string            `json:"hashUnencryptedObject"`
	Metadata                map[string]string `json:"metadata"`
	HashUnencryptedMetadata string            `json:"hashUnencryptedMetadata"`
	HashEncryptedMetadata   string            `json:"hashEncryptedMetadata"`
	DataObjectHash          string            `json:"dataObjectHash"`
}

// DataArray https://github.com/obada-foundation/standard/blob/main/1.0-information-architecture/1.2-asset-data-model.md#123-asset-data-array
type DataArray struct {
	VersionHash string   `json:"versionHash"`
	RootHash    string   `json:"rootHash"`
	Objects     []Object `json:"dataArray"`
}

// DataArrayVersions stores version as a key
type DataArrayVersions map[int]DataArray
