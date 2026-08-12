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

	// The in-memory chain may be empty after startup even though the
	// persisted blockchain already contains the current canonical tip.
	// Synchronize the in-memory genesis/tip before committing synced blocks.
	if len(BlockchainSnapshot()) == 0 {
		if latest.Height == 0 {
			if err := CommitBlock(latest); err != nil &&
				!errors.Is(err, ErrBlockAlreadyCommitted) {
				return fmt.Errorf("failed to initialize committed genesis: %w", err)
			}
		} else {
			if err := RecoverBlockchainFromDisk(); err != nil {
				return fmt.Errorf(
					"failed to recover persisted blockchain before sync: %w",
					err,
				)
			}
		}
	}

	previous := latest

	// Phase 1:
	// Validate the complete batch before mutating persistent or in-memory state.
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

	// Phase 2:
	// Persist and commit each validated block in canonical order.
	for _, block := range blocks {
		if err := SaveBlock(block); err != nil {
			return fmt.Errorf(
				"failed to save synced block %d: %w",
				block.Height,
				err,
			)
		}

		if err := CommitBlock(block); err != nil {
			if !errors.Is(err, ErrBlockAlreadyCommitted) {
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
