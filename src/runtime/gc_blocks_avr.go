//go:build (gc.conservative || gc.precise) && avr

package runtime

// gcMask is the word type of a bitmap.
// It is intended to match the architecture's GPR width.
// AVR uses 8-bit GPRs.
type gcMask = uint8
