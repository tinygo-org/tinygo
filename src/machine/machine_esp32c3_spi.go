//go:build esp32c3

package machine

// On the C3 variant, SPI2 is a general purpose SPI controller. SPI0 and SPI1
// are used internally to access the ESP32-C3’s attached flash memory. Due to
// different registers between SPI2 and the other SPI ports, this driver
// currently supports only the the general purpose FSPI SPI2 controller.
// https://docs.espressif.com/projects/esp-idf/en/latest/esp32c3/api-reference/peripherals/spi_master.html

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

const spiClockFreq = 80_000_000 // APB_CLK, MST_CLK_SEL 1. See ESP-IDF hal/esp32c3/include/hal/spi_ll.h spi_ll_master_cal_clock.

const (
	FSPICLK_IN_IDX  = uint32(63)
	FSPICLK_OUT_IDX = uint32(63)
	FSPIQ_IN_IDX    = uint32(64)
	FSPIQ_OUT_IDX   = uint32(64)
	FSPID_IN_IDX    = uint32(65)
	FSPID_OUT_IDX   = uint32(65)
	FSPIHD_IN_IDX   = uint32(66)
	FSPIHD_OUT_IDX  = uint32(66)
	FSPIWP_IN_IDX   = uint32(67)
	FSPIWP_OUT_IDX  = uint32(67)
	FSPICS0_IN_IDX  = uint32(68)
	FSPICS0_OUT_IDX = uint32(68)
	FSPICS1_OUT_IDX = uint32(69)
	FSPICS2_OUT_IDX = uint32(70)
	FSPICS3_OUT_IDX = uint32(71)
	FSPICS4_OUT_IDX = uint32(72)
	FSPICS5_OUT_IDX = uint32(73)
)

// Serial Peripheral Interface on the ESP32-C3.
type SPI struct {
	Bus *esp.SPI2_Type
}

var (
	// SPI0 and SPI1 are reserved for use by the caching system etc.
	SPI2 = &SPI{esp.SPI2}
	SPI0 = SPI2
)

func spiBusFor(bus *esp.SPI2_Type) (spiBus, bool) {
	switch bus {
	case esp.SPI2:
		return spiBus{
			sckOut: FSPICLK_OUT_IDX,
			sdoOut: FSPID_OUT_IDX,
			sdiIn:  FSPIQ_IN_IDX,
			csOut:  FSPICS0_OUT_IDX,
		}, true
	default:
		return spiBus{}, false
	}
}

func (spi *SPI) configure(config SPIConfig, bus spiBus, mode spiMode) error {
	// periph module reset
	esp.SYSTEM.SetPERIP_RST_EN0_SPI2_RST(1)
	esp.SYSTEM.SetPERIP_RST_EN0_SPI2_RST(0)

	// periph module enable
	esp.SYSTEM.SetPERIP_CLK_EN0_SPI2_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_SPI2_RST(0)

	// init the spi2 bus
	spi.Bus.SLAVE.Set(0)
	spi.Bus.MISC.Set(0)
	spi.Bus.USER.Set(0)
	spi.Bus.USER1.Set(0)
	spi.Bus.CTRL.Set(0)
	spi.Bus.CLK_GATE.Set(0)
	spi.Bus.DMA_CONF.Set(0)
	spi.Bus.SetDMA_CONF_RX_AFIFO_RST(1)
	spi.Bus.SetDMA_CONF_BUF_AFIFO_RST(1)
	spi.Bus.CLOCK.Set(0)

	// clear data buf
	spi.Bus.SetW0(0)
	spi.Bus.SetW1(0)
	spi.Bus.SetW2(0)
	spi.Bus.SetW3(0)
	spi.Bus.SetW4(0)
	spi.Bus.SetW5(0)
	spi.Bus.SetW6(0)
	spi.Bus.SetW7(0)
	spi.Bus.SetW8(0)
	spi.Bus.SetW9(0)
	spi.Bus.SetW10(0)
	spi.Bus.SetW11(0)
	spi.Bus.SetW12(0)
	spi.Bus.SetW13(0)
	spi.Bus.SetW14(0)
	spi.Bus.SetW15(0)

	// start the spi2 bus
	spi.Bus.SetCLK_GATE_CLK_EN(1)
	spi.Bus.SetCLK_GATE_MST_CLK_SEL(1)
	spi.Bus.SetCLK_GATE_MST_CLK_ACTIVE(1)
	spi.Bus.SetDMA_CONF_SLV_TX_SEG_TRANS_CLR_EN(1)
	spi.Bus.SetDMA_CONF_SLV_RX_SEG_TRANS_CLR_EN(1)
	spi.Bus.SetDMA_CONF_DMA_SLV_SEG_TRANS_EN(0)
	spi.Bus.SetUSER_USR_MOSI(1)
	spi.Bus.SetUSER_USR_MISO(1)
	spi.Bus.SetUSER_DOUTDIN(1)

	// set spi2 data mode
	spi.Bus.SetMISC_CK_IDLE_EDGE(mode.idleEdge())
	spi.Bus.SetUSER_CK_OUT_EDGE(mode.outEdge())

	// set spi2 bit order
	if config.LSBFirst {
		spi.Bus.SetCTRL_WR_BIT_ORDER(1) // LSB first
		spi.Bus.SetCTRL_RD_BIT_ORDER(1)
	} else {
		spi.Bus.SetCTRL_WR_BIT_ORDER(0) // MSB first
		spi.Bus.SetCTRL_RD_BIT_ORDER(0)
	}

	// configure SPI bus clock
	spi.Bus.CLOCK.Set(freqToClockDiv(config.Frequency))

	return nil
}

func (spi *SPI) start() {
	spi.Bus.SetCMD_UPDATE(1)
	for spi.Bus.GetCMD_UPDATE() != 0 {
	}
	spi.Bus.SetCMD_USR(1)
	for spi.Bus.GetCMD_USR() != 0 {
	}
}

func (spi *SPI) setBitLen(n uint32) {
	spi.Bus.SetMS_DLEN_MS_DATA_BITLEN(n)
}

func (spi *SPI) buffer() *[16]volatile.Register32 {
	return (*[16]volatile.Register32)(unsafe.Pointer(&spi.Bus.W0))
}
