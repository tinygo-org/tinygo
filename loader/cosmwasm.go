package loader

// mergeCosmwasmPaths overrides packages that need to be modified in order to support
// cosmwasm - specifically it overrides packages that contain operations that involve floats.
func mergeCosmwasmPaths(m map[string]bool) {
	m["strconv/"] = true
}
