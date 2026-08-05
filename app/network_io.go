package app

import (
	"encoding/json"
	"net"
)

func SendJSON(conn net.Conn, data any) error {
	return json.NewEncoder(conn).Encode(data)
}

func ReceiveJSON(conn net.Conn, data any) error {
	return json.NewDecoder(conn).Decode(data)
}
