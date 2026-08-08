package app

import "errors"

const MaxBlocksPerSyncRequest = 100

func ValidateBlockRequest(req BlockRequest) error {
	if req.FromHeight < 1 {
		return errors.New("invalid from height")
	}

	if req.ToHeight < req.FromHeight {
		return errors.New("invalid height range")
	}

	if req.ToHeight-req.FromHeight+1 > MaxBlocksPerSyncRequest {
		return errors.New("block sync request too large")
	}

	return nil
}
