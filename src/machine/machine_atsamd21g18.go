// +build sam,atsamd21g18a

// Peripheral abstraction layer for the atsamd21.
//
// Datasheet:
// http://ww1.microchip.com/downloads/en/DeviceDoc/SAMD21-Family-DataSheet-DS40001882D.pdf
//
package machine

import (
	"device/arm"
	"device/sam"
	"errors"
)

const CPU_FREQUENCY = 48000000

type GPIOMode uint8

const (
	GPIO_ANALOG       = 1
	GPIO_SERCOM       = 2
	GPIO_SERCOM_ALT   = 3
	GPIO_TIMER        = 4
	GPIO_TIMER_ALT    = 5
	GPIO_COM          = 6
	GPIO_AC_CLK       = 7
	GPIO_DIGITAL      = 8
	GPIO_INPUT        = 9
	GPIO_INPUT_PULLUP = 10
	GPIO_OUTPUT       = 11
	GPIO_PWM          = GPIO_TIMER
	GPIO_PWM_ALT      = GPIO_TIMER_ALT
)

// Configure this pin with the given configuration.
func (p GPIO) Configure(config GPIOConfig) {
	switch config.Mode {
	case GPIO_OUTPUT:
		sam.PORT.DIRSET0 = (1 << p.Pin)
		// output is also set to input enable so pin can read back its own value
		p.setPinCfg(sam.PORT_PINCFG0_INEN)

	case GPIO_INPUT:
		sam.PORT.DIRCLR0 = (1 << p.Pin)
		p.setPinCfg(sam.PORT_PINCFG0_INEN)

	case GPIO_SERCOM:
		if p.Pin&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
			p.setPMux(val | (GPIO_SERCOM << sam.PORT_PMUX0_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
			p.setPMux(val | (GPIO_SERCOM << sam.PORT_PMUX0_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_PINCFG0_PMUXEN | sam.PORT_PINCFG0_DRVSTR | sam.PORT_PINCFG0_INEN)
	}
}

// Get returns the current value of a GPIO pin.
func (p GPIO) Get() bool {
	return (sam.PORT.IN0>>p.Pin)&1 > 0
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p GPIO) Set(high bool) {
	if high {
		sam.PORT.OUTSET0 = (1 << p.Pin)
	} else {
		sam.PORT.OUTCLR0 = (1 << p.Pin)
	}
}

// getPMux returns the value for the correct PMUX register for this pin.
func (p GPIO) getPMux() sam.RegValue8 {
	return getPMux(p.Pin)
}

// setPMux sets the value for the correct PMUX register for this pin.
func (p GPIO) setPMux(val sam.RegValue8) {
	setPMux(p.Pin, val)
}

// getPinCfg returns the value for the correct PINCFG register for this pin.
func (p GPIO) getPinCfg() sam.RegValue8 {
	return getPinCfg(p.Pin)
}

// setPinCfg sets the value for the correct PINCFG register for this pin.
func (p GPIO) setPinCfg(val sam.RegValue8) {
	setPinCfg(p.Pin, val)
}

// UART on the SAMD21.
type UART struct {
	Buffer *RingBuffer
	Bus    *sam.SERCOM_USART_Type
}

var (
	// The first hardware serial port on the SAMD21. Uses the SERCOM0 interface.
	UART0 = UART{Bus: sam.SERCOM0_USART, Buffer: NewRingBuffer()}

	// The second hardware serial port on the SAMD21. Uses the SERCOM1 interface.
	UART1 = UART{Bus: sam.SERCOM1_USART, Buffer: NewRingBuffer()}
)

const (
	sampleRate16X  = 16
	lsbFirst       = 1
	sercomRXPad0   = 0
	sercomRXPad1   = 1
	sercomRXPad2   = 2
	sercomRXPad3   = 3
	sercomTXPad0   = 0 // Only for UART
	sercomTXPad2   = 1 // Only for UART
	sercomTXPad023 = 2 // Only for UART with TX on PAD0, RTS on PAD2 and CTS on PAD3
)

// Configure the UART.
func (uart UART) Configure(config UARTConfig) {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// determine pins
	if config.TX == 0 {
		// use default pins
		config.TX = UART_TX_PIN
		config.RX = UART_RX_PIN
	}

	// determine pads
	var txpad, rxpad int
	switch config.TX {
	case UART_TX_PIN:
		txpad = sercomTXPad2
	case D10:
		txpad = sercomTXPad2
	case D11:
		txpad = sercomTXPad0
	default:
		panic("Invalid TX pin for UART")
	}

	switch config.RX {
	case UART_RX_PIN:
		rxpad = sercomRXPad3
	case D10:
		rxpad = sercomRXPad2
	case D11:
		rxpad = sercomRXPad0
	case D12:
		rxpad = sercomRXPad3
	case D13:
		rxpad = sercomRXPad1
	default:
		panic("Invalid RX pin for UART")
	}

	// configure pins
	GPIO{config.TX}.Configure(GPIOConfig{Mode: GPIO_SERCOM})
	GPIO{config.RX}.Configure(GPIOConfig{Mode: GPIO_SERCOM})

	// reset SERCOM0
	uart.Bus.CTRLA |= sam.SERCOM_USART_CTRLA_SWRST
	for (uart.Bus.CTRLA&sam.SERCOM_USART_CTRLA_SWRST) > 0 ||
		(uart.Bus.SYNCBUSY&sam.SERCOM_USART_SYNCBUSY_SWRST) > 0 {
	}

	// set UART mode/sample rate
	// SERCOM_USART_CTRLA_MODE(mode) |
	// SERCOM_USART_CTRLA_SAMPR(sampleRate);
	uart.Bus.CTRLA = (sam.SERCOM_USART_CTRLA_MODE_USART_INT_CLK << sam.SERCOM_USART_CTRLA_MODE_Pos) |
		(1 << sam.SERCOM_USART_CTRLA_SAMPR_Pos) // sample rate of 16x

	// Set baud rate
	uart.SetBaudRate(config.BaudRate)

	// setup UART frame
	// SERCOM_USART_CTRLA_FORM( (parityMode == SERCOM_NO_PARITY ? 0 : 1) ) |
	// dataOrder << SERCOM_USART_CTRLA_DORD_Pos;
	uart.Bus.CTRLA |= (0 << sam.SERCOM_USART_CTRLA_FORM_Pos) | // no parity
		(lsbFirst << sam.SERCOM_USART_CTRLA_DORD_Pos) // data order

	// set UART stop bits/parity
	// SERCOM_USART_CTRLB_CHSIZE(charSize) |
	// 	nbStopBits << SERCOM_USART_CTRLB_SBMODE_Pos |
	// 	(parityMode == SERCOM_NO_PARITY ? 0 : parityMode) << SERCOM_USART_CTRLB_PMODE_Pos; //If no parity use default value
	uart.Bus.CTRLB |= (0 << sam.SERCOM_USART_CTRLB_CHSIZE_Pos) | // 8 bits is 0
		(0 << sam.SERCOM_USART_CTRLB_SBMODE_Pos) | // 1 stop bit is zero
		(0 << sam.SERCOM_USART_CTRLB_PMODE_Pos) // no parity

	// set UART pads. This is not same as pins...
	//  SERCOM_USART_CTRLA_TXPO(txPad) |
	//   SERCOM_USART_CTRLA_RXPO(rxPad);
	uart.Bus.CTRLA |= sam.RegValue((txpad << sam.SERCOM_USART_CTRLA_TXPO_Pos) |
		(rxpad << sam.SERCOM_USART_CTRLA_RXPO_Pos))

	// Enable Transceiver and Receiver
	//sercom->USART.CTRLB.reg |= SERCOM_USART_CTRLB_TXEN | SERCOM_USART_CTRLB_RXEN ;
	uart.Bus.CTRLB |= (sam.SERCOM_USART_CTRLB_TXEN | sam.SERCOM_USART_CTRLB_RXEN)

	// Enable USART1 port.
	// sercom->USART.CTRLA.bit.ENABLE = 0x1u;
	uart.Bus.CTRLA |= sam.SERCOM_USART_CTRLA_ENABLE
	for (uart.Bus.SYNCBUSY & sam.SERCOM_USART_SYNCBUSY_ENABLE) > 0 {
	}

	// setup interrupt on receive
	uart.Bus.INTENSET = sam.SERCOM_USART_INTENSET_RXC

	// Enable RX IRQ.
	if config.TX == UART_TX_PIN {
		// UART0
		arm.EnableIRQ(sam.IRQ_SERCOM0)
	} else {
		// UART1
		arm.EnableIRQ(sam.IRQ_SERCOM1)
	}
}

// SetBaudRate sets the communication speed for the UART.
func (uart UART) SetBaudRate(br uint32) {
	// Asynchronous fractional mode (Table 24-2 in datasheet)
	//   BAUD = fref / (sampleRateValue * fbaud)
	// (multiply by 8, to calculate fractional piece)
	// uint32_t baudTimes8 = (SystemCoreClock * 8) / (16 * baudrate);
	baud := (CPU_FREQUENCY * 8) / (sampleRate16X * br)

	// sercom->USART.BAUD.FRAC.FP   = (baudTimes8 % 8);
	// sercom->USART.BAUD.FRAC.BAUD = (baudTimes8 / 8);
	uart.Bus.BAUD = sam.RegValue16(((baud % 8) << sam.SERCOM_USART_BAUD_FRAC_MODE_FP_Pos) |
		((baud / 8) << sam.SERCOM_USART_BAUD_FRAC_MODE_BAUD_Pos))
}

// WriteByte writes a byte of data to the UART.
func (uart UART) WriteByte(c byte) error {
	// wait until ready to receive
	for (uart.Bus.INTFLAG & sam.SERCOM_USART_INTFLAG_DRE) == 0 {
	}
	uart.Bus.DATA = sam.RegValue16(c)
	return nil
}

//go:export SERCOM0_IRQHandler
func handleUART0() {
	// should reset IRQ
	UART0.Receive(byte((UART0.Bus.DATA & 0xFF)))
	UART0.Bus.INTFLAG |= sam.SERCOM_USART_INTFLAG_RXC
}

//go:export SERCOM1_IRQHandler
func handleUART1() {
	// should reset IRQ
	UART1.Receive(byte((UART1.Bus.DATA & 0xFF)))
	UART1.Bus.INTFLAG |= sam.SERCOM_USART_INTFLAG_RXC
}

// I2C on the SAMD21.
type I2C struct {
	Bus *sam.SERCOM_I2CM_Type
}

// Since the I2C interfaces on the SAMD21 use the SERCOMx peripherals,
// you can have multiple ones. we currently only implement one.
var (
	I2C0 = I2C{Bus: sam.SERCOM3_I2CM}
)

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       uint8
	SDA       uint8
}

const (
	// Default rise time in nanoseconds, based on 4.7K ohm pull up resistors
	riseTimeNanoseconds = 125

	// wire bus states
	wireUnknownState = 0
	wireIdleState    = 1
	wireOwnerState   = 2
	wireBusyState    = 3

	// wire commands
	wireCmdNoAction    = 0
	wireCmdRepeatStart = 1
	wireCmdRead        = 2
	wireCmdStop        = 3
)

const i2cTimeout = 1000

// Configure is intended to setup the I2C interface.
func (i2c I2C) Configure(config I2CConfig) {
	// Default I2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}

	// reset SERCOM3
	i2c.Bus.CTRLA |= sam.SERCOM_I2CM_CTRLA_SWRST
	for (i2c.Bus.CTRLA&sam.SERCOM_I2CM_CTRLA_SWRST) > 0 ||
		(i2c.Bus.SYNCBUSY&sam.SERCOM_I2CM_SYNCBUSY_SWRST) > 0 {
	}

	// Set i2c master mode
	//SERCOM_I2CM_CTRLA_MODE( I2C_MASTER_OPERATION )
	i2c.Bus.CTRLA = (sam.SERCOM_I2CM_CTRLA_MODE_I2C_MASTER << sam.SERCOM_I2CM_CTRLA_MODE_Pos) // |

	i2c.SetBaudRate(config.Frequency)

	// Enable I2CM port.
	// sercom->USART.CTRLA.bit.ENABLE = 0x1u;
	i2c.Bus.CTRLA |= sam.SERCOM_I2CM_CTRLA_ENABLE
	for (i2c.Bus.SYNCBUSY & sam.SERCOM_I2CM_SYNCBUSY_ENABLE) > 0 {
	}

	// set bus idle mode
	i2c.Bus.STATUS |= (wireIdleState << sam.SERCOM_I2CM_STATUS_BUSSTATE_Pos)
	for (i2c.Bus.SYNCBUSY & sam.SERCOM_I2CM_SYNCBUSY_SYSOP) > 0 {
	}

	// enable pins
	GPIO{SDA_PIN}.Configure(GPIOConfig{Mode: GPIO_SERCOM})
	GPIO{SCL_PIN}.Configure(GPIOConfig{Mode: GPIO_SERCOM})
}

// SetBaudRate sets the communication speed for the I2C.
func (i2c I2C) SetBaudRate(br uint32) {
	// Synchronous arithmetic baudrate, via Arduino SAMD implementation:
	// SystemCoreClock / ( 2 * baudrate) - 5 - (((SystemCoreClock / 1000000) * WIRE_RISE_TIME_NANOSECONDS) / (2 * 1000));
	baud := CPU_FREQUENCY/(2*br) - 5 - (((CPU_FREQUENCY / 1000000) * riseTimeNanoseconds) / (2 * 1000))
	i2c.Bus.BAUD = sam.RegValue(baud)
}

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c I2C) Tx(addr uint16, w, r []byte) error {
	var err error
	if len(w) != 0 {
		// send start/address for write
		i2c.sendAddress(addr, true)

		// wait until transmission complete
		timeout := i2cTimeout
		for (i2c.Bus.INTFLAG & sam.SERCOM_I2CM_INTFLAG_MB) == 0 {
			timeout--
			if timeout == 0 {
				return errors.New("I2C timeout on ready to write data")
			}
		}

		// ACK received (0: ACK, 1: NACK)
		if (i2c.Bus.STATUS & sam.SERCOM_I2CM_STATUS_RXNACK) > 0 {
			return errors.New("I2C write error: expected ACK not NACK")
		}

		// write data
		for _, b := range w {
			err = i2c.WriteByte(b)
			if err != nil {
				return err
			}
		}

		err = i2c.signalStop()
		if err != nil {
			return err
		}
	}
	if len(r) != 0 {
		// send start/address for read
		i2c.sendAddress(addr, false)

		// wait transmission complete
		for (i2c.Bus.INTFLAG & sam.SERCOM_I2CM_INTFLAG_SB) == 0 {
			// If the slave NACKS the address, the MB bit will be set.
			// In that case, send a stop condition and return error.
			if (i2c.Bus.INTFLAG & sam.SERCOM_I2CM_INTFLAG_MB) > 0 {
				i2c.Bus.CTRLB |= (wireCmdStop << sam.SERCOM_I2CM_CTRLB_CMD_Pos) // Stop condition
				return errors.New("I2C read error: expected ACK not NACK")
			}
		}

		// ACK received (0: ACK, 1: NACK)
		if (i2c.Bus.STATUS & sam.SERCOM_I2CM_STATUS_RXNACK) > 0 {
			return errors.New("I2C read error: expected ACK not NACK")
		}

		// read first byte
		r[0] = i2c.readByte()
		for i := 1; i < len(r); i++ {
			// Send an ACK
			i2c.Bus.CTRLB &^= sam.SERCOM_I2CM_CTRLB_ACKACT

			i2c.signalRead()

			// Read data and send the ACK
			r[i] = i2c.readByte()
		}

		// Send NACK to end transmission
		i2c.Bus.CTRLB |= sam.SERCOM_I2CM_CTRLB_ACKACT

		err = i2c.signalStop()
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteByte writes a single byte to the I2C bus.
func (i2c I2C) WriteByte(data byte) error {
	// Send data byte
	i2c.Bus.DATA = sam.RegValue8(data)

	// wait until transmission successful
	timeout := i2cTimeout
	for (i2c.Bus.INTFLAG & sam.SERCOM_I2CM_INTFLAG_MB) == 0 {
		// check for bus error
		if (sam.SERCOM3_I2CM.STATUS & sam.SERCOM_I2CM_STATUS_BUSERR) > 0 {
			return errors.New("I2C bus error")
		}
		timeout--
		if timeout == 0 {
			return errors.New("I2C timeout on write data")
		}
	}

	if (i2c.Bus.STATUS & sam.SERCOM_I2CM_STATUS_RXNACK) > 0 {
		return errors.New("I2C write error: expected ACK not NACK")
	}

	return nil
}

// sendAddress sends the address and start signal
func (i2c I2C) sendAddress(address uint16, write bool) error {
	data := (address << 1)
	if !write {
		data |= 1 // set read flag
	}

	// wait until bus ready
	timeout := i2cTimeout
	for (i2c.Bus.STATUS&(wireIdleState<<sam.SERCOM_I2CM_STATUS_BUSSTATE_Pos)) == 0 &&
		(i2c.Bus.STATUS&(wireOwnerState<<sam.SERCOM_I2CM_STATUS_BUSSTATE_Pos)) == 0 {
		timeout--
		if timeout == 0 {
			return errors.New("I2C timeout on bus ready")
		}
	}
	i2c.Bus.ADDR = sam.RegValue(data)

	return nil
}

func (i2c I2C) signalStop() error {
	i2c.Bus.CTRLB |= (wireCmdStop << sam.SERCOM_I2CM_CTRLB_CMD_Pos) // Stop command
	timeout := i2cTimeout
	for (i2c.Bus.SYNCBUSY & sam.SERCOM_I2CM_SYNCBUSY_SYSOP) > 0 {
		timeout--
		if timeout == 0 {
			return errors.New("I2C timeout on signal stop")
		}
	}
	return nil
}

func (i2c I2C) signalRead() error {
	i2c.Bus.CTRLB |= (wireCmdRead << sam.SERCOM_I2CM_CTRLB_CMD_Pos) // Read command
	timeout := i2cTimeout
	for (i2c.Bus.SYNCBUSY & sam.SERCOM_I2CM_SYNCBUSY_SYSOP) > 0 {
		timeout--
		if timeout == 0 {
			return errors.New("I2C timeout on signal read")
		}
	}
	return nil
}

func (i2c I2C) readByte() byte {
	for (i2c.Bus.INTFLAG & sam.SERCOM_I2CM_INTFLAG_SB) == 0 {
	}
	return byte(i2c.Bus.DATA)
}

// PWM
const period = 0xFFFF

// InitPWM initializes the PWM interface.
func InitPWM() {
	// turn on timer clocks used for PWM
	sam.PM.APBCMASK |= sam.PM_APBCMASK_TCC0_ | sam.PM_APBCMASK_TCC1_ | sam.PM_APBCMASK_TCC2_

	// Use GCLK0 for TCC0/TCC1
	sam.GCLK.CLKCTRL = sam.RegValue16((sam.GCLK_CLKCTRL_ID_TCC0_TCC1 << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	for (sam.GCLK.STATUS & sam.GCLK_STATUS_SYNCBUSY) > 0 {
	}

	// Use GCLK0 for TCC2/TC3
	sam.GCLK.CLKCTRL = sam.RegValue16((sam.GCLK_CLKCTRL_ID_TCC2_TC3 << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	for (sam.GCLK.STATUS & sam.GCLK_STATUS_SYNCBUSY) > 0 {
	}
}

// Configure configures a PWM pin for output.
func (pwm PWM) Configure() {
	// figure out which TCCX timer for this pin
	timer := pwm.getTimer()

	// disable timer
	timer.CTRLA &^= sam.TCC_CTRLA_ENABLE
	// Wait for synchronization
	for (timer.SYNCBUSY & sam.TCC_SYNCBUSY_ENABLE) > 0 {
	}

	// Use "Normal PWM" (single-slope PWM)
	timer.WAVE |= sam.TCC_WAVE_WAVEGEN_NPWM
	// Wait for synchronization
	for (timer.SYNCBUSY & sam.TCC_SYNCBUSY_WAVE) > 0 {
	}

	// Set the period (the number to count to (TOP) before resetting timer)
	//TCC0->PER.reg = period;
	timer.PER = period
	// Wait for synchronization
	for (timer.SYNCBUSY & sam.TCC_SYNCBUSY_PER) > 0 {
	}

	// Set pin as output
	sam.PORT.DIRSET0 = (1 << pwm.Pin)
	// Set pin to low
	sam.PORT.OUTCLR0 = (1 << pwm.Pin)

	// Enable the port multiplexer for pin
	pwm.setPinCfg(sam.PORT_PINCFG0_PMUXEN)

	// Connect TCCX timer to pin.
	// we normally use the F channel aka ALT
	pwmConfig := GPIO_PWM_ALT

	// in the case of PA6 or PA7 we have to use E channel
	if pwm.Pin == 6 || pwm.Pin == 7 {
		pwmConfig = GPIO_PWM
	}

	if pwm.Pin&1 > 0 {
		// odd pin, so save the even pins
		val := pwm.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
		pwm.setPMux(val | sam.RegValue8(pwmConfig<<sam.PORT_PMUX0_PMUXO_Pos))
	} else {
		// even pin, so save the odd pins
		val := pwm.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
		pwm.setPMux(val | sam.RegValue8(pwmConfig<<sam.PORT_PMUX0_PMUXE_Pos))
	}
}

// Set turns on the duty cycle for a PWM pin using the provided value.
func (pwm PWM) Set(value uint16) {
	// figure out which TCCX timer for this pin
	timer := pwm.getTimer()

	// disable output
	timer.CTRLA &^= sam.TCC_CTRLA_ENABLE

	// Wait for synchronization
	for (timer.SYNCBUSY & sam.TCC_SYNCBUSY_ENABLE) > 0 {
	}

	// Set PWM signal to output duty cycle
	pwm.setChannel(sam.RegValue(value))

	// Wait for synchronization on all channels
	for (timer.SYNCBUSY & (sam.TCC_SYNCBUSY_CC0 |
		sam.TCC_SYNCBUSY_CC1 |
		sam.TCC_SYNCBUSY_CC2 |
		sam.TCC_SYNCBUSY_CC3)) > 0 {
	}

	// enable
	timer.CTRLA |= sam.TCC_CTRLA_ENABLE
	// Wait for synchronization
	for (timer.SYNCBUSY & sam.TCC_SYNCBUSY_ENABLE) > 0 {
	}
}

// getPMux returns the value for the correct PMUX register for this pin.
func (pwm PWM) getPMux() sam.RegValue8 {
	return getPMux(pwm.Pin)
}

// setPMux sets the value for the correct PMUX register for this pin.
func (pwm PWM) setPMux(val sam.RegValue8) {
	setPMux(pwm.Pin, val)
}

// getPinCfg returns the value for the correct PINCFG register for this pin.
func (pwm PWM) getPinCfg() sam.RegValue8 {
	return getPinCfg(pwm.Pin)
}

// setPinCfg sets the value for the correct PINCFG register for this pin.
func (pwm PWM) setPinCfg(val sam.RegValue8) {
	setPinCfg(pwm.Pin, val)
}

// getPMux returns the value for the correct PMUX register for this pin.
func getPMux(p uint8) sam.RegValue8 {
	pin := p >> 1
	switch pin {
	case 0:
		return sam.PORT.PMUX0_0
	case 1:
		return sam.PORT.PMUX0_1
	case 2:
		return sam.PORT.PMUX0_2
	case 3:
		return sam.PORT.PMUX0_3
	case 4:
		return sam.PORT.PMUX0_4
	case 5:
		return sam.PORT.PMUX0_5
	case 6:
		return sam.PORT.PMUX0_6
	case 7:
		return sam.PORT.PMUX0_7
	case 8:
		return sam.PORT.PMUX0_8
	case 9:
		return sam.PORT.PMUX0_9
	case 10:
		return sam.PORT.PMUX0_10
	case 11:
		return sam.PORT.PMUX0_11
	case 12:
		return sam.PORT.PMUX0_12
	case 13:
		return sam.PORT.PMUX0_13
	case 14:
		return sam.PORT.PMUX0_14
	case 15:
		return sam.PORT.PMUX0_15
	default:
		return 0
	}
}

// setPMux sets the value for the correct PMUX register for this pin.
func setPMux(p uint8, val sam.RegValue8) {
	pin := p >> 1
	switch pin {
	case 0:
		sam.PORT.PMUX0_0 = val
	case 1:
		sam.PORT.PMUX0_1 = val
	case 2:
		sam.PORT.PMUX0_2 = val
	case 3:
		sam.PORT.PMUX0_3 = val
	case 4:
		sam.PORT.PMUX0_4 = val
	case 5:
		sam.PORT.PMUX0_5 = val
	case 6:
		sam.PORT.PMUX0_6 = val
	case 7:
		sam.PORT.PMUX0_7 = val
	case 8:
		sam.PORT.PMUX0_8 = val
	case 9:
		sam.PORT.PMUX0_9 = val
	case 10:
		sam.PORT.PMUX0_10 = val
	case 11:
		sam.PORT.PMUX0_11 = val
	case 12:
		sam.PORT.PMUX0_12 = val
	case 13:
		sam.PORT.PMUX0_13 = val
	case 14:
		sam.PORT.PMUX0_14 = val
	case 15:
		sam.PORT.PMUX0_15 = val
	}
}

// getPinCfg returns the value for the correct PINCFG register for this pin.
func getPinCfg(p uint8) sam.RegValue8 {
	switch p {
	case 0:
		return sam.PORT.PINCFG0_0
	case 1:
		return sam.PORT.PINCFG0_1
	case 2:
		return sam.PORT.PINCFG0_2
	case 3:
		return sam.PORT.PINCFG0_3
	case 4:
		return sam.PORT.PINCFG0_4
	case 5:
		return sam.PORT.PINCFG0_5
	case 6:
		return sam.PORT.PINCFG0_6
	case 7:
		return sam.PORT.PINCFG0_7
	case 8:
		return sam.PORT.PINCFG0_8
	case 9:
		return sam.PORT.PINCFG0_9
	case 10:
		return sam.PORT.PINCFG0_10
	case 11:
		return sam.PORT.PINCFG0_11
	case 12:
		return sam.PORT.PINCFG0_12
	case 13:
		return sam.PORT.PINCFG0_13
	case 14:
		return sam.PORT.PINCFG0_14
	case 15:
		return sam.PORT.PINCFG0_15
	case 16:
		return sam.PORT.PINCFG0_16
	case 17:
		return sam.PORT.PINCFG0_17
	case 18:
		return sam.PORT.PINCFG0_18
	case 19:
		return sam.PORT.PINCFG0_19
	case 20:
		return sam.PORT.PINCFG0_20
	case 21:
		return sam.PORT.PINCFG0_21
	case 22:
		return sam.PORT.PINCFG0_22
	case 23:
		return sam.PORT.PINCFG0_23
	case 24:
		return sam.PORT.PINCFG0_24
	case 25:
		return sam.PORT.PINCFG0_25
	case 26:
		return sam.PORT.PINCFG0_26
	case 27:
		return sam.PORT.PINCFG0_27
	case 28:
		return sam.PORT.PINCFG0_28
	case 29:
		return sam.PORT.PINCFG0_29
	case 30:
		return sam.PORT.PINCFG0_30
	case 31:
		return sam.PORT.PINCFG0_31
	default:
		return 0
	}
}

// setPinCfg sets the value for the correct PINCFG register for this pin.
func setPinCfg(p uint8, val sam.RegValue8) {
	switch p {
	case 0:
		sam.PORT.PINCFG0_0 = val
	case 1:
		sam.PORT.PINCFG0_1 = val
	case 2:
		sam.PORT.PINCFG0_2 = val
	case 3:
		sam.PORT.PINCFG0_3 = val
	case 4:
		sam.PORT.PINCFG0_4 = val
	case 5:
		sam.PORT.PINCFG0_5 = val
	case 6:
		sam.PORT.PINCFG0_6 = val
	case 7:
		sam.PORT.PINCFG0_7 = val
	case 8:
		sam.PORT.PINCFG0_8 = val
	case 9:
		sam.PORT.PINCFG0_9 = val
	case 10:
		sam.PORT.PINCFG0_10 = val
	case 11:
		sam.PORT.PINCFG0_11 = val
	case 12:
		sam.PORT.PINCFG0_12 = val
	case 13:
		sam.PORT.PINCFG0_13 = val
	case 14:
		sam.PORT.PINCFG0_14 = val
	case 15:
		sam.PORT.PINCFG0_15 = val
	case 16:
		sam.PORT.PINCFG0_16 = val
	case 17:
		sam.PORT.PINCFG0_17 = val
	case 18:
		sam.PORT.PINCFG0_18 = val
	case 19:
		sam.PORT.PINCFG0_19 = val
	case 20:
		sam.PORT.PINCFG0_20 = val
	case 21:
		sam.PORT.PINCFG0_21 = val
	case 22:
		sam.PORT.PINCFG0_22 = val
	case 23:
		sam.PORT.PINCFG0_23 = val
	case 24:
		sam.PORT.PINCFG0_24 = val
	case 25:
		sam.PORT.PINCFG0_25 = val
	case 26:
		sam.PORT.PINCFG0_26 = val
	case 27:
		sam.PORT.PINCFG0_27 = val
	case 28:
		sam.PORT.PINCFG0_28 = val
	case 29:
		sam.PORT.PINCFG0_29 = val
	case 30:
		sam.PORT.PINCFG0_30 = val
	case 31:
		sam.PORT.PINCFG0_31 = val
	}
}

// getTimer returns the timer to be used for PWM on this pin
func (pwm PWM) getTimer() *sam.TCC_Type {
	switch pwm.Pin {
	case 6:
		return sam.TCC1
	case 7:
		return sam.TCC1
	case 8:
		return sam.TCC1
	case 9:
		return sam.TCC1
	case 14:
		return sam.TCC0
	case 15:
		return sam.TCC0
	case 16:
		return sam.TCC0
	case 17:
		return sam.TCC0
	case 18:
		return sam.TCC0
	case 19:
		return sam.TCC0
	case 20:
		return sam.TCC0
	case 21:
		return sam.TCC0
	default:
		return nil // not supported on this pin
	}
}

// setChannel sets the value for the correct channel for PWM on this pin
func (pwm PWM) setChannel(val sam.RegValue) {
	switch pwm.Pin {
	case 6:
		pwm.getTimer().CC0 = val
	case 7:
		pwm.getTimer().CC1 = val
	case 8:
		pwm.getTimer().CC0 = val
	case 9:
		pwm.getTimer().CC1 = val
	case 14:
		pwm.getTimer().CC0 = val
	case 15:
		pwm.getTimer().CC1 = val
	case 16:
		pwm.getTimer().CC2 = val
	case 17:
		pwm.getTimer().CC3 = val
	case 18:
		pwm.getTimer().CC2 = val
	case 19:
		pwm.getTimer().CC3 = val
	case 20:
		pwm.getTimer().CC2 = val
	case 21:
		pwm.getTimer().CC3 = val
	default:
		return // not supported on this pin
	}
}
