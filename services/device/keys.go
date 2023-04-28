package device

import "fmt"

const (
	prefix = "devices:"
)

func makeUSNKey(userID, usn string) []byte {
	return []byte(fmt.Sprintf(prefix+"%s:usn:%s", userID, usn))
}

func makeDIDKey(userID, did string) []byte {
	return []byte(fmt.Sprintf(prefix+"%s:%s", userID, did))
}

// nolint:unused //need refactoring
func makeSecretKey(userID, did string) []byte {
	return []byte(fmt.Sprintf(prefix+"%s:%s:secret", userID, did))
}

func makeAddressKey(userID, address, did string) []byte {
	return []byte(fmt.Sprintf(prefix+"%s:%s:%s", userID, address, did))
}

// nolint:unused //need refactoring
func makeUserDevicesKey(userID string) []byte {
	return []byte(fmt.Sprintf(prefix+"%s", userID))
}
