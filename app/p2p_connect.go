package app

import (
	"net"
)

func ConnectPeer(address string) error {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	LogInfo("=================================")
	LogInfo("Connected To Peer")
	LogInfo("Peer : " + conn.RemoteAddr().String())
	LogInfo("=================================")

	conn.Close()

	return nil
}
