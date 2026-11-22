//go:build esp32c3 || esp32p4

package machine

import "device/esp"

// spiBus holds the GPIO matrix signals of one SPI bus.
type spiBus struct {
	sckOut, sdoOut, sdiIn, csOut uint32
}

// Configure configures the SPI bus for use.
func (spi *SPI) Configure(config SPIConfig) error {
	bus, ok := spiBusFor(spi.Bus)
	if !ok {
		return ErrInvalidSPIBus
	}
	if config.Frequency == 0 {
		config.Frequency = 4 * MHz
	}
	// A zero CS is GPIO0, so treat it as unset.
	if config.CS == 0 {
		config.CS = NoPin
	}
	mode := spiModeFor(config.Mode)
	if mode&spiModeValid == 0 {
		return ErrInvalidSPIMode
	}
	enableSPIClock(spi.Bus)
	if err := spi.configure(config, bus, mode); err != nil {
		return err
	}
	// Route the pins through the GPIO matrix.
	if config.SDI != NoPin {
		config.SDI.Configure(PinConfig{Mode: PinInput})
		inFunc(bus.sdiIn).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(config.SDI))
	}
	if config.SDO != NoPin {
		config.SDO.Configure(PinConfig{Mode: PinOutput})
		config.SDO.outFunc().Set(bus.sdoOut)
	}
	if config.SCK != NoPin {
		config.SCK.Configure(PinConfig{Mode: PinOutput})
		config.SCK.outFunc().Set(bus.sckOut)
	}
	if config.CS != NoPin {
		config.CS.Configure(PinConfig{Mode: PinOutput})
		config.CS.outFunc().Set(bus.csOut)
	}
	return nil
}

// Transfer writes and reads a single byte.
func (spi *SPI) Transfer(w byte) (byte, error) {
	buf := spi.buffer()
	spi.setBitLen(7)
	buf[0].Set(uint32(w))
	spi.start()
	return byte(buf[0].Get()), nil
}

// Tx writes w and reads into r. The shorter slice is padded with zeros or
// its extra input is discarded.
func (spi *SPI) Tx(w, r []byte) error {
	toTransfer := len(w)
	if len(r) > toTransfer {
		toTransfer = len(r)
	}
	buf := spi.buffer()
	for toTransfer > 0 {
		// Chunk 64 bytes at a time.
		chunkSize := toTransfer
		if chunkSize > 64 {
			chunkSize = 64
		}
		spiTxFillBuffer(buf, w)
		spi.setBitLen(uint32(chunkSize)*8 - 1)
		spi.start()

		// Read rx buffer.
		rxSize := 64
		if rxSize > len(r) {
			rxSize = len(r)
		}
		for i := 0; i < rxSize; i++ {
			r[i] = byte(buf[i/4].Get() >> ((i % 4) * 8))
		}

		// Drop the bytes that were sent and received.
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
