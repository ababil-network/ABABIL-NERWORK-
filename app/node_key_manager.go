package app

func InitNodeKey() error {

	key, err := LoadNodeKey()
	if err == nil {
		LocalNodeKey = key
		LogInfo("Existing node key loaded.")
		return nil
	}

	LogInfo("Node key not found.")
	LogInfo("Node key generation will be implemented next.")

	return nil
}
