//go:build py32

package machine

// This variable must be updated after any change to the system clock from user code.
// Don't forget to re-initialize peripherals that depend on the system clock, such as:
// - runtime.ConfigureSystemTimer() for the system timer
// - machine.DefaultUART.Configure() if UART is in use
var CPUFrequencyHz uint32 = 24_000_000

func CPUFrequency() uint32 {
	return CPUFrequencyHz
}
