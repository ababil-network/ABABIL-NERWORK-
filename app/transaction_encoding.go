package app

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

const TransactionVersion uint8 = 1

// CanonicalTransactionBytes returns the exact deterministic byte
// representation used for transaction hashing and signing.
//
// Hash, Signature and PublicKey are deliberately excluded because
// they are derived from the canonical transaction data.
func CanonicalTransactionBytes(tx Transaction) ([]byte, error) {
	if tx.ID == "" {
		return nil, errors.New("transaction ID is empty")
	}

	if tx.From == "" {
		return nil, errors.New("transaction sender is empty")
	}

	if tx.To == "" {
		return nil, errors.New("transaction receiver is empty")
	}

	if ChainID == "" {
		return nil, errors.New("chain ID is empty")
	}

	var buf bytes.Buffer

	// Domain separation.
	if err := writeString(&buf, ChainID); err != nil {
		return nil, err
	}

	// Transaction version.
	if err := buf.WriteByte(TransactionVersion); err != nil {
		return nil, err
	}

	// Transaction identity and participants.
	if err := writeString(&buf, tx.ID); err != nil {
		return nil, err
	}

	if err := writeString(&buf, tx.From); err != nil {
		return nil, err
	}

	if err := writeString(&buf, tx.To); err != nil {
		return nil, err
	}

	// Monetary and execution fields.
	if err := writeUint64(&buf, tx.Amount); err != nil {
		return nil, err
	}

	if err := writeUint64(&buf, tx.GasLimit); err != nil {
		return nil, err
	}

	if err := writeUint64(&buf, tx.GasPrice); err != nil {
		return nil, err
	}

	if err := writeUint64(&buf, tx.Fee); err != nil {
		return nil, err
	}

	if err := writeUint64(&buf, tx.Nonce); err != nil {
		return nil, err
	}

	// Timestamp is encoded as Unix nanoseconds.
	// This deliberately ignores Go's internal monotonic clock component.
	if err := writeInt64(&buf, tx.Timestamp.UTC().UnixNano()); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeString(buf *bytes.Buffer, value string) error {
	data := []byte(value)

	if uint64(len(data)) > uint64(^uint32(0)) {
		return errors.New("string too large")
	}

	if err := binary.Write(buf, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}

	_, err := buf.Write(data)
	return err
}

func writeUint64(buf *bytes.Buffer, value uint64) error {
	return binary.Write(buf, binary.BigEndian, value)
}

func writeInt64(buf *bytes.Buffer, value int64) error {
	return binary.Write(buf, binary.BigEndian, value)
}

// GenerateTransactionHash creates the deterministic transaction hash.
func GenerateTransactionHash(tx Transaction) (string, error) {
	data, err := CanonicalTransactionBytes(tx)
	if err != nil {
		return "", err
	}

	hash := crypto.Keccak256(data)

	return encodeHash(hash), nil
}

func encodeHash(hash []byte) string {
	const hexAlphabet = "0123456789abcdef"

	result := make([]byte, len(hash)*2)

	for i, b := range hash {
		result[i*2] = hexAlphabet[b>>4]
		result[i*2+1] = hexAlphabet[b&0x0f]
	}

	return string(result)
}

// Ensure time.Time remains part of the transaction API while
// making the encoding rules explicit.
var _ time.Time
