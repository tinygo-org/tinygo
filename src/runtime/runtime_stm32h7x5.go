// +build stm32h7x5

package runtime

// This file implements the common types and operations available when running
// in either the Cortex-M7 or Cortex-M4 context.

import (
	"device/stm32"
	"runtime/volatile"
	"unsafe"
)

func initSystem() {

	// Reset core registers
	initCore()

	// Install interrupt vector table (VTOR)
	initVectors()

	// Copy data/bss sections from flash to RAM
	preinit()

	// Configure memory access (MPU, I/D-CACHE)
	initMemory()

	// Initialize hardware semaphore (HSEM)
	initSync()

	// Configure cache accelerators (ART on Cortex-M4)
	initAccel()
}

func initClocks() {

	// Initialize SysTick with timebase out of reset -- needed for supporting
	// timeout functionality when adjusting core clock frequencies below.
	initSysTick(stm32.ClockFreq())

	// Configure core clocks to final frequency:
	//   Cortex-M7: 480 MHz
	//   Cortex-M4: 240 MHz
	initCoreClocks() // Also updates SysTick with new core frequencies
}

func initPeripherals() {

}

const (
	tickPriority = 16   // NVIC priority number
	tickFreqHz   = 1000 // 1 kHz
	nsecsPerTick = 1000000000 / tickFreqHz
)

var (
	tickCount  volatile.Register64
	cycleCount volatile.Register32
)

var (
	// We use the ARMv7-M Debug facilities for counting CPU cycles, specifically
	// the Data Watchpoint and Trace Unit (DWT), register CYCCNT.
	// Usage of these capabilities is specified in the ARMv7-M Architecture
	// Reference Manual (https://developer.arm.com/documentation/ddi0403/latest/),
	// in the chapters indicated below.

	// C1.6.5 Debug Exception and Monitor Control Register, DEMCR
	DEM_CR = (*volatile.Register32)(unsafe.Pointer(uintptr(0xE000EDFC)))
	// C1.8.7 Control register, DWT_CTRL
	DWT_CR = (*volatile.Register32)(unsafe.Pointer(uintptr(0xE0001000)))
	// C1.8.8 Cycle Count register, DWT_CYCCNT
	DWT_CYCCNT = (*volatile.Register32)(unsafe.Pointer(uintptr(0xE0001004)))
)

func initSysTick(clk stm32.RCC_CLK_Type) {

	initCoreSysTick(clk)

	// turn on cycle counter
	DEM_CR.SetBits(0x01000000) // enable debugging & monitoring blocks
	DWT_CR.SetBits(0x00000001) // cycle count register
	cycleCount.Set(DWT_CYCCNT.Get())
}

//go:export SysTick_Handler
func tick() {
	tickCount.Set(tickCount.Get() + 1)
	cycleCount.Set(DWT_CYCCNT.Get())
}

func ticksToNanoseconds(t timeUnit) int64 { return int64(t) * nsecsPerTick }
func nanosecondsToTicks(n int64) timeUnit { return timeUnit(n / nsecsPerTick) }

// number of ticks (microseconds) since start.
//go:linkname ticks runtime.ticks
func ticks() timeUnit { return timeUnit(tickCount.Get()) }

// current CPU cycle count reported by Cortex-M DWT unit
//go:linkname ticks runtime.cycles
func cycles() uint32 { return cycleCount.Get() }

func sleepTicks(d timeUnit) {}
