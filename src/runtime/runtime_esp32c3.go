//go:build esp32c3

package runtime

import (
	"device/esp"
	"device/riscv"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// This is the function called on startup after the flash (IROM/DROM) is
// initialized and the stack pointer has been set.
//
//export main
func main() {
	// This initialization configures the following things:
	// * It disables all watchdog timers. They might be useful at some point in
	//   the future, but will need integration into the scheduler. For now,
	//   they're all disabled.
	// * It sets the CPU frequency to 160MHz, which is the maximum speed allowed
	//   for this CPU. Lower frequencies might be possible in the future, but
	//   running fast and sleeping quickly is often also a good strategy to save
	//   power.
	// TODO: protect certain memory regions, especially the area below the stack
	// to protect against stack overflows. See
	// esp_cpu_configure_region_protection in ESP-IDF.

	// Disable Timer 0 watchdog.
	esp.TIMG0.WDTCONFIG0.Set(0)

	// Disable RTC watchdog.
	esp.RTC_CNTL.WDTWPROTECT.Set(0x50D83AA1)
	esp.RTC_CNTL.WDTCONFIG0.Set(0)

	// Disable super watchdog.
	esp.RTC_CNTL.SWD_WPROTECT.Set(0x8F1D312A)
	esp.RTC_CNTL.SWD_CONF.Set(esp.RTC_CNTL_SWD_CONF_SWD_DISABLE)

	// Change CPU frequency from 20MHz to 80MHz, by switching from the XTAL to
	// the PLL clock source (see table "CPU Clock Frequency" in the reference
	// manual).
	esp.SYSTEM.SYSCLK_CONF.Set(1 << esp.SYSTEM_SYSCLK_CONF_SOC_CLK_SEL_Pos)

	// Change CPU frequency from 80MHz to 160MHz by setting SYSTEM_CPUPERIOD_SEL
	// to 1 (see table "CPU Clock Frequency" in the reference manual).
	// Note: we might not want to set SYSTEM_CPU_WAIT_MODE_FORCE_ON to save
	// power. It is set here to keep the default on reset.
	esp.SYSTEM.CPU_PER_CONF.Set(esp.SYSTEM_CPU_PER_CONF_CPU_WAIT_MODE_FORCE_ON | esp.SYSTEM_CPU_PER_CONF_PLL_FREQ_SEL | 1<<esp.SYSTEM_CPU_PER_CONF_CPUPERIOD_SEL_Pos)

	clearbss()

	// Configure interrupt handler
	interruptInit()

	// Initialize main system timer used for time.Now.
	initTimer()

	// Initialize timer alarm interrupt for the scheduler.
	initTimerInterrupt()

	// Initialize the heap, call main.main, etc.
	run()

	// Fallback: if main ever returns, hang the CPU.
	exit(0)
}

func init() {
	// Initialize UART.
	machine.InitSerial()
}

func abort() {
	// lock up forever
	for {
		riscv.Asm("wfi")
	}
}

// interruptInit initialize the interrupt controller and called from runtime once.
func interruptInit() {
	mie := riscv.DisableInterrupts()

	// Reset all interrupt source priorities to zero.
	priReg := &esp.INTERRUPT_CORE0.CPU_INT_PRI_1
	for i := 0; i < 31; i++ {
		priReg.Set(0)
		priReg = (*volatile.Register32)(unsafe.Add(unsafe.Pointer(priReg), 4))
	}

	// default threshold for interrupts is 5
	esp.INTERRUPT_CORE0.CPU_INT_THRESH.Set(5)

	// Set the interrupt address.
	// Set MODE field to 1 - a vector base address (only supported by ESP32C3)
	// Note that this address must be aligned to 256 bytes.
	riscv.MTVEC.Set((uintptr(unsafe.Pointer(&_vector_table))) | 1)

	riscv.EnableInterrupts(mie)
}

// CPU interrupt number used for the TIMG0 timer alarm.
const timerAlarmCPUInterrupt = 9

var interruptPending volatile.Register8

func signalInterrupt() {
	interruptPending.Set(1)
}

// initTimerInterrupt routes the TIMG0 timer 0 alarm interrupt to a CPU
// interrupt and registers a handler that signals timerWakeup.
func initTimerInterrupt() {
	// Map the TIMG0 T0 peripheral interrupt to a CPU interrupt line.
	esp.INTERRUPT_CORE0.TG_T0_INT_MAP.Set(timerAlarmCPUInterrupt)

	// Enable T0 interrupt at the timer group level.
	esp.TIMG0.INT_ENA_TIMERS.SetBits(1)

	// Register the interrupt handler (compile-time wiring).
	interrupt.New(timerAlarmCPUInterrupt, func(interrupt.Interrupt) {
		// Clear the timer interrupt at the peripheral level.
		esp.TIMG0.INT_CLR_TIMERS.Set(1)
	})

	// Manually enable the CPU interrupt with correct ordering:
	// 1) clear any stale pending bit first
	// 2) set edge-triggered
	// 3) set priority above threshold
	// 4) enable the interrupt last
	mie := riscv.DisableInterrupts()

	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.SetBits(1 << timerAlarmCPUInterrupt)
	esp.INTERRUPT_CORE0.CPU_INT_CLEAR.ClearBits(1 << timerAlarmCPUInterrupt)

	esp.INTERRUPT_CORE0.CPU_INT_TYPE.SetBits(1 << timerAlarmCPUInterrupt)

	priReg := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.INTERRUPT_CORE0.CPU_INT_PRI_0), timerAlarmCPUInterrupt*4))
	priReg.Set(10)

	riscv.Asm("fence")

	esp.INTERRUPT_CORE0.CPU_INT_ENABLE.SetBits(1 << timerAlarmCPUInterrupt)

	riscv.EnableInterrupts(mie)
}

// sleepTicks spins until the given number of ticks have elapsed, using the
// TIMG0 alarm interrupt to avoid busy-waiting for the entire duration.
func sleepTicks(d timeUnit) {
	machine.FlushSerial()
	target := ticks() + d
	for ticks() < target {
		// Set the alarm to fire at the target tick count (or as close
		// as the 54-bit counter allows).
		interruptPending.Set(0)

		esp.TIMG0.T0ALARMLO.Set(uint32(target))
		esp.TIMG0.T0ALARMHI.Set(uint32(target >> 32))

		// Enable the alarm (auto-clears when alarm fires).
		esp.TIMG0.T0CONFIG.SetBits(esp.TIMG_T0CONFIG_ALARM_EN)

		// Wait for any interrupt (timer alarm or other) or a timeout.
		for interruptPending.Get() == 0 {
			if ticks() >= target {
				return
			}
		}
	}
}

//go:extern _vector_table
var _vector_table [0]uintptr
