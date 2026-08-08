package app

import (
	"encoding/json"
	"net"
	"strconv"
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
	case MessageBlockRequest:
		LogInfo("Dispatch : Block Request")

		var req BlockRequest

		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			return err
		}

		if err := ValidateBlockRequest(req); err != nil {
			return err
		}

		LogInfo("Block Sync Request : " +
			strconv.Itoa(req.FromHeight) + " -> " +
			strconv.Itoa(req.ToHeight))

	case MessageBlockResponse:
		LogInfo("Dispatch : Block Response")

		var response BlockResponse

		if err := json.Unmarshal(msg.Payload, &response); err != nil {
			return err
		}

		LogInfo("Blocks Received : " +
			strconv.Itoa(len(response.Blocks)))
		if err := CommitSyncedBlocks(response.Blocks); err != nil {
			return err
		}

		LogInfo("Synced Blocks Committed")

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
