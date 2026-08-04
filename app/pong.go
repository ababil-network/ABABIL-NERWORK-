package app

import "time"

type PongMessage struct {
	Timestamp int64
}

func NewPong() PongMessage {
	return PongMessage{
		Timestamp: time.Now().Unix(),
	}
}
