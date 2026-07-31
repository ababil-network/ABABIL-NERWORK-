package app

import "errors"

const MaxTransactionsPerBlock = 5000

func ValidateBlock(block Block, previous Block) error {

	// Height check
	if block.Height != previous.Height+1 {
		return errors.New("invalid block height")
	}

	// Empty block check
	if len(block.Transactions) == 0 {
		return errors.New("empty block")
	}

	// Maximum transaction limit
	if len(block.Transactions) > MaxTransactionsPerBlock {
		return errors.New("too many transactions")
	}

        // Previous Hash check
        if block.PreviousHash != previous.Hash {
                return errors.New("invalid previous hash")
        }

        // Duplicate Transaction check
        seen := make(map[string]bool)
        for _, tx := range block.Transactions {
                if seen[tx.Hash] {
                        return errors.New("duplicate transaction")
                }
                seen[tx.Hash] = true
        }

	// Block hash check
	if block.Hash == "" {
		return errors.New("invalid block hash")
	}

	return nil
}
