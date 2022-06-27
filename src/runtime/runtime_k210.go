//go:build k210
// +build k210

// This file implements target-specific things for the K210 chip as used in the
// MAix Bit with Mic.

package runtime

import (
	"device/kendryte"
	"device/riscv"
	"machine"
	"runtime/volatile"
	"unsafe"
)

type timeUnit int64

//export main
func main() {

	// Both harts should disable all interrupts on startup.
	initPLIC()

	// Only use one hart for the moment.
	if riscv.MHARTID.Get() != 0 {
		abort()
	}

	// Reset all interrupt source priorities to zero.
	for i := 0; i < kendryte.IRQ_max; i++ {
		kendryte.PLIC.PRIORITY[i].Set(0)
	}

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
	exit(0)
}

func initPLIC() {
	hartId := riscv.MHARTID.Get()

	// Zero the PLIC enable bits at startup.
	for i := 0; i < ((kendryte.IRQ_max + 32) / 32); i++ {
		kendryte.PLIC.TARGET_ENABLES[hartId].ENABLE[i].Set(0)
	}

	// Zero the PLIC threshold bits to allow all interrupts.
	kendryte.PLIC.TARGETS[hartId].THRESHOLD.Set(0)
}

//go:extern handleInterruptASM
var handleInterruptASM [0]uintptr

//export handleInterrupt
func handleInterrupt() {
	cause := riscv.MCAUSE.Get()
	code := uint64(cause &^ (1 << 63))
	if cause&(1<<63) != 0 {
		// Topmost bit is set, which means that it is an interrupt.
		switch code {
		case 7: // Machine timer interrupt
			// Signal timeout.
			timerWakeup.Set(1)
			// Disable the timer, to avoid triggering the interrupt right after
			// this interrupt returns.
			riscv.MIE.ClearBits(1 << 7) // MTIE bit
		case 11: // Machine external interrupt
			hartId := riscv.MHARTID.Get()

			// Claim this interrupt.
			id := kendryte.PLIC.TARGETS[hartId].CLAIM.Get()
			// Call the interrupt handler, if any is registered for this ID.
			kendryte.HandleInterrupt(int(id))
			// Complete this interrupt.
			kendryte.PLIC.TARGETS[hartId].CLAIM.Set(id)
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
	// Enable APB0 clock.
	kendryte.SYSCTL.CLK_EN_CENT.SetBits(kendryte.SYSCTL_CLK_EN_CENT_APB0_CLK_EN)

	// Enable FPIOA peripheral.
	kendryte.SYSCTL.CLK_EN_PERI.SetBits(kendryte.SYSCTL_CLK_EN_PERI_FPIOA_CLK_EN)

	machine.InitSerial()
	machine.Serial.Configure(machine.UARTConfig{})
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for machine.Serial.Buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

func buffered() int {
	return machine.Serial.Buffered()
}

var timerWakeup volatile.Register8

func ticks() timeUnit {
	highBits := uint32(kendryte.CLINT.MTIME.Get() >> 32)
	for {
		lowBits := uint32(kendryte.CLINT.MTIME.Get() & 0xffffffff)
		newHighBits := uint32(kendryte.CLINT.MTIME.Get() >> 32)
		if newHighBits == highBits {
			return timeUnit(lowBits) | (timeUnit(highBits) << 32)
		}
		highBits = newHighBits
	}
}

func sleepTicks(d timeUnit) {
	target := uint64(ticks() + d)
	kendryte.CLINT.MTIMECMP[0].Set(target)
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
func handleException(code uint64) {
	// For a list of exception codes, see:
	// https://content.riscv.org/wp-content/uploads/2019/08/riscv-privileged-20190608-1.pdf#page=49
	print("fatal error: exception with mcause=")
	print(code)
	print(" pc=")
	print(riscv.MEPC.Get())
	println()
	abort()
}
