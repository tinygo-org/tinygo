//go:build (cortexm && qemu) || (tinygo.riscv && virt)

package runtime

// stdoutWrite is the address of the emulated UART data register on each machine.

func putchar(c byte) {
	stdoutWrite.Set(uint8(c))
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}
