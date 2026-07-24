package app

// AppModules contains all registered blockchain modules.
//
// Cosmos SDK modules will be registered here
// during the production implementation.
type AppModules struct{}

// NewAppModules creates the default module registry.
func NewAppModules() *AppModules {
	return &AppModules{}
}
