package sdkgo

import (
	"fmt"
	"github.com/obada-foundation/sdkgo/hash"
	"github.com/obada-foundation/sdkgo/properties"
)

// NewObit creates new obit
func (sdk *Sdk) NewObit(dto ObitDto) (Obit, error) {
	var o Obit
	o.debug = sdk.debug
	o.logger = sdk.logger

	err := sdk.validate.Struct(dto)
	if err != nil {
		return o, err
	}

	if sdk.debug {
		sdk.logger.Printf("<|%s|> => NewObit(%v)", "<|Creating new Obit|>", dto)
	}

	snProp, err := properties.NewStringProperty(
		"Making serialNumberHash",
		dto.SerialNumberHash,
		sdk.logger,
		sdk.debug,
	)

	if err != nil {
		return o, err
	}

	manufacturerProp, err := properties.NewStringProperty(
		"Making manufacturer hash",
		dto.Manufacturer,
		sdk.logger,
		sdk.debug,
	)

	if err != nil {
		return o, err
	}

	pnProp, err := properties.NewStringProperty(
		"Making partNumber hash",
		dto.PartNumber,
		sdk.logger,
		sdk.debug,
	)

	if err != nil {
		return o, err
	}

	trustAnchorToken, err := properties.NewStringProperty(
		"Making Trust Anchor Token hash",
		dto.TrustAnchorToken,
		sdk.logger,
		sdk.debug,
	)

	if err != nil {
		return o, err
	}

	obitIDProp, err := properties.NewObitIDProperty(snProp, manufacturerProp, pnProp, sdk.logger, sdk.debug)

	if err != nil {
		return o, err
	}

	documentsProp := properties.NewDocumentsCollection(
		sdk.logger,
		sdk.debug,
	)

	if err != nil {
		return o, err
	}

	o.obitID = obitIDProp
	o.serialNumberHash = snProp
	o.manufacturer = manufacturerProp
	o.partNumber = pnProp
	o.trustAnchorToken = trustAnchorToken
	o.documents = documentsProp

	return o, nil
}

// GetObitID returns obit ID
func (o Obit) GetObitID() properties.ObitID {
	return o.obitID
}

// GetSerialNumberHash returns serial number hash Obit property
func (o Obit) GetSerialNumberHash() properties.StringProperty {
	return o.serialNumberHash
}

// GetPartNumber returns part number Obit property
func (o Obit) GetPartNumber() properties.StringProperty {
	return o.partNumber
}

// GetManufacturer returns manufacturer Obit property
func (o Obit) GetManufacturer() properties.StringProperty {
	return o.manufacturer
}

// GetTrustAnchorToken returns OBADA Obit obd DID
func (o Obit) GetTrustAnchorToken() properties.StringProperty {
	return o.trustAnchorToken
}

// GetDocuments returns Obit documents
func (o Obit) GetDocuments() properties.Documents {
	return o.documents
}

// GetChecksum returns Obit checksum
func (o Obit) GetChecksum(parentChecksum *hash.Hash) (hash.Hash, error) {
	var checksum hash.Hash

	if o.debug {
		o.logger.Println("\n\n<|Obit checksum calculation|>")
	}

	documentsHash, err := o.documents.GetHash()
	if err != nil {
		return checksum, nil
	}

	sum := o.obitID.GetHash().GetDec() +
		o.serialNumberHash.GetHash().GetDec() +
		o.manufacturer.GetHash().GetDec() +
		o.partNumber.GetHash().GetDec() +
		o.trustAnchorToken.GetHash().GetDec() +
		documentsHash.GetDec()

	if o.debug {
		o.logger.Println(fmt.Sprintf(
			"(%d + %d + %d + %d + %d + %d) -> %d -> Dec2Hex(%d) -> %s",
			o.obitID.GetHash().GetDec(),
			o.serialNumberHash.GetHash().GetDec(),
			o.manufacturer.GetHash().GetDec(),
			o.partNumber.GetHash().GetDec(),
			o.trustAnchorToken.GetHash().GetDec(),
			documentsHash.GetDec(),
			sum,
			sum,
			fmt.Sprintf("%x", sum),
		))
	}

	if parentChecksum != nil {
		prhDec := parentChecksum.GetDec()

		if o.debug {
			o.logger.Println(fmt.Sprintf("(%d + %d) -> %d", sum, prhDec, sum+prhDec))
		}

		sum += prhDec
	}

	checksum, err = hash.NewHash([]byte(fmt.Sprintf("%x", sum)), o.logger, o.debug)
	if err != nil {
		return checksum, err
	}

	if o.debug {
		o.logger.Printf("Checksum(%v) -> %q", o, checksum.GetHash())
	}

	return checksum, nil
}
