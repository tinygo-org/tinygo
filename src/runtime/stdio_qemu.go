//go:build (cortexm && qemu) || (tinygo.riscv && virt)

package runtime

import (
	"runtime/volatile"
	"unsafe"
)

// UART0 registers, which QEMU connects to its stdin and stdout with -nographic.
// See https://github.com/qemu/qemu/blob/master/hw/char/pl011.c and hw/char/serial.c
var (
	pl011Data         = (*volatile.Register8)(unsafe.Pointer(uintptr(0x4000c000)))
	pl011Flag         = (*volatile.Register32)(unsafe.Pointer(uintptr(0x4000c018)))
	ns16550Data       = (*volatile.Register8)(unsafe.Pointer(uintptr(0x10000000)))
	ns16550LineStatus = (*volatile.Register8)(unsafe.Pointer(uintptr(0x10000005)))
)

func putchar(c byte) {
	if qemuPL011 {
		pl011Data.Set(c)
	} else {
		ns16550Data.Set(c)
	}
}

func getchar() byte {
	for !uartReceived() {
		Gosched()
	}
	if qemuPL011 {
		return pl011Data.Get()
	}
	return ns16550Data.Get()
}

func buffered() int {
	if uartReceived() {
		return 1
	}
	return 0
}

// uartReceived reports whether a byte is waiting. The PL011 clears UARTFR
// bit 4 (RXFE) and the NS16550A sets line status bit 0 (data ready).
func uartReceived() bool {
	if qemuPL011 {
		return pl011Flag.Get()&(1<<4) == 0
	}
	return ns16550LineStatus.Get()&1 != 0
}
