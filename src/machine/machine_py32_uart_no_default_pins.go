//go:build py32 && !default_uart_pins

package machine

func configureDefaultUARTPins() {
	// There are no default UART pins for this board. Configure them in
	// application main() using ConfigureUARTPin().
}
