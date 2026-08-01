package app

import (
	"encoding/json"
	"net"
)

func SendHandshake(conn net.Conn) error {

	hs := Handshake{
		ProtocolVersion: 1,
		ChainID:         7777,
		Network:         "ABABIL Network",
		NodeName:        "ABABIL Node",
		NodeVersion:     "0.1.0",
	}

	return json.NewEncoder(conn).Encode(hs)
}
