//go:build avr

package runtime

// AVR prologues push callee-saved registers in ascending order, which stores a
// saved pair high byte first. The scan must also try the byte swapped word.
const gcScanSwapped = true
