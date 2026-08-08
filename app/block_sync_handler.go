package app

import "fmt"

func HandleBlockRequest(req BlockRequest) (BlockResponse, error) {
	if err := ValidateBlockRequest(req); err != nil {
		return BlockResponse{}, err
	}

	response := BlockResponse{
		Blocks: make([]Block, 0, req.ToHeight-req.FromHeight+1),
	}

	for height := req.FromHeight; height <= req.ToHeight; height++ {
		block, err := LoadBlock(height)
		if err != nil {
			return BlockResponse{}, fmt.Errorf("failed to load block %d: %w", height, err)
		}

		response.Blocks = append(response.Blocks, block)
	}

	return response, nil
}
