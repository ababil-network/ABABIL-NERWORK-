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

	// Cryptographically verify the signature before recovering the signer.
	// This prevents malformed or tampered signatures from being accepted
	// merely because public-key recovery succeeds.
	recoveredPublicKey, err := crypto.SigToPub(hashBytes, signature)
	if err != nil {
		return false
	}

	recoveredPublicKeyBytes := crypto.FromECDSAPub(recoveredPublicKey)

	if !crypto.VerifySignature(
		recoveredPublicKeyBytes,
		hashBytes,
		signature[:64],
	) {
		return false
	}

	encodedRecoveredPublicKey := hex.EncodeToString(
		recoveredPublicKeyBytes,
	)

	if !strings.EqualFold(encodedRecoveredPublicKey, tx.PublicKey) {
		return false
	}

	recoveredAddress := crypto.PubkeyToAddress(*recoveredPublicKey)

	return strings.EqualFold(
		recoveredAddress.Hex(),
		tx.From,
	)
}
