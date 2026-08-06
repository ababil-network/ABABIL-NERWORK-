package app

import (
	"encoding/json"
	"net"
)

func SendJSON(conn net.Conn, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return WriteFrame(conn, payload)
}

func ReceiveJSON(conn net.Conn, data any) error {
	payload, err := ReadFrame(conn)
	if err != nil {
		return err
	}

	return json.Unmarshal(payload, data)
}
