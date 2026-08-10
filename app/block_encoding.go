package app

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

const BlockVersion uint8 = 1

// CanonicalBlockBytes returns the exact deterministic byte representation
// used to calculate a block hash.
//
// Hash is deliberately excluded because it is derived from these fields.
func CanonicalBlockBytes(block Block) ([]byte, error) {
	if block.Height < 0 {
		return nil, errors.New("invalid block height")
	}

	if block.PreviousHash == "" {
		return nil, errors.New("previous block hash is empty")
	}

	if block.Timestamp == "" {
		return nil, errors.New("block timestamp is empty")
	}

	timestamp, err := time.Parse(time.RFC3339Nano, block.Timestamp)
	if err != nil {
		return nil, errors.New("invalid block timestamp")
	}

	var buf bytes.Buffer

	// Domain separation.
	if err := writeString(&buf, ChainID); err != nil {
		return nil, err
	}

	// Block version.
	if err := buf.WriteByte(BlockVersion); err != nil {
		return nil, err
	}

	// Block height.
	if err := writeInt64(&buf, int64(block.Height)); err != nil {
		return nil, err
	}

	// Previous block hash.
	if err := writeString(&buf, block.PreviousHash); err != nil {
		return nil, err
	}

	// Canonical timestamp.
	if err := writeInt64(&buf, timestamp.UTC().UnixNano()); err != nil {
		return nil, err
	}

	// Transaction count.
	if uint64(len(block.Transactions)) > uint64(^uint32(0)) {
		return nil, errors.New("too many transactions")
	}

	if err := binary.Write(&buf, binary.BigEndian, uint32(len(block.Transactions))); err != nil {
		return nil, err
	}

	// Each transaction is represented by its canonical unsigned bytes.
	for _, tx := range block.Transactions {
		txBytes, err := CanonicalTransactionBytes(tx)
		if err != nil {
			return nil, err
		}

		if uint64(len(txBytes)) > uint64(^uint32(0)) {
			return nil, errors.New("transaction encoding too large")
		}

		if err := binary.Write(&buf, binary.BigEndian, uint32(len(txBytes))); err != nil {
			return nil, err
		}

		if _, err := buf.Write(txBytes); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// GenerateBlockHash creates the deterministic Keccak-256 hash of a block.
func GenerateBlockHash(block Block) (string, error) {
	data, err := CanonicalBlockBytes(block)
	if err != nil {
		return "", err
	}

	return encodeHash(crypto.Keccak256(data)), nil
}
