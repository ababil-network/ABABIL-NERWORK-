package app

import (
	"errors"
	"strings"
)

const MaxTransactionsPerBlock = 5000

func ValidateBlock(block Block, previous Block) error {
	// Height check.
	if block.Height != previous.Height+1 {
		return errors.New("invalid block height")
	}

	// Empty block check.
	if len(block.Transactions) == 0 {
		return errors.New("empty block")
	}

	// Maximum transaction limit.
	if len(block.Transactions) > MaxTransactionsPerBlock {
		return errors.New("too many transactions")
	}

	// Previous hash check.
	if block.PreviousHash != previous.Hash {
		return errors.New("invalid previous hash")
	}

	// Timestamp must be valid.
	if block.Timestamp == "" {
		return errors.New("invalid block timestamp")
	}

	if _, err := parseBlockTimestamp(block.Timestamp); err != nil {
		return err
	}

	// Duplicate transaction check.
	seen := make(map[string]struct{}, len(block.Transactions))

	for _, tx := range block.Transactions {
		if tx.Hash == "" {
			return errors.New("invalid transaction hash")
		}

		hash := strings.ToLower(tx.Hash)

		if _, exists := seen[hash]; exists {
			return errors.New("duplicate transaction")
		}

		seen[hash] = struct{}{}
	}

	// Block hash must exist.
	if block.Hash == "" {
		return errors.New("invalid block hash")
	}

	// Recalculate the block hash from canonical block data.
	expectedHash, err := GenerateBlockHash(block)
	if err != nil {
		return err
	}

	if !strings.EqualFold(block.Hash, expectedHash) {
		return errors.New("block hash mismatch")
	}

	return nil
}
