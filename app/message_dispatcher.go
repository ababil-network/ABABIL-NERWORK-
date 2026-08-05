package app

import (
	"net"
)

func DispatchMessage(conn net.Conn) error {
	var msg NetworkMessage

	err := ReceiveJSON(conn, &msg)
	if err != nil {
		return err
	}

	switch msg.Type {

	case MessageTransaction:
		LogInfo("Dispatch : Transaction")

	case MessageBlock:
		LogInfo("Dispatch : Block")

	case MessageHandshake:
		LogInfo("Dispatch : Handshake")

	case MessagePing:
		LogInfo("Dispatch : Ping")

	case MessagePong:
		LogInfo("Dispatch : Pong")

	default:
		return ErrUnknownMessage
	}

	return nil
}
