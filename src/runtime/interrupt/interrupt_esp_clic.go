//go:build esp32p4

package interrupt

import (
	"device/riscv"
	"errors"
	"runtime/volatile"
	"unsafe"
)

//go:extern tinygo_saved_ra
var tinygo_saved_ra uintptr

// interruptDepth counts nested interrupt handlers per core. In() uses it
// instead of reading mcause, because on the CLIC mcause must retain the
// pre-interrupt priority so that `mret` can restore the CPU threshold.
var interruptDepth [2]uint32

// In returns whether the system is currently in an interrupt. The CLIC needs
// mcause to stay valid for mret, so use the software nesting count instead.
func In() bool {
	return interruptDepth[currentCPU()&1] != 0
}

// CLIC register access for the ESP32-P4.
// The CLIC is at 0x20800000, with per-interrupt registers at 0x20801000.
// Each interrupt source has 4 bytes: [ip, ie, attr, ctl].
const clicBase = 0x20800000
const clicIntBase = clicBase + 0x1000

// clicExtIntrNumOffset mirrors the CLIC_EXT_INTR_NUM_OFFSET in ESP-IDF: the
// first 16 CLIC sources are internal, so a CPU interrupt number n lives at
// CLIC source n+16 (and mcause reports n+16 on dispatch).
const clicExtIntrNumOffset = 16

// clicSource maps a CPU interrupt number (as used for INTMTX routing and by
// the rest of TinyGo) to its CLIC per-source register index.
func clicSource(n uint32) uint32 {
	return n + clicExtIntrNumOffset
}

// CLIC priority encoding (NLBITS=3 on ESP32-P4): bits [7:5] hold the level and
// bits [4:0] are reserved but must be written as 1 (ESP-IDF NLBITS_TO_BYTE).
const (
	clicNLBITS      = 3
	clicMaxPriority = (1 << clicNLBITS) - 1
	clicLowBits     = (1 << (8 - clicNLBITS)) - 1
)

func clicEncodePriority(level uint8) uint8 {
	if level > clicMaxPriority {
		level = clicMaxPriority
	}
	return level<<(8-clicNLBITS) | clicLowBits
}

func clicSetPriority(n uint32, level uint8) {
	reg := (*volatile.Register8)(unsafe.Pointer(uintptr(clicIntBase + clicSource(n)*4 + 3)))
	reg.Set(clicEncodePriority(level))
}

func clicGetPriority(n uint32) uint8 {
	reg := (*volatile.Register8)(unsafe.Pointer(uintptr(clicIntBase + clicSource(n)*4 + 3)))
	return reg.Get() >> (8 - clicNLBITS)
}

func clicSetEnable(n uint32, enable bool) {
	reg := (*volatile.Register8)(unsafe.Pointer(uintptr(clicIntBase + clicSource(n)*4 + 1)))
	if enable {
		reg.Set(1)
	} else {
		reg.Set(0)
	}
}

func clicClearPending(n uint32) {
	reg := (*volatile.Register8)(unsafe.Pointer(uintptr(clicIntBase + clicSource(n)*4)))
	reg.Set(0)
}

// Enable registers a CPU interrupt.
// The ESP32-P4 CLIC supports up to 48 interrupt sources.
func (i Interrupt) Enable() error {
	if i.num < 1 || i.num > 47 {
		return errors.New("interrupt for ESP32-P4 must be in range of 1 through 47")
	}
	// The CLIC is per core and peripherals are routed to core 0, so core 1
	// asks core 0 to enable the line.
	if currentCPU() != 0 {
		enableOnCore0(uint32(i.num))
		return nil
	}
	enableLocal(uint32(i.num))
	return nil
}

// enableLocal enables the given CPU interrupt in the CLIC of the calling core.
func enableLocal(num uint32) {
	mask := riscv.DisableInterrupts()

	src := clicSource(num)

	// Enable interrupt via CLIC.
	clicSetEnable(num, true)

	// Set rising-edge triggered. TRIG field is at bits[2:1]; rising edge =
	// 0b01 << 1 = 0x02. MODE bits [7:6] are read-only and fixed to 0b11, and
	// SHV (bit 0) stays 0 so dispatch uses the classic MTVEC jump table.
	attrReg := (*volatile.Register8)(unsafe.Pointer(uintptr(clicIntBase + src*4 + 2)))
	attrReg.Set(0xC2) // machine mode, rising-edge triggered, non-vectored

	// Set priority above threshold.
	clicSetPriority(num, defaultThreshold)

	// Clear pending.
	clicClearPending(num)

	riscv.Asm("fence")
	riscv.EnableInterrupts(mask)
}

// Adding pseudo function calls that is replaced by the compiler with the actual
// functions registered through interrupt.New.
//
//go:linkname callHandlers runtime/interrupt.callHandlers
func callHandlers(num int)

//go:linkname signalInterrupt runtime.signalInterrupt
func signalInterrupt()

//go:linkname currentCPU runtime.currentCPU
func currentCPU() uint32

//go:linkname enableOnCore0 runtime.enableInterruptOnCore0
func enableOnCore0(num uint32)

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
	IRQNUM_32
	IRQNUM_33
	IRQNUM_34
	IRQNUM_35
	IRQNUM_36
	IRQNUM_37
	IRQNUM_38
	IRQNUM_39
	IRQNUM_40
	IRQNUM_41
	IRQNUM_42
	IRQNUM_43
	IRQNUM_44
	IRQNUM_45
	IRQNUM_46
	IRQNUM_47
)

const (
	defaultThreshold = 5
	// Priority 0 disables an interrupt.
	disableThreshold = 0
)

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
	case IRQNUM_32:
		callHandlers(IRQNUM_32)
	case IRQNUM_33:
		callHandlers(IRQNUM_33)
	case IRQNUM_34:
		callHandlers(IRQNUM_34)
	case IRQNUM_35:
		callHandlers(IRQNUM_35)
	case IRQNUM_36:
		callHandlers(IRQNUM_36)
	case IRQNUM_37:
		callHandlers(IRQNUM_37)
	case IRQNUM_38:
		callHandlers(IRQNUM_38)
	case IRQNUM_39:
		callHandlers(IRQNUM_39)
	case IRQNUM_40:
		callHandlers(IRQNUM_40)
	case IRQNUM_41:
		callHandlers(IRQNUM_41)
	case IRQNUM_42:
		callHandlers(IRQNUM_42)
	case IRQNUM_43:
		callHandlers(IRQNUM_43)
	case IRQNUM_44:
		callHandlers(IRQNUM_44)
	case IRQNUM_45:
		callHandlers(IRQNUM_45)
	case IRQNUM_46:
		callHandlers(IRQNUM_46)
	case IRQNUM_47:
		callHandlers(IRQNUM_47)
	}
}

//export handleInterrupt
func handleInterrupt() {
	mcause := riscv.MCAUSE.Get()
	exception := mcause&(1<<31) == 0

	// The CLIC reports the interrupt vector as source = cpu_number + 16 in
	// mcause; recover the CPU interrupt number used throughout TinyGo.
	interruptNumber := uint32(mcause&0x3f) - clicExtIntrNumOffset

	if !exception && interruptNumber > 0 {
		// Save MSTATUS & MEPC, which could be overwritten by another CPU interrupt.
		mstatus := riscv.MSTATUS.Get()
		mepc := riscv.MEPC.Get()

		// Temporarily disable this interrupt by lowering its CLIC priority
		// below the threshold.
		thresholdSave := clicGetPriority(interruptNumber)
		clicSetPriority(interruptNumber, disableThreshold)
		riscv.Asm("fence")

		// Clear pending.
		clicClearPending(interruptNumber)

		// Enable CPU interrupts for nesting.
		riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)

		// Call registered interrupt handler(s).
		core := currentCPU() & 1
		interruptDepth[core]++
		callHandler(int(interruptNumber))
		interruptDepth[core]--

		// Signal to sleepTicks that an interrupt has occurred.
		signalInterrupt()

		// Disable CPU interrupts.
		riscv.MSTATUS.ClearBits(riscv.MSTATUS_MIE)

		// Restore interrupt priority to enable interrupt again.
		clicSetPriority(interruptNumber, thresholdSave)
		riscv.Asm("fence")

		// Restore MCAUSE to the value captured at trap entry.  On the CLIC,
		// mcause[23:16] holds the interrupt level that was active before this
		// interrupt, and `mret` restores the CPU threshold from it.  Zeroing
		// mcause (as a non-CLIC target may do) leaves the level stuck at the
		// serviced interrupt's priority, so the next equal-priority interrupt
		// can never dispatch again.
		riscv.MCAUSE.Set(mcause)

		// Restore MSTATUS & MEPC.
		riscv.MSTATUS.Set(mstatus)
		riscv.MEPC.Set(mepc)
	} else {
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
