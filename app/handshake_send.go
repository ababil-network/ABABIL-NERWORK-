package app

import (
	"net"
	"time"
)

func SendHandshake(conn net.Conn) error {

	hs := Handshake{
		ProtocolVersion: 1,
		ChainID:         7777,
		Network:         "ABABIL Network",

		NodeID: "ababil-node",

		NodeName:    "ABABIL Node",
		NodeVersion: "0.1.0",

		Timestamp: time.Now().Unix(),
	}

	return SendJSON(conn, hs)
}
