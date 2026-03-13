//go:build esp32c6

package runtime

import (
	"device/esp"
	"device/riscv"
	"machine"
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
	esp.LP_WDT.SWD_WPROTECT.Set(0x8F1D312A)
	esp.LP_WDT.SWD_CONF.SetBits(1 << 30) // SWD_DISABLE bit

	// Change CPU frequency to 160MHz.
	// Switch from XTAL to PLL clock source (SOC_CLK_SEL = 1 for SPLL).
	esp.PCR.SYSCLK_CONF.Set(1 << 16) // SOC_CLK_SEL = 1 (SPLL)

	// Set CPU divider to div1 for 160MHz from PLL.
	// CPU_HS_DIV_NUM = 0 means div1 of clk_hproot (160MHz PLL).
	esp.PCR.CPU_FREQ_CONF.Set(0 << 8) // CPU_HS_DIV_NUM = 0

	clearbss()

	// Configure interrupt handler
	interruptInit()

	// Initialize main system timer used for time.Now.
	initTimer()

	// Initialize the heap, call main.main, etc.
	run()

	// Fallback: if main ever returns, hang the CPU.
	exit(0)
}

func init() {
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
	priReg := &esp.INTPRI.CPU_INT_PRI_1
	for i := 0; i < 31; i++ {
		priReg.Set(0)
		priReg = (*volatile.Register32)(unsafe.Add(unsafe.Pointer(priReg), 4))
	}

	// default threshold for interrupts is 5
	esp.INTPRI.CPU_INT_THRESH.Set(5)

	// Set the interrupt address.
	// Set MODE field to 1 - a vector base address (only supported by ESP32-C6)
	// Note that this address must be aligned to 256 bytes.
	riscv.MTVEC.Set((uintptr(unsafe.Pointer(&_vector_table))) | 1)

	riscv.EnableInterrupts(mie)
}

//go:extern _vector_table
var _vector_table [0]uintptr

// Serial I/O via USB-Serial-JTAG controller.
// The ESP32-C6-DevKitC connects its USB port to the internal USB-Serial-JTAG
// peripheral, not to UART0. The ROM bootloader also uses this path, so output
// appears on the same port as boot messages.

func putchar(c byte) {
	// Wait for the USB-Serial-JTAG TX FIFO to have space.
	// Use a timeout to avoid hanging if no host is reading.
	for i := 0; i < 10000; i++ {
		if esp.USB_DEVICE.EP1_CONF.Get()&0x2 != 0 { // SERIAL_IN_EP_DATA_FREE
			break
		}
	}
	// Write byte directly to EP1 FIFO. We must NOT use the generated
	// SetEP1_RDWR_BYTE accessor because it does a read-modify-write,
	// and reading EP1 pops a byte from the RX FIFO.
	esp.USB_DEVICE.EP1.Set(uint32(c))
	// Signal that data has been written to the FIFO.
	esp.USB_DEVICE.EP1_CONF.Set(1) // WR_DONE
}

func getchar() byte {
	// Not implemented for USB-Serial-JTAG yet.
	for {
		riscv.Asm("wfi")
	}
}

func buffered() int {
	return 0
}
