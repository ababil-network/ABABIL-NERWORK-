package app

func TestP2PConnect() {

	LogInfo("=================================")
	LogInfo("P2P Connection Test")
	LogInfo("=================================")

	err := ConnectPeer("127.0.0.1:26656")
	if err != nil {
		LogError(err.Error())
		return
	}

	LogInfo("P2P Test Success")
}
