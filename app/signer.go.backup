package app

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

type SignedTransaction struct {
	Hash      string
	Signature string
	PublicKey string
}

func SignTransaction(hash string) SignedTransaction {
	if NodeWallet == nil {
		return SignedTransaction{}
	}

	return SignTransactionWithPrivateKey(hash, NodeWallet.PrivateKey)
}

func SignTransactionWithPrivateKey(hash string, privateKeyHex string) SignedTransaction {
	if hash == "" || privateKeyHex == "" {
		return SignedTransaction{}
	}

	digest, err := hex.DecodeString(hash)
	if err != nil || len(digest) != 32 {
		return SignedTransaction{}
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return SignedTransaction{}
	}

	signature, err := crypto.Sign(digest, privateKey)
	if err != nil {
		return SignedTransaction{}
	}

	publicKey := crypto.FromECDSAPub(&privateKey.PublicKey)

	return SignedTransaction{
		Hash:      hash,
		Signature: hex.EncodeToString(signature),
		PublicKey: hex.EncodeToString(publicKey),
	}
}

func ValidatePrivateKeyForAddress(privateKeyHex, address string) error {
	if privateKeyHex == "" {
		return errors.New("private key is empty")
	}

	if !IsValidAddress(address) {
		return errors.New("invalid address")
	}

	key, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return errors.New("invalid private key")
	}

	derived := crypto.PubkeyToAddress(key.PublicKey)

	if derived != crypto.PubkeyToAddress(key.PublicKey) {
		return errors.New("address derivation failure")
	}

	if !equalAddress(derived.Hex(), address) {
		return errors.New("private key does not belong to address")
	}

	return nil
}

func equalAddress(a, b string) bool {
	return len(a) == len(b) &&
		strings.EqualFold(a, b)
}
