package app

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateWalletAddress() string {
	b := make([]byte, 20)

	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return "ababil1" + hex.EncodeToString(b)
}
