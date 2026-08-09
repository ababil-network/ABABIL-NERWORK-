package app

import "testing"

func TestCanonicalTransactionHashChangesWithAmount(t *testing.T) {
	tx := Transaction{
		ID:       "tx-test-1",
		From:     "0x1111111111111111111111111111111111111111",
		To:       "0x2222222222222222222222222222222222222222",
		Amount:   100,
		GasLimit: 21000,
		GasPrice: 1,
		Fee:      21000,
		Nonce:    1,
	}

	tx.Timestamp = tx.Timestamp.UTC()

	hash1, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}

	tx.Amount = 101

	hash2, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 == hash2 {
		t.Fatal("transaction hash did not change after amount changed")
	}
}

func TestCanonicalTransactionHashStable(t *testing.T) {
	tx := Transaction{
		ID:       "tx-test-2",
		From:     "0x1111111111111111111111111111111111111111",
		To:       "0x2222222222222222222222222222222222222222",
		Amount:   100,
		GasLimit: 21000,
		GasPrice: 1,
		Fee:      21000,
		Nonce:    1,
	}

	hash1, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 != hash2 {
		t.Fatal("canonical transaction hash is not deterministic")
	}
}
