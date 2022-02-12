//go:build esp32c3
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

// Adding pseudo function calls that is replaced by the compiler with the actual
// functions registered through interrupt.New.
//go:linkname callHandlers runtime/interrupt.callHandlers
func callHandlers(num int)

const (
	IRQNUM_1 = 1 + iota
	IRQNUM_2
	IRQNUM_3
	IRQNUM_4
	IRQNUM_5
	IRQNUM_6
	IRQNUM_7
	IRQNUM_8
	IRQNUM_9
	IRQNUM_10
	IRQNUM_11
	IRQNUM_12
	IRQNUM_13
	IRQNUM_14
	IRQNUM_15
	IRQNUM_16
	IRQNUM_17
	IRQNUM_18
	IRQNUM_19
	IRQNUM_20
	IRQNUM_21
	IRQNUM_22
	IRQNUM_23
	IRQNUM_24
	IRQNUM_25
	IRQNUM_26
	IRQNUM_27
	IRQNUM_28
	IRQNUM_29
	IRQNUM_30
	IRQNUM_31
)

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
		switch interruptNumber {
		// case IRQNUM_1:
		// 	callHandlers(IRQNUM_1)
		// case 2:
		// 	callHandlers(IRQNUM_2)
		// case 3:
		// 	callHandlers(IRQNUM_3)
		// case 4:
		// 	callHandlers(IRQNUM_4)
		// case IRQNUM_5:
		// 	callHandlers(IRQNUM_5)
		case IRQNUM_6:
			callHandlers(IRQNUM_6)
			// case 7:
			// 	callHandlers(IRQNUM_7)
			// case 8:
			// 	callHandlers(IRQNUM_8)
			// case 9:
			// 	callHandlers(IRQNUM_9)
			// case 10:
			// 	callHandlers(IRQNUM_10)
			// case 11:
			// 	callHandlers(IRQNUM_11)
			// case 12:
			// 	callHandlers(IRQNUM_12)
			// case 13:
			// 	callHandlers(IRQNUM_13)
			// case 14:
			// 	callHandlers(IRQNUM_14)
			// case 15:
			// 	callHandlers(IRQNUM_15)
			// case 16:
			// 	callHandlers(IRQNUM_16)
			// case 17:
			// 	callHandlers(IRQNUM_17)
			// case 18:
			// 	callHandlers(IRQNUM_18)
			// case 19:
			// 	callHandlers(IRQNUM_19)
			// case 20:
			// 	callHandlers(IRQNUM_20)
			// case 21:
			// 	callHandlers(IRQNUM_21)
			// case 22:
			// 	callHandlers(IRQNUM_22)
			// case 23:
			// 	callHandlers(IRQNUM_23)
			// case 24:
			// 	callHandlers(IRQNUM_24)
			// case 25:
			// 	callHandlers(IRQNUM_25)
			// case 26:
			// 	callHandlers(IRQNUM_26)
			// case 27:
			// 	callHandlers(IRQNUM_27)
			// case 28:
			// 	callHandlers(IRQNUM_28)
			// case 29:
			// 	callHandlers(IRQNUM_29)
			// case 30:
			// 	callHandlers(IRQNUM_30)
			// case IRQNUM_31:
			// 	callHandlers(IRQNUM_31)
		}

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
	println("*** Exception:     pc:", riscv.MEPC.Get())
	println("*** Exception:   code:", uint32(mcause&0x1f))
	println("*** Exception: mcause:", mcause)
	for {
		riscv.Asm("wfi")
	}
}
