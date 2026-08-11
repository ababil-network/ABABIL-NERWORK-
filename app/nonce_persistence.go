package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

var ErrNonceStateCorrupted = errors.New("nonce state corrupted")

type PersistentNonce struct {
	Address string `json:"address"`
	Nonce   uint64 `json:"nonce"`
}

type PersistentNonceState struct {
	Version uint32            `json:"version"`
	Nonces  []PersistentNonce `json:"nonces"`
}

const CurrentNonceStateVersion uint32 = 1

func nonceStatePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(home, ".ababil", "data", "state")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(dir, "nonces.json"), nil
}

func BuildPersistentNonceState() PersistentNonceState {
	NodeNonce.mu.Lock()
	defer NodeNonce.mu.Unlock()

	nonces := make([]PersistentNonce, 0, len(NodeNonce.nonces))

	for address, nonce := range NodeNonce.nonces {
		nonces = append(nonces, PersistentNonce{
			Address: address,
			Nonce:   nonce,
		})
	}

	sort.Slice(nonces, func(i, j int) bool {
		return nonces[i].Address < nonces[j].Address
	})

	return PersistentNonceState{
		Version: CurrentNonceStateVersion,
		Nonces:  nonces,
	}
}

func ValidatePersistentNonceState(state PersistentNonceState) error {
	if state.Version != CurrentNonceStateVersion {
		return ErrNonceStateCorrupted
	}

	seen := make(map[string]struct{}, len(state.Nonces))

	for _, entry := range state.Nonces {
		if entry.Address == "" {
			return ErrNonceStateCorrupted
		}

		if _, exists := seen[entry.Address]; exists {
			return ErrNonceStateCorrupted
		}

		seen[entry.Address] = struct{}{}
	}

	return nil
}

func SavePersistentNonceState(state PersistentNonceState) error {
	if err := ValidatePersistentNonceState(state); err != nil {
		return err
	}

	path, err := nonceStatePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	temp := path + ".tmp"

	if err := os.WriteFile(temp, data, 0600); err != nil {
		return err
	}

	if err := os.Rename(temp, path); err != nil {
		_ = os.Remove(temp)
		return err
	}

	return nil
}

func LoadPersistentNonceState() (PersistentNonceState, error) {
	path, err := nonceStatePath()
	if err != nil {
		return PersistentNonceState{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return PersistentNonceState{}, err
	}

	var state PersistentNonceState

	if err := json.Unmarshal(data, &state); err != nil {
		return PersistentNonceState{}, ErrNonceStateCorrupted
	}

	if err := ValidatePersistentNonceState(state); err != nil {
		return PersistentNonceState{}, err
	}

	return state, nil
}

func ApplyPersistentNonceState(state PersistentNonceState) error {
	if err := ValidatePersistentNonceState(state); err != nil {
		return err
	}

	NodeNonce.mu.Lock()
	defer NodeNonce.mu.Unlock()

	NodeNonce.nonces = make(map[string]uint64, len(state.Nonces))

	for _, entry := range state.Nonces {
		NodeNonce.nonces[entry.Address] = entry.Nonce
	}

	return nil
}
