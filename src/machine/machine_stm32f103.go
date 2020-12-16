// +build stm32,stm32f103

package machine

// Peripheral abstraction layer for the stm32.

import (
	"device/stm32"
	"unsafe"
)

func CPUFrequency() uint32 {
	return 72000000
}

const (
	PinInput       PinMode = 0 // Input mode
	PinOutput10MHz PinMode = 1 // Output mode, max speed 10MHz
	PinOutput2MHz  PinMode = 2 // Output mode, max speed 2MHz
	PinOutput50MHz PinMode = 3 // Output mode, max speed 50MHz
	PinOutput      PinMode = PinOutput2MHz

	PinInputModeAnalog     PinMode = 0  // Input analog mode
	PinInputModeFloating   PinMode = 4  // Input floating mode
	PinInputModePullUpDown PinMode = 8  // Input pull up/down mode
	PinInputModeReserved   PinMode = 12 // Input mode (reserved)

	PinOutputModeGPPushPull   PinMode = 0  // Output mode general purpose push/pull
	PinOutputModeGPOpenDrain  PinMode = 4  // Output mode general purpose open drain
	PinOutputModeAltPushPull  PinMode = 8  // Output mode alt. purpose push/pull
	PinOutputModeAltOpenDrain PinMode = 12 // Output mode alt. purpose open drain
)

// Configure this pin with the given I/O settings.
// stm32f1xx uses different technique for setting the GPIO pins than the stm32f407
func (p Pin) Configure(config PinConfig) {
	// Configure the GPIO pin.
	p.enableClock()
	port := p.getPort()
	pin := uint8(p) % 16
	pos := (pin % 8) * 4
	if pin < 8 {
		port.CRL.ReplaceBits(uint32(config.Mode), 0xf, pos)
	} else {
		port.CRH.ReplaceBits(uint32(config.Mode), 0xf, pos)
	}
}

func (p Pin) getPort() *stm32.GPIO_Type {
	switch p / 16 {
	case 0:
		return stm32.GPIOA
	case 1:
		return stm32.GPIOB
	case 2:
		return stm32.GPIOC
	case 3:
		return stm32.GPIOD
	case 4:
		return stm32.GPIOE
	case 5:
		return stm32.GPIOF
	case 6:
		return stm32.GPIOG
	default:
		panic("machine: unknown port")
	}
}

// enableClock enables the clock for this desired GPIO port.
func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPAEN)
	case 1:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPBEN)
	case 2:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPCEN)
	case 3:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPDEN)
	case 4:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPEEN)
	case 5:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPFEN)
	case 6:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPGEN)
	default:
		panic("machine: unknown port")
	}
}

// Enable peripheral clock. Expand to include all the desired peripherals
func enableAltFuncClock(bus unsafe.Pointer) {
	if bus == unsafe.Pointer(stm32.USART1) {
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	} else if bus == unsafe.Pointer(stm32.USART2) {
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART2EN)
	} else if bus == unsafe.Pointer(stm32.I2C1) {
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C1EN)
	} else if bus == unsafe.Pointer(stm32.SPI1) {
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	}
}

//---------- UART related code

// Configure the TX and RX pins
func (uart *UART) configurePins(config UARTConfig) {

	// pins
	switch config.TX {
	case UART_ALT_TX_PIN:
		// use alternate TX/RX pins via AFIO mapping
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_AFIOEN)
		if uart.Bus == stm32.USART1 {
			stm32.AFIO.MAPR.SetBits(stm32.AFIO_MAPR_USART1_REMAP)
		} else if uart.Bus == stm32.USART2 {
			stm32.AFIO.MAPR.SetBits(stm32.AFIO_MAPR_USART2_REMAP)
		}
	default:
		// use standard TX/RX pins PA9 and PA10
	}
	config.TX.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltPushPull})
	config.RX.Configure(PinConfig{Mode: PinInputModeFloating})
}

// Determine the divisor for USARTs to get the given baudrate
func (uart *UART) getBaudRateDivisor(br uint32) uint32 {

	// Note: PCLK2 (from APB2) used for USART1 and PCLK1 for USART2, 3, 4, 5
	var divider uint32
	if uart.Bus == stm32.USART1 {
		// first divide by PCLK2 prescaler (div 1) and then desired baudrate
		divider = CPUFrequency() / br
	} else {
		// first divide by PCLK1 prescaler (div 2) and then desired baudrate
		divider = CPUFrequency() / 2 / br
	}
	return divider
}

// Register names vary by ST processor, these are for STM F103xx
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.DR
	uart.txReg = &uart.Bus.DR
	uart.statusReg = &uart.Bus.SR
	uart.txEmptyFlag = stm32.USART_SR_TXE
}

//---------- SPI related types and code

type SPI struct {
	Bus *stm32.SPI_Type
}

// There are 3 SPI interfaces on the STM32F103xx.
// Since the first interface is named SPI1, both SPI0 and SPI1 refer to SPI1.
// TODO: implement SPI2 and SPI3.
var (
	SPI1 = SPI{Bus: stm32.SPI1}
	SPI0 = SPI1
)

// Set baud rate for SPI
func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var conf uint32

	// set frequency dependent on PCLK2 prescaler (div 1)
	switch config.Frequency {
	case 125000:
		// Note: impossible to achieve lower frequency with current PCLK2!
		conf |= stm32.SPI_CR1_BR_Div256
	case 250000:
		conf |= stm32.SPI_CR1_BR_Div256
	case 500000:
		conf |= stm32.SPI_CR1_BR_Div128
	case 1000000:
		conf |= stm32.SPI_CR1_BR_Div64
	case 2000000:
		conf |= stm32.SPI_CR1_BR_Div32
	case 4000000:
		conf |= stm32.SPI_CR1_BR_Div16
	case 8000000:
		conf |= stm32.SPI_CR1_BR_Div8
	default:
		conf |= stm32.SPI_CR1_BR_Div256
	}
	return conf
}

// Configure SPI pins for input output and clock
func (spi SPI) configurePins(config SPIConfig) {
	config.SCK.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltPushPull})
	config.SDO.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltPushPull})
	config.SDI.Configure(PinConfig{Mode: PinInputModeFloating})
}

//---------- I2C related types and code

type I2C struct {
	Bus *stm32.I2C_Type
}

// There are 2 I2C interfaces on the STM32F103xx.
// Since the first interface is named I2C1, both I2C0 and I2C1 refer to I2C1.
// TODO: implement I2C2.
var (
	I2C1 = I2C{Bus: stm32.I2C1}
	I2C0 = I2C1
)

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

// Configure is intended to setup the I2C interface.
func (i2c I2C) Configure(config I2CConfig) error {
	// Default I2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}

	// enable clock for I2C
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C1EN)

	// I2C1 pins
	switch config.SDA {
	case PB9:
		config.SCL = PB8
		// use alternate I2C1 pins PB8/PB9 via AFIO mapping
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_AFIOEN)
		stm32.AFIO.MAPR.SetBits(stm32.AFIO_MAPR_I2C1_REMAP)
	default:
		// use default I2C1 pins PB6/PB7
		config.SDA = SDA_PIN
		config.SCL = SCL_PIN
	}

	config.SDA.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltOpenDrain})
	config.SCL.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltOpenDrain})

	// Disable the selected I2C peripheral to configure
	i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_PE)

	// pclk1 clock speed is main frequency divided by PCLK1 prescaler (div 2)
	pclk1 := CPUFrequency() / 2

	// set freqency range to PCLK1 clock speed in MHz
	// aka setting the value 36 means to use 36 MHz clock
	pclk1Mhz := pclk1 / 1000000
	i2c.Bus.CR2.SetBits(pclk1Mhz)

	switch config.Frequency {
	case TWI_FREQ_100KHZ:
		// Normal mode speed calculation
		ccr := pclk1 / (config.Frequency * 2)
		i2c.Bus.CCR.Set(ccr)

		// duty cycle 2
		i2c.Bus.CCR.ClearBits(stm32.I2C_CCR_DUTY)

		// frequency standard mode
		i2c.Bus.CCR.ClearBits(stm32.I2C_CCR_F_S)

		// Set Maximum Rise Time for standard mode
		i2c.Bus.TRISE.Set(pclk1Mhz)

	case TWI_FREQ_400KHZ:
		// Fast mode speed calculation
		ccr := pclk1 / (config.Frequency * 3)
		i2c.Bus.CCR.Set(ccr)

		// duty cycle 2
		i2c.Bus.CCR.ClearBits(stm32.I2C_CCR_DUTY)

		// frequency fast mode
		i2c.Bus.CCR.SetBits(stm32.I2C_CCR_F_S)

		// Set Maximum Rise Time for fast mode
		i2c.Bus.TRISE.Set(((pclk1Mhz * 300) / 1000))
	}

	// re-enable the selected I2C peripheral
	i2c.Bus.CR1.SetBits(stm32.I2C_CR1_PE)

	return nil
}

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c I2C) Tx(addr uint16, w, r []byte) error {
	var err error
	if len(w) != 0 {
		// start transmission for writing
		err = i2c.signalStart()
		if err != nil {
			return err
		}

		// send address
		err = i2c.sendAddress(uint8(addr), true)
		if err != nil {
			return err
		}

		for _, b := range w {
			err = i2c.WriteByte(b)
			if err != nil {
				return err
			}
		}

		// sending stop here for write
		err = i2c.signalStop()
		if err != nil {
			return err
		}
	}
	if len(r) != 0 {
		// re-start transmission for reading
		err = i2c.signalStart()
		if err != nil {
			return err
		}

		// 1 byte
		switch len(r) {
		case 1:
			// send address
			err = i2c.sendAddress(uint8(addr), false)
			if err != nil {
				return err
			}

			// Disable ACK of received data
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)

			// clear timeout here
			timeout := i2cTimeout
			for !i2c.Bus.SR2.HasBits(stm32.I2C_SR2_MSL | stm32.I2C_SR2_BUSY) {
				timeout--
				if timeout == 0 {
					return errI2CWriteTimeout
				}
			}

			// Generate stop condition
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

			timeout = i2cTimeout
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_RxNE) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// Read and return data byte from I2C data register
			r[0] = byte(i2c.Bus.DR.Get())

			// wait for stop
			return i2c.waitForStop()

		case 2:
			// enable pos
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_POS)

			// Enable ACK of received data
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_ACK)

			// send address
			err = i2c.sendAddress(uint8(addr), false)
			if err != nil {
				return err
			}

			// clear address here
			timeout := i2cTimeout
			for !i2c.Bus.SR2.HasBits(stm32.I2C_SR2_MSL | stm32.I2C_SR2_BUSY) {
				timeout--
				if timeout == 0 {
					return errI2CWriteTimeout
				}
			}

			// Disable ACK of received data
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)

			// wait for btf. we need a longer timeout here than normal.
			timeout = 1000
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// Generate stop condition
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

			// read the 2 bytes by reading twice.
			r[0] = byte(i2c.Bus.DR.Get())
			r[1] = byte(i2c.Bus.DR.Get())

			// wait for stop
			err = i2c.waitForStop()

			//disable pos
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_POS)

			return err

		case 3:
			// Enable ACK of received data
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_ACK)

			// send address
			err = i2c.sendAddress(uint8(addr), false)
			if err != nil {
				return err
			}

			// clear address here
			timeout := i2cTimeout
			for !i2c.Bus.SR2.HasBits(stm32.I2C_SR2_MSL | stm32.I2C_SR2_BUSY) {
				timeout--
				if timeout == 0 {
					return errI2CWriteTimeout
				}
			}

			// Enable ACK of received data
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_ACK)

			// wait for btf. we need a longer timeout here than normal.
			timeout = 1000
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// Disable ACK of received data
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)

			// read the first byte
			r[0] = byte(i2c.Bus.DR.Get())

			timeout = 1000
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// Generate stop condition
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

			// read the last 2 bytes by reading twice.
			r[1] = byte(i2c.Bus.DR.Get())
			r[2] = byte(i2c.Bus.DR.Get())

			// wait for stop
			return i2c.waitForStop()

		default:
			// more than 3 bytes of data to read

			// send address
			err = i2c.sendAddress(uint8(addr), false)
			if err != nil {
				return err
			}

			// clear address here
			timeout := i2cTimeout
			for !i2c.Bus.SR2.HasBits(stm32.I2C_SR2_MSL | stm32.I2C_SR2_BUSY) {
				timeout--
				if timeout == 0 {
					return errI2CWriteTimeout
				}
			}

			for i := 0; i < len(r)-3; i++ {
				// Enable ACK of received data
				i2c.Bus.CR1.SetBits(stm32.I2C_CR1_ACK)

				// wait for btf. we need a longer timeout here than normal.
				timeout = 1000
				for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
					timeout--
					if timeout == 0 {
						return errI2CReadTimeout
					}
				}

				// read the next byte
				r[i] = byte(i2c.Bus.DR.Get())
			}

			// wait for btf. we need a longer timeout here than normal.
			timeout = 1000
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// Disable ACK of received data
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)

			// get third from last byte
			r[len(r)-3] = byte(i2c.Bus.DR.Get())

			// Generate stop condition
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

			// get second from last byte
			r[len(r)-2] = byte(i2c.Bus.DR.Get())

			timeout = i2cTimeout
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_RxNE) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// get last byte
			r[len(r)-1] = byte(i2c.Bus.DR.Get())

			// wait for stop
			return i2c.waitForStop()
		}
	}

	return nil
}

const i2cTimeout = 1000

// signalStart sends a start signal.
func (i2c I2C) signalStart() error {
	// Wait until I2C is not busy
	timeout := i2cTimeout
	for i2c.Bus.SR2.HasBits(stm32.I2C_SR2_BUSY) {
		timeout--
		if timeout == 0 {
			return errI2CSignalStartTimeout
		}
	}

	// clear stop
	i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_STOP)

	// Generate start condition
	i2c.Bus.CR1.SetBits(stm32.I2C_CR1_START)

	// Wait for I2C EV5 aka SB flag.
	timeout = i2cTimeout
	for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_SB) {
		timeout--
		if timeout == 0 {
			return errI2CSignalStartTimeout
		}
	}

	return nil
}

// signalStop sends a stop signal and waits for it to succeed.
func (i2c I2C) signalStop() error {
	// Generate stop condition
	i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

	// wait for stop
	return i2c.waitForStop()
}

// waitForStop waits after a stop signal.
func (i2c I2C) waitForStop() error {
	// Wait until I2C is stopped
	timeout := i2cTimeout
	for i2c.Bus.SR1.HasBits(stm32.I2C_SR1_STOPF) {
		timeout--
		if timeout == 0 {
			return errI2CSignalStopTimeout
		}
	}

	return nil
}

// Send address of device we want to talk to
func (i2c I2C) sendAddress(address uint8, write bool) error {
	data := (address << 1)
	if !write {
		data |= 1 // set read flag
	}

	i2c.Bus.DR.Set(uint32(data))

	// Wait for I2C EV6 event.
	// Destination device acknowledges address
	timeout := i2cTimeout
	if write {
		// EV6 which is ADDR flag.
		for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_ADDR) {
			timeout--
			if timeout == 0 {
				return errI2CWriteTimeout
			}
		}

		timeout = i2cTimeout
		for !i2c.Bus.SR2.HasBits(stm32.I2C_SR2_MSL | stm32.I2C_SR2_BUSY | stm32.I2C_SR2_TRA) {
			timeout--
			if timeout == 0 {
				return errI2CWriteTimeout
			}
		}
	} else {
		// I2C_EVENT_MASTER_RECEIVER_MODE_SELECTED which is ADDR flag.
		for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_ADDR) {
			timeout--
			if timeout == 0 {
				return errI2CWriteTimeout
			}
		}
	}

	return nil
}

// WriteByte writes a single byte to the I2C bus.
func (i2c I2C) WriteByte(data byte) error {
	// Send data byte
	i2c.Bus.DR.Set(uint32(data))

	// Wait for I2C EV8_2 when data has been physically shifted out and
	// output on the bus.
	// I2C_EVENT_MASTER_BYTE_TRANSMITTED is TXE flag.
	timeout := i2cTimeout
	for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_TxE) {
		timeout--
		if timeout == 0 {
			return errI2CWriteTimeout
		}
	}

	return nil
}
