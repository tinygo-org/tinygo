//go:build esp32c3 || esp32p4

package machine

// spiBus holds the GPIO matrix signals of one SPI bus.
type spiBus struct {
	sckOut, sdoOut, sdiIn, csOut uint32
}

// Configure and make the SPI peripheral ready to use.
func (spi *SPI) Configure(config SPIConfig) error {
	bus, ok := spiBusFor(spi.Bus)
	if !ok {
		return ErrInvalidSPIBus
	}
	// A zero CS is GPIO0, so treat it as unset.
	if config.CS == 0 {
		config.CS = NoPin
	}
	return spi.configure(config, bus)
}
