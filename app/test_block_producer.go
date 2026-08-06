package app

func TestBlockProducer() {

	InitMempool()

	tx := NewTransaction(
		"0xABA1111111111111111111111111111111111111",
		"0xABA2222222222222222222222222222222222222",
		100,
	)

	NodeMempool.AddTransaction(tx)

	block := ProduceBlock(1, "GENESIS_BLOCK")

	LogInfo("New Block Produced")
	LogInfo("Block Height : 1")
	LogInfo("Transactions : 1")
	LogInfo("Block Hash : " + block.Hash)
}
