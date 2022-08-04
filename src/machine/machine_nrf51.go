//go:build nrf51
// +build nrf51

package machine

import (
	"tinygo.org/x/device/nrf"
)

// Get peripheral and pin number for this GPIO pin.
func (p Pin) getPortPin() (*nrf.GPIO_Type, uint32) {
	return nrf.GPIO, uint32(p)
}

func (uart *UART) setPins(tx, rx Pin) {
	nrf.UART0.PSELTXD.Set(uint32(tx))
	nrf.UART0.PSELRXD.Set(uint32(rx))
}

func (i2c *I2C) setPins(scl, sda Pin) {
	i2c.Bus.PSELSCL.Set(uint32(scl))
	i2c.Bus.PSELSDA.Set(uint32(sda))
}

// SPI on the NRF.
type SPI struct {
	Bus *nrf.SPI_Type
}

// There are 2 SPI interfaces on the NRF51.
var (
	SPI0 = SPI{Bus: nrf.SPI0}
	SPI1 = SPI{Bus: nrf.SPI1}
)

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin
	SDO       Pin
	SDI       Pin
	LSBFirst  bool
	Mode      uint8
}

// Configure is intended to setup the SPI interface.
func (spi SPI) Configure(config SPIConfig) {
	// Disable bus to configure it
	spi.Bus.ENABLE.Set(nrf.SPI_ENABLE_ENABLE_Disabled)

	// set frequency
	var freq uint32

	if config.Frequency == 0 {
		config.Frequency = 4000000 // 4MHz
	}

	switch {
	case config.Frequency >= 8000000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_M8
	case config.Frequency >= 4000000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_M4
	case config.Frequency >= 2000000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_M2
	case config.Frequency >= 1000000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_M1
	case config.Frequency >= 500000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_K500
	case config.Frequency >= 250000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_K250
	default: // below 250kHz, default to the lowest speed available
		freq = nrf.SPI_FREQUENCY_FREQUENCY_K125
	}
	spi.Bus.FREQUENCY.Set(freq)

	var conf uint32

	// set bit transfer order
	if config.LSBFirst {
		conf = (nrf.SPI_CONFIG_ORDER_LsbFirst << nrf.SPI_CONFIG_ORDER_Pos)
	}

	// set mode
	switch config.Mode {
	case 0:
		conf &^= (nrf.SPI_CONFIG_CPOL_ActiveHigh << nrf.SPI_CONFIG_CPOL_Pos)
		conf &^= (nrf.SPI_CONFIG_CPHA_Leading << nrf.SPI_CONFIG_CPHA_Pos)
	case 1:
		conf &^= (nrf.SPI_CONFIG_CPOL_ActiveHigh << nrf.SPI_CONFIG_CPOL_Pos)
		conf |= (nrf.SPI_CONFIG_CPHA_Trailing << nrf.SPI_CONFIG_CPHA_Pos)
	case 2:
		conf |= (nrf.SPI_CONFIG_CPOL_ActiveLow << nrf.SPI_CONFIG_CPOL_Pos)
		conf &^= (nrf.SPI_CONFIG_CPHA_Leading << nrf.SPI_CONFIG_CPHA_Pos)
	case 3:
		conf |= (nrf.SPI_CONFIG_CPOL_ActiveLow << nrf.SPI_CONFIG_CPOL_Pos)
		conf |= (nrf.SPI_CONFIG_CPHA_Trailing << nrf.SPI_CONFIG_CPHA_Pos)
	default: // to mode
		conf &^= (nrf.SPI_CONFIG_CPOL_ActiveHigh << nrf.SPI_CONFIG_CPOL_Pos)
		conf &^= (nrf.SPI_CONFIG_CPHA_Leading << nrf.SPI_CONFIG_CPHA_Pos)
	}
	spi.Bus.CONFIG.Set(conf)

	// set pins
	if config.SCK == 0 && config.SDO == 0 && config.SDI == 0 {
		config.SCK = SPI0_SCK_PIN
		config.SDO = SPI0_SDO_PIN
		config.SDI = SPI0_SDI_PIN
	}
	spi.Bus.PSELSCK.Set(uint32(config.SCK))
	spi.Bus.PSELMOSI.Set(uint32(config.SDO))
	spi.Bus.PSELMISO.Set(uint32(config.SDI))

	// Re-enable bus now that it is configured.
	spi.Bus.ENABLE.Set(nrf.SPI_ENABLE_ENABLE_Enabled)
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi SPI) Transfer(w byte) (byte, error) {
	spi.Bus.TXD.Set(uint32(w))
	for spi.Bus.EVENTS_READY.Get() == 0 {
	}
	r := spi.Bus.RXD.Get()
	spi.Bus.EVENTS_READY.Set(0)

	// TODO: handle SPI errors
	return byte(r), nil
}

// Tx handles read/write operation for SPI interface. Since SPI is a syncronous write/read
// interface, there must always be the same number of bytes written as bytes read.
// The Tx method knows about this, and offers a few different ways of calling it.
//
// This form sends the bytes in tx buffer, putting the resulting bytes read into the rx buffer.
// Note that the tx and rx buffers must be the same size:
//
//	spi.Tx(tx, rx)
//
// This form sends the tx buffer, ignoring the result. Useful for sending "commands" that return zeros
// until all the bytes in the command packet have been received:
//
//	spi.Tx(tx, nil)
//
// This form sends zeros, putting the result into the rx buffer. Good for reading a "result packet":
//
//	spi.Tx(nil, rx)
func (spi SPI) Tx(w, r []byte) error {
	var err error

	switch {
	case len(w) == 0:
		// read only, so write zero and read a result.
		for i := range r {
			r[i], err = spi.Transfer(0)
			if err != nil {
				return err
			}
		}
	case len(r) == 0:
		// write only
		spi.Bus.TXD.Set(uint32(w[0]))
		w = w[1:]
		for _, b := range w {
			spi.Bus.TXD.Set(uint32(b))
			for spi.Bus.EVENTS_READY.Get() == 0 {
			}
			spi.Bus.EVENTS_READY.Set(0)
			_ = spi.Bus.RXD.Get()
		}
		for spi.Bus.EVENTS_READY.Get() == 0 {
		}
		spi.Bus.EVENTS_READY.Set(0)
		_ = spi.Bus.RXD.Get()

	default:
		// write/read
		if len(w) != len(r) {
			return ErrTxInvalidSliceSize
		}

		for i, b := range w {
			r[i], err = spi.Transfer(b)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
