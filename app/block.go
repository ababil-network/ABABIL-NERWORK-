package app

import (
	"fmt"
	"time"
)

type Block struct {
	Height    int
	Hash      string
	Timestamp string
}

func CreateGenesisBlock() Block {
	block := Block{
		Height:    0,
		Hash:      "GENESIS_BLOCK",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	fmt.Println("[INFO] Genesis Block Created")
	fmt.Println("[INFO] Height :", block.Height)
	fmt.Println("[INFO] Hash   :", block.Hash)
	fmt.Println("[INFO] Time   :", block.Timestamp)

	return block
}
