package types

import (
	"github.com/obada-foundation/sdkgo/asset"
)

// ContentType supported content types
type ContentType string

// nolint
const (
	DIDJSON   ContentType = "application/did+json"
	DIDJSONLD ContentType = "application/did+ld+json"
	JSONLD    ContentType = "application/ld+json"
	JSON      ContentType = "application/json"

	DIDSchemaJSONLD                  = "https://www.w3.org/ns/did/v1"
	ResolutionSchemaJSONLD           = "https://w3id.org/did-resolution/v1"
	Ed25519VerificationKey2020JSONLD = "https://w3id.org/security/suites/ed25519-2020/v1"
	Ed25519VerificationKey2018JSONLD = "https://w3id.org/security/suites/ed25519-2018/v1"
	JsonWebKey2020JSONLD             = "https://w3id.org/security/suites/jws-2020/v1"
)

// RegisterDID describes payload for new DID registration
type RegisterDID struct {
	DID                string               `json:"did"`
	VerificationMethod []VerificationMethod `json:"verificationMethod,omitempty"`
	Authentication     []string             `json:"authentication,omitempty"`
	Service            []Service            `json:"service,omitempty"`
}

// SaveMetadata payload for new DID saving metadata
type SaveMetadata struct {
	Objects   []asset.Object `json:"objects"`
	Signature string         `json:"signature"`
}

// DIDDocument structure that expresses the DID Document
type DIDDocument struct {
	Context              []string                `json:"@context,omitempty" example:"https://www.w3.org/ns/did/v1"`
	ID                   string                  `json:"id,omitempty" example:"did:obada:123"`
	Controller           []string                `json:"controller,omitempty" example:"did:obada:123"`
	VerificationMethod   []VerificationMethod    `json:"verificationMethod,omitempty"`
	Authentication       []string                `json:"authentication,omitempty" example:"did:obada:123#key-1"`
	AssertionMethod      []string                `json:"assertionMethod,omitempty"`
	CapabilityInvocation []string                `json:"capabilityInvocation,omitempty"`
	CapabilityDelegation []string                `json:"capability_delegation,omitempty"`
	KeyAgreement         []string                `json:"keyAgreement,omitempty"`
	Service              []Service               `json:"service,omitempty"`
	AlsoKnownAs          []string                `json:"alsoKnownAs,omitempty"`
	Metadata             Metadata                `json:"metadata,omitempty"`
	MetadataHistory      asset.DataArrayVersions `json:"-"`
}

// VerificationMethod structure that expresses the Verification Method
type VerificationMethod struct {
	Context            []string    `json:"@context,omitempty"`
	ID                 string      `json:"id,omitempty"`
	Type               string      `json:"type,omitempty"`
	Controller         string      `json:"controller,omitempty"`
	PublicKeyJwk       interface{} `json:"publicKeyJwk,omitempty"`
	PublicKeyMultibase string      `json:"publicKeyMultibase,omitempty"`
	PublicKeyBase58    string      `json:"publicKeyBase58,omitempty"`
}

// Service structure that expresses the Service
type Service struct {
	Context         []string `json:"@context,omitempty"`
	ID              string   `json:"id,omitempty" example:"did:obada:123#service-1"`
	Type            string   `json:"type,omitempty" example:"did-communication"`
	ServiceEndpoint []string `json:"serviceEndpoint,omitempty" example:"https://www.tradeloop.com/usn/125"`
}

// Metadata actually represent asset objects (no formal spec yet)
type Metadata struct {
	VersionID   int            `json:"versionID"`
	VersionHash string         `json:"versionHash"`
	RootHash    string         `json:"rootHash"`
	Objects     []asset.Object `json:"objects"`
}

// NewDIDDoc creates an empty DID Document
func NewDIDDoc() DIDDocument {
	return DIDDocument{
		Context: []string{
			DIDSchemaJSONLD,
			Ed25519VerificationKey2018JSONLD,
		},
		ID:                 "",
		VerificationMethod: make([]VerificationMethod, 0),
		Authentication:     make([]string, 0),
		Metadata: Metadata{
			VersionID: 0,
			Objects:   make([]asset.Object, 0),
		},
	}
}
