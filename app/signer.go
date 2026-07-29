package app

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
)

type SignedTransaction struct {
	Hash      string
	Signature string
	PublicKey string
}

func SignTransaction(hash string) SignedTransaction {

	privateKeyBytes, err := hex.DecodeString(NodeWallet.PrivateKey)
	if err != nil {
		return SignedTransaction{}
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return SignedTransaction{}
	}

	signature, err := crypto.Sign(
		crypto.Keccak256([]byte(hash)),
		privateKey,
	)
	if err != nil {
		return SignedTransaction{}
	}

	publicKey := privateKey.Public().(*ecdsa.PublicKey)

	return SignedTransaction{
		Hash:      hash,
		Signature: hex.EncodeToString(signature),
		PublicKey: hex.EncodeToString(crypto.FromECDSAPub(publicKey)),
	}
}
