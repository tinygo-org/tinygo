//go:build tinygo.riscv32

package runtime

// GOARCH reports the real CPU architecture. The Go toolchain and tinygo
// info still use GOARCH=arm for this target.
const GOARCH = "riscv32"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32
