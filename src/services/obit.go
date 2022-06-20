package services

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/obada-foundation/client-helper/utils"
	"github.com/obada-foundation/sdkgo"
	"github.com/obada-foundation/sdkgo/properties"
	"go.uber.org/zap"
)

type ObitService struct {
	logger *zap.SugaredLogger
	db     *badger.DB
	sdk    *sdkgo.Sdk
}

func NewObitService(logger *zap.SugaredLogger, db *badger.DB, sdk *sdkgo.Sdk) *ObitService {
	return &ObitService{
		db:     db,
		logger: logger,
		sdk:    sdk,
	}
}

type NFTPayload struct {
	SerialNumber     string `json:"serial_number"`
	Manufacturer     string `json:"manufacturer"`
	PartNumber       string `json:"part_number"`
	TrustAnchorToken string `json:"trust_anchor_token"`
}

type LocalNFT struct {
	Usn              string `json:"usn"`
	DID              string `json:"did"`
	Owner            string `json:"owner"`
	Checksum         string `json:"checksum"`
	SerialNumberHash string `json:"serial_number_hash"`
	Manufacturer     string `json:"manufacturer"`
	PartNumber       string `json:"part_number"`
	TrustAnchorToken string `json:"trust_anchor_token"`
}

func (svc *ObitService) GenerateObit(serialNumber, manufacturer, partNumber string) (*properties.ObitID, error) {
	snh, err := utils.HashStr(serialNumber)
	if err != nil {
		return nil, err
	}

	obit, err := svc.sdk.NewObitID(sdkgo.ObitIDDto{
		SerialNumberHash: snh,
		Manufacturer:     manufacturer,
		PartNumber:       partNumber,
	})
	if err != nil {
		return nil, err
	}

	return &obit, nil
}

func (svc *ObitService) createSdkObit(sdk *sdkgo.Sdk, payload NFTPayload) (sdkgo.Obit, error) {
	var obit sdkgo.Obit

	serialNumberHash, err := utils.HashStr(payload.SerialNumber)
	if err != nil {
		return obit, err
	}

	obit, err = sdk.NewObit(sdkgo.ObitDto{
		ObitIDDto: sdkgo.ObitIDDto{
			SerialNumberHash: serialNumberHash,
			Manufacturer:     payload.Manufacturer,
			PartNumber:       payload.PartNumber,
		},
		TrustAnchorToken: payload.TrustAnchorToken,
	})

	return obit, err
}

func (svc *ObitService) MakeLocalNFT(payload NFTPayload, catchLogs bool) (LocalNFT, bytes.Buffer, error) {
	var capturedLogs bytes.Buffer
	var err error
	var localNFT LocalNFT

	sdk := svc.sdk

	if catchLogs {
		sdklogger := log.New(&capturedLogs, "", 0)
		sdk, err = sdkgo.NewSdk(sdklogger, catchLogs)

		if err != nil {
			return localNFT, capturedLogs, err
		}
	}

	obit, err := svc.createSdkObit(sdk, payload)
	if err != nil {
		return localNFT, capturedLogs, err
	}

	checksum, err := obit.GetChecksum(nil)
	if err != nil {
	}

	did := obit.GetObitID()

	localNFT = LocalNFT{
		Usn:              did.GetUsn(),
		DID:              did.GetDid(),
		Checksum:         checksum.GetHash(),
		SerialNumberHash: obit.GetSerialNumberHash().GetValue(),
		Manufacturer:     obit.GetManufacturer().GetValue(),
		PartNumber:       obit.GetPartNumber().GetValue(),
		TrustAnchorToken: obit.GetTrustAnchorToken().GetValue(),
	}

	return localNFT, capturedLogs, nil
}

func makeUSNKey(usn string) []byte {
	return []byte(fmt.Sprintf("usn:%s", usn))
}

func makeDIDKey(DID string) []byte {
	return []byte(fmt.Sprintf("%s", DID))
}

// Save
func (svc *ObitService) Save(payload NFTPayload, address string) (LocalNFT, error) {
	localNFT, _, err := svc.MakeLocalNFT(payload, false)
	if err != nil {
		return localNFT, err
	}

	localNFT.Owner = address

	err = svc.db.Update(func(txn *badger.Txn) error {
		var buff bytes.Buffer

		enc := gob.NewEncoder(&buff)

		if er := enc.Encode(localNFT); er != nil {
			return er
		}

		DIDKey := makeDIDKey(localNFT.DID)

		DIDEntry := badger.NewEntry(DIDKey, buff.Bytes())
		err = txn.SetEntry(DIDEntry)

		USNEntry := badger.NewEntry(makeUSNKey(localNFT.Usn), DIDKey)
		err = txn.SetEntry(USNEntry)

		return err
	})

	return localNFT, nil
}

func (svc *ObitService) GetByDID(DID string) (LocalNFT, error) {
	var localNFT LocalNFT

	err := svc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(makeDIDKey(DID))
		if err != nil {
			return err
		}

		localNFTBytes, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		var buff bytes.Buffer
		if _, err := buff.Write(localNFTBytes); err != nil {
			return err
		}

		dec := gob.NewDecoder(&buff)
		if err := dec.Decode(&localNFT); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return localNFT, err
	}

	return localNFT, nil
}

func (svc *ObitService) GetByUSN(usn string) (LocalNFT, error) {
	var localNFT LocalNFT

	err := svc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(makeUSNKey(usn))
		if err != nil {
			return err
		}

		DIDBytes, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		localNFT, err = svc.GetByDID(string(DIDBytes))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return localNFT, err
	}

	return localNFT, nil

}

// Get
func (svc *ObitService) Get(key string) (LocalNFT, error) {
	svc.logger.Debug(len(key))
	if len(key) == 8 {
		return svc.GetByUSN(key)
	}

	return svc.GetByDID(key)
}

// Search
func (svc *ObitService) Search(query string) ([]LocalNFT, error) {
	localNFTs := make([]LocalNFT, 0)

	return localNFTs, nil
}
