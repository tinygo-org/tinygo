// +build stm32h7

package runtime

import "device/arm"

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

// The HSE frequency is derived from STM32H7 Nucleo(s) and Arduino Portenta H7,
// but it may not be the same on every device with an STM32H7.

// +---------+----------+
// |    HSE  |  25 MHz  | High-speed  (External)
// |    LSE  |  32 kHz  | Low-speed   (External)
// |    HSI  |  64 MHz  | High-speed  (Internal)
// |    CSI  |   4 MHz  | Low-power   (Internal)
// |    LSI  |  32 kHz  | Low-speed   (Internal)
// +---------+----------+
// |  HSI48  |  48 MHz  | <- USB PHY
// +---------+----------+

type timeUnit int64

const asyncScheduler = false

//go:extern _svectors
var _svectors [0]uint8

//go:extern _evectors
var _evectors [0]uint8

//go:extern _svtor
var _svtor [0]uint8 // vectors location in RAM

func postinit() {}

//export Reset_Handler
func main() {

	// disable interrupts
	irq := arm.DisableInterrupts()

	// Initialize system and perform any core-specific bootstrapping:
	//   1. Reset core registers
	//   2. Install interrupt vector table (VTOR)
	//   3. Initialize flash -> RAM
	//   4. TODO: Enable MPU:
	//       a. Define memory layout and access schema
	//       b. Basic interface implemented in device/stm32/stm32h7x7_cmx_mpu.go,
	//          which can be used to implement 4a above.
	//   5. Enable CPU instruction and data caches (I/D-Cache, Cortex-M7 only)
	//   6. Enable hardware semaphore (HSEM)
	//   7. At this point, execution paths diverge widely:
	//   (Cortex-M4)
	//       a. Enter deep sleep and wait until boot signal received
	//       b. Receive boot signal (e.g., from Cortex-M7 core)
	//       c. Enable instruction cache via ART accelerator
	//       d. Continue to clock/timer initialization
	//   (Cortex-M7)
	//       a. Continue to clock/timer initialization
	initSystem()

	// Rreenable interrupts
	arm.EnableInterrupts(irq)

	// NOTE: There are separate SysTick timers, one for each core.

	// (Cortex-M4)
	//   1. Enable SysTick with current core frequency (HCLK, 240 MHz).
	//        * The Cortex-M4 core only reaches this point once it has been gated
	//          by the Cortex-M7 core, which always occurs after system clocks
	//          have been initialized. Thus, the Cortex-M4 SysTick only needs to
	//          be configured once, since it already has the final core frequency.
	//   2. TODO: Enable peripheral sleep timer(s)
	//       a. The CPU debug cycle counter (CYCCNT), part of ARM Coresight Debug,
	//          Watch, & Trace (DWT) unit, is already enabled. This should be used
	//          in combination with sleep timers to more accurately handle timings
	//          with sub-microsecond precision.
	// (Cortex-M7)
	//   1. Enable SysTick with default core frequency (HSI, 64 MHz)
	//        * An initial SysTick is required prior to clock initializations, as
	//          we use timeout logic to verify switching clock sources.
	//   2. TODO: Enable peripheral sleep timer(s)
	//       a. The CPU debug cycle counter (CYCCNT), part of ARM Coresight Debug,
	//          Watch, & Trace (DWT) unit, is already enabled. This should be used
	//          in combination with sleep timers to more accurately handle timings
	//          with sub-microsecond precision.
	//   3. Configure clock frequencies used by both cores (480 MHz, 240 MHz):
	//       a. Initialize low-speed crystal (LSE, 32 kHz) required for sleep and
	//          other low-power modes.
	//       b. Change system clock source to CSI (4 MHz) temporarily while
	//          modifying main PLL (PLL1).
	//       c. Configure internal voltage regular.
	//       d. Enable HSE (25 MHz), and select HSE as main PLL (PLL1) source.
	//       e. Configure main PLL (PLL1, 480 MHz) with new HSE clock source.
	//       f. Enable main PLL (PLL1), change system clock source to PLL1, and
	//          configure bus clock divisors (SYS 480 MHz, AHB 240 MHz).
	//       g. TODO: Enable USB regulator and VBUS level detection.
	//   4. Reconfigure SysTick with new core frequencies (PLL[HSE], 480 MHz).
	//   5. Enable Cortex-M4 clock gate.
	initClocks()

	// configure GPIO and default peripherals
	initPeripherals()

	run()
	abort()
}

func waitForEvents() {
	arm.Asm("wfe")
}

func putchar(c byte) {}
