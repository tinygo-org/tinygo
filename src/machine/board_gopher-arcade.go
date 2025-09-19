//go:build gopher_arcade

package machine

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 8000000
}

const (
	P5 Pin = PB0
	P6 Pin = PB1
	P7 Pin = PB2
	P2 Pin = PB3
	P3 Pin = PB4
	P1 Pin = PB5

	LED          = P1
	BUTTON_LEFT  = P7
	BUTTON_RIGHT = P5
	SPEAKER      = P6
)
