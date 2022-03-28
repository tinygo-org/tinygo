//go:build fe310
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

//export main
func main() {
	// Zero the PLIC enable bits on startup: they are not zeroed at reset.
	sifive.PLIC.ENABLE[0].Set(0)
	sifive.PLIC.ENABLE[1].Set(0)

	// Zero the threshold value to allow all priorities of interrupts.
	sifive.PLIC.THRESHOLD.Set(0)

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
			id := sifive.PLIC.CLAIM.Get()
			// Call the interrupt handler, if any is registered for this ID.
			sifive.HandleInterrupt(int(id))
			// Complete this interrupt.
			sifive.PLIC.CLAIM.Set(id)
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
	// Configure PLL to output 320MHz.
	//   R=2:  divide 16MHz to 8MHz
	//   F=80: multiply 8MHz by 80 to get 640MHz (80/2-1=39)
	//   Q=2:  divide 640MHz by 2 to get 320MHz
	// This makes the main CPU run at 320MHz.
	sifive.PRCI.PLLCFG.Set(sifive.PRCI_PLLCFG_PLLR_R2<<sifive.PRCI_PLLCFG_PLLR_Pos | 39<<sifive.PRCI_PLLCFG_PLLF_Pos | sifive.PRCI_PLLCFG_PLLQ_Q2<<sifive.PRCI_PLLCFG_PLLQ_Pos | sifive.PRCI_PLLCFG_SEL | sifive.PRCI_PLLCFG_REFSEL)

	// Turn off HFROSC to save power
	sifive.PRCI.HFROSCCFG.ClearBits(sifive.PRCI_HFROSCCFG_ENABLE)

	// Enable the RTC.
	sifive.RTC.RTCCFG.Set(sifive.RTC_RTCCFG_ENALWAYS)

	// Configure the UART.
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
	// For a list of exception codes, see:
	// https://content.riscv.org/wp-content/uploads/2019/08/riscv-privileged-20190608-1.pdf#page=49
	print("fatal error: exception with mcause=")
	print(code)
	print(" pc=")
	print(riscv.MEPC.Get())
	println()
	abort()
}
