package app

import "fmt"

func CommitSyncedBlocks(blocks []Block) error {
	if len(blocks) == 0 {
		return fmt.Errorf("empty block response")
	}

	latest, err := GetLatestBlock()
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	previous := latest

	for _, block := range blocks {
		if err := ValidateBlock(block, previous); err != nil {
			return fmt.Errorf("invalid synced block %d: %w", block.Height, err)
		}

		if err := SaveBlock(block); err != nil {
			return fmt.Errorf("failed to save synced block %d: %w", block.Height, err)
		}

		CommitBlock(block)

		previous = block
	}

	return nil
}
