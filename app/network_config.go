package app

import "time"

type NetworkConfig struct {
	MaxPeers         int32
	HandshakeTimeout time.Duration
	ReconnectDelay   time.Duration
	HeartbeatDelay   time.Duration
}

var NodeNetworkConfig = NetworkConfig{
	MaxPeers:         100,
	HandshakeTimeout: 15 * time.Second,
	ReconnectDelay:   30 * time.Second,
	HeartbeatDelay:   30 * time.Second,
}
