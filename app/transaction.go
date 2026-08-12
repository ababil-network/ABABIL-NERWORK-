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
	GasLimit  uint64
	GasPrice  uint64
	Fee       uint64
	Nonce     uint64
	Hash      string
	Signature string
	PublicKey string
	Timestamp time.Time
}

func GenerateTransactionID() string {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}

// NewTransaction creates a transaction using the final ABABIL
// load-based USD-equivalent fee architecture.
//
// GasPrice is retained in the transaction structure for compatibility,
// but it is no longer used to determine the final transaction fee.
//
// The final fee is:
//
//	network load
//	    -> USD-equivalent fee
//	    -> validated ABABIL reference price
//	    -> native ABABIL fee
func NewTransaction(from, to string, amount uint64) Transaction {
	gasLimit := DefaultGasLimit

	fee, err := CalculateFinalNativeFee()
	if err != nil {
		return Transaction{}
	}
	gasPrice := uint64(0)

	tx := Transaction{
		ID:        GenerateTransactionID(),
		From:      from,
		To:        to,
		Amount:    amount,
		GasLimit:  gasLimit,
		GasPrice:  gasPrice,
		Fee:       fee,
		Nonce:     NodeNonce.Get(from) + 1,
		Timestamp: time.Now().UTC(),
	}

	if tx.ID == "" {
		return Transaction{}
	}

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		return Transaction{}
	}

	tx.Hash = hash

	signed := SignTransaction(tx.Hash)

	if signed.Signature == "" || signed.PublicKey == "" {
		return Transaction{}
	}

	tx.Signature = signed.Signature
	tx.PublicKey = signed.PublicKey

	return tx
}
