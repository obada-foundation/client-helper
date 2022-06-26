package account

import "fmt"

func accountKey(ID string) []byte {
	return []byte(fmt.Sprintf("accounts:%s", ID))
}

func walletKey(ID string) []byte {
	return []byte(fmt.Sprintf("accounts:%s:wallet", ID))
}
