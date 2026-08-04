package app

import (
	"net"
)

func ConnectPeer(address string) error {

	d := net.Dialer{
		Timeout: NodeNetworkConfig.HandshakeTimeout,
	}

	conn, err := d.Dial("tcp", address)

	if err != nil {
		return err
	}

	LogInfo("=================================")
	LogInfo("Connected To Peer")
	LogInfo("Peer : " + conn.RemoteAddr().String())
	LogInfo("=================================")
	go HandlePeer(conn)
	conn.Close()

	return nil
}
