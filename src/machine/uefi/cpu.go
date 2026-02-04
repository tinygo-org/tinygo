//go:build uefi

package uefi

// Ticks returns a high-resolution monotonic counter.
//   - amd64: RDTSC (Time Stamp Counter)
//   - arm64: CNTVCT_EL0 (generic timer virtual count)
//
//export uefiTicks
func Ticks() uint64

// CpuPause hints to the CPU that we are in a spin-wait loop.
//   - amd64: PAUSE instruction
//   - arm64: YIELD instruction
//
//export uefiCpuPause
func CpuPause()

// ReadRandom reads a 64-bit hardware random number.
// Returns the random value and true if successful, or 0 and false if
// the hardware RNG is not ready (entropy exhausted).
// Check HasRNGSupport() before calling this function.
//   - amd64: RDRAND instruction
//   - arm64: RNDR register (ARMv8.5-A)
//
//export uefiReadRandom
func ReadRandom() (value uint64, ok bool)

// hasRNG returns true if the CPU has a hardware RNG.
//   - amd64: checks CPUID leaf 1 ECX bit 30 (RDRAND)
//   - arm64: checks ID_AA64ISAR0_EL1 bits [63:60] (RNDR)
//
//export uefiHasRNG
func hasRNG() bool

var hasRNGSupport *bool

// HasRNGSupport returns true if the CPU has a hardware RNG.
func HasRNGSupport() bool {
	if hasRNGSupport == nil {
		supported := hasRNG()
		hasRNGSupport = &supported
	}
	return *hasRNGSupport
}
