package app

import "time"

func StartHeartbeatWorker() {

	go func() {

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {

			peers := GetPeers()

			for _, peer := range peers {

				NodeHealth.Update(peer.Address)
			}
		}
	}()
}
