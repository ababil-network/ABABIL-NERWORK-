package app

import "testing"

func TestMempoolAddTransaction(t *testing.T) {

	mempool := NewMempool()

	tx := Transaction{
		Hash: "tx1",
	}

	mempool.AddTransaction(tx)

	if mempool.Count() != 1 {
		t.Fatal("transaction was not added")
	}
}

func TestMempoolRemoveProcessed(t *testing.T) {

	mempool := NewMempool()

	tx1 := Transaction{
		Hash: "tx1",
	}

	tx2 := Transaction{
		Hash: "tx2",
	}

	mempool.AddTransaction(tx1)
	mempool.AddTransaction(tx2)

	mempool.RemoveProcessedTransactions([]Transaction{tx1})

	if mempool.Count() != 1 {
		t.Fatal("processed transaction was not removed")
	}

	if mempool.Transactions[0].Hash != "tx2" {
		t.Fatal("wrong transaction remains")
	}
}
func TestMempoolDuplicateTransaction(t *testing.T) {

	mempool := NewMempool()

	tx := Transaction{
		Hash: "tx1",
	}

	mempool.AddTransaction(tx)
	mempool.AddTransaction(tx)

	if mempool.Count() != 1 {
		t.Fatal("duplicate transaction accepted")
	}
}
func TestMempoolLimit(t *testing.T) {

	mempool := NewMempool()

	for i := 0; i < MaxMempoolTransactions+100; i++ {

		tx := Transaction{
			Hash: "tx" + string(rune(i)),
		}

		mempool.AddTransaction(tx)
	}

	if mempool.Count() > MaxMempoolTransactions {
		t.Fatal("mempool limit exceeded")
	}
}
