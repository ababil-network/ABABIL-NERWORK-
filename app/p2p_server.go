package app

import (
	"net"
	"sync/atomic"
)

var P2PListener net.Listener

const MaxPeers int32 = 100

var CurrentPeers int32

func StartP2PServer() error {

	ln, err := net.Listen("tcp", ":26656")
	if err != nil {
		return err
	}

	P2PListener = ln

	LogInfo("=================================")
	LogInfo("P2P Server Started")
	LogInfo("Listening : 0.0.0.0:26656")
	LogInfo("=================================")

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {

				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					continue
				}

				LogInfo("P2P Server Stopped")
				return
			}
			if atomic.LoadInt32(&CurrentPeers) >= MaxPeers {
				conn.Close()
				continue
			}

			atomic.AddInt32(&CurrentPeers, 1)

			ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
			if err != nil {
				conn.Close()
				atomic.AddInt32(&CurrentPeers, -1)
				continue
			}

			if !NodeRateLimiter.Allow(ip) {

				duration := NodeBan.HandleViolation(ip)

				if duration == 0 {
					LogInfo("Warning : " + ip)
				} else {
					LogInfo("Peer Banned : " + ip)
				}

				conn.Close()
				atomic.AddInt32(&CurrentPeers, -1)
				continue
			}

			LogInfo("Peer Connected : " + conn.RemoteAddr().String())

			go HandlePeer(conn)
		}
	}()
	go ConnectSeeds()

	StartReconnectWorker()

	LogInfo("Seed Discovery Started")
	LogInfo("Reconnect Worker Started")

	return nil

}
