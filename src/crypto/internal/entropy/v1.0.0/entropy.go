// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Stripped-down version of the entropy package for TinyGo.
//
// The upstream Go 1.26 package allocates a [1<<25]byte (32 MiB) global buffer
// for CPU jitter-based SP 800-90B entropy collection. This is fine on systems
// with virtual memory where BSS pages are lazily backed, but it causes a fatal
// "RAM overflowed" error on microcontrollers and other memory-constrained
// targets.
//
// Because FIPS 140-3 jitter entropy is never used on TinyGo targets (the DRBG
// falls through to sysrand.Read when fips140.Enabled is false), this overlay
// replaces the 32 MiB buffer with a zero-size type.
package entropy

// Version returns the version of the entropy source.
func Version() string {
	return "v1.0.0"
}

// ScratchBuffer is a large buffer in upstream Go (32 MiB). TinyGo replaces it
// with a zero-size type since the CPU jitter entropy source is not used.
type ScratchBuffer [0]byte

// Seed returns a 384-bit seed with full entropy.
// On TinyGo targets this is never called because FIPS mode is not enabled.
func Seed(memory *ScratchBuffer) ([48]byte, error) {
	panic("entropy: CPU jitter entropy source is not supported on TinyGo targets")
}

// Samples collects entropy samples. Not supported on TinyGo targets.
func Samples(samples []uint8, memory *ScratchBuffer) error {
	panic("entropy: CPU jitter entropy source is not supported on TinyGo targets")
}

// SHA384 computes SHA-384 over 1024 bytes. Not supported on TinyGo targets.
func SHA384(p *[1024]byte) [48]byte {
	panic("entropy: CPU jitter entropy source is not supported on TinyGo targets")
}

// TestingOnlySHA384 computes SHA-384 over arbitrary bytes. Not supported on TinyGo targets.
func TestingOnlySHA384(p []byte) [48]byte {
	panic("entropy: CPU jitter entropy source is not supported on TinyGo targets")
}

// RepetitionCountTest implements the repetition count test from SP 800-90B.
func RepetitionCountTest(samples []uint8) error {
	panic("entropy: CPU jitter entropy source is not supported on TinyGo targets")
}

// AdaptiveProportionTest implements the adaptive proportion test from SP 800-90B.
func AdaptiveProportionTest(samples []uint8) error {
	panic("entropy: CPU jitter entropy source is not supported on TinyGo targets")
}
