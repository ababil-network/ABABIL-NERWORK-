package app

import (
	"encoding/json"
	"net"
)

func BroadcastBlock(block Block, peer string) error {

	conn, err := net.Dial("tcp", peer)
	if err != nil {
		return err
	}
	defer conn.Close()

	return json.NewEncoder(conn).Encode(block)
}
