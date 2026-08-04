package app

import "time"

func StartReconnectWorker() {

	go func() {

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			ConnectSeeds()
		}
	}()
}
