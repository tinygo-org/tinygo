//go:build !avr

package reflectlite

// typeIDField holds the type ID that interface lowering assigns to some types.
// It must match hasTypeIDField in compiler/interface.go.
type typeIDField uint8
