//go:build esp32c6

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

	// Disable Timer Group 0 watchdog (unlock first).
	esp.TIMG0.WDTWPROTECT.Set(0x50D83AA1)
	esp.TIMG0.WDTCONFIG0.Set(0)

	// Disable Timer Group 1 watchdog (unlock first).
	esp.TIMG1.WDTWPROTECT.Set(0x50D83AA1)
	esp.TIMG1.WDTCONFIG0.Set(0)

	// Disable LP watchdog (write-protect key first).
	esp.LP_WDT.WDTWPROTECT.Set(0x50D83AA1)
	esp.LP_WDT.WDTCONFIG0.Set(0)

	// Disable super watchdog.
	esp.LP_WDT.SWD_WPROTECT.Set(0x50D83AA1)
	esp.LP_WDT.SWD_CONF.SetBits(1 << 30) // SWD_DISABLE bit

	// Change CPU frequency to 160MHz from SPLL (480MHz).
	//
	// Clock tree: SPLL (480MHz) → HP root → CPU / AHB / APB
	//
	// Set dividers BEFORE switching the clock source so the first PLL
	// cycle already arrives divided:
	//   HP root  = SPLL / (HS_DIV_NUM+1)      = 480 / 3 = 160 MHz
	//   CPU      = HP root / (CPU_HS_DIV_NUM+1) = 160 / 1 = 160 MHz
	//   AHB      = HP root / (AHB_HS_DIV_NUM+1) = 160 / 4 =  40 MHz
	//   APB      = AHB / (APB_HS_DIV_NUM+1)     =  40 / 1 =  40 MHz
	esp.PCR.CPU_FREQ_CONF.Set(0 << 8) // CPU_HS_DIV_NUM = 0 (div1)
	esp.PCR.AHB_FREQ_CONF.Set(3 << 8) // AHB_HS_DIV_NUM = 3 (div4)
	esp.PCR.APB_FREQ_CONF.Set(0 << 8) // APB_HS_DIV_NUM = 0 (div1)

	// Switch to PLL: SOC_CLK_SEL = 1 (SPLL), HS_DIV_NUM = 2 (div3).
	esp.PCR.SYSCLK_CONF.Set(1<<16 | 2<<8)

	// Select the Timer Group 0 timer clock source.
	//
	// The shared timekeeping code (runtime_esp32xx.go) assumes the TIMG0 timer
	// counts at 40MHz: an 80MHz source divided by the prescaler of 2 set in
	// initTimer, giving 25ns/tick. On the ESP32-C6 the timer group timer clock
	// is not the APB clock; it is selected via PCR and defaults to the 40MHz
	// XTAL (TG0_TIMER_CLK_SEL = 0). With the /2 prescaler that yields a 20MHz
	// tick, so every delay would be twice as long as intended (time.Sleep and
	// therefore blinky run at half speed).
	//
	// Select PLL_F80M (80MHz) as the source so the timer counts at 40MHz and
	// the 25ns/tick assumption holds. Clock source encoding (esp-idf
	// timer_ll_set_clock_source): 0 = XTAL, 1 = PLL_F80M, 2 = RC_FAST.
	esp.PCR.SetTIMERGROUP0_CONF_TG0_CLK_EN(1)
	esp.PCR.SetTIMERGROUP0_TIMER_CLK_CONF_TG0_TIMER_CLK_SEL(1)
	esp.PCR.SetTIMERGROUP0_TIMER_CLK_CONF_TG0_TIMER_CLK_EN(1)

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
	machine.InitSerial()
}

func abort() {
	for {
		riscv.Asm("wfi")
	}
}

// interruptInit initializes the interrupt controller.
func interruptInit() {
	mie := riscv.DisableInterrupts()

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

	// On the ESP32-C6 the CPU receives PLIC interrupts through the standard
	// RISC-V mie CSR (machine interrupt-enable, 0x304). Unlike the ESP32-C3's
	// INTC, the PLIC's per-line interrupts are gated by mie, so it must be
	// enabled for any interrupt to reach the CPU. esp-hal does the same thing
	// for the PLIC: `csrw mie, 0xffffffff`.
	riscv.MIE.Set(0xffffffff)

	// Globally enable machine-mode interrupts (MSTATUS.MIE).
	//
	// Unlike the ESP32-C3, whose ROM bootloader hands control to the
	// application with MSTATUS.MIE already set, the ESP32-C6 ROM leaves it
	// cleared. Restoring the previous state (riscv.EnableInterrupts(mie))
	// would therefore leave interrupts globally disabled, so no interrupt is
	// ever delivered to the CPU: timer alarms silently fall back to the
	// polling loop in sleepTicks and peripheral RX interrupts (such as the
	// USB Serial/JTAG controller) never fire. Set the bit explicitly instead.
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
	// Map the TIMG0 T0 peripheral interrupt to a CPU interrupt line.
	// On C6, INTERRUPT_CORE0 is for peripheral→CPU interrupt mapping.
	esp.INTERRUPT_CORE0.TG0_T0_INTR_MAP.Set(timerAlarmCPUInterrupt)

	// Enable T0 interrupt at the timer group level.
	esp.TIMG0.INT_ENA_TIMERS.SetBits(1)

	// Register the interrupt handler.
	interrupt.New(timerAlarmCPUInterrupt, func(interrupt.Interrupt) {
		esp.TIMG0.INT_CLR_TIMERS.Set(1)
	})

	// Enable the CPU interrupt:
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

// plicType maps the ESP32-C6 PLIC (Platform-Level Interrupt Controller)
// machine-mode registers. The C6 uses the PLIC — not the INTPRI/INTC block
// that the ESP32-C3 uses — as its CPU interrupt controller
// (SOC_INT_PLIC_SUPPORTED). The INTPRI registers at 0x600c5000 are vestigial
// backward-compatibility aliases on the C6 and are not wired to the CPU, so
// writing them never delivers an interrupt.
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
