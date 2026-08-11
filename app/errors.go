package app

import "errors"

var (
	ErrInvalidProtocol          = errors.New("invalid protocol version")
	ErrInvalidChainID           = errors.New("invalid chain id")
	ErrInvalidNetwork           = errors.New("invalid network")
	ErrInvalidNodeID            = errors.New("invalid node id")
	ErrUnknownMessage           = errors.New("unknown network message")
	ErrInvalidNonce             = errors.New("invalid nonce")
	ErrTransactionValueOverflow = errors.New("transaction value overflow")
	ErrMessageTooLarge          = errors.New("message exceeds maximum size")
)
