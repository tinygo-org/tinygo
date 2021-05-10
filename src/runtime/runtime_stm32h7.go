// +build stm32h7

package runtime

import "device/arm"

const asyncScheduler = false

type timeUnit int64

func postinit() {}

//export Reset_Handler
func main() {

	// disable interrupts
	irq := arm.DisableInterrupts()

	// Initialize system and perform any core-specific bootstrapping:
	//   1. Reset core registers
	//   2. Install interrupt vector table (VTOR)
	//   3. Initialize flash -> RAM
	//   4. TODO: Enable MPU
	//       a. Define memory layout and access schema
	//       b. Basic interface implemented in device/stm32/stm32h7x7_cmx_mpu.go,
	//          which can be used to implement 4a above.
	//   5. Enable CPU instruction and data caches (I/D-Cache, Cortex-M7 only)
	//   6. Enable hardware semaphore (HSEM)
	//   7. At this point, execution paths diverge widely:
	//   (Cortex-M4)
	//        a. Enter deep sleep and wait until boot signal received
	//        b. Receive boot signal (e.g., from Cortex-M7 core)
	//        c. Enable instruction cache via ART accelerator
	//        d. Continue to clock/timer initialization
	//   (Cortex-M7)
	//        a. Continue to clock/timer initialization
	initSystem()

	// reenable interrupts
	arm.EnableInterrupts(irq)

	// configure core clocks and oscillators
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
