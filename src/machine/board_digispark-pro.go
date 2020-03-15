// +build digispark_pro

package machine

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 16000000
}

const (
	P0  = PB0
	P1  = PB1
	P2  = PB2
	P3  = PB6
	P4  = PB3
	P5  = PA7
	P6  = PA0
	P7  = PA1
	P8  = PA2
	P9  = PA3
	P10 = PA4
	P11 = PA5
	P12 = PA6

	LED = P1
)
