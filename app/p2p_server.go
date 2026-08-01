package app

import (
	"net"
)

var P2PListener net.Listener

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
				continue
			}

			LogInfo("Peer Connected : " + conn.RemoteAddr().String())

			go HandleTransaction(conn)
		}
	}()

	return nil
}
