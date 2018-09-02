// +build amd64

package runtime

const GOARCH = "amd64"

// The length type used inside strings and slices.
type lenType uint32

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 64
