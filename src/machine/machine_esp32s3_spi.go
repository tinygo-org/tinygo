//go:build esp32s3

package machine

// ESP32-S3 SPI support based on ESP-IDF HAL
// Simple but correct implementation following spi_ll.h
// SPI0 = hardware SPI2 (FSPI), SPI1 = hardware SPI3 (HSPI)
// https://docs.espressif.com/projects/esp-idf/en/latest/esp32s3/api-reference/peripherals/spi_master.html

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

	// ESP32-S3 PLL clock frequency (same as ESP32-C3)
	pplClockFreq = 80e6

	// Default SPI frequency - maximum safe speed
	SPI_DEFAULT_FREQUENCY = 80e6 // 80MHz
)

// ESP32-S3 default SPI pins that support IO MUX direct connection
const (
	// SPI2 (FSPI) default pins - support IO MUX function 4
	SPI2_DEFAULT_SCK  = GPIO12 // SCK
	SPI2_DEFAULT_MOSI = GPIO11 // SDO (MOSI)
	SPI2_DEFAULT_MISO = GPIO13 // SDI (MISO)
	SPI2_DEFAULT_CS   = GPIO10 // CS

	// SPI3 (HSPI) default pins - support IO MUX function 4
	SPI3_DEFAULT_SCK  = GPIO36 // SCK
	SPI3_DEFAULT_MOSI = GPIO35 // SDO (MOSI)
	SPI3_DEFAULT_MISO = GPIO37 // SDI (MISO)
	SPI3_DEFAULT_CS   = GPIO34 // CS

	// IO MUX function number for SPI direct connection
	SPI_IOMUX_FUNC = 4
)

// ESP32-S3 GPIO Matrix signal indices for SPI - CORRECTED from ESP-IDF gpio_sig_map.h
const (
	// SPI2 (FSPI) signals - Hardware SPI2 - CORRECT VALUES from ESP-IDF
	SPI2_CLK_OUT_IDX = uint32(101) // FSPICLK_OUT_IDX
	SPI2_CLK_IN_IDX  = uint32(101) // FSPICLK_IN_IDX
	SPI2_Q_OUT_IDX   = uint32(102) // FSPIQ_OUT_IDX (MISO)
	SPI2_Q_IN_IDX    = uint32(102) // FSPIQ_IN_IDX
	SPI2_D_OUT_IDX   = uint32(103) // FSPID_OUT_IDX (MOSI)
	SPI2_D_IN_IDX    = uint32(103) // FSPID_IN_IDX
	SPI2_CS0_OUT_IDX = uint32(110) // FSPICS0_OUT_IDX

	// SPI3 (HSPI) signals - Hardware SPI3 - CORRECTED from ESP-IDF gpio_sig_map.h
	// Source: /esp-idf/components/soc/esp32s3/include/soc/gpio_sig_map.h
	SPI3_CLK_OUT_IDX = uint32(66) // Line 136: SPI3_CLK_OUT_IDX
	SPI3_CLK_IN_IDX  = uint32(66) // Line 135: SPI3_CLK_IN_IDX
	SPI3_Q_OUT_IDX   = uint32(67) // Line 138: SPI3_Q_OUT_IDX (MISO)
	SPI3_Q_IN_IDX    = uint32(67) // Line 137: SPI3_Q_IN_IDX
	SPI3_D_OUT_IDX   = uint32(68) // Line 140: SPI3_D_OUT_IDX (MOSI)
	SPI3_D_IN_IDX    = uint32(68) // Line 139: SPI3_D_IN_IDX
	SPI3_CS0_OUT_IDX = uint32(71) // Line 146: SPI3_CS0_OUT_IDX
)

// Serial Peripheral Interface on the ESP32-S3.
type SPI struct {
	Bus   interface{}
	busID uint8
}

var (
	SPI0 = &SPI{Bus: esp.SPI2, busID: 2} // Primary SPI (FSPI)
	SPI1 = &SPI{Bus: esp.SPI3, busID: 3} // Secondary SPI (HSPI)
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

// Configure and make the SPI peripheral ready to use.
// Implementation following ESP-IDF HAL with GPIO Matrix routing
func (spi *SPI) Configure(config SPIConfig) error {

	// Set default frequency if not specified
	if config.Frequency == 0 {
		config.Frequency = SPI_DEFAULT_FREQUENCY // Default to maximum safe speed
	}

	// Get GPIO Matrix signal indices for this SPI bus
	var sckOutIdx, mosiOutIdx, misoInIdx, csOutIdx uint32
	switch spi.busID {
	case 2: // SPI2 (FSPI)
		sckOutIdx = SPI2_CLK_OUT_IDX
		mosiOutIdx = SPI2_D_OUT_IDX
		misoInIdx = SPI2_Q_IN_IDX
		csOutIdx = SPI2_CS0_OUT_IDX
	case 3: // SPI3 (HSPI)
		sckOutIdx = SPI3_CLK_OUT_IDX
		mosiOutIdx = SPI3_D_OUT_IDX
		misoInIdx = SPI3_Q_IN_IDX
		csOutIdx = SPI3_CS0_OUT_IDX
	default:
		return ErrInvalidSPIBus
	}

	// Check if we can use IO MUX direct connection for better performance
	if isDefaultSPIPins(spi.busID, config) {
		// Use IO MUX direct connection - better signal quality and performance
		// Configure pins using IO MUX direct connection (SPI function)
		if config.SCK != NoPin {
			config.SCK.configure(PinConfig{Mode: PinOutput}, SPI_IOMUX_FUNC)
		}
		if config.SDO != NoPin {
			config.SDO.configure(PinConfig{Mode: PinOutput}, SPI_IOMUX_FUNC)
		}
		if config.SDI != NoPin {
			config.SDI.configure(PinConfig{Mode: PinInput}, SPI_IOMUX_FUNC)
		}
		if config.CS != NoPin {
			config.CS.configure(PinConfig{Mode: PinOutput}, SPI_IOMUX_FUNC)
		}
	} else {
		// Use GPIO Matrix routing - more flexible but slightly slower
		// Configure SDI (MISO) pin
		if config.SDI != NoPin {
			config.SDI.Configure(PinConfig{Mode: PinInput})
			inFunc(misoInIdx).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(config.SDI))
		}

		// Configure SDO (MOSI) pin
		if config.SDO != NoPin {
			config.SDO.Configure(PinConfig{Mode: PinOutput})
			config.SDO.outFunc().Set(mosiOutIdx)
		}

		// Configure SCK (Clock) pin
		if config.SCK != NoPin {
			config.SCK.Configure(PinConfig{Mode: PinOutput})
			config.SCK.outFunc().Set(sckOutIdx)
		}

		// Configure CS (Chip Select) pin
		if config.CS != NoPin {
			config.CS.Configure(PinConfig{Mode: PinOutput})
			config.CS.outFunc().Set(csOutIdx)
		}
	}

	// Enable peripheral clock and reset
	// Without bootloader, we need to be more explicit about clock initialization
	switch spi.busID {
	case 2: // Hardware SPI2 (FSPI)
		esp.SYSTEM.SetPERIP_CLK_EN0_SPI2_CLK_EN(1)
		esp.SYSTEM.SetPERIP_RST_EN0_SPI2_RST(1)
		esp.SYSTEM.SetPERIP_RST_EN0_SPI2_RST(0)
	case 3: // Hardware SPI3 (HSPI)
		esp.SYSTEM.SetPERIP_CLK_EN0_SPI3_CLK_EN(1)
		esp.SYSTEM.SetPERIP_RST_EN0_SPI3_RST(1)
		esp.SYSTEM.SetPERIP_RST_EN0_SPI3_RST(0)
	}

	// Get bus handle - both SPI2 and SPI3 use SPI2_Type
	bus, ok := spi.Bus.(*esp.SPI2_Type)
	if !ok {
		return ErrInvalidSPIBus
	}

	// Reset timing: cs_setup_time = 0, cs_hold_time = 0
	bus.USER1.Set(0)

	// Use all 64 bytes of the buffer
	bus.SetUSER_USR_MISO_HIGHPART(0)
	bus.SetUSER_USR_MOSI_HIGHPART(0)

	// Disable unneeded interrupts and clear all USER bits first
	bus.SLAVE.Set(0)
	bus.USER.Set(0)

	// Clear other important registers like ESP32-C3
	bus.MISC.Set(0)
	bus.CTRL.Set(0)
	bus.CLOCK.Set(0)

	// Clear data buffers like ESP32-C3
	bus.W0.Set(0)
	bus.W1.Set(0)
	bus.W2.Set(0)
	bus.W3.Set(0)

	// Configure master clock gate - CRITICAL: need CLK_EN bit!
	bus.SetCLK_GATE_CLK_EN(1)         // Enable basic SPI clock (bit 0)
	bus.SetCLK_GATE_MST_CLK_ACTIVE(1) // Enable master clock (bit 1)
	bus.SetCLK_GATE_MST_CLK_SEL(1)    // Select master clock (bit 2)

	// Configure DMA following ESP-IDF HAL
	// Reset DMA configuration
	bus.DMA_CONF.Set(0)
	// Set DMA segment transaction clear enable bits
	bus.SetDMA_CONF_SLV_TX_SEG_TRANS_CLR_EN(1)
	bus.SetDMA_CONF_SLV_RX_SEG_TRANS_CLR_EN(1)
	// dma_seg_trans_en = 0 (already 0 from DMA_CONF.Set(0))

	// Configure master mode
	bus.SetUSER_USR_MOSI(1)     // Enable MOSI
	bus.SetUSER_USR_MISO(1)     // Enable MISO
	bus.SetUSER_DOUTDIN(1)      // Full-duplex mode
	bus.SetCTRL_WR_BIT_ORDER(0) // MSB first
	bus.SetCTRL_RD_BIT_ORDER(0) // MSB first

	// CRITICAL: Enable clock output (from working test)
	bus.SetMISC_CK_DIS(0) // Enable CLK output - THIS IS KEY!

	// Configure SPI mode (CPOL/CPHA) following ESP-IDF HAL
	switch config.Mode {
	case SPI_MODE0:
		// CPOL=0, CPHA=0 (default)
	case SPI_MODE1:
		bus.SetUSER_CK_OUT_EDGE(1) // CPHA=1
	case SPI_MODE2:
		bus.SetMISC_CK_IDLE_EDGE(1) // CPOL=1
		bus.SetUSER_CK_OUT_EDGE(1)  // CPHA=1
	case SPI_MODE3:
		bus.SetMISC_CK_IDLE_EDGE(1) // CPOL=1
	}

	// Configure SPI bus clock using ESP32-C3 algorithm for better accuracy
	bus.CLOCK.Set(freqToClockDiv(config.Frequency))

	return nil
}

// Transfer writes/reads a single byte using the SPI interface.
// Implementation following ESP-IDF HAL spi_ll_user_start with proper USER register setup
func (spi *SPI) Transfer(w byte) (byte, error) {
	// Both SPI2 and SPI3 use SPI2_Type
	bus, ok := spi.Bus.(*esp.SPI2_Type)
	if !ok {
		return 0, errors.New("invalid SPI bus type")
	}

	// Set transfer length (8 bits = 7 in register)
	bus.SetMS_DLEN_MS_DATA_BITLEN(7)

	// Clear any pending interrupt flags BEFORE starting transaction
	bus.SetDMA_INT_CLR_TRANS_DONE_INT_CLR(1)

	// Write data to buffer (use W0 register)
	bus.W0.Set(uint32(w))

	// CRITICAL: Apply configuration before transmission (like ESP-IDF spi_ll_apply_config)
	bus.SetCMD_UPDATE(1)
	for bus.GetCMD_UPDATE() != 0 {
		// Wait for config to be applied
	}

	// Start transaction following ESP-IDF HAL spi_ll_user_start
	bus.SetCMD_USR(1)

	// Wait for completion using CMD_USR flag (like ESP32-C3 approach)
	// Hardware clears CMD_USR when transaction is complete
	timeout := 100000
	for bus.GetCMD_USR() != 0 && timeout > 0 {
		timeout--
		// Wait for CMD_USR to be cleared by hardware
	}

	if timeout == 0 {
		return 0, errors.New("SPI transfer timeout")
	}

	// Read received data from W0 register
	result := byte(bus.W0.Get() & 0xFF)
	return result, nil
}

// Tx handles read/write operation for SPI interface. Since SPI is a synchronous write/read
// interface, there must always be the same number of bytes written as bytes read.
// This is accomplished by sending zero bits if r is bigger than w or discarding
// the incoming data if w is bigger than r.
// Optimized implementation ported from ESP32-C3 for better performance.
func (spi *SPI) Tx(w, r []byte) error {
	toTransfer := len(w)
	if len(r) > toTransfer {
		toTransfer = len(r)
	}

	// Get bus handle - both SPI2 and SPI3 use SPI2_Type
	bus, ok := spi.Bus.(*esp.SPI2_Type)
	if !ok {
		return ErrInvalidSPIBus
	}

	for toTransfer > 0 {
		// Chunk 64 bytes at a time.
		chunkSize := toTransfer
		if chunkSize > 64 {
			chunkSize = 64
		}

		// Fill tx buffer.
		transferWords := (*[16]volatile.Register32)(unsafe.Add(unsafe.Pointer(&bus.W0), 0))
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
		bus.SetMS_DLEN_MS_DATA_BITLEN(uint32(chunkSize)*8 - 1)

		bus.SetCMD_UPDATE(1)
		for bus.GetCMD_UPDATE() != 0 {
		}

		bus.SetCMD_USR(1)
		for bus.GetCMD_USR() != 0 {
		}

		// Read rx buffer.
		rxSize := chunkSize
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

// Compute the SPI bus frequency from the APB clock frequency.
// Note: APB clock is always 80MHz on ESP32-S3, independent of CPU frequency.
// Ported from ESP32-C3 implementation for better accuracy.
func freqToClockDiv(hz uint32) uint32 {
	// Use APB clock frequency (80MHz), not CPU frequency!
	// SPI peripheral is connected to APB bus which stays at 80MHz
	const apbFreq = pplClockFreq // 80MHz

	if hz >= apbFreq { // maximum frequency
		return 1 << 31
	}
	if hz < (apbFreq / (16 * 64)) { // minimum frequency
		return 15<<18 | 63<<12 | 31<<6 | 63 // pre=15, n=63
	}

	// iterate looking for an exact match
	// or iterate all 16 prescaler options
	// looking for the smallest error
	var bestPre, bestN, bestErr uint32
	bestN = 1
	bestErr = 0xffffffff
	q := uint32(float32(apbFreq)/float32(hz) + float32(0.5))
	for p := uint32(0); p < 16; p++ {
		n := q/(p+1) - 1
		if n < 1 { // prescaler became too large, stop enum
			break
		}
		if n > 63 { // prescaler too small, skip to next
			continue
		}

		freq := apbFreq / ((p + 1) * (n + 1))
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

// isDefaultSPIPins checks if the given pins match the default SPI pin configuration
// that supports IO MUX direct connection for better performance
func isDefaultSPIPins(busID uint8, config SPIConfig) bool {
	switch busID {
	case 2: // SPI2 (FSPI)
		return config.SCK == SPI2_DEFAULT_SCK &&
			config.SDO == SPI2_DEFAULT_MOSI &&
			config.SDI == SPI2_DEFAULT_MISO &&
			(config.CS == SPI2_DEFAULT_CS || config.CS == NoPin)
	case 3: // SPI3 (HSPI)
		return config.SCK == SPI3_DEFAULT_SCK &&
			config.SDO == SPI3_DEFAULT_MOSI &&
			config.SDI == SPI3_DEFAULT_MISO &&
			(config.CS == SPI3_DEFAULT_CS || config.CS == NoPin)
	default:
		return false
	}
}
