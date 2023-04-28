package account

import "fmt"

const (
	prefix = "profiles:"
)

func profileKey(id string) []byte {
	return []byte(fmt.Sprintf("%s%s", prefix, id))
}

func walletKey(id string) []byte {
	return []byte(fmt.Sprintf("%s%s:wallet", prefix, id))
}

func accountImportedIdx(profileID string) []byte {
	return []byte(fmt.Sprintf("%s%s:account-import-index", prefix, profileID))
}

func keyringAccountKey(profileID string, index uint) string {
	return fmt.Sprintf("%s_%d", profileID, index)
}

func keyringAccountImportedKey(profileID string, index uint) string {
	return fmt.Sprintf("%s_imported_%d", profileID, index)
}

func keyringAccountPrefix(profileID string) string {
	return fmt.Sprintf("%s_", profileID)
}

func accountHDKey(profileID, accountAddress string) []byte {
	return []byte(fmt.Sprintf("%s%s:hd-accounts:%s", prefix, profileID, accountAddress))
}

func accountImportedKey(profileID, accountAddress string) []byte {
	return []byte(fmt.Sprintf("%s%s:imported-accounts:%s", prefix, profileID, accountAddress))
}
