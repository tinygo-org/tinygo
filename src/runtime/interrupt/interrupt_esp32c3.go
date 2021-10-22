// +build esp32c3

package interrupt

import (
	"device/esp"
	"device/riscv"
	"errors"
	"runtime/volatile"
	"unsafe"
)

// Enable register CPU interrupt with interrupt.Interrupt.
// The ESP32-C3 has 31 CPU independent interrupts.
// The Interrupt.New(x, f) (x = [1..31]) attaches CPU interrupt to function f.
// Caller must map the selected interrupt using following sequence (for example using id 5):
//
//    // map interrupt 5 to my XXXX module
//    esp.INTERRUPT_CORE0.XXXX_INTERRUPT_PRO_MAP.Set( 5 )
//    _ = Interrupt.New(5, func(interrupt.Interrupt) {
//        ...
//    }).Enable()
func (i Interrupt) Enable() error {
	if i.num < 1 && i.num > 31 {
		return errors.New("interrupt for ESP32-C3 must be in range of 1 through 31")
	}
	mask := riscv.DisableInterrupts()
	defer riscv.EnableInterrupts(mask)

	// enable CPU interrupt number i.num
	esp.INTERRUPT_CORE0.CPU_INT_ENABLE.SetBits(1 << i.num)

	// Set pulse interrupt type (rising edge detection)
	esp.INTERRUPT_CORE0.CPU_INT_TYPE.SetBits(1 << i.num)

	// Set default threshold to 5
	reg := (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.INTERRUPT_CORE0.CPU_INT_PRI_0)) + uintptr(i.num)*4)))
	reg.Set(5)

	// Reset interrupt before reenabling
	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.SetBits(1 << i.num)
	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.ClearBits(1 << i.num)

	// we must wait for any pending write operations to complete
	riscv.Asm("fence")
	return nil
}

//export handleInterrupt
func handleInterrupt() {
	mcause := riscv.MCAUSE.Get()
	exception := mcause&(1<<31) == 0
	interruptNumber := uint32(mcause & 0x1f)

	if !exception && interruptNumber > 0 {
		// save mepc, which could be overwritten by another CPU interrupt
		mepc := riscv.MEPC.Get()

		// disable interrupt
		interruptBit := uint32(1 << interruptNumber)
		esp.INTERRUPT_CORE0.CPU_INT_ENABLE.ClearBits(interruptBit)

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
		riscv.MSTATUS.SetBits(0x8)

		// Call registered interrupt handler(s)
		callInterruptHandler(int(interruptNumber))

		// disable CPU interrupts
		riscv.MSTATUS.ClearBits(0x8)

		// mpie must be set to 1 to resume interrupts after 'MRET'
		riscv.MSTATUS.SetBits(0x80)

		// restore MEPC
		riscv.MEPC.Set(mepc)

		// enable this interrupt
		esp.INTERRUPT_CORE0.CPU_INT_ENABLE.SetBits(interruptBit)

		// do not enable CPU interrupts now
		// the 'MRET' in src/device/riscv/handleinterrupt.S will copies the state of MPIE back into MIE, and subsequently clears MPIE.
		// riscv.MSTATUS.SetBits(0x8)
	} else {
		// Topmost bit is clear, so it is an exception of some sort.
		// We could implement support for unsupported instructions here (such as
		// misaligned loads). However, for now we'll just print a fatal error.
		handleException(mcause)
	}
}

func handleException(mcause uintptr) {
	// TODO need to get location of actual MEPC from the stack stash created in src/device/riscv/handleinterrupt.S
	println("*** Exception:     pc:", riscv.MEPC.Get())
	println("*** Exception:   code:", uint32(mcause&0x1f))
	println("*** Exception: mcause:", mcause)
	for {
		riscv.Asm("wfi")
	}
}

// callInterruptHandler is a compiler-generated function that calls the
// appropriate interrupt handler for the given interrupt ID.
//go:linkname callInterruptHandler runtime.callInterruptHandler
func callInterruptHandler(id int)
