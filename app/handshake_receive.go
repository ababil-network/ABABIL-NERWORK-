package app

import (
	"net"
)

func ReceiveHandshake(conn net.Conn) (*Handshake, error) {

	var hs Handshake

	err := ReceiveJSON(conn, &hs)
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
	if hs.NodeID == "" {
		return nil, ErrInvalidNodeID
	}

	LogInfo("=================================")
	LogInfo("Handshake Verified")
	LogInfo("Node : " + hs.NodeName)
	LogInfo("Version : " + hs.NodeVersion)
	LogInfo("Node ID : " + hs.NodeID)
	LogInfo("=================================")

	return &hs, nil
}
