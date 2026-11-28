//go:build (cortexm && qemu) || (tinygo.riscv && virt)

package runtime

// Each machine defines uartData and uartReceived for its UART0, which QEMU
// connects to its own stdin and stdout with -nographic.

func putchar(c byte) {
	uartData.Set(uint8(c))
}

func getchar() byte {
	for !uartReceived() {
		Gosched()
	}
	return uartData.Get()
}

func buffered() int {
	if uartReceived() {
		return 1
	}
	return 0
}
