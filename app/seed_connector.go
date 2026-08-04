package app

func ConnectSeeds() {

	seeds := NodeSeed.List()

	for _, seed := range seeds {

		if HasPeer(seed) {
			continue
		}

		go func(addr string) {
			err := ConnectPeer(addr)
			if err == nil {
				AddPeer(addr)
				LogInfo("Seed Connected : " + addr)
			}
		}(seed)
	}
}
