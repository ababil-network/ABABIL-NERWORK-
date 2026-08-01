package app

func VerifyHandshake(hs *Handshake) error {

	if hs.ProtocolVersion != 1 {
		return ErrInvalidProtocol
	}

	if hs.ChainID != 7777 {
		return ErrInvalidChainID
	}

	if hs.Network != "ABABIL Network" {
		return ErrInvalidNetwork
	}

	return nil
}
