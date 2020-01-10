// +build fe310

package machine

import (
	"device/sifive"
	"runtime/interrupt"
)

func CPUFrequency() uint32 {
	return 16000000
}

type PinMode uint8

const (
	PinInput PinMode = iota
	PinOutput
	PinPWM
	PinSPI
	PinI2C = PinSPI
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	sifive.GPIO0.INPUT_EN.SetBits(1 << uint8(p))
	switch config.Mode {
	case PinOutput:
		sifive.GPIO0.OUTPUT_EN.SetBits(1 << uint8(p))
	case PinPWM:
		sifive.GPIO0.IOF_EN.SetBits(1 << uint8(p))
		sifive.GPIO0.IOF_SEL.SetBits(1 << uint8(p))
	case PinSPI:
		sifive.GPIO0.IOF_EN.SetBits(1 << uint8(p))
		sifive.GPIO0.IOF_SEL.ClearBits(1 << uint8(p))
	}
}

// Set the pin to high or low.
func (p Pin) Set(high bool) {
	if high {
		sifive.GPIO0.PORT.SetBits(1 << uint8(p))
	} else {
		sifive.GPIO0.PORT.ClearBits(1 << uint8(p))
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	val := sifive.GPIO0.VALUE.Get() & (1 << uint8(p))
	println(sifive.GPIO0.VALUE.Get())
	return (val > 0)
}

type UART struct {
	Bus    *sifive.UART_Type
	Buffer *RingBuffer
}

var (
	UART0 = UART{Bus: sifive.UART0, Buffer: NewRingBuffer()}
)

func (uart UART) Configure(config UARTConfig) {
	// Assuming a 16Mhz Crystal (which is Y1 on the HiFive1), the divisor for a
	// 115200 baud rate is 138.
	sifive.UART0.DIV.Set(138)
	sifive.UART0.TXCTRL.Set(sifive.UART_TXCTRL_ENABLE)
	sifive.UART0.RXCTRL.Set(sifive.UART_RXCTRL_ENABLE)
	sifive.UART0.IE.Set(sifive.UART_IE_RXWM) // enable the receive interrupt (only)
	intr := interrupt.New(sifive.IRQ_UART0, UART0.handleInterrupt)
	intr.SetPriority(5)
	intr.Enable()
}

func (uart *UART) handleInterrupt(interrupt.Interrupt) {
	rxdata := uart.Bus.RXDATA.Get()
	c := byte(rxdata)
	if uint32(c) != rxdata {
		// The rxdata has other bits set than just the low 8 bits. This probably
		// means that the 'empty' flag is set, which indicates there is no data
		// to be read and the byte is garbage. Ignore this byte.
		return
	}
	uart.Receive(c)
}

func (uart UART) WriteByte(c byte) {
	for sifive.UART0.TXDATA.Get()&sifive.UART_TXDATA_FULL != 0 {
	}

	sifive.UART0.TXDATA.Set(uint32(c))
}

// SPI on the FE310. The normal SPI0 is actually a quad-SPI meant for flash, so it is best
// to use SPI1 or SPI2 port for most applications.
type SPI struct {
	Bus *sifive.QSPI_Type
}

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin
	MOSI      Pin
	MISO      Pin
	LSBFirst  bool
	Mode      uint8
}

// Configure is intended to setup the SPI interface.
func (spi SPI) Configure(config SPIConfig) error {
	// Use default pins if not set.
	if config.SCK == 0 && config.MOSI == 0 && config.MISO == 0 {
		config.SCK = SPI0_SCK_PIN
		config.MOSI = SPI0_MOSI_PIN
		config.MISO = SPI0_MISO_PIN
	}

	// enable pins for SPI
	config.SCK.Configure(PinConfig{Mode: PinSPI})
	config.MOSI.Configure(PinConfig{Mode: PinSPI})
	config.MISO.Configure(PinConfig{Mode: PinSPI})

	// set default frequency
	if config.Frequency == 0 {
		config.Frequency = 4000000
	}

	// div = (SPI_CFG(dev)->f_sys / (2 * frequency)) - 1;
	div := CPUFrequency()/(2*config.Frequency) - 1
	spi.Bus.DIV.Set(div)

	// set mode
	switch config.Mode {
	case 0:
		spi.Bus.MODE.ClearBits(sifive.QSPI_MODE_PHASE)
		spi.Bus.MODE.ClearBits(sifive.QSPI_MODE_POLARITY)
	case 1:
		spi.Bus.MODE.SetBits(sifive.QSPI_MODE_PHASE)
		spi.Bus.MODE.ClearBits(sifive.QSPI_MODE_POLARITY)
	case 2:
		spi.Bus.MODE.ClearBits(sifive.QSPI_MODE_PHASE)
		spi.Bus.MODE.SetBits(sifive.QSPI_MODE_POLARITY)
	case 3:
		spi.Bus.MODE.SetBits(sifive.QSPI_MODE_PHASE | sifive.QSPI_MODE_POLARITY)
	default: // to mode 0
		spi.Bus.MODE.ClearBits(sifive.QSPI_MODE_PHASE)
		spi.Bus.MODE.ClearBits(sifive.QSPI_MODE_POLARITY)
	}

	// frame length
	spi.Bus.FMT.SetBits(8 << sifive.QSPI_FMT_LENGTH_Pos)

	// Set single line operation, by clearing all bits
	spi.Bus.FMT.ClearBits(sifive.QSPI_FMT_PROTOCOL_Msk)

	// set bit transfer order
	if config.LSBFirst {
		spi.Bus.FMT.SetBits(sifive.QSPI_FMT_ENDIAN)
	} else {
		spi.Bus.FMT.ClearBits(sifive.QSPI_FMT_ENDIAN)
	}

	return nil
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi SPI) Transfer(w byte) (byte, error) {
	// wait for tx ready
	for spi.Bus.TXDATA.HasBits(sifive.QSPI_TXDATA_FULL) {
	}

	// write data
	spi.Bus.TXDATA.Set(uint32(w))

	// wait until receive has data
	for spi.Bus.RXDATA.HasBits(sifive.QSPI_RXDATA_EMPTY) {
	}

	// return data
	return byte(spi.Bus.RXDATA.Get() & sifive.QSPI_RXDATA_DATA_Msk), nil
}
