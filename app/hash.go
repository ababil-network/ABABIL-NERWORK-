package app

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
)

func GenerateHash(data string) string {
	hash := crypto.Keccak256([]byte(data))
	return hex.EncodeToString(hash)
}
