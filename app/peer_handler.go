package app

import (
	"net"
	"sync/atomic"
	"time"
)

func HandlePeer(conn net.Conn) {
	defer func() {

		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		NodeConnections.Remove(ip)

		conn.Close()

		if atomic.LoadInt32(&CurrentPeers) > 0 {
			atomic.AddInt32(&CurrentPeers, -1)
		}
	}()

	conn.SetDeadline(time.Now().Add(NodeNetworkConfig.HandshakeTimeout))

	_, err := ReceiveHandshake(conn)
	if err != nil {
		LogError(err.Error())
		return
	}

	LogInfo("=================================")
	LogInfo("Peer Session Started")
	LogInfo("Remote : " + conn.RemoteAddr().String())
	LogInfo("=================================")
	for {

		conn.SetDeadline(time.Now().Add(NodeNetworkConfig.HeartbeatDelay))

		time.Sleep(5 * time.Second)

		if NodeBan.IsBanned(conn.RemoteAddr().String()) {
			LogInfo("Peer Disconnected : " + conn.RemoteAddr().String())
			return
		}

	}
}
