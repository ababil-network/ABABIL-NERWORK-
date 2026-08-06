package app

type BlockRequest struct {
	FromHeight int `json:"from_height"`
	ToHeight   int `json:"to_height"`
}
