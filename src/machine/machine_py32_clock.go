//go:build py32

package machine

var CPUFrequencyHz uint32

func CPUFrequency() uint32 {
	return CPUFrequencyHz
}

// SetCPUFrequency sets the CPU frequency in hertz.
// Called by runtime.main() with the default frequency, and can be called by user code to change the frequency at runtime.
// Note that peripherals may need re-initialization after a frequency change:
// - runtime.ConfigureSystemTimer() for the system timer
// - machine.DefaultUART.Configure() if UART is in use
func SetCPUFrequency(frequency uint32) {
	CPUFrequencyHz = frequency
}
