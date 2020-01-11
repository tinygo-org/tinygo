// +build digispark

package machine

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 16000000
}

const (
	P0 Pin = 0
	P1 Pin = 1
	P2 Pin = 2
	P3 Pin = 3
	P4 Pin = 4
	P5 Pin = 5

	LED = P1
)
