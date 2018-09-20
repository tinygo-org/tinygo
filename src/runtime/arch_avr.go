// +build avr

package runtime

const GOARCH = "avr"

// The length type used inside strings and slices.
type lenType uint16

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 8
