//go:build esp32c3

package machine

// On the C3 variant, SPI2 is a general purpose SPI controller. SPI0 and SPI1
// are used internally to access the ESP32-C3’s attached flash memory. Due to
// different registers between SPI2 and the other SPI ports, this driver
// currently supports only the the general purpose FSPI SPI2 controller.
// https://docs.espressif.com/projects/esp-idf/en/latest/esp32c3/api-reference/peripherals/spi_master.html

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

const (
	SPI_MODE0 = uint8(0)
	SPI_MODE1 = uint8(1)
	SPI_MODE2 = uint8(2)
	SPI_MODE3 = uint8(3)

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

// Serial Peripheral Interface on the ESP32-C3.
type SPI struct {
	Bus *esp.SPI2_Type
}

var (
	// SPI0 and SPI1 are reserved for use by the caching system etc.
	SPI2 = SPI{esp.SPI2}
)

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin   // Serial Clock
	SDO       Pin   // Serial Data Out (MOSI)
	SDI       Pin   // Serial Data In  (MISO)
	CS        Pin   // Chip Select (optional)
	LSBFirst  bool  // MSB is default
	Mode      uint8 // SPI_MODE0 is default
}

// Compute the SPI bus frequency from the CPU frequency.
func freqToClockDiv(hz uint32) uint32 {
	fcpu := CPUFrequency()
	if hz >= fcpu { // maximum frequency
		return 1 << 31
	}
	if hz < (fcpu / (16 * 64)) { // minimum frequency
		return 15<<18 | 63<<12 | 31<<6 | 63 // pre=15, n=63
	}

	// iterate looking for an exact match
	// or iterate all 16 prescaler options
	// looking for the smallest error
	var bestPre, bestN, bestErr uint32
	bestN = 1
	bestErr = 0xffffffff
	q := uint32(float32(pplClockFreq)/float32(hz) + float32(0.5))
	for p := uint32(0); p < 16; p++ {
		n := q/(p+1) - 1
		if n < 1 { // prescaler became too large, stop enum
			break
		}
		if n > 63 { // prescaler too small, skip to next
			continue
		}

		freq := fcpu / ((p + 1) * (n + 1))
		if freq == hz { // exact match
			return p<<18 | n<<12 | (n/2)<<6 | n
		}

		var err uint32
		if freq < hz {
			err = hz - freq
		} else {
			err = freq - hz
		}
		if err < bestErr {
			bestErr = err
			bestPre = p
			bestN = n
		}
	}

	return bestPre<<18 | bestN<<12 | (bestN/2)<<6 | bestN
}

// Configure and make the SPI peripheral ready to use.
func (spi SPI) Configure(config SPIConfig) error {
	// right now this is only setup to work for the esp32c3 spi2 bus
	if spi.Bus != esp.SPI2 {
		return ErrInvalidSPIBus
	}

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
	switch config.Mode {
	case SPI_MODE0:
		spi.Bus.SetMISC_CK_IDLE_EDGE(0)
		spi.Bus.SetUSER_CK_OUT_EDGE(0)
	case SPI_MODE1:
		spi.Bus.SetMISC_CK_IDLE_EDGE(0)
		spi.Bus.SetUSER_CK_OUT_EDGE(1)
	case SPI_MODE2:
		spi.Bus.SetMISC_CK_IDLE_EDGE(1)
		spi.Bus.SetUSER_CK_OUT_EDGE(1)
	case SPI_MODE3:
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

	// configure esp32c3 gpio pin matrix
	config.SDI.Configure(PinConfig{Mode: PinInput})
	inFunc(FSPIQ_IN_IDX).Set(esp.GPIO_FUNC_IN_SEL_CFG_SIG_IN_SEL | uint32(config.SDI))
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

// Transfer writes/reads a single byte using the SPI interface. If you need to
// transfer larger amounts of data, Tx will be faster.
func (spi SPI) Transfer(w byte) (byte, error) {
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

// Tx handles read/write operation for SPI interface. Since SPI is a syncronous write/read
// interface, there must always be the same number of bytes written as bytes read.
// This is accomplished by sending zero bits if r is bigger than w or discarding
// the incoming data if w is bigger than r.
func (spi SPI) Tx(w, r []byte) error {
	toTransfer := len(w)
	if len(r) > toTransfer {
		toTransfer = len(r)
	}

	for toTransfer > 0 {
		// Chunk 64 bytes at a time.
		chunkSize := toTransfer
		if chunkSize > 64 {
			chunkSize = 64
		}

		// Fill tx buffer.
		transferWords := (*[16]volatile.Register32)(unsafe.Pointer(uintptr(unsafe.Pointer(&spi.Bus.W0))))
		if len(w) >= 64 {
			// We can fill the entire 64-byte transfer buffer with data.
			// This loop is slightly faster than the loop below.
			for i := 0; i < 16; i++ {
				word := uint32(w[i*4]) | uint32(w[i*4+1])<<8 | uint32(w[i*4+2])<<16 | uint32(w[i*4+3])<<24
				transferWords[i].Set(word)
			}
		} else {
			// We can't fill the entire transfer buffer, so we need to be a bit
			// more careful.
			// Note that parts of the transfer buffer that aren't used still
			// need to be set to zero, otherwise we might be transferring
			// garbage from a previous transmission if w is smaller than r.
			for i := 0; i < 16; i++ {
				var word uint32
				if i*4+3 < len(w) {
					word |= uint32(w[i*4+3]) << 24
				}
				if i*4+2 < len(w) {
					word |= uint32(w[i*4+2]) << 16
				}
				if i*4+1 < len(w) {
					word |= uint32(w[i*4+1]) << 8
				}
				if i*4+0 < len(w) {
					word |= uint32(w[i*4+0]) << 0
				}
				transferWords[i].Set(word)
			}
		}

		// Do the transfer.
		spi.Bus.SetMS_DLEN_MS_DATA_BITLEN(uint32(chunkSize)*8 - 1)

		spi.Bus.SetCMD_UPDATE(1)
		for spi.Bus.GetCMD_UPDATE() != 0 {
		}

		spi.Bus.SetCMD_USR(1)
		for spi.Bus.GetCMD_USR() != 0 {
		}

		// Read rx buffer.
		rxSize := 64
		if rxSize > len(r) {
			rxSize = len(r)
		}
		for i := 0; i < rxSize; i++ {
			r[i] = byte(transferWords[i/4].Get() >> ((i % 4) * 8))
		}

		// Cut off some part of the output buffer so the next iteration we will
		// only send the remaining bytes.
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
