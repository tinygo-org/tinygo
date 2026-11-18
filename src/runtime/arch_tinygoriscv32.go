//go:build tinygo.riscv32

package runtime

// GOARCH reports the real CPU architecture. The target definition still uses
// GOARCH=arm for package resolution; changing that is a separate change.
const GOARCH = "riscv"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32
