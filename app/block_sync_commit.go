package app

import (
	"errors"
	"fmt"
	"os"
)

func CommitSyncedBlocks(blocks []Block) error {
	if len(blocks) == 0 {
		return fmt.Errorf("empty block response")
	}

	latest, err := GetLatestBlock()
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	previous := latest

	// Phase 1: validate the complete batch before mutating state.
	for _, block := range blocks {
		if err := ValidateBlock(block, previous); err != nil {
			return fmt.Errorf(
				"invalid synced block %d: %w",
				block.Height,
				err,
			)
		}

		existing, err := LoadBlock(block.Height)

		switch {
		case err == nil:
			if existing.Hash != block.Hash {
				return fmt.Errorf(
					"conflicting stored block at height %d",
					block.Height,
				)
			}

		case errors.Is(err, os.ErrNotExist):
			// Expected for a new synced block.

		default:
			return fmt.Errorf(
				"failed to inspect stored block %d: %w",
				block.Height,
				err,
			)
		}

		previous = block
	}

	// Phase 2: persist and commit.
	for _, block := range blocks {
		if err := SaveBlock(block); err != nil {
			return fmt.Errorf(
				"failed to save synced block %d: %w",
				block.Height,
				err,
			)
		}

		if err := CommitBlock(block); err != nil {
			if err != ErrBlockAlreadyCommitted {
				return fmt.Errorf(
					"failed to commit synced block %d: %w",
					block.Height,
					err,
				)
			}
		}

		if NodeMempool != nil {
			NodeMempool.RemoveProcessedTransactions(block.Transactions)
		}
	}

	return nil
}
