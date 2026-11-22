//go:build esp32p4

package runtime

import (
	"device/esp"
	"device/riscv"
	clic "internal/esp"
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
func interruptInit() {
	riscv.DisableInterrupts()
	clic.Init()

	// MODE 1 is vectored. The table must be aligned to 256 bytes.
	riscv.MTVEC.Set(uintptr(unsafe.Pointer(&_vector_table)) | 1)

	// The ROM can leave MSTATUS.MIE clear.
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
	esp.INTERRUPT_CORE0.TIMERGRP0_T0_INT_MAP.Set(timerAlarmCPUInterrupt + clic.MatrixOffset)
	esp.TIMG0.INT_ENA_TIMERS.SetBits(1)
	interrupt.New(timerAlarmCPUInterrupt, func(interrupt.Interrupt) {
		esp.TIMG0.INT_CLR_TIMERS.Set(1)
	})

	mask := riscv.DisableInterrupts()
	clic.Enable(timerAlarmCPUInterrupt, 7)
	riscv.Asm("fence")
	riscv.EnableInterrupts(mask)
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
