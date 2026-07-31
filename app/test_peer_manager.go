package app

func TestPeerManager() {

	AddPeer("192.168.1.10:26656")
	AddPeer("192.168.1.11:26656")
	AddPeer("192.168.1.12:26656")

	LogInfo("=================================")
	LogInfo("Peer Manager Test")
	LogInfo("=================================")
	LogInfo("Total Peers : 3")

	for _, peer := range GetPeers() {
		LogInfo("Peer : " + peer.Address)
	}

	RemovePeer("192.168.1.11:26656")

	LogInfo("=================================")
	LogInfo("After Remove Peer")
	LogInfo("Total Peers : 2")

	for _, peer := range GetPeers() {
		LogInfo("Peer : " + peer.Address)
	}
}
