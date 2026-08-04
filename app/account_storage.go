package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func SaveAccount(account Account) error {
	if account.Address == "" {
		return os.ErrInvalid
	}

	if !IsValidAddress(account.Address) {
		return os.ErrInvalid
	}
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
	if _, err := os.Stat(file); err == nil {
		return os.ErrExist
	}
	return os.WriteFile(file, data, 0644)
}
