package app

import (
	"io"
	"net"
	"strconv"
)

func HandleBlock(conn net.Conn) {
	defer conn.Close()

	var block Block

	if err := ReceiveJSON(conn, &block); err != nil {
		if err == io.EOF {
			return
		}

		LogError(err.Error())
		return
	}

	LogInfo("=================================")
	LogInfo("Block Received")
	LogInfo("Height : " + strconv.Itoa(block.Height))
	LogInfo("Hash : " + block.Hash)
	LogInfo("=================================")

	latest, err := GetLatestBlock()
	if err != nil {
		LogError(err.Error())
		return
	}

	if err := ValidateBlock(block, latest); err != nil {
		LogError(err.Error())
		return
	}

	// Persist first. The in-memory chain is changed only after persistence
	// succeeds.
	if err := SaveBlock(block); err != nil {
		LogError("failed to persist block: " + err.Error())
		return
	}

	if err := CommitBlock(block); err != nil {
		// An already committed identical block is idempotent. A conflicting
		// block must never be silently accepted.
		if err == ErrBlockAlreadyCommitted {
			LogInfo("Block already committed")
			return
		}

		LogError("failed to commit block: " + err.Error())
		return
	}

	// Only remove transactions after the block has successfully passed
	// validation and has been persisted/committed.
	if NodeMempool != nil {
		NodeMempool.RemoveProcessedTransactions(block.Transactions)
	}

	LogInfo("Block Successfully Added")
}
