package app

import (
	"testing"
	"time"
)

func TestApplyTransactionRejectsInvalidTransactionWithoutStateChange(t *testing.T) {
	originalBalances := WalletBalances
	originalPool := CurrentRewardPool
	originalHistory := RewardHistory

	defer func() {
		WalletBalances = originalBalances
		CurrentRewardPool = originalPool
		RewardHistory = originalHistory
	}()

	from := "0x1111111111111111111111111111111111111111"
	to := "0x2222222222222222222222222222222222222222"

	WalletBalances = []WalletBalance{
		{
			Address: from,
			Balance: 100000,
		},
	}

	tx := Transaction{
		ID:        "atomic-invalid-test",
		From:      from,
		To:        to,
		Amount:    100,
		GasLimit:  DefaultGasLimit,
		GasPrice:  DefaultGasPrice,
		Fee:       DefaultGasLimit * DefaultGasPrice,
		Nonce:     1,
		Timestamp: time.Now().UTC(),
	}

	// Deliberately invalidate the transaction.
	tx.Hash = "invalid-hash"

	beforeSender := GetBalance(from)
	beforeReceiver := GetBalance(to)
	beforePool := CurrentRewardPool
	beforeHistory := len(RewardHistory)

	err := ApplyTransaction(tx)
	if err == nil {
		t.Fatal("expected invalid transaction to be rejected")
	}

	if GetBalance(from) != beforeSender {
		t.Fatal("sender balance changed after rejected transaction")
	}

	if GetBalance(to) != beforeReceiver {
		t.Fatal("receiver balance changed after rejected transaction")
	}

	if CurrentRewardPool != beforePool {
		t.Fatal("reward pool changed after rejected transaction")
	}

	if len(RewardHistory) != beforeHistory {
		t.Fatal("reward history changed after rejected transaction")
	}

	if NodeReplay.Exists(tx.Hash) {
		t.Fatal("rejected transaction was added to replay state")
	}
}
