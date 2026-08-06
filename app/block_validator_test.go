package app

import "testing"

func TestValidateBlock(t *testing.T) {

	previous := Block{
		Height: 1,
		Hash:   "previous_hash",
	}

	validBlock := Block{
		Height:       2,
		PreviousHash: "previous_hash",
		Hash:         "block_hash",
		Transactions: []Transaction{
			{
				Hash: "tx1",
			},
		},
	}

	err := ValidateBlock(validBlock, previous)
	if err != nil {
		t.Fatal("valid block rejected:", err)
	}
}

func TestInvalidBlockHeight(t *testing.T) {

	previous := Block{
		Height: 1,
		Hash:   "previous_hash",
	}

	block := Block{
		Height:       5,
		PreviousHash: "previous_hash",
		Hash:         "block_hash",
		Transactions: []Transaction{
			{
				Hash: "tx1",
			},
		},
	}

	err := ValidateBlock(block, previous)

	if err == nil {
		t.Fatal("invalid height accepted")
	}
}
func TestInvalidPreviousHash(t *testing.T) {

	previous := Block{
		Height: 1,
		Hash:   "previous_hash",
	}

	block := Block{
		Height:       2,
		PreviousHash: "wrong_hash",
		Hash:         "block_hash",
		Transactions: []Transaction{
			{
				Hash: "tx1",
			},
		},
	}

	err := ValidateBlock(block, previous)

	if err == nil {
		t.Fatal("invalid previous hash accepted")
	}
}

func TestEmptyBlock(t *testing.T) {

	previous := Block{
		Height: 1,
		Hash:   "previous_hash",
	}

	block := Block{
		Height:       2,
		PreviousHash: "previous_hash",
		Hash:         "block_hash",
		Transactions: []Transaction{},
	}

	err := ValidateBlock(block, previous)

	if err == nil {
		t.Fatal("empty block accepted")
	}
}
func TestDuplicateTransaction(t *testing.T) {

	previous := Block{
		Height: 1,
		Hash:   "previous_hash",
	}

	block := Block{
		Height:       2,
		PreviousHash: "previous_hash",
		Hash:         "block_hash",
		Transactions: []Transaction{
			{
				Hash: "tx1",
			},
			{
				Hash: "tx1",
			},
		},
	}

	err := ValidateBlock(block, previous)

	if err == nil {
		t.Fatal("duplicate transaction accepted")
	}
}
