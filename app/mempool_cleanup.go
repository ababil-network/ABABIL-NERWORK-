package app

import "time"

func StartMempoolCleanup(mempool *Mempool) {

	ticker := time.NewTicker(time.Minute)

	go func() {

		for range ticker.C {

			mempool.RemoveExpiredTransactions()

			LogInfo("Expired mempool transactions cleaned")
		}
	}()
}
