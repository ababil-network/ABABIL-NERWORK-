package app

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const CurrentStateVersion uint32 = 1

var (
	ErrInvalidStateVersion = errors.New("invalid state version")
	ErrStateChecksum       = errors.New("state checksum mismatch")
	ErrStateCorrupted      = errors.New("state corrupted")
)

type PersistentWalletState struct {
	Address string `json:"address"`
	Balance uint64 `json:"balance"`
}

type PersistentState struct {
	Version uint32                  `json:"version"`
	Height  uint64                  `json:"height"`
	Wallets []PersistentWalletState `json:"wallets"`
}

type persistedStateEnvelope struct {
	State    PersistentState `json:"state"`
	Checksum string          `json:"checksum"`
}

func stateDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(home, ".ababil", "data", "state")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return dir, nil
}

func stateFilePath() (string, error) {
	dir, err := stateDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "state.json"), nil
}

func canonicalStateBytes(state PersistentState) ([]byte, error) {
	return json.Marshal(state)
}

func calculateStateChecksum(state PersistentState) (string, error) {
	data, err := canonicalStateBytes(state)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)

	return fmt.Sprintf("%x", sum[:]), nil
}

func snapshotWalletState() []PersistentWalletState {
	walletBalanceMu.RLock()
	defer walletBalanceMu.RUnlock()

	wallets := make([]PersistentWalletState, 0, len(WalletBalances))

	for _, wallet := range WalletBalances {
		wallets = append(wallets, PersistentWalletState{
			Address: wallet.Address,
			Balance: wallet.Balance,
		})
	}

	sort.Slice(wallets, func(i, j int) bool {
		return wallets[i].Address < wallets[j].Address
	})

	return wallets
}

func BuildPersistentState(height uint64) PersistentState {
	return PersistentState{
		Version: CurrentStateVersion,
		Height:  height,
		Wallets: snapshotWalletState(),
	}
}

func ValidatePersistentState(state PersistentState) error {
	if state.Version != CurrentStateVersion {
		return ErrInvalidStateVersion
	}

	seen := make(map[string]struct{}, len(state.Wallets))

	for _, wallet := range state.Wallets {
		if wallet.Address == "" {
			return ErrStateCorrupted
		}

		if _, exists := seen[wallet.Address]; exists {
			return ErrStateCorrupted
		}

		seen[wallet.Address] = struct{}{}
	}

	return nil
}

func SavePersistentState(state PersistentState) error {
	if err := ValidatePersistentState(state); err != nil {
		return err
	}

	path, err := stateFilePath()
	if err != nil {
		return err
	}

	checksum, err := calculateStateChecksum(state)
	if err != nil {
		return err
	}

	envelope := persistedStateEnvelope{
		State:    state,
		Checksum: checksum,
	}

	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return err
	}

	tempPath := path + ".tmp"

	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return err
	}

	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return err
	}

	return nil
}

func LoadPersistentState() (PersistentState, error) {
	path, err := stateFilePath()
	if err != nil {
		return PersistentState{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return PersistentState{}, err
	}

	var envelope persistedStateEnvelope

	if err := json.Unmarshal(data, &envelope); err != nil {
		return PersistentState{}, ErrStateCorrupted
	}

	if err := ValidatePersistentState(envelope.State); err != nil {
		return PersistentState{}, err
	}

	expectedChecksum, err := calculateStateChecksum(envelope.State)
	if err != nil {
		return PersistentState{}, err
	}

	if envelope.Checksum != expectedChecksum {
		return PersistentState{}, ErrStateChecksum
	}

	return envelope.State, nil
}

func ApplyPersistentState(state PersistentState) error {
	if err := ValidatePersistentState(state); err != nil {
		return err
	}

	wallets := make([]WalletBalance, 0, len(state.Wallets))

	for _, wallet := range state.Wallets {
		wallets = append(wallets, WalletBalance{
			Address: wallet.Address,
			Balance: wallet.Balance,
		})
	}

	walletBalanceMu.Lock()
	WalletBalances = wallets
	walletBalanceIndex = nil
	ensureWalletBalanceIndexLocked()
	walletBalanceMu.Unlock()

	return nil
}
