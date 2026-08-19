package app

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

func InitNodeKey() error {
	key, err := LoadNodeKey()
	if err == nil {
		if key.PrivateKey == "" || key.PublicKey == "" || key.NodeID == "" {
			return fmt.Errorf("stored node key is incomplete")
		}

		LocalNodeKey = key
		LogInfo("Existing node key loaded.")
		return nil
	}

	LogInfo("Node key not found. Generating new node identity...")

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate node key: %w", err)
	}

	privateBytes := crypto.FromECDSA(privateKey)
	publicBytes := crypto.FromECDSAPub(&privateKey.PublicKey)

	nodeIDHash := sha256.Sum256(publicBytes)

	key = &NodeKey{
		PrivateKey: hex.EncodeToString(privateBytes),
		PublicKey:  hex.EncodeToString(publicBytes),
		NodeID:     hex.EncodeToString(nodeIDHash[:]),
	}

	if err := SaveNodeKey(key); err != nil {
		return fmt.Errorf("failed to save node key: %w", err)
	}

	LocalNodeKey = key

	LogInfo("New node identity generated and persisted.")
	return nil
}
