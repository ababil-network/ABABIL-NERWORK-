package app

import "errors"

var (
	ErrInvalidProtocol = errors.New("invalid protocol version")
	ErrInvalidChainID  = errors.New("invalid chain id")
	ErrInvalidNetwork  = errors.New("invalid network")
	ErrUnknownMessage  = errors.New("unknown network message")
	ErrInvalidNonce    = errors.New("invalid nonce")
)
