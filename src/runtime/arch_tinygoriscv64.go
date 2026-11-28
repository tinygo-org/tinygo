//go:build tinygo.riscv64

package runtime

// GOARCH reports the real CPU architecture.
const GOARCH = "riscv64"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 64
