package app

import (
	"encoding/json"
	"net"
)

type HandshakeResponse struct {
	Accepted bool   `json:"accepted"`
	Message  string `json:"message"`
}

func SendHandshakeResponse(conn net.Conn, accepted bool, message string) error {

	resp := HandshakeResponse{
		Accepted: accepted,
		Message:  message,
	}

	return json.NewEncoder(conn).Encode(resp)
}
