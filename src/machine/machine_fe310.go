// +build fe310

package machine

import (
	"device/sifive"
	"runtime/interrupt"
	"unsafe"
)

func CPUFrequency() uint32 {
	return 320000000 // 320MHz
}

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
	return (val > 0)
}

type UART struct {
	Bus    *sifive.UART_Type
	Buffer *RingBuffer
}

var (
	UART0  = &_UART0
	_UART0 = UART{Bus: sifive.UART0, Buffer: NewRingBuffer()}
)

func (uart *UART) Configure(config UARTConfig) {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	// The divisor is:
	//   fbaud = fin / (div + 1)
	// Restating to get the divisor:
	//   div = fin / fbaud - 1
	// But we're using integers, so we should take care of rounding:
	//   div = (fin + fbaud/2) / fbaud - 1
	divisor := (CPUFrequency()+config.BaudRate/2)/config.BaudRate - 1
	sifive.UART0.DIV.Set(divisor)
	sifive.UART0.TXCTRL.Set(sifive.UART_TXCTRL_ENABLE)
	sifive.UART0.RXCTRL.Set(sifive.UART_RXCTRL_ENABLE)
	sifive.UART0.IE.Set(sifive.UART_IE_RXWM) // enable the receive interrupt (only)
	intr := interrupt.New(sifive.IRQ_UART0, _UART0.handleInterrupt)
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

func (uart *UART) WriteByte(c byte) {
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
	SDO       Pin
	SDI       Pin
	LSBFirst  bool
	Mode      uint8
}

// Configure is intended to setup the SPI interface.
func (spi SPI) Configure(config SPIConfig) error {
	// Use default pins if not set.
	if config.SCK == 0 && config.SDO == 0 && config.SDI == 0 {
		config.SCK = SPI0_SCK_PIN
		config.SDO = SPI0_SDO_PIN
		config.SDI = SPI0_SDI_PIN
	}

	// enable pins for SPI
	config.SCK.Configure(PinConfig{Mode: PinSPI})
	config.SDO.Configure(PinConfig{Mode: PinSPI})
	config.SDI.Configure(PinConfig{Mode: PinSPI})

	// set default frequency
	if config.Frequency == 0 {
		config.Frequency = 4000000 // 4MHz
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
	data := spi.Bus.RXDATA.Get()
	for data&sifive.QSPI_RXDATA_EMPTY > 0 {
		data = spi.Bus.RXDATA.Get()
	}

	// return data
	return byte(data), nil
}

// I2C on the FE310-G002.
type I2C struct {
	Bus sifive.I2C_Type
}

var (
	I2C0 = (*I2C)(unsafe.Pointer(sifive.I2C0))
)

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

// Configure is intended to setup the I2C interface.
func (i2c *I2C) Configure(config I2CConfig) error {
	var i2cClockFrequency uint32 = 32000000
	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}

	if config.SDA == 0 && config.SCL == 0 {
		config.SDA = I2C0_SDA_PIN
		config.SCL = I2C0_SCL_PIN
	}

	var prescaler = i2cClockFrequency/(5*config.Frequency) - 1

	// disable controller before setting the prescale registers
	i2c.Bus.CTR.ClearBits(sifive.I2C_CTR_EN)

	// set prescaler registers
	i2c.Bus.PRER_LO.Set(uint32(prescaler & 0xff))
	i2c.Bus.PRER_HI.Set(uint32((prescaler >> 8) & 0xff))

	// enable controller
	i2c.Bus.CTR.SetBits(sifive.I2C_CTR_EN)

	config.SDA.Configure(PinConfig{Mode: PinI2C})
	config.SCL.Configure(PinConfig{Mode: PinI2C})

	return nil
}

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c *I2C) Tx(addr uint16, w, r []byte) error {
	var err error
	if len(w) != 0 {
		// send start/address for write
		i2c.sendAddress(addr, true)

		// ACK received (0: ACK, 1: NACK)
		if i2c.Bus.CR_SR.HasBits(sifive.I2C_SR_RX_ACK) {
			return errI2CAckExpected
		}

		// write data
		for _, b := range w {
			err = i2c.writeByte(b)
			if err != nil {
				return err
			}
		}
	}
	if len(r) != 0 {
		// send start/address for read
		i2c.sendAddress(addr, false)

		// ACK received (0: ACK, 1: NACK)
		if i2c.Bus.CR_SR.HasBits(sifive.I2C_SR_RX_ACK) {
			return errI2CAckExpected
		}

		// read first byte
		r[0] = i2c.readByte()
		for i := 1; i < len(r); i++ {
			// send an ACK
			i2c.Bus.CR_SR.Set(^uint32(sifive.I2C_CR_ACK))

			// read data and send the ACK
			r[i] = i2c.readByte()
		}

		// send NACK to end transmission
		i2c.Bus.CR_SR.Set(sifive.I2C_CR_ACK)
	}

	// generate stop condition
	i2c.Bus.CR_SR.Set(sifive.I2C_CR_STO)
	return nil
}

// Writes a single byte to the I2C bus.
func (i2c *I2C) writeByte(data byte) error {
	// Send data byte
	i2c.Bus.TXR_RXR.Set(uint32(data))

	i2c.Bus.CR_SR.Set(sifive.I2C_CR_WR)

	// wait until transmission complete
	for i2c.Bus.CR_SR.HasBits(sifive.I2C_SR_TIP) {
	}

	// ACK received (0: ACK, 1: NACK)
	if i2c.Bus.CR_SR.HasBits(sifive.I2C_SR_RX_ACK) {
		return errI2CAckExpected
	}

	return nil
}

// Reads a single byte from the I2C bus.
func (i2c *I2C) readByte() byte {
	i2c.Bus.CR_SR.Set(sifive.I2C_CR_RD)

	// wait until transmission complete
	for i2c.Bus.CR_SR.HasBits(sifive.I2C_SR_TIP) {
	}

	return byte(i2c.Bus.TXR_RXR.Get())
}

// Sends the address and start signal.
func (i2c *I2C) sendAddress(address uint16, write bool) error {
	data := (address << 1)
	if !write {
		data |= 1 // set read flag in transmit register
	}

	// write address to transmit register
	i2c.Bus.TXR_RXR.Set(uint32(data))

	// generate start condition
	i2c.Bus.CR_SR.Set((sifive.I2C_CR_STA | sifive.I2C_CR_WR))

	// wait until transmission complete
	for i2c.Bus.CR_SR.HasBits(sifive.I2C_SR_TIP) {
	}

	return nil
}
