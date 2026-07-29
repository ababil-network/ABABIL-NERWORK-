package app

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Transaction struct {
	ID         string
	From       string
	To         string
	Amount     uint64
	Nonce      uint64
	Hash       string
	Signature  string
	PublicKey  string
	Timestamp  time.Time
}

func GenerateTransactionID() string {
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}
func NewTransaction(from, to string, amount uint64) Transaction {

	tx := Transaction{
		ID:        GenerateTransactionID(),
		From:      from,
		To:        to,
		Amount:    amount,
		Nonce:     1,
		Timestamp: time.Now().UTC(),
	}

        tx.Hash = GenerateHash(tx.ID + tx.From + tx.To)

        signed := SignTransaction(tx.Hash)

        tx.Signature = signed.Signature
        tx.PublicKey = signed.PublicKey

        return tx
}
