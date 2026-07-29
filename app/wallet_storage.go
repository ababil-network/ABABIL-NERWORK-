package app

import (
	"encoding/json"
	"os"
)

const WalletFile = "wallet.json"

func SaveWallet(wallet *Wallet) error {

	data, err := json.MarshalIndent(wallet, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(WalletFile, data, 0600)
}

func LoadWallet() (*Wallet, error) {

	data, err := os.ReadFile(WalletFile)
	if err != nil {
		return nil, err
	}

	var wallet Wallet

	if err := json.Unmarshal(data, &wallet); err != nil {
		return nil, err
	}

	return &wallet, nil
}
