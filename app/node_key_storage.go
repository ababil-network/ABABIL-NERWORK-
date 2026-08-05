package app

import (
	"encoding/json"
	"os"
)

const NodeKeyFile = "node_key.json"

func SaveNodeKey(key *NodeKey) error {

	data, err := json.MarshalIndent(key, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(NodeKeyFile, data, 0600)
}

func LoadNodeKey() (*NodeKey, error) {

	data, err := os.ReadFile(NodeKeyFile)
	if err != nil {
		return nil, err
	}

	var key NodeKey

	if err := json.Unmarshal(data, &key); err != nil {
		return nil, err
	}

	return &key, nil
}
