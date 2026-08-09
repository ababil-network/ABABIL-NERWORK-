package app

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
)

func VerifyTransaction(tx SignedTransaction) bool {

	signature, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return false
	}

	hash := crypto.Keccak256([]byte(tx.Hash))

	publicKey, err := crypto.SigToPub(hash, signature)
	if err != nil {
		return false
	}

	recovered := hex.EncodeToString(
		crypto.FromECDSAPub(publicKey),
	)

	return recovered == tx.PublicKey
}
