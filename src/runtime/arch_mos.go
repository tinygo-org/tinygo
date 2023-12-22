//go:build tinygo.mos

package runtime

const GOARCH = "mos" // riscv pretends to be arm

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 8
