package wallet

import (
	"fmt"
)

func masterKey(accountID, kid string) []byte {
	return []byte(fmt.Sprintf("accounts:%s:master-keys:%s", accountID, kid))
}

func privateKey(accountID, kid, pubKey string) []byte {
	return []byte(fmt.Sprintf("accounts:%s:master-keys:%s:private-keys:%s", accountID, kid, pubKey))
}
