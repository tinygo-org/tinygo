//go:build esp32c3

package interrupt

import (
	"device/esp"
	"device/riscv"
	"runtime/volatile"
	"unsafe"
)

//go:extern tinygo_saved_ra
var tinygo_saved_ra uintptr

// enableLine enables CPU interrupt line num as a rising edge interrupt with
// the default priority.
func enableLine(num int) {
	mask := riscv.DisableInterrupts()

	// enable CPU interrupt number num
	esp.INTERRUPT_CORE0.CPU_INT_ENABLE.SetBits(1 << num)

	// Set pulse interrupt type (rising edge detection)
	esp.INTERRUPT_CORE0.CPU_INT_TYPE.SetBits(1 << num)

	// Set default threshold to defaultThreshold
	reg := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.INTERRUPT_CORE0.CPU_INT_PRI_0), num*4))
	reg.Set(defaultThreshold)

	// Reset interrupt before reenabling
	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.SetBits(1 << num)
	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.ClearBits(1 << num)

	// we must wait for any pending write operations to complete
	riscv.Asm("fence")
	riscv.EnableInterrupts(mask)
}

const (
	// First / last allocatable CPU interrupt for peripherals.
	firstCPUInt = 1
	lastCPUInt  = 31
)

const (
	defaultThreshold = 5
	// Priority 0 disables an interrupt on ESP32-C3 (per the TRM,
	// priority 0 = masked).  During handling we set the current
	// interrupt's priority to 0 so it cannot re-fire while MIE is
	// re-enabled for nesting.  This is critical for level-triggered
	// interrupts where the hardware line stays asserted until the
	// peripheral source is serviced by the handler.
	disableThreshold = 0
)

//export handleInterrupt
func handleInterrupt() {
	mcause := riscv.MCAUSE.Get()
	exception := mcause&(1<<31) == 0
	interruptNumber := uint32(mcause & 0x1f)

	if !exception && interruptNumber > 0 {
		// save MSTATUS & MEPC, which could be overwritten by another CPU interrupt
		mstatus := riscv.MSTATUS.Get()
		mepc := riscv.MEPC.Get()
		// Using threshold to temporary disable this interrupts.
		// FYI: using CPU interrupt enable bit make runtime to loose interrupts.
		reg := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.INTERRUPT_CORE0.CPU_INT_PRI_0), interruptNumber*4))
		thresholdSave := reg.Get()
		reg.Set(disableThreshold)
		riscv.Asm("fence")

		interruptBit := uint32(1 << interruptNumber)

		// reset pending status interrupt
		if esp.INTERRUPT_CORE0.CPU_INT_TYPE.Get()&interruptBit != 0 {
			// this is edge type interrupt
			esp.INTERRUPT_CORE0.CPU_INT_CLEAR.SetBits(interruptBit)
			esp.INTERRUPT_CORE0.CPU_INT_CLEAR.ClearBits(interruptBit)
		} else {
			// this is level type interrupt
			esp.INTERRUPT_CORE0.CPU_INT_CLEAR.ClearBits(interruptBit)
		}

		// enable CPU interrupts
		riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)

		// Call registered interrupt handler(s)
		callHandler(int(interruptNumber))

		// Signal to sleepTicks that an interrupt has occurred.
		signalInterrupt()

		// disable CPU interrupts
		riscv.MSTATUS.ClearBits(riscv.MSTATUS_MIE)

		// restore interrupt threshold to enable interrupt again
		reg.Set(thresholdSave)
		riscv.Asm("fence")

		// Zero MCAUSE so that interrupt.In() returns false once we
		// return to normal (non-interrupt) code.  Other RISC-V targets
		// (FE310, K210) do the same.
		riscv.MCAUSE.Set(0)

		// restore MSTATUS & MEPC
		riscv.MSTATUS.Set(mstatus)
		riscv.MEPC.Set(mepc)

		// do not enable CPU interrupts now
		// the 'MRET' in src/device/riscv/handleinterrupt.S will copies the state of MPIE back into MIE, and subsequently clears MPIE.
		// riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)
	} else {
		// Topmost bit is clear, so it is an exception of some sort.
		// We could implement support for unsupported instructions here (such as
		// misaligned loads). However, for now we'll just print a fatal error.
		handleException(mcause)
	}
}

func handleException(mcause uintptr) {
	println("*** Exception:     pc:", riscv.MEPC.Get())
	println("*** Exception:   code:", uint32(mcause&0x1f))
	println("*** Exception: mcause:", mcause)
	println("*** Exception:     ra:", tinygo_saved_ra)
	switch uint32(mcause & 0x1f) {
	case riscv.InstructionAccessFault:
		println("***    virtual address:", riscv.MTVAL.Get())
	case riscv.IllegalInstruction:
		println("***            opcode:", riscv.MTVAL.Get())
	case riscv.LoadAccessFault:
		println("***      read address:", riscv.MTVAL.Get())
	case riscv.StoreOrAMOAccessFault:
		println("***     write address:", riscv.MTVAL.Get())
	}
	for {
		riscv.Asm("wfi")
	}
}
