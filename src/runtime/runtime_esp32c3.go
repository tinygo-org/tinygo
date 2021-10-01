// +build esp32c3

package runtime

import (
	"device/esp"
	"device/riscv"
	"runtime/volatile"
	"unsafe"
)

const (
	INT_CODE_INSTR_ACCESS_FAULT = 0x1 // PMP Instruction access fault
	INT_CODE_ILL_INSTR          = 0x2 // Illegal Instruction
	INT_CODE_BRK                = 0x3 // Hardware Breakpoint/Watchpoint or EBREAK
	INT_CODE_LOAD_FAULT         = 0x5 // PMP Load access fault
	INT_CODE_STORE_FAULT        = 0x7 // PMP Store access fault
	INT_CODE_USR_CALL           = 0x8 // ECALL from U mode
	INT_CODE_MACH_CALL          = 0xb // ECALL from M mode
)

// This is the function called on startup after the flash (IROM/DROM) is
// initialized and the stack pointer has been set.
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
	esp.RTC_CNTL.RTC_WDTWPROTECT.Set(0x50D83AA1)
	esp.RTC_CNTL.RTC_WDTCONFIG0.Set(0)

	// Disable super watchdog.
	esp.RTC_CNTL.RTC_SWD_WPROTECT.Set(0x8F1D312A)
	esp.RTC_CNTL.RTC_SWD_CONF.Set(esp.RTC_CNTL_RTC_SWD_CONF_SWD_DISABLE)

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
	initInterrupt()

	// Initialize main system timer used for time.Now.
	initTimer()

	// Initialize the heap, call main.main, etc.
	run()

	// Fallback: if main ever returns, hang the CPU.
	abort()
}

func abort() {
	// lock up forever
	for {
		riscv.Asm("wfi")
	}
}

//go:extern _vector_table
var _vector_table [0]uintptr

func initInterrupt() {
	mie := riscv.DisableInterrupts()

	// Reset all interrupt source priorities to zero.
	priReg := &esp.INTERRUPT_CORE0.CPU_INT_PRI_1
	for i := 0; i < 31; i++ {
		priReg.Set(0)
		addr := uintptr(unsafe.Pointer(priReg)) + 4
		priReg = (*volatile.Register32)(unsafe.Pointer(addr))
	}

	println("_vector_table:", &_vector_table)

	// Set the interrupt address.
	// Set MODE field to 1 - a vector base address (only supported by ESP32C3)
	// Note that this address must be aligned to 256 bytes.
	riscv.MTVEC.Set((uintptr(unsafe.Pointer(&_vector_table))) | 1)

	riscv.EnableInterrupts(mie)
}

//export handleInterrupt
func handleInterrupt() {
	cause := riscv.MCAUSE.Get()
	code := uint32(cause & 0xf)
	if cause&(1<<31) == 0 {
		handleException(code)
		return
	}

	// Topmost bit is set, which means that it is an interrupt.
	println("INTR: code:", code)
	// switch code {
	// case 7: // Machine timer interrupt
	// 	// Signal timeout.
	// 	timerWakeup.Set(1)
	// 	// Disable the timer, to avoid triggering the interrupt right after
	// 	// this interrupt returns.
	// 	riscv.MIE.ClearBits(1 << 7) // MTIE bit
	// case 11: // Machine external interrupt
	// 	hartId := riscv.MHARTID.Get()

	// 	// Claim this interrupt.
	// 	id := kendryte.PLIC.TARGETS[hartId].CLAIM.Get()
	// 	// Call the interrupt handler, if any is registered for this ID.
	// 	callInterruptHandler(int(id))
	// 	// Complete this interrupt.
	// 	kendryte.PLIC.TARGETS[hartId].CLAIM.Set(id)
	// }
}

//export handleException
func handleException(code uint32) {
	println("*** Exception: code:")
}
