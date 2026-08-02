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
	GasLimit   uint64
	GasPrice   uint64
	Fee        uint64
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

	gasLimit := uint64(21000)

	gasPrice := NodeDynamicFee.GasPrice()

	fee := NodeDynamicFee.CalculateFee(gasLimit)

	// Daily Free Transaction
	if NodeFreeTransaction.Remaining(from) > 0 {
		gasPrice = 0
		fee = 0
	}

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

	tx.Hash = GenerateHash(tx.ID + tx.From + tx.To)

	signed := SignTransaction(tx.Hash)

	tx.Signature = signed.Signature
	tx.PublicKey = signed.PublicKey

	return tx
}
