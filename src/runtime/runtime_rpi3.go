// +build rpi3
package runtime

import (
	dev "device/arm64"
	"machine"
	"runtime/volatile"
	"unsafe"
)

const tickMicros = int64(1)

type timeUnit int64

//go:export sleepticks sleepticks
func sleepTicks(n timeUnit) {
	WaitMuSec(uint32(n))
}

func ticks() timeUnit {
	return timeUnit(SysTimer())
}

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_start_bss)
	for ptr != unsafe.Pointer(&_end_bss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 4)
	}

	machine.MiniUART.Configure(machine.UARTConfig{})
}

// main on bare metal (bootloader uses this path)
//go:export main
func main() {
	preinit()
	primaryMain()
}

func primaryMain() {
	initAll()
	callMain()
	dev.Abort()
}

//go:extern
func putchar(c byte) {
	//MiniUARTSend(c) if you prefer the mini uart
	machine.MiniUART.WriteByte(c)
}

// wait a given number of CPU cycles (at least)
func WaitCycles(n int) {
	for n > 0 {
		dev.Asm("nop")
		n--
	}
}

//
// Wait MuSec waits for at least n musecs based on the system timer.
//
func WaitMuSec(n uint32) {
	var f, t, r uint64
	dev.AsmFull(`mrs x28, cntfrq_el0
		str x28,{f}
		mrs x27, cntpct_el0
		str x27,{t}`, map[string]interface{}{"f": &f, "t": &t})
	//expires at t
	t += ((f / 1000) * uint64(n)) / 1000
	for r < t {
		dev.AsmFull(`mrs x27, cntpct_el0
			str x27,{r}`, map[string]interface{}{"r": &r})
	}
}

//
// SysTimer gets the 64 bit timer's value.
//
func SysTimer() uint64 {
	h := uint32(0xffffffff)
	var l uint32

	// the reads from MMIO are two separate 32 bit reads
	h = volatile.LoadUint32((*uint32)(dev.SYSTMR_HI))
	l = volatile.LoadUint32((*uint32)(dev.SYSTMR_LO))
	//the read of hi can fail
	again := volatile.LoadUint32((*uint32)(dev.SYSTMR_HI))
	if h != again {
		h = volatile.LoadUint32((*uint32)(dev.SYSTMR_HI))
		l = volatile.LoadUint32((*uint32)(dev.SYSTMR_LO))
	}
	high := uint64(h << 32)
	return high | uint64(l)
}

// abort is called by panic().
func abort() {
	dev.Abort()
}

// asyncScheduler I *think* causes us to get the coroutine scheduler.
const asyncScheduler = false

//go:extern _start_bss
var _start_bss unsafe.Pointer

//go:extern _end_bss
var _end_bss unsafe.Pointer

//go:extern _heap_end
var _heap_end = unsafe.Pointer(uintptr(0x800000))

//go:extern _heap_start
var _heap_start = unsafe.Pointer(uintptr(0x400000))

//go:extern __bss_size
var __bss_size unsafe.Pointer
