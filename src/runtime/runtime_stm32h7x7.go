// +build stm32h7x7

package runtime

// This file implements the common types and operations available when running
// in either the Cortex-M7 or Cortex-M4 context.

// +-------------+-----------+
// |   SYSCLK    |  480 MHz  |  <-  M7 CPU, M7 SysTick
// |   AHBCLK    |  240 MHz  |  <-  M4 CPU, M4 SysTick, AXI Periphs
// |  APB1CLK    |  120 MHz  |  <-  APB1 Periphs
// |  APB1CLK(*) |  240 MHz  |  <-  APB1 xTIMs
// |  APB2CLK    |  120 MHz  |  <-  APB2 Periphs
// |  APB2CLK(*) |  240 MHz  |  <-  APB2 xTIMs
// |  APB3CLK    |  120 MHz  |  <-  APB3 Periphs
// |  APB4CLK    |  120 MHz  |  <-  APB4 Periphs, APB4 xTIMs
// |  AHB4CLK    |  240 MHz  |  <-  AHB4 Periphs
// +-------------+-----------+

// +---------+----------+
// |    HSE  |  25 MHz  | High-speed, External
// |    LSE  |  32 kHz  | Low-speed, External
// |    HSI  |  64 MHz  | High-speed Internal
// |    CSI  |   4 MHz  | Low-power, Internal
// |    LSI  |  32 kHz  | Low-speed, Internal
// +---------+----------+
// |  HSI48  |  48 MHz  |  <-  USB PHY
// +---------+----------+

import (
	"device/stm32"
	"runtime/volatile"
	"unsafe"
)

type hsemID uint8

const (
	hsemRNG hsemID = iota
	hsemPKA
	hsemFLASH
	hsemRCC
	hsemSTOP
	hsemGPIO
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
	initSysTick(stm32.RCC.ClockFreq())

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
