package asset

import (
	"fmt"
	"log"
	"sort"

	sdkhash "github.com/obada-foundation/sdkgo/hash"
	"github.com/obada-foundation/sdkgo/properties"
)

// VersionHash compute the version hash from asset data
func VersionHash(logger *log.Logger, objs []Object) (sdkhash.Hash, error) {
	objectHashes := make([]sdkhash.Hash, 0, len(objs))
	var h sdkhash.Hash

	for _, obj := range objs {

		if len(obj.Metadata) == 0 {
			return h, fmt.Errorf("empty metadata")
		}

		if obj.HashUnencryptedObject == "" {

			if logger != nil {
				logger.Printf("missing dataobject hash: %+v", obj)
			}
			return h, fmt.Errorf("missing dataobject hash")
		}

		m, err := properties.NewMetadata(obj.Metadata, logger)
		if err != nil {
			return h, err
		}

		metadataHash, err := m.Hash()
		if err != nil {
			return h, err
		}

		objectHash, err := DataObjectHash(
			obj.URL,
			obj.HashUnencryptedObject,
			metadataHash.GetHash(),
			logger,
		)

		if err != nil {
			return h, err
		}

		objectHashes = append(objectHashes, objectHash)
	}

	versionSum := sdkhash.SumHashes(logger, objectHashes...)

	return sdkhash.NewHash([]byte(fmt.Sprintf("%x", versionSum)), logger)
}

// RootHash compute the root hash from asset data
func RootHash(snapshots DataArrayVersions, logger *log.Logger) (sdkhash.Hash, error) {
	var rootHash sdkhash.Hash

	if len(snapshots) == 0 {
		return rootHash, fmt.Errorf("cannot compute root hash because of no snapshots")
	}

	keys := make([]int, 0, len(snapshots))
	for version := range snapshots {
		keys = append(keys, version)
	}

	sort.Ints(keys)

	// prevent for computing hashes with versions 1,3,123 etc but expect 1,2,3...123,n
	if !isIncremental(keys) {
		return rootHash, fmt.Errorf("snapshot versions are not incremental: %+v", keys)
	}

	for _, version := range keys {
		versionDataArray := snapshots[version]

		if len(versionDataArray.Objects) == 0 {
			return rootHash, fmt.Errorf("cannot compute root hash because of empty data array")
		}

		versionHash, err := VersionHash(logger, versionDataArray.Objects)
		if err != nil {
			return versionHash, err
		}

		if len(snapshots) == 1 {
			return versionHash, nil
		}

		if version == 1 {
			rootHash = versionHash
			continue
		}

		if versionHash.GetHash() == rootHash.GetHash() {
			return rootHash, fmt.Errorf("data in a version %d are the same as in previous version", version)
		}

		rootHash, err = sdkhash.NewHash([]byte(fmt.Sprintf("%s%s", rootHash.GetHash(), versionHash.GetHash())), logger)
		if err != nil {
			return rootHash, err
		}
	}

	return rootHash, nil
}

// DataObjectHash wait until Rohi finalize params now we have a lot of not needed entropy
func DataObjectHash(url, hashUnencryptedObject, hashUnencryptedMetadata string, logger *log.Logger) (sdkhash.Hash, error) {
	strToHash := fmt.Sprintf(url + hashUnencryptedObject + hashUnencryptedMetadata)

	return sdkhash.NewHash([]byte(strToHash), logger)
}

func isIncremental(keys []int) bool {
	n := len(keys)

	if keys[0] != 1 {
		return false
	}

	for i := 1; i < n; i++ {
		if keys[i] != keys[i-1]+1 {
			return false
		}
		if keys[i] > n {
			return false
		}
	}

	return true
}
