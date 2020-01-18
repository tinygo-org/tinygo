// +build fe310

// This file implements target-specific things for the FE310 chip as used in the
// HiFive1.

package runtime

import (
	"machine"
	"unsafe"

	"device/riscv"
	"device/sifive"
	"runtime/volatile"
)

type timeUnit int64

//go:extern _sbss
var _sbss unsafe.Pointer

//go:extern _ebss
var _ebss unsafe.Pointer

//go:extern _sdata
var _sdata unsafe.Pointer

//go:extern _sidata
var _sidata unsafe.Pointer

//go:extern _edata
var _edata unsafe.Pointer

//go:export main
func main() {
	// Zero the PLIC enable bits on startup: they are not zeroed at reset.
	sifive.PLIC.ENABLE[0].Set(0)
	sifive.PLIC.ENABLE[1].Set(0)

	// Set the interrupt address.
	// Note that this address must be aligned specially, otherwise the MODE bits
	// of MTVEC won't be zero.
	riscv.MTVEC.Set(uintptr(unsafe.Pointer(&handleInterruptASM)))

	// Enable global interrupts now that they've been set up.
	riscv.MSTATUS.SetBits(1 << 3) // MIE

	preinit()
	initAll()
	callMain()
	abort()
}

//go:extern handleInterruptASM
var handleInterruptASM [0]uintptr

//export handleInterrupt
func handleInterrupt() {
	cause := riscv.MCAUSE.Get()
	code := uint(cause &^ (1 << 31))
	if cause&(1<<31) != 0 {
		// Topmost bit is set, which means that it is an interrupt.
		switch code {
		case 7: // Machine timer interrupt
			// Signal timeout.
			timerWakeup.Set(1)
			// Disable the timer, to avoid triggering the interrupt right after
			// this interrupt returns.
			riscv.MIE.ClearBits(1 << 7) // MTIE bit
		}
	} else {
		// Topmost bit is clear, so it is an exception of some sort.
		// We could implement support for unsupported instructions here (such as
		// misaligned loads). However, for now we'll just print a fatal error.
		handleException(code)
	}
}

func init() {
	pric_init()
	machine.UART0.Configure(machine.UARTConfig{})
}

func pric_init() {
	// Make sure the HFROSC is on
	sifive.PRCI.HFROSCCFG.SetBits(sifive.PRCI_HFROSCCFG_ENABLE)

	// Run off 16 MHz Crystal for accuracy.
	sifive.PRCI.PLLCFG.SetBits(sifive.PRCI_PLLCFG_REFSEL | sifive.PRCI_PLLCFG_BYPASS)
	sifive.PRCI.PLLCFG.SetBits(sifive.PRCI_PLLCFG_SEL)

	// Turn off HFROSC to save power
	sifive.PRCI.HFROSCCFG.ClearBits(sifive.PRCI_HFROSCCFG_ENABLE)

	// Enable the RTC.
	sifive.RTC.RTCCFG.Set(sifive.RTC_RTCCFG_ENALWAYS)
}

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 4)
	}

	// Initialize .data: global variables initialized from flash.
	src := unsafe.Pointer(&_sidata)
	dst := unsafe.Pointer(&_sdata)
	for dst != unsafe.Pointer(&_edata) {
		*(*uint32)(dst) = *(*uint32)(src)
		dst = unsafe.Pointer(uintptr(dst) + 4)
		src = unsafe.Pointer(uintptr(src) + 4)
	}
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

const asyncScheduler = false

var timerWakeup volatile.Register8

func ticks() timeUnit {
	// Combining the low bits and the high bits yields a time span of over 270
	// years without counter rollover.
	highBits := sifive.CLINT.MTIMEH.Get()
	for {
		lowBits := sifive.CLINT.MTIME.Get()
		newHighBits := sifive.CLINT.MTIMEH.Get()
		if newHighBits == highBits {
			// High bits stayed the same.
			return timeUnit(lowBits) | (timeUnit(highBits) << 32)
		}
		// Retry, because there was a rollover in the low bits (happening every
		// 1.5 days).
		highBits = newHighBits
	}
}

func sleepTicks(d timeUnit) {
	target := uint64(ticks() + d)
	sifive.CLINT.MTIMECMPH.Set(uint32(target >> 32))
	sifive.CLINT.MTIMECMP.Set(uint32(target))
	riscv.MIE.SetBits(1 << 7) // MTIE
	for {
		if timerWakeup.Get() != 0 {
			timerWakeup.Set(0)
			// Disable timer.
			break
		}
		riscv.Asm("wfi")
	}
}

// handleException is called from the interrupt handler for any exception.
// Exceptions can be things like illegal instructions, invalid memory
// read/write, and similar issues.
func handleException(code uint) {
	if faultHandler != nil {
		faultHandler(FaultInfo{
			PC:   riscv.MEPC.Get(),
			SP:   0,
			Code: int(code),
		})
	}
	// For a list of exception codes, see:
	// https://content.riscv.org/wp-content/uploads/2019/08/riscv-privileged-20190608-1.pdf#page=49
	print("fatal error: exception with mcause=")
	print(code)
	print(" pc=")
	print(riscv.MEPC.Get())
	println()
	abort()
}
