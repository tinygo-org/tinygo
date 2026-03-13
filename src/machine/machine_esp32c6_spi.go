//go:build esp32c6

package machine

// On the C6 variant, SPI2 is a general purpose SPI controller. SPI0 and SPI1
// are used internally to access the ESP32-C6's attached flash memory. Due to
// different registers between SPI2 and the other SPI ports, this driver
// currently supports only the general purpose FSPI SPI2 controller.

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

// GPIO matrix signal indices for SPI2 on ESP32-C6 (from gpio_sig_map.h).
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

var (
	ErrInvalidSPIBus  = errors.New("machine: SPI bus is invalid")
	ErrInvalidSPIMode = errors.New("machine: SPI mode is invalid")
)

// Serial Peripheral Interface on the ESP32-C6.
type SPI struct {
	Bus *esp.SPI2_Type
}

var (
	// SPI0 and SPI1 are reserved for use by the caching system etc.
	SPI2 = &SPI{esp.SPI2}
	SPI0 = SPI2
)

// Configure and make the SPI peripheral ready to use.
func (spi *SPI) Configure(config SPIConfig) error {
	if spi.Bus != esp.SPI2 {
		return ErrInvalidSPIBus
	}

	// Enable SPI2 clock and reset via PCR.
	esp.PCR.SPI2_CONF.SetBits(1 << 0)  // SPI2_CLK_EN
	esp.PCR.SPI2_CONF.SetBits(1 << 1)  // SPI2_RST_EN
	esp.PCR.SPI2_CONF.ClearBits(1 << 1) // clear reset

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
	switch config.Mode {
	case Mode0:
		spi.Bus.SetMISC_CK_IDLE_EDGE(0)
		spi.Bus.SetUSER_CK_OUT_EDGE(0)
	case Mode1:
		spi.Bus.SetMISC_CK_IDLE_EDGE(0)
		spi.Bus.SetUSER_CK_OUT_EDGE(1)
	case Mode2:
		spi.Bus.SetMISC_CK_IDLE_EDGE(1)
		spi.Bus.SetUSER_CK_OUT_EDGE(1)
	case Mode3:
		spi.Bus.SetMISC_CK_IDLE_EDGE(1)
		spi.Bus.SetUSER_CK_OUT_EDGE(0)
	default:
		return ErrInvalidSPIMode
	}

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

	// configure gpio pin matrix
	config.SDI.Configure(PinConfig{Mode: PinInput})
	inFunc(FSPIQ_IN_IDX).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(config.SDI))
	config.SDO.Configure(PinConfig{Mode: PinOutput})
	config.SDO.outFunc().Set(FSPID_OUT_IDX)
	config.SCK.Configure(PinConfig{Mode: PinOutput})
	config.SCK.outFunc().Set(FSPICLK_OUT_IDX)
	if config.CS != NoPin {
		config.CS.Configure(PinConfig{Mode: PinOutput})
		config.CS.outFunc().Set(FSPICS0_OUT_IDX)
	}

	return nil
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi *SPI) Transfer(w byte) (byte, error) {
	spi.Bus.SetMS_DLEN_MS_DATA_BITLEN(7)

	spi.Bus.SetW0(uint32(w))

	// Send/receive byte.
	spi.Bus.SetCMD_UPDATE(1)
	for spi.Bus.GetCMD_UPDATE() != 0 {
	}

	spi.Bus.SetCMD_USR(1)
	for spi.Bus.GetCMD_USR() != 0 {
	}

	// The received byte is stored in W0.
	return byte(spi.Bus.GetW0()), nil
}

// Tx handles read/write operation for SPI interface.
func (spi *SPI) Tx(w, r []byte) error {
	toTransfer := len(w)
	if len(r) > toTransfer {
		toTransfer = len(r)
	}

	for toTransfer > 0 {
		chunkSize := toTransfer
		if chunkSize > 64 {
			chunkSize = 64
		}

		transferWords := (*[16]volatile.Register32)(unsafe.Pointer(uintptr(unsafe.Pointer(&spi.Bus.W0))))
		spiTxFillBuffer(transferWords, w)

		spi.Bus.SetMS_DLEN_MS_DATA_BITLEN(uint32(chunkSize)*8 - 1)

		spi.Bus.SetCMD_UPDATE(1)
		for spi.Bus.GetCMD_UPDATE() != 0 {
		}

		spi.Bus.SetCMD_USR(1)
		for spi.Bus.GetCMD_USR() != 0 {
		}

		rxSize := 64
		if rxSize > len(r) {
			rxSize = len(r)
		}
		for i := 0; i < rxSize; i++ {
			r[i] = byte(transferWords[i/4].Get() >> ((i % 4) * 8))
		}

		if len(w) < chunkSize {
			w = nil
		} else {
			w = w[chunkSize:]
		}
		if len(r) < chunkSize {
			r = nil
		} else {
			r = r[chunkSize:]
		}
		toTransfer -= chunkSize
	}

	return nil
}
