package app

import "encoding/json"

func EncodeMessage(msg NetworkMessage) ([]byte, error) {
	return json.Marshal(msg)
}

func DecodeMessage(data []byte) (NetworkMessage, error) {
	var msg NetworkMessage

	err := json.Unmarshal(data, &msg)

	return msg, err
}
