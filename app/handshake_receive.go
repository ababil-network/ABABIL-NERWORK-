package app

import (
	"encoding/json"
	"net"
)

func ReceiveHandshake(conn net.Conn) (*Handshake, error) {

	var hs Handshake

	err := json.NewDecoder(conn).Decode(&hs)
	if err != nil {
		return nil, err
	}

	if hs.ProtocolVersion != 1 {
		return nil, ErrInvalidProtocol
	}

	if hs.ChainID != 7777 {
		return nil, ErrInvalidChainID
	}

	if hs.Network != "ABABIL Network" {
		return nil, ErrInvalidNetwork
	}

	LogInfo("=================================")
	LogInfo("Handshake Verified")
	LogInfo("Node : " + hs.NodeName)
	LogInfo("Version : " + hs.NodeVersion)
	LogInfo("=================================")

	return &hs, nil
}
