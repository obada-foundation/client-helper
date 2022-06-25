package device

import "fmt"

func makeUSNKey(userID, usn string) []byte {
	return []byte(fmt.Sprintf("devices:%s:usn:%s", userID, usn))
}

func makeDIDKey(userID, DID string) []byte {
	return []byte(fmt.Sprintf("devices:%s:%s", userID, DID))
}

func makeSecretKey(userID, DID string) []byte {
	return []byte(fmt.Sprintf("devices:%s:%s:secret", userID, DID))
}
