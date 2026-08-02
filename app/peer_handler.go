package app

import "net"

func HandlePeer(conn net.Conn) {
	defer conn.Close()

	_, err := ReceiveHandshake(conn)
	if err != nil {
		LogError(err.Error())
		return
	}

	LogInfo("=================================")
	LogInfo("Peer Session Started")
	LogInfo("Remote : " + conn.RemoteAddr().String())
	LogInfo("=================================")
}
