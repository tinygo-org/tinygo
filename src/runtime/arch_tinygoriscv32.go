//go:build tinygo.riscv32

package runtime

// GOARCH reports the real CPU architecture. The Go toolchain still uses arm
// through the go-toolchain-goarch target property.
const GOARCH = "riscv"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32
