// +build wasm

package runtime

const GOARCH = "wasm"

// The length type used inside strings and slices.
type lenType uint32

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32
