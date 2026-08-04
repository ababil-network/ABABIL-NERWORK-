package app

import "time"

type PingMessage struct {
	Timestamp int64
}

func NewPing() PingMessage {
	return PingMessage{
		Timestamp: time.Now().Unix(),
	}
}
