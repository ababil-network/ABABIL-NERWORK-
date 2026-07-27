package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func SaveTransaction(tx Transaction) error {
if err := ValidateTransaction(tx); err != nil {
    return err
}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".ababil", "data", "transactions")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		return err
	}

	file := filepath.Join(dir, tx.ID+".json")

	return os.WriteFile(file, data, 0644)
}
