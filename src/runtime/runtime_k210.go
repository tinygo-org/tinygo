// +build k210

// This file implements target-specific things for the K210 chip as used in the
// MAix Bit with Mic.

package runtime

import (
	"unsafe"

	"device/riscv"
	"runtime/volatile"
)

type timeUnit int64

func postinit() {}

//export main
func main() {
	// todo

	// Set the interrupt address.
	// Note that this address must be aligned specially, otherwise the MODE bits
	// of MTVEC won't be zero.
	riscv.MTVEC.Set(uintptr(unsafe.Pointer(&handleInterruptASM)))

	// Reset the MIE register and enable external interrupts.
	// It must be reset here because it not zeroed at startup.
	riscv.MIE.Set(1 << 11) // bit 11 is for machine external interrupts

	// Enable global interrupts now that they've been set up.
	riscv.MSTATUS.SetBits(1 << 3) // MIE

	preinit()
	initPeripherals()
	run()
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
		case 11: // Machine external interrupt
			// Claim this interrupt.
			//id := sifive.PLIC.CLAIM.Get()
			// Call the interrupt handler, if any is registered for this ID.
			callInterruptHandler(int(0))
			// Complete this interrupt.
			//sifive.PLIC.CLAIM.Set(id)
		}
	} else {
		// Topmost bit is clear, so it is an exception of some sort.
		// We could implement support for unsupported instructions here (such as
		// misaligned loads). However, for now we'll just print a fatal error.
		handleException(code)
	}
}

// initPeripherals configures periperhals the way the runtime expects them.
func initPeripherals() {
	// todo
}

func putchar(c byte) {
	//machine.UART0.WriteByte(c)
}

const asyncScheduler = false

var timerWakeup volatile.Register8

func ticks() timeUnit {
	// todo
}

func sleepTicks(d timeUnit) {
	// todo
}

// handleException is called from the interrupt handler for any exception.
// Exceptions can be things like illegal instructions, invalid memory
// read/write, and similar issues.
func handleException(code uint) {
	// For a list of exception codes, see:
	// https://content.riscv.org/wp-content/uploads/2019/08/riscv-privileged-20190608-1.pdf#page=49
	print("fatal error: exception with mcause=")
	print(code)
	print(" pc=")
	print(riscv.MEPC.Get())
	println()
	abort()
}

// callInterruptHandler is a compiler-generated function that calls the
// appropriate interrupt handler for the given interrupt ID.
func callInterruptHandler(id int)
