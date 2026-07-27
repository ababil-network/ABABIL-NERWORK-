package app

func TestMempool() {

	mempool := NewMempool()

	tx := NewTransaction(
		"0xABA1111111111111111111111111111111111111",
		"0xABA2222222222222222222222222222222222222",
		100,
	)

	mempool.AddTransaction(tx)

	LogInfo("Mempool Created")
	LogInfo("Pending Transactions : 1")
}
