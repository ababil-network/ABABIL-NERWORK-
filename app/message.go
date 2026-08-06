package app

type MessageType uint8

const (
	MessageTransaction   MessageType = 1
	MessageBlock         MessageType = 2
	MessageHandshake     MessageType = 3
	MessagePing          MessageType = 4
	MessagePong          MessageType = 5
	MessageBlockRequest  MessageType = 6
	MessageBlockResponse MessageType = 7
)

type NetworkMessage struct {
	Type    MessageType `json:"type"`
	Payload []byte      `json:"payload"`
}
