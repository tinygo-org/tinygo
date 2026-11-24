//go:build avr

package reflectlite

// On AVR numMethod directly follows the kind byte, so there is no type ID.
type typeIDField struct{}
