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

			LogInfo("Peer Connected : " + conn.RemoteAddr().String())

			go HandlePeer(conn)
		}
	}()

	return nil
}
