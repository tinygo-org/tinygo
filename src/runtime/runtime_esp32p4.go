//go:build esp32p4

package runtime

import (
	"device/esp"
	"device/riscv"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// TIMG0 runs from APB (80MHz): divider 2 gives 40MHz ticks, 25ns each.
// T0CONFIG.USE_XTAL is left clear, so the source is APB, not XTAL.
const (
	timerDivider = 2
	nsPerTick    = 25
)

// This is the function called on startup after the flash (IROM/DROM) is
// initialized and the stack pointer has been set.
//
//export main
func main() {
	// NOTE: Watchdog timers and peripheral clocks (UART0, TIMG0) were already
	// configured in call_start_cpu0 assembly before jumping here. Assembly is
	// used to avoid TinyGo's BSS-global issue where uintptr address literals
	// are stored in BSS globals that the ROM zeroes, causing load faults.

	clearbss()

	// Configure interrupt handler.
	interruptInit()

	// Initialize main system timer used for time.Now.
	initTimer()

	// Initialize the heap, call main.main, etc.
	run()

	// Fallback: if main ever returns, hang the CPU.
	exit(0)
}

func init() {
	// The PSRAM driver allocates, so it runs after the heap is ready.
	startPSRAM()
	machine.InitSerial()
	checkPSRAM()
}

// interruptInit initializes the interrupt controller.
//
// The ESP32-P4 uses a hybrid interrupt architecture:
//   - INTMTX (at 0x500D6000): routes peripheral IRQs to CPU interrupt lines
//   - CLIC (at 0x20800000): provides per-source vectored dispatch and priority
//
// For initial bring-up, we use the INTMTX routing with a CLIC configuration
// modeled after the ESP32-C6's PLIC handler pattern.
func interruptInit() {
	mie := riscv.DisableInterrupts()

	// Reset all CLIC interrupt priorities to zero.
	for i := 1; i < 48; i++ {
		clicSetPriority(uint32(i), 0)
	}

	// Default threshold is 0, matching ESP-IDF's RVHAL_INTR_ENABLE_THRESH:
	// any interrupt with priority >= 1 is delivered, while priority 0 masks
	// it (used by the handler to avoid re-entry).
	clicSetThreshold(0)

	// Set the interrupt vector base address.
	// Set MODE field to 1 — a vector base address (classic vectored table whose
	// entries are jump instructions, e.g. `j handleInterruptASM`).
	// Note that this address must be aligned to 256 bytes.
	riscv.MTVEC.Set((uintptr(unsafe.Pointer(&_vector_table))) | 1)

	// Enable machine-mode interrupts via mie CSR.
	riscv.MIE.Set(0xffffffff)

	// Globally enable machine-mode interrupts (MSTATUS.MIE).
	// The ROM bootloader may leave this cleared (as the C6 ROM does).
	_ = mie
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
	// Map the TIMG0 T0 peripheral interrupt to the CLIC source for CPU
	// interrupt number `timerAlarmCPUInterrupt`.  The INTMTX map register holds
	// the CLIC source index, which is cpu_number + CLIC_EXT_INTR_NUM_OFFSET
	// (the first 16 CLIC sources are internal/local).
	esp.INTERRUPT_CORE0.TIMERGRP0_T0_INT_MAP.Set(timerAlarmCPUInterrupt + clicExtIntrNumOffset)

	// Enable T0 interrupt at the timer group level.
	esp.TIMG0.INT_ENA_TIMERS.SetBits(1)

	// Register the interrupt handler.
	interrupt.New(timerAlarmCPUInterrupt, func(interrupt.Interrupt) {
		esp.TIMG0.INT_CLR_TIMERS.Set(1)
	})

	// Enable the CPU interrupt via the CLIC.
	mie := riscv.DisableInterrupts()

	// Clear any stale pending bit.
	clicClearPending(timerAlarmCPUInterrupt)

	// Set edge-triggered.
	clicSetTrigger(timerAlarmCPUInterrupt, 1) // 1 = edge

	// Set priority above threshold.
	clicSetPriority(timerAlarmCPUInterrupt, 7)

	riscv.Asm("fence")

	// Enable this interrupt in the CLIC.
	clicSetEnable(timerAlarmCPUInterrupt, true)

	riscv.EnableInterrupts(mie)
}

// sleepTicks spins until the given number of ticks have elapsed. It does not
// use the TIMG0 alarm interrupt.
func sleepTicks(d timeUnit) {
	printlock()
	machine.FlushSerial()
	printunlock()
	core := currentCPU()
	target := ticks() + d

	for ticks() < target && !sleepWoken(core) {
	}
}

//go:extern _vector_table
var _vector_table [0]uintptr

// CLIC register access functions.
//
// The ESP32-P4 CLIC is at 0x20800000 with the following layout:
//   - 0x20800000: CLIC base (interrupt pending/enable/attr/ctl per source)
//   - 0x20800008: CLIC threshold (mintthresh, memory-mapped on pre-v3)
//   - 0x20801000: CLIC control
//
// Each interrupt source has 4 bytes of control at base + 0x1000 + N*4:
//
//	byte 0: clicintip  (pending)
//	byte 1: clicintie  (enable)
//	byte 2: clicintattr (trigger type, shv)
//	byte 3: clicintctl (priority/level)
const clicBase = 0x20800000
const clicIntBase = clicBase + 0x1000

// clicExtIntrNumOffset is the CLIC_EXT_INTR_NUM_OFFSET used by ESP-IDF: the
// first 16 CLIC sources are internal (local) interrupts, so an externally
// routed peripheral interrupt number n lives at CLIC source n+16.
// The CLIC also reports this source (n+16) in mcause on dispatch, so the ISR
// must subtract this offset to recover the CPU interrupt number.
const clicExtIntrNumOffset = 16

// clicSource returns the CLIC per-source register index for a CPU interrupt
// number n.  Verified on hardware: source n+16 accepts writes, source n does
// not (reserved), which is why interrupts never previously dispatched.
func clicSource(n uint32) uint32 {
	return n + clicExtIntrNumOffset
}

// CLIC priority encoding (NLBITS=3 on ESP32-P4):
//
//	bits [7:5] = priority level (0-7)
//	bits [4:0] = reserved, read and written as 1
//
// ESP-IDF's NLBITS_TO_BYTE() uses the same encoding.  Writing a raw level
// shifted by 5 leaves the low bits zero, which is not the format the hardware
// expects; conversely reading the CTL byte yields the encoded value, not the
// level, so get/set must agree on the representation.
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

func clicSetEnable(n uint32, enable bool) {
	reg := (*volatile.Register8)(unsafe.Pointer(uintptr(clicIntBase + clicSource(n)*4 + 1)))
	if enable {
		reg.Set(1)
	} else {
		reg.Set(0)
	}
}

func clicSetTrigger(n uint32, edgeTriggered uint8) {
	reg := (*volatile.Register8)(unsafe.Pointer(uintptr(clicIntBase + clicSource(n)*4 + 2)))
	// MODE bits [7:6] are read-only and fixed to 0b11.  Set SHV=0 so
	// dispatch uses the classic MTVEC jump table, and configure TRIG[2:1]:
	// 00 = level, 01 = rising edge.
	v := uint8(0xC0)
	if edgeTriggered != 0 {
		v |= 0x02
	}
	reg.Set(v)
}

func clicClearPending(n uint32) {
	reg := (*volatile.Register8)(unsafe.Pointer(uintptr(clicIntBase + clicSource(n)*4)))
	reg.Set(0)
}

func clicSetThreshold(threshold uint8) {
	reg := (*volatile.Register32)(unsafe.Pointer(uintptr(clicBase + 0x8)))
	// CLIC_CPU_INT_THRESH field is at bits[31:24] of CLIC_INT_THRESH_REG,
	// using the same NLBITS encoding as an interrupt priority.  Interrupts
	// with priority <= threshold are masked.
	reg.Set(uint32(clicEncodePriority(threshold)) << 24)
}

func clicGetPriority(n uint32) uint8 {
	reg := (*volatile.Register8)(unsafe.Pointer(uintptr(clicIntBase + clicSource(n)*4 + 3)))
	return reg.Get() >> (8 - clicNLBITS)
}
