package app

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)
type Transaction struct {
	ID        string
	From      string
	To        string
	Amount    uint64
	Timestamp time.Time
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
	return Transaction{
		ID:        GenerateTransactionID(),
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().UTC(),
	}
}
