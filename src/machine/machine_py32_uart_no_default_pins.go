//go:build py32 && !default_uart_pins

package machine

func defaultUARTPins() (Pin, Pin) {
	return NoPin, NoPin
}
