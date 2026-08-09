package app

import (
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

func VerifyTransaction(tx SignedTransaction) bool {
	signature, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return false
	}

	// Ethereum secp256k1 signature: R || S || V.
	if len(signature) != 65 {
		return false
	}

	hashBytes, err := hex.DecodeString(tx.Hash)
	if err != nil {
		return false
	}

	if len(hashBytes) != 32 {
		return false
	}

	publicKey, err := crypto.SigToPub(hashBytes, signature)
	if err != nil {
		return false
	}

	recoveredPublicKey := hex.EncodeToString(
		crypto.FromECDSAPub(publicKey),
	)

	return strings.EqualFold(recoveredPublicKey, tx.PublicKey)
}

// VerifyTransactionSender verifies that the transaction signature
// belongs to the address specified by tx.From.
func VerifyTransactionSender(tx Transaction) bool {
	signature, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return false
	}

	if len(signature) != 65 {
		return false
	}

	hashBytes, err := hex.DecodeString(tx.Hash)
	if err != nil {
		return false
	}

	if len(hashBytes) != 32 {
		return false
	}

	publicKey, err := crypto.SigToPub(hashBytes, signature)
	if err != nil {
		return false
	}

	recoveredPublicKey := hex.EncodeToString(
		crypto.FromECDSAPub(publicKey),
	)

	if !strings.EqualFold(recoveredPublicKey, tx.PublicKey) {
		return false
	}

	recoveredAddress := crypto.PubkeyToAddress(*publicKey)

	return strings.EqualFold(
		recoveredAddress.Hex(),
		tx.From,
	)
}
