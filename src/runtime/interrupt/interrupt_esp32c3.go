//go:build esp32c3

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
//	// map interrupt 5 to my XXXX module
//	esp.INTERRUPT_CORE0.XXXX_INTERRUPT_PRO_MAP.Set( 5 )
//	_ = Interrupt.New(5, func(interrupt.Interrupt) {
//	    ...
//	}).Enable()
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

	// Set default threshold to defaultThreshold
	reg := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.INTERRUPT_CORE0.CPU_INT_PRI_0), i.num*4))
	reg.Set(defaultThreshold)

	// Reset interrupt before reenabling
	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.SetBits(1 << i.num)
	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.ClearBits(1 << i.num)

	// we must wait for any pending write operations to complete
	riscv.Asm("fence")
	return nil
}

// Adding pseudo function calls that is replaced by the compiler with the actual
// functions registered through interrupt.New.
//
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

const (
	defaultThreshold = 5
	disableThreshold = 10
)

//go:inline
func callHandler(n int) {
	switch n {
	case IRQNUM_1:
		callHandlers(IRQNUM_1)
	case IRQNUM_2:
		callHandlers(IRQNUM_2)
	case IRQNUM_3:
		callHandlers(IRQNUM_3)
	case IRQNUM_4:
		callHandlers(IRQNUM_4)
	case IRQNUM_5:
		callHandlers(IRQNUM_5)
	case IRQNUM_6:
		callHandlers(IRQNUM_6)
	case IRQNUM_7:
		callHandlers(IRQNUM_7)
	case IRQNUM_8:
		callHandlers(IRQNUM_8)
	case IRQNUM_9:
		callHandlers(IRQNUM_9)
	case IRQNUM_10:
		callHandlers(IRQNUM_10)
	case IRQNUM_11:
		callHandlers(IRQNUM_11)
	case IRQNUM_12:
		callHandlers(IRQNUM_12)
	case IRQNUM_13:
		callHandlers(IRQNUM_13)
	case IRQNUM_14:
		callHandlers(IRQNUM_14)
	case IRQNUM_15:
		callHandlers(IRQNUM_15)
	case IRQNUM_16:
		callHandlers(IRQNUM_16)
	case IRQNUM_17:
		callHandlers(IRQNUM_17)
	case IRQNUM_18:
		callHandlers(IRQNUM_18)
	case IRQNUM_19:
		callHandlers(IRQNUM_19)
	case IRQNUM_20:
		callHandlers(IRQNUM_20)
	case IRQNUM_21:
		callHandlers(IRQNUM_21)
	case IRQNUM_22:
		callHandlers(IRQNUM_22)
	case IRQNUM_23:
		callHandlers(IRQNUM_23)
	case IRQNUM_24:
		callHandlers(IRQNUM_24)
	case IRQNUM_25:
		callHandlers(IRQNUM_25)
	case IRQNUM_26:
		callHandlers(IRQNUM_26)
	case IRQNUM_27:
		callHandlers(IRQNUM_27)
	case IRQNUM_28:
		callHandlers(IRQNUM_28)
	case IRQNUM_29:
		callHandlers(IRQNUM_29)
	case IRQNUM_30:
		callHandlers(IRQNUM_30)
	case IRQNUM_31:
		callHandlers(IRQNUM_31)
	}
}

//export handleInterrupt
func handleInterrupt() {
	mcause := riscv.MCAUSE.Get()
	exception := mcause&(1<<31) == 0
	interruptNumber := uint32(mcause & 0x1f)

	if !exception && interruptNumber > 0 {
		// save MSTATUS & MEPC, which could be overwritten by another CPU interrupt
		mstatus := riscv.MSTATUS.Get()
		mepc := riscv.MEPC.Get()
		// Useing threshold to temporary disable this interrupts.
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
		riscv.MSTATUS.SetBits(1 << 3)

		// Call registered interrupt handler(s)
		callHandler(int(interruptNumber))

		// disable CPU interrupts
		riscv.MSTATUS.ClearBits(1 << 3)

		// restore interrupt threshold to enable interrupt again
		reg.Set(thresholdSave)
		riscv.Asm("fence")

		// restore MSTATUS & MEPC
		riscv.MSTATUS.Set(mstatus)
		riscv.MEPC.Set(mepc)

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
	switch uint32(mcause & 0x1f) {
	case 1:
		println("***    virtual addess:", riscv.MTVAL.Get())
	case 2:
		println("***            opcode:", riscv.MTVAL.Get())
	case 5:
		println("***      read address:", riscv.MTVAL.Get())
	case 7:
		println("***     write address:", riscv.MTVAL.Get())
	}
	for {
		riscv.Asm("wfi")
	}
}
