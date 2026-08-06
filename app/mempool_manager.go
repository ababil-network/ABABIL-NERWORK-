package app

var NodeMempool *Mempool

func InitMempool() {

	NodeMempool = NewMempool()

	StartMempoolCleanup(NodeMempool)

	LogInfo("Global Mempool Initialized")
}
