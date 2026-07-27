package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func SaveAccount(account Account) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".ababil", "data", "accounts")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return err
	}

	file := filepath.Join(dir, account.Address+".json")

	return os.WriteFile(file, data, 0644)
}
