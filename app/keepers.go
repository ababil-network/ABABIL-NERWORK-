package app

// AppKeepers holds all blockchain keepers.
//
// Cosmos SDK keepers will be added here
// during the production implementation.
type AppKeepers struct{}

// NewAppKeepers creates the default keeper set.
func NewAppKeepers() *AppKeepers {
	return &AppKeepers{}
}
