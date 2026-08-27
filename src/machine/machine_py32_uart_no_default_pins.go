//go:build py32 && !py32_default_uart_pins

package machine

func defaultUARTPins() (Pin, Pin) {
	return NoPin, NoPin
}
