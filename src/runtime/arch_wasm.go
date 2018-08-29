// +build wasm

package runtime

const GOARCH = "wasm"

// The length type used inside strings and slices.
type lenType uint32
