package app

import "time"

type NetworkConfig struct {
	MaxPeers         int32
	HandshakeTimeout time.Duration
	ReconnectDelay   time.Duration
	HeartbeatDelay   time.Duration

	MaxMessageSize int

	MaxSendRate    int64
	MaxReceiveRate int64

	MaxInboundQueue  int
	MaxOutboundQueue int

	MaxConnectionsPerIP int
}

var NodeNetworkConfig = NetworkConfig{
	MaxPeers:         100,
	HandshakeTimeout: 15 * time.Second,
	ReconnectDelay:   30 * time.Second,
	HeartbeatDelay:   30 * time.Second,
	MaxMessageSize:   1024 * 1024,
	MaxSendRate:      5 * 1024 * 1024,
	MaxReceiveRate:   5 * 1024 * 1024,

	MaxInboundQueue:  1024,
	MaxOutboundQueue: 1024,

	MaxConnectionsPerIP: 5,
}
