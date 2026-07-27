package app

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func GenerateBlockHash() string {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}

func ProduceBlock(height int, mempool *Mempool) Block {

	block := Block{
		Height:    height,
		Hash:      GenerateBlockHash(),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return block
}
