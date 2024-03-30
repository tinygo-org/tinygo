//go:build rp2040

package machine

import (
	"device/rp"
	"runtime/volatile"
	"unsafe"
)

const deviceName = rp.Device

//go:linkname machineInit runtime.machineInit
func machineInit() {
	// Reset all peripherals to put system into a known state,
	// except for QSPI pads and the XIP IO bank, as this is fatal if running from flash
	// and the PLLs, as this is fatal if clock muxing has not been reset on this boot
	// and USB, syscfg, as this disturbs USB-to-SWD on core 1
	bits := ^uint32(rp.RESETS_RESET_IO_QSPI |
		rp.RESETS_RESET_PADS_QSPI |
		rp.RESETS_RESET_PLL_USB |
		rp.RESETS_RESET_USBCTRL |
		rp.RESETS_RESET_SYSCFG |
		rp.RESETS_RESET_PLL_SYS)
	resetBlock(bits)

	// Remove reset from peripherals which are clocked only by clkSys and
	// clkRef. Other peripherals stay in reset until we've configured clocks.
	bits = ^uint32(rp.RESETS_RESET_ADC |
		rp.RESETS_RESET_RTC |
		rp.RESETS_RESET_SPI0 |
		rp.RESETS_RESET_SPI1 |
		rp.RESETS_RESET_UART0 |
		rp.RESETS_RESET_UART1 |
		rp.RESETS_RESET_USBCTRL)
	unresetBlockWait(bits)

	clocks.init()

	// Peripheral clocks should now all be running
	unresetBlockWait(RESETS_RESET_Msk)
}

//go:linkname ticks runtime.machineTicks
func ticks() uint64 {
	return timer.timeElapsed()
}

//go:linkname lightSleep runtime.machineLightSleep
func lightSleep(ticks uint64) {
	timer.lightSleep(ticks)
}

// CurrentCore returns the core number the call was made from.
func CurrentCore() int {
	return int(rp.SIO.CPUID.Get())
}

// NumCores returns number of cores available on the device.
func NumCores() int { return 2 }

// ChipVersion returns the version of the chip. 1 is returned for B0 and B1
// chip.
func ChipVersion() uint8 {
	const (
		SYSINFO_BASE                  = 0x40000000
		SYSINFO_CHIP_ID_OFFSET        = 0x00000000
		SYSINFO_CHIP_ID_REVISION_BITS = 0xf0000000
		SYSINFO_CHIP_ID_REVISION_LSB  = 28
	)

	// First register of sysinfo is chip id
	chipID := *(*uint32)(unsafe.Pointer(uintptr(SYSINFO_BASE + SYSINFO_CHIP_ID_OFFSET)))
	// Version 1 == B0/B1
	version := (chipID & SYSINFO_CHIP_ID_REVISION_BITS) >> SYSINFO_CHIP_ID_REVISION_LSB
	return uint8(version)
}

// Single DMA channel. See rp.DMA_Type.
type dmaChannel struct {
	READ_ADDR   volatile.Register32
	WRITE_ADDR  volatile.Register32
	TRANS_COUNT volatile.Register32
	CTRL_TRIG   volatile.Register32
	_           [12]volatile.Register32 // aliases
}

// Static assignment of DMA channels to peripherals.
// Allocating them statically is good enough for now. If lots of peripherals use
// DMA, these might need to be assigned at runtime.
const (
	spi0DMAChannel = iota
	spi1DMAChannel
)

// DMA channels usable on the RP2040.
var dmaChannels = (*[12]dmaChannel)(unsafe.Pointer(rp.DMA))
