// +build rpi3

package runtime

import (
	dev "device/rpi3"
	"runtime/volatile"
	"unsafe"
)

//const GOOS = "linux"
const tickMicros = int64(1)
const asyncScheduler = false

type timeUnit int64

var z timeUnit
var x byte

//go:export sleepticks sleepticks
func sleepTicks(n timeUnit) {
	dev.WaitMuSec(uint32(n))
}

func ticks() timeUnit {
	return timeUnit(dev.SysTimer())
}

func preinit() {

	dev.UART0Init()
	//MinuUARTInit() if you prefer the MiniUart

	// XXX what should we do for a bootloader?
	heapStart := 0x90000
	heapEnd = 0xAFFF8
	heapptr = uintptr(heapStart)
	globalsStart = 0xB0000
	globalsEnd = 0xB0FF8

}

// main on non bootloader (the bootloader itself uses this path)
//go:export main
func main() {
	preinit()
	primaryMain()
}

func primaryMain() {
	initAll()
	callMain()
	abort()
}

//go:export mainFromBootloader
func mainFromBootloader() {
	heapStart := dev.ReadRegister("x2")
	heapEnd = heapStart + 0x8000
	heapptr = uintptr(heapStart)
	// not sure what to set these two to
	//globalsStart = 0xB0000
	//globalsEnd = 0xB0FF8
	t := dev.ReadRegister("x3")
	dev.SetStartTime(uint32(t))
	primaryMain()
}

//go:extern
func putchar(c byte) {
	c++
	c--
	//MiniUARTSend(c) if you prefer the mini uart
	dev.UART0Send(c)
}

// just send to device code which ends up calling WFE
func abort() {
	dev.Abort()
}

// Implement memset for LLVM and compiler-rt.
//go:export memset
func libc_memset(ptr unsafe.Pointer, c byte, size uintptr) {
	for i := uintptr(0); i < size; i++ {
		*(*byte)(unsafe.Pointer(uintptr(ptr) + i)) = c
	}
}

// Implement memmove for LLVM and compiler-rt.
//go:export memmove
func libc_memmove(dst, src unsafe.Pointer, size uintptr) {
	memmove(dst, src, size)
}

// wait a given number of CPU cycles (at least)
func WaitCycles(n int) {
	for n > 0 {
		dev.Asm("nop")
		n--
	}
}

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

func SysTimer() uint64 {
	h := uint32(0xffffffff)
	var l uint32

	// the reads from MMIO are are two separate 32 bit reads
	h = volatile.LoadUint32((*uint32)(dev.SYSTMR_HI))
	l = volatile.LoadUint32((*uint32)(dev.SYSTMR_LO))
	//the read of h can fail
	again := volatile.LoadUint32((*uint32)(dev.SYSTMR_HI))
	if h != again {
		h = volatile.LoadUint32((*uint32)(dev.SYSTMR_HI))
		l = volatile.LoadUint32((*uint32)(dev.SYSTMR_LO))
	}
	high := uint64(h << 32)
	return high | uint64(l)
}

// just send to device coode which ends up calling WFE
func abort() {
	dev.Abort()
}
