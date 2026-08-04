package app

import (
	"net"
	"sync/atomic"
	"time"
)

func HandlePeer(conn net.Conn) {
	defer func() {
		conn.Close()

		if atomic.LoadInt32(&CurrentPeers) > 0 {
			atomic.AddInt32(&CurrentPeers, -1)
		}
	}()

	conn.SetDeadline(time.Now().Add(15 * time.Second))

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
