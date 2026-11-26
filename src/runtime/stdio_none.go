//go:build (avr && attiny) || avrtiny || (arm7tdmi && !mgbadebug)

package runtime

// These targets have no UART, so console output is dropped and no input arrives.

func putchar(c byte) {
}

func getchar() byte {
	return 0
}

func buffered() int {
	return 0
}
