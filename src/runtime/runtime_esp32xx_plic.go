//go:build esp32c6 || esp32h2

package runtime

import (
	"device/esp"
	"device/riscv"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

func init() {
	machine.InitSerial()
}

func abort() {
	for {
		riscv.Asm("wfi")
	}
}

// interruptInit initializes the interrupt controller.
func interruptInit() {
	riscv.DisableInterrupts()

	// Reset all interrupt priorities to zero in the PLIC.
	for i := 1; i < 32; i++ {
		plic.MXINT_PRI[i].Set(0)
	}

	// Default threshold for interrupts is 5.
	plic.MXINT_THRESH.Set(5)

	// Set the interrupt address.
	// Set MODE field to 1 - a vector base address.
	// Note that this address must be aligned to 256 bytes.
	riscv.MTVEC.Set((uintptr(unsafe.Pointer(&_vector_table))) | 1)

	// The mie CSR gates each PLIC interrupt line, so enable all lines.
	// esp-hal does the same with `csrw mie, 0xffffffff`.
	riscv.MIE.Set(0xffffffff)

	// The ROM can leave MSTATUS.MIE cleared, so set it here.
	// Without it, no interrupt gets to the CPU.
	riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)
}

// CPU interrupt number used for the TIMG0 timer alarm.
const timerAlarmCPUInterrupt = 9

var interruptPending volatile.Register8

func signalInterrupt() {
	interruptPending.Set(1)
}

// initTimerInterrupt routes the TIMG0 timer 0 alarm interrupt to a CPU
// interrupt and registers a handler.
func initTimerInterrupt() {
	// Map the TIMG0 T0 peripheral interrupt to a CPU interrupt line.
	esp.INTERRUPT_CORE0.TG0_T0_INTR_MAP.Set(timerAlarmCPUInterrupt)

	// Enable T0 interrupt at the timer group level.
	esp.TIMG0.INT_ENA_TIMERS.SetBits(1)

	// Register the interrupt handler.
	interrupt.New(timerAlarmCPUInterrupt, func(interrupt.Interrupt) {
		esp.TIMG0.INT_CLR_TIMERS.Set(1)
	})

	// Enable the CPU interrupt.
	mie := riscv.DisableInterrupts()

	// Clear any stale pending bit.
	plic.MXINT_CLEAR.SetBits(1 << timerAlarmCPUInterrupt)
	plic.MXINT_CLEAR.ClearBits(1 << timerAlarmCPUInterrupt)

	// Set edge-triggered.
	plic.MXINT_TYPE.SetBits(1 << timerAlarmCPUInterrupt)

	// Set priority above threshold.
	plic.MXINT_PRI[timerAlarmCPUInterrupt].Set(10)

	riscv.Asm("fence")

	plic.MXINT_ENABLE.SetBits(1 << timerAlarmCPUInterrupt)

	riscv.EnableInterrupts(mie)
}

// sleepTicks spins until the given number of ticks have elapsed, using the
// TIMG0 alarm interrupt to avoid busy-waiting for the entire duration.
func sleepTicks(d timeUnit) {
	machine.FlushSerial()
	target := ticks() + d
	for ticks() < target {
		interruptPending.Set(0)

		esp.TIMG0.T0ALARMLO.Set(uint32(target))
		esp.TIMG0.T0ALARMHI.Set(uint32(target >> 32))

		// Enable the alarm (auto-clears when alarm fires).
		esp.TIMG0.T0CONFIG.SetBits(esp.TIMG_T0CONFIG_ALARM_EN)

		for interruptPending.Get() == 0 {
			if ticks() >= target {
				return
			}
		}
	}
}

//go:extern _vector_table
var _vector_table [0]uintptr

// plicType maps the machine-mode registers of the PLIC.
// See ESP-IDF components/soc/esp32c6/register/soc/plic_reg.h.
type plicType struct {
	MXINT_ENABLE     volatile.Register32     // 0x00 bit N enables CPU interrupt line N
	MXINT_TYPE       volatile.Register32     // 0x04 bit N: 1=edge, 0=level
	MXINT_CLEAR      volatile.Register32     // 0x08 edge acknowledge
	MXINT_EIP_STATUS volatile.Register32     // 0x0C pending status (read-only)
	MXINT_PRI        [32]volatile.Register32 // 0x10..0x8C per-line priority (4 bits)
	MXINT_THRESH     volatile.Register32     // 0x90 priority threshold (8 bits)
}

// plic points at the PLIC machine-mode register block (DR_REG_PLIC_MX_BASE).
var plic = (*plicType)(unsafe.Pointer(uintptr(0x20001000)))
