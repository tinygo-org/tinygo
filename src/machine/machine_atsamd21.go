// +build sam,atsamd21

// Peripheral abstraction layer for the atsamd21.
//
// Datasheet:
// http://ww1.microchip.com/downloads/en/DeviceDoc/SAMD21-Family-DataSheet-DS40001882D.pdf
//
package machine

import (
	"bytes"
	"device/arm"
	"device/sam"
	"encoding/binary"
	"errors"
	"unsafe"
)

const CPU_FREQUENCY = 48000000

type PinMode uint8

const (
	PinAnalog    PinMode = 1
	PinSERCOM    PinMode = 2
	PinSERCOMAlt PinMode = 3
	PinTimer     PinMode = 4
	PinTimerAlt  PinMode = 5
	PinCom       PinMode = 6
	//PinAC_CLK        PinMode = 7
	PinDigital       PinMode = 8
	PinInput         PinMode = 9
	PinInputPullup   PinMode = 10
	PinOutput        PinMode = 11
	PinPWM           PinMode = PinTimer
	PinPWMAlt        PinMode = PinTimerAlt
	PinInputPulldown PinMode = 12
)

// Hardware pins
const (
	PA00 Pin = 0
	PA01 Pin = 1
	PA02 Pin = 2
	PA03 Pin = 3
	PA04 Pin = 4
	PA05 Pin = 5
	PA06 Pin = 6
	PA07 Pin = 7
	PA08 Pin = 8
	PA09 Pin = 9
	PA10 Pin = 10
	PA11 Pin = 11
	PA12 Pin = 12
	PA13 Pin = 13
	PA14 Pin = 14
	PA15 Pin = 15
	PA16 Pin = 16
	PA17 Pin = 17
	PA18 Pin = 18
	PA19 Pin = 19
	PA20 Pin = 20
	PA21 Pin = 21
	PA22 Pin = 22
	PA23 Pin = 23
	PA24 Pin = 24
	PA25 Pin = 25
	PA26 Pin = 26
	PA27 Pin = 27
	PA28 Pin = 28
	PA29 Pin = 29
	PA30 Pin = 30
	PA31 Pin = 31
	PB00 Pin = 32
	PB01 Pin = 33
	PB02 Pin = 34
	PB03 Pin = 35
	PB04 Pin = 36
	PB05 Pin = 37
	PB06 Pin = 38
	PB07 Pin = 39
	PB08 Pin = 40
	PB09 Pin = 41
	PB10 Pin = 42
	PB11 Pin = 43
	PB12 Pin = 44
	PB13 Pin = 45
	PB14 Pin = 46
	PB15 Pin = 47
	PB16 Pin = 48
	PB17 Pin = 49
	PB18 Pin = 50
	PB19 Pin = 51
	PB20 Pin = 52
	PB21 Pin = 53
	PB22 Pin = 54
	PB23 Pin = 55
	PB24 Pin = 56
	PB25 Pin = 57
	PB26 Pin = 58
	PB27 Pin = 59
	PB28 Pin = 60
	PB29 Pin = 61
	PB30 Pin = 62
	PB31 Pin = 63
)

// InitADC initializes the ADC.
func InitADC() {
	// ADC Bias Calibration
	// #define ADC_FUSES_BIASCAL_ADDR      (NVMCTRL_OTP4 + 4)
	// #define ADC_FUSES_BIASCAL_Pos       3            /**< \brief (NVMCTRL_OTP4) ADC Bias Calibration */
	// #define ADC_FUSES_BIASCAL_Msk       (0x7u << ADC_FUSES_BIASCAL_Pos)
	// #define ADC_FUSES_BIASCAL(value)    ((ADC_FUSES_BIASCAL_Msk & ((value) << ADC_FUSES_BIASCAL_Pos)))
	// #define ADC_FUSES_LINEARITY_0_ADDR  NVMCTRL_OTP4
	// #define ADC_FUSES_LINEARITY_0_Pos   27           /**< \brief (NVMCTRL_OTP4) ADC Linearity bits 4:0 */
	// #define ADC_FUSES_LINEARITY_0_Msk   (0x1Fu << ADC_FUSES_LINEARITY_0_Pos)
	// #define ADC_FUSES_LINEARITY_0(value) ((ADC_FUSES_LINEARITY_0_Msk & ((value) << ADC_FUSES_LINEARITY_0_Pos)))
	// #define ADC_FUSES_LINEARITY_1_ADDR  (NVMCTRL_OTP4 + 4)
	// #define ADC_FUSES_LINEARITY_1_Pos   0            /**< \brief (NVMCTRL_OTP4) ADC Linearity bits 7:5 */
	// #define ADC_FUSES_LINEARITY_1_Msk   (0x7u << ADC_FUSES_LINEARITY_1_Pos)
	// #define ADC_FUSES_LINEARITY_1(value) ((ADC_FUSES_LINEARITY_1_Msk & ((value) << ADC_FUSES_LINEARITY_1_Pos)))

	biasFuse := *(*uint32)(unsafe.Pointer(uintptr(0x00806020) + 4))
	bias := uint16(biasFuse>>3) & uint16(0x7)

	// ADC Linearity bits 4:0
	linearity0Fuse := *(*uint32)(unsafe.Pointer(uintptr(0x00806020)))
	linearity := uint16(linearity0Fuse>>27) & uint16(0x1f)

	// ADC Linearity bits 7:5
	linearity1Fuse := *(*uint32)(unsafe.Pointer(uintptr(0x00806020) + 4))
	linearity |= uint16(linearity1Fuse) & uint16(0x7) << 5

	// set calibration
	sam.ADC.CALIB.Set((bias << 8) | linearity)

	// Wait for synchronization
	waitADCSync()

	// Divide Clock by 32 with 12 bits resolution as default
	sam.ADC.CTRLB.Set((sam.ADC_CTRLB_PRESCALER_DIV32 << sam.ADC_CTRLB_PRESCALER_Pos) |
		(sam.ADC_CTRLB_RESSEL_12BIT << sam.ADC_CTRLB_RESSEL_Pos))

	// Sampling Time Length
	sam.ADC.SAMPCTRL.Set(5)

	// Wait for synchronization
	waitADCSync()

	// Use internal ground
	sam.ADC.INPUTCTRL.Set(sam.ADC_INPUTCTRL_MUXNEG_GND << sam.ADC_INPUTCTRL_MUXNEG_Pos)

	// Averaging (see datasheet table in AVGCTRL register description)
	sam.ADC.AVGCTRL.Set((sam.ADC_AVGCTRL_SAMPLENUM_1 << sam.ADC_AVGCTRL_SAMPLENUM_Pos) |
		(0x0 << sam.ADC_AVGCTRL_ADJRES_Pos))

	// Analog Reference is AREF pin (3.3v)
	sam.ADC.INPUTCTRL.SetBits(sam.ADC_INPUTCTRL_GAIN_DIV2 << sam.ADC_INPUTCTRL_GAIN_Pos)

	// 1/2 VDDANA = 0.5 * 3V3 = 1.65V
	sam.ADC.REFCTRL.SetBits(sam.ADC_REFCTRL_REFSEL_INTVCC1 << sam.ADC_REFCTRL_REFSEL_Pos)
}

// Configure configures a ADCPin to be able to be used to read data.
func (a ADC) Configure() {
	a.Pin.Configure(PinConfig{Mode: PinAnalog})
	return
}

// Get returns the current value of a ADC pin, in the range 0..0xffff.
func (a ADC) Get() uint16 {
	ch := a.getADCChannel()

	// Selection for the positive ADC input
	sam.ADC.INPUTCTRL.ClearBits(sam.ADC_INPUTCTRL_MUXPOS_Msk)
	waitADCSync()
	sam.ADC.INPUTCTRL.SetBits(uint32(ch << sam.ADC_INPUTCTRL_MUXPOS_Pos))
	waitADCSync()

	// Select internal ground for ADC input
	sam.ADC.INPUTCTRL.ClearBits(sam.ADC_INPUTCTRL_MUXNEG_Msk)
	waitADCSync()
	sam.ADC.INPUTCTRL.SetBits(sam.ADC_INPUTCTRL_MUXNEG_GND << sam.ADC_INPUTCTRL_MUXNEG_Pos)
	waitADCSync()

	// Enable ADC
	sam.ADC.CTRLA.SetBits(sam.ADC_CTRLA_ENABLE)
	waitADCSync()

	// Start conversion
	sam.ADC.SWTRIG.SetBits(sam.ADC_SWTRIG_START)
	waitADCSync()

	// Clear the Data Ready flag
	sam.ADC.INTFLAG.SetBits(sam.ADC_INTFLAG_RESRDY)
	waitADCSync()

	// Start conversion again, since first conversion after reference voltage changed is invalid.
	sam.ADC.SWTRIG.SetBits(sam.ADC_SWTRIG_START)
	waitADCSync()

	// Waiting for conversion to complete
	for !sam.ADC.INTFLAG.HasBits(sam.ADC_INTFLAG_RESRDY) {
	}
	val := sam.ADC.RESULT.Get()

	// Disable ADC
	sam.ADC.CTRLA.ClearBits(sam.ADC_CTRLA_ENABLE)
	waitADCSync()

	return uint16(val) << 4 // scales from 12 to 16-bit result
}

func (a ADC) getADCChannel() uint8 {
	switch a.Pin {
	case PA02:
		return 0
	case PB08:
		return 2
	case PB09:
		return 3
	case PA04:
		return 4
	case PA05:
		return 5
	case PA06:
		return 6
	case PA07:
		return 7
	case PB02:
		return 10
	case PB03:
		return 11
	case PA09:
		return 17
	case PA11:
		return 19
	default:
		return 0
	}
}

func waitADCSync() {
	for sam.ADC.STATUS.HasBits(sam.ADC_STATUS_SYNCBUSY) {
	}
}

// UART on the SAMD21.
type UART struct {
	Buffer *RingBuffer
	Bus    *sam.SERCOM_USART_Type
	Mode   PinMode
	IRQVal uint32
}

var (
	// UART0 is actually a USB CDC interface.
	UART0 = USBCDC{Buffer: NewRingBuffer()}
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

	spiTXPad0SCK1 = 0
	spiTXPad2SCK3 = 1
	spiTXPad3SCK1 = 2
	spiTXPad0SCK3 = 3
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
	case PA10:
		txpad = sercomTXPad2
	case PA18:
		txpad = sercomTXPad2
	case PA16:
		txpad = sercomTXPad0
	case PA22:
		txpad = sercomTXPad0
	default:
		panic("Invalid TX pin for UART")
	}

	switch config.RX {
	case PA11:
		rxpad = sercomRXPad3
	case PA18:
		rxpad = sercomRXPad2
	case PA16:
		rxpad = sercomRXPad0
	case PA19:
		rxpad = sercomRXPad3
	case PA17:
		rxpad = sercomRXPad1
	case PA23:
		rxpad = sercomRXPad1
	default:
		panic("Invalid RX pin for UART")
	}

	// configure pins
	config.TX.Configure(PinConfig{Mode: uart.Mode})
	config.RX.Configure(PinConfig{Mode: uart.Mode})

	// reset SERCOM0
	uart.Bus.CTRLA.SetBits(sam.SERCOM_USART_CTRLA_SWRST)
	for uart.Bus.CTRLA.HasBits(sam.SERCOM_USART_CTRLA_SWRST) ||
		uart.Bus.SYNCBUSY.HasBits(sam.SERCOM_USART_SYNCBUSY_SWRST) {
	}

	// set UART mode/sample rate
	// SERCOM_USART_CTRLA_MODE(mode) |
	// SERCOM_USART_CTRLA_SAMPR(sampleRate);
	uart.Bus.CTRLA.Set((sam.SERCOM_USART_CTRLA_MODE_USART_INT_CLK << sam.SERCOM_USART_CTRLA_MODE_Pos) |
		(1 << sam.SERCOM_USART_CTRLA_SAMPR_Pos)) // sample rate of 16x

	// Set baud rate
	uart.SetBaudRate(config.BaudRate)

	// setup UART frame
	// SERCOM_USART_CTRLA_FORM( (parityMode == SERCOM_NO_PARITY ? 0 : 1) ) |
	// dataOrder << SERCOM_USART_CTRLA_DORD_Pos;
	uart.Bus.CTRLA.SetBits((0 << sam.SERCOM_USART_CTRLA_FORM_Pos) | // no parity
		(lsbFirst << sam.SERCOM_USART_CTRLA_DORD_Pos)) // data order

	// set UART stop bits/parity
	// SERCOM_USART_CTRLB_CHSIZE(charSize) |
	// 	nbStopBits << SERCOM_USART_CTRLB_SBMODE_Pos |
	// 	(parityMode == SERCOM_NO_PARITY ? 0 : parityMode) << SERCOM_USART_CTRLB_PMODE_Pos; //If no parity use default value
	uart.Bus.CTRLB.SetBits((0 << sam.SERCOM_USART_CTRLB_CHSIZE_Pos) | // 8 bits is 0
		(0 << sam.SERCOM_USART_CTRLB_SBMODE_Pos) | // 1 stop bit is zero
		(0 << sam.SERCOM_USART_CTRLB_PMODE_Pos)) // no parity

	// set UART pads. This is not same as pins...
	//  SERCOM_USART_CTRLA_TXPO(txPad) |
	//   SERCOM_USART_CTRLA_RXPO(rxPad);
	uart.Bus.CTRLA.SetBits(uint32((txpad << sam.SERCOM_USART_CTRLA_TXPO_Pos) |
		(rxpad << sam.SERCOM_USART_CTRLA_RXPO_Pos)))

	// Enable Transceiver and Receiver
	//sercom->USART.CTRLB.reg |= SERCOM_USART_CTRLB_TXEN | SERCOM_USART_CTRLB_RXEN ;
	uart.Bus.CTRLB.SetBits(sam.SERCOM_USART_CTRLB_TXEN | sam.SERCOM_USART_CTRLB_RXEN)

	// Enable USART1 port.
	// sercom->USART.CTRLA.bit.ENABLE = 0x1u;
	uart.Bus.CTRLA.SetBits(sam.SERCOM_USART_CTRLA_ENABLE)
	for uart.Bus.SYNCBUSY.HasBits(sam.SERCOM_USART_SYNCBUSY_ENABLE) {
	}

	// setup interrupt on receive
	uart.Bus.INTENSET.Set(sam.SERCOM_USART_INTENSET_RXC)

	// Enable RX IRQ.
	arm.EnableIRQ(uart.IRQVal)
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
	uart.Bus.BAUD.Set(uint16(((baud % 8) << sam.SERCOM_USART_BAUD_FRAC_MODE_FP_Pos) |
		((baud / 8) << sam.SERCOM_USART_BAUD_FRAC_MODE_BAUD_Pos)))
}

// WriteByte writes a byte of data to the UART.
func (uart UART) WriteByte(c byte) error {
	// wait until ready to receive
	for !uart.Bus.INTFLAG.HasBits(sam.SERCOM_USART_INTFLAG_DRE) {
	}
	uart.Bus.DATA.Set(uint16(c))
	return nil
}

// defaultUART1Handler handles the UART1 IRQ.
func defaultUART1Handler() {
	// should reset IRQ
	UART1.Receive(byte((UART1.Bus.DATA.Get() & 0xFF)))
	UART1.Bus.INTFLAG.SetBits(sam.SERCOM_USART_INTFLAG_RXC)
}

// I2C on the SAMD21.
type I2C struct {
	Bus     *sam.SERCOM_I2CM_Type
	SCL     Pin
	SDA     Pin
	PinMode PinMode
}

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
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

	// reset SERCOM
	i2c.Bus.CTRLA.SetBits(sam.SERCOM_I2CM_CTRLA_SWRST)
	for i2c.Bus.CTRLA.HasBits(sam.SERCOM_I2CM_CTRLA_SWRST) ||
		i2c.Bus.SYNCBUSY.HasBits(sam.SERCOM_I2CM_SYNCBUSY_SWRST) {
	}

	// Set i2c master mode
	//SERCOM_I2CM_CTRLA_MODE( I2C_MASTER_OPERATION )
	i2c.Bus.CTRLA.Set(sam.SERCOM_I2CM_CTRLA_MODE_I2C_MASTER << sam.SERCOM_I2CM_CTRLA_MODE_Pos) // |

	i2c.SetBaudRate(config.Frequency)

	// Enable I2CM port.
	// sercom->USART.CTRLA.bit.ENABLE = 0x1u;
	i2c.Bus.CTRLA.SetBits(sam.SERCOM_I2CM_CTRLA_ENABLE)
	for i2c.Bus.SYNCBUSY.HasBits(sam.SERCOM_I2CM_SYNCBUSY_ENABLE) {
	}

	// set bus idle mode
	i2c.Bus.STATUS.SetBits(wireIdleState << sam.SERCOM_I2CM_STATUS_BUSSTATE_Pos)
	for i2c.Bus.SYNCBUSY.HasBits(sam.SERCOM_I2CM_SYNCBUSY_SYSOP) {
	}

	// enable pins
	i2c.SDA.Configure(PinConfig{Mode: i2c.PinMode})
	i2c.SCL.Configure(PinConfig{Mode: i2c.PinMode})
}

// SetBaudRate sets the communication speed for the I2C.
func (i2c I2C) SetBaudRate(br uint32) {
	// Synchronous arithmetic baudrate, via Arduino SAMD implementation:
	// SystemCoreClock / ( 2 * baudrate) - 5 - (((SystemCoreClock / 1000000) * WIRE_RISE_TIME_NANOSECONDS) / (2 * 1000));
	baud := CPU_FREQUENCY/(2*br) - 5 - (((CPU_FREQUENCY / 1000000) * riseTimeNanoseconds) / (2 * 1000))
	i2c.Bus.BAUD.Set(baud)
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
		for !i2c.Bus.INTFLAG.HasBits(sam.SERCOM_I2CM_INTFLAG_MB) {
			timeout--
			if timeout == 0 {
				return errors.New("I2C timeout on ready to write data")
			}
		}

		// ACK received (0: ACK, 1: NACK)
		if i2c.Bus.STATUS.HasBits(sam.SERCOM_I2CM_STATUS_RXNACK) {
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
		for !i2c.Bus.INTFLAG.HasBits(sam.SERCOM_I2CM_INTFLAG_SB) {
			// If the slave NACKS the address, the MB bit will be set.
			// In that case, send a stop condition and return error.
			if i2c.Bus.INTFLAG.HasBits(sam.SERCOM_I2CM_INTFLAG_MB) {
				i2c.Bus.CTRLB.SetBits(wireCmdStop << sam.SERCOM_I2CM_CTRLB_CMD_Pos) // Stop condition
				return errors.New("I2C read error: expected ACK not NACK")
			}
		}

		// ACK received (0: ACK, 1: NACK)
		if i2c.Bus.STATUS.HasBits(sam.SERCOM_I2CM_STATUS_RXNACK) {
			return errors.New("I2C read error: expected ACK not NACK")
		}

		// read first byte
		r[0] = i2c.readByte()
		for i := 1; i < len(r); i++ {
			// Send an ACK
			i2c.Bus.CTRLB.ClearBits(sam.SERCOM_I2CM_CTRLB_ACKACT)

			i2c.signalRead()

			// Read data and send the ACK
			r[i] = i2c.readByte()
		}

		// Send NACK to end transmission
		i2c.Bus.CTRLB.SetBits(sam.SERCOM_I2CM_CTRLB_ACKACT)

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
	i2c.Bus.DATA.Set(data)

	// wait until transmission successful
	timeout := i2cTimeout
	for !i2c.Bus.INTFLAG.HasBits(sam.SERCOM_I2CM_INTFLAG_MB) {
		// check for bus error
		if sam.SERCOM3_I2CM.STATUS.HasBits(sam.SERCOM_I2CM_STATUS_BUSERR) {
			return errors.New("I2C bus error")
		}
		timeout--
		if timeout == 0 {
			return errors.New("I2C timeout on write data")
		}
	}

	if i2c.Bus.STATUS.HasBits(sam.SERCOM_I2CM_STATUS_RXNACK) {
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
	for !i2c.Bus.STATUS.HasBits(wireIdleState<<sam.SERCOM_I2CM_STATUS_BUSSTATE_Pos) &&
		!i2c.Bus.STATUS.HasBits(wireOwnerState<<sam.SERCOM_I2CM_STATUS_BUSSTATE_Pos) {
		timeout--
		if timeout == 0 {
			return errors.New("I2C timeout on bus ready")
		}
	}
	i2c.Bus.ADDR.Set(uint32(data))

	return nil
}

func (i2c I2C) signalStop() error {
	i2c.Bus.CTRLB.SetBits(wireCmdStop << sam.SERCOM_I2CM_CTRLB_CMD_Pos) // Stop command
	timeout := i2cTimeout
	for i2c.Bus.SYNCBUSY.HasBits(sam.SERCOM_I2CM_SYNCBUSY_SYSOP) {
		timeout--
		if timeout == 0 {
			return errors.New("I2C timeout on signal stop")
		}
	}
	return nil
}

func (i2c I2C) signalRead() error {
	i2c.Bus.CTRLB.SetBits(wireCmdRead << sam.SERCOM_I2CM_CTRLB_CMD_Pos) // Read command
	timeout := i2cTimeout
	for i2c.Bus.SYNCBUSY.HasBits(sam.SERCOM_I2CM_SYNCBUSY_SYSOP) {
		timeout--
		if timeout == 0 {
			return errors.New("I2C timeout on signal read")
		}
	}
	return nil
}

func (i2c I2C) readByte() byte {
	for !i2c.Bus.INTFLAG.HasBits(sam.SERCOM_I2CM_INTFLAG_SB) {
	}
	return byte(i2c.Bus.DATA.Get())
}

// I2S on the SAMD21.

// I2S
type I2S struct {
	Bus *sam.I2S_Type
}

// Configure is used to configure the I2S interface. You must call this
// before you can use the I2S bus.
func (i2s I2S) Configure(config I2SConfig) {
	// handle defaults
	if config.SCK == 0 {
		config.SCK = I2S_SCK_PIN
		config.WS = I2S_WS_PIN
		config.SD = I2S_SD_PIN
	}

	if config.AudioFrequency == 0 {
		config.AudioFrequency = 48000
	}

	if config.DataFormat == I2SDataFormatDefault {
		if config.Stereo {
			config.DataFormat = I2SDataFormat16bit
		} else {
			config.DataFormat = I2SDataFormat32bit
		}
	}

	// Turn on clock for I2S
	sam.PM.APBCMASK.SetBits(sam.PM_APBCMASK_I2S_)

	// setting clock rate for sample.
	division_factor := CPU_FREQUENCY / (config.AudioFrequency * uint32(config.DataFormat))

	// Switch Generic Clock Generator 3 to DFLL48M.
	sam.GCLK.GENDIV.Set((sam.GCLK_CLKCTRL_GEN_GCLK3 << sam.GCLK_GENDIV_ID_Pos) |
		(division_factor << sam.GCLK_GENDIV_DIV_Pos))
	waitForSync()

	sam.GCLK.GENCTRL.Set((sam.GCLK_CLKCTRL_GEN_GCLK3 << sam.GCLK_GENCTRL_ID_Pos) |
		(sam.GCLK_GENCTRL_SRC_DFLL48M << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_IDC |
		sam.GCLK_GENCTRL_GENEN)
	waitForSync()

	// Use Generic Clock Generator 3 as source for I2S.
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_I2S_0 << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK3 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	waitForSync()

	// reset the device
	i2s.Bus.CTRLA.SetBits(sam.I2S_CTRLA_SWRST)
	for i2s.Bus.SYNCBUSY.HasBits(sam.I2S_SYNCBUSY_SWRST) {
	}

	// disable device before continuing
	for i2s.Bus.SYNCBUSY.HasBits(sam.I2S_SYNCBUSY_ENABLE) {
	}
	i2s.Bus.CTRLA.ClearBits(sam.I2S_CTRLA_ENABLE)

	// setup clock
	if config.ClockSource == I2SClockSourceInternal {
		// TODO: make sure correct for I2S output

		// set serial clock select pin
		i2s.Bus.CLKCTRL0.SetBits(sam.I2S_CLKCTRL_SCKSEL)

		// set frame select pin
		i2s.Bus.CLKCTRL0.SetBits(sam.I2S_CLKCTRL_FSSEL)
	} else {
		// Configure FS generation from SCK clock.
		i2s.Bus.CLKCTRL0.ClearBits(sam.I2S_CLKCTRL_FSSEL)
	}

	if config.Standard == I2StandardPhilips {
		// set 1-bit delay
		i2s.Bus.CLKCTRL0.SetBits(sam.I2S_CLKCTRL_BITDELAY)
	} else {
		// set 0-bit delay
		i2s.Bus.CLKCTRL0.ClearBits(sam.I2S_CLKCTRL_BITDELAY)
	}

	// set number of slots.
	if config.Stereo {
		i2s.Bus.CLKCTRL0.SetBits(1 << sam.I2S_CLKCTRL_NBSLOTS_Pos)
	} else {
		i2s.Bus.CLKCTRL0.ClearBits(1 << sam.I2S_CLKCTRL_NBSLOTS_Pos)
	}

	// set slot size
	switch config.DataFormat {
	case I2SDataFormat8bit:
		i2s.Bus.CLKCTRL0.SetBits(sam.I2S_CLKCTRL_SLOTSIZE_8)

	case I2SDataFormat16bit:
		i2s.Bus.CLKCTRL0.SetBits(sam.I2S_CLKCTRL_SLOTSIZE_16)

	case I2SDataFormat24bit:
		i2s.Bus.CLKCTRL0.SetBits(sam.I2S_CLKCTRL_SLOTSIZE_24)

	case I2SDataFormat32bit:
		i2s.Bus.CLKCTRL0.SetBits(sam.I2S_CLKCTRL_SLOTSIZE_32)
	}

	// configure pin for clock
	config.SCK.Configure(PinConfig{Mode: PinCom})

	// configure pin for WS, if needed
	if config.WS != NoPin {
		config.WS.Configure(PinConfig{Mode: PinCom})
	}

	// now set serializer data size.
	switch config.DataFormat {
	case I2SDataFormat8bit:
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_DATASIZE_8)

	case I2SDataFormat16bit:
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_DATASIZE_16)

	case I2SDataFormat24bit:
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_DATASIZE_24)

	case I2SDataFormat32bit:
	case I2SDataFormatDefault:
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_DATASIZE_32)
	}

	// set serializer slot adjustment
	if config.Standard == I2SStandardLSB {
		// adjust right
		i2s.Bus.SERCTRL1.ClearBits(sam.I2S_SERCTRL_SLOTADJ)
	} else {
		// adjust left
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_SLOTADJ)

		// reverse bit order?
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_BITREV)
	}

	// set serializer mode.
	if config.Mode == I2SModePDM {
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_SERMODE_PDM2)
	} else {
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_SERMODE_RX)
	}

	// configure data pin
	config.SD.Configure(PinConfig{Mode: PinCom})

	// re-enable
	i2s.Bus.CTRLA.SetBits(sam.I2S_CTRLA_ENABLE)
	for i2s.Bus.SYNCBUSY.HasBits(sam.I2S_SYNCBUSY_ENABLE) {
	}

	// enable i2s clock
	i2s.Bus.CTRLA.SetBits(sam.I2S_CTRLA_CKEN0)
	for i2s.Bus.SYNCBUSY.HasBits(sam.I2S_SYNCBUSY_CKEN0) {
	}

	// enable i2s serializer
	i2s.Bus.CTRLA.SetBits(sam.I2S_CTRLA_SEREN1)
	for i2s.Bus.SYNCBUSY.HasBits(sam.I2S_SYNCBUSY_SEREN1) {
	}
}

// Read data from the I2S bus into the provided slice.
// The I2S bus must already have been configured correctly.
func (i2s I2S) Read(p []uint32) (n int, err error) {
	i := 0
	for i = 0; i < len(p); i++ {
		// Wait until ready
		for !i2s.Bus.INTFLAG.HasBits(sam.I2S_INTFLAG_RXRDY1) {
		}

		for i2s.Bus.SYNCBUSY.HasBits(sam.I2S_SYNCBUSY_DATA1) {
		}

		// read data
		p[i] = i2s.Bus.DATA1.Get()

		// indicate read complete
		i2s.Bus.INTFLAG.Set(sam.I2S_INTFLAG_RXRDY1)
	}

	return i, nil
}

// Write data to the I2S bus from the provided slice.
// The I2S bus must already have been configured correctly.
func (i2s I2S) Write(p []uint32) (n int, err error) {
	i := 0
	for i = 0; i < len(p); i++ {
		// Wait until ready
		for !i2s.Bus.INTFLAG.HasBits(sam.I2S_INTFLAG_TXRDY1) {
		}

		for i2s.Bus.SYNCBUSY.HasBits(sam.I2S_SYNCBUSY_DATA1) {
		}

		// write data
		i2s.Bus.DATA1.Set(p[i])

		// indicate write complete
		i2s.Bus.INTFLAG.Set(sam.I2S_INTFLAG_TXRDY1)
	}

	return i, nil
}

// Close the I2S bus.
func (i2s I2S) Close() error {
	// Sync wait
	for i2s.Bus.SYNCBUSY.HasBits(sam.I2S_SYNCBUSY_ENABLE) {
	}

	// disable I2S
	i2s.Bus.CTRLA.ClearBits(sam.I2S_CTRLA_ENABLE)

	return nil
}

func waitForSync() {
	for sam.GCLK.STATUS.HasBits(sam.GCLK_STATUS_SYNCBUSY) {
	}
}

// SPI
type SPI struct {
	Bus *sam.SERCOM_SPI_Type
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
func (spi SPI) Configure(config SPIConfig) {
	config.SCK = SPI0_SCK_PIN
	config.MOSI = SPI0_MOSI_PIN
	config.MISO = SPI0_MISO_PIN

	doPad := spiTXPad2SCK3
	diPad := sercomRXPad0

	// set default frequency
	if config.Frequency == 0 {
		config.Frequency = 4000000
	}

	// Disable SPI port.
	spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPI_CTRLA_ENABLE)
	for spi.Bus.SYNCBUSY.HasBits(sam.SERCOM_SPI_SYNCBUSY_ENABLE) {
	}

	// enable pins
	config.SCK.Configure(PinConfig{Mode: PinSERCOMAlt})
	config.MOSI.Configure(PinConfig{Mode: PinSERCOMAlt})
	config.MISO.Configure(PinConfig{Mode: PinSERCOMAlt})

	// reset SERCOM
	spi.Bus.CTRLA.SetBits(sam.SERCOM_SPI_CTRLA_SWRST)
	for spi.Bus.CTRLA.HasBits(sam.SERCOM_SPI_CTRLA_SWRST) ||
		spi.Bus.SYNCBUSY.HasBits(sam.SERCOM_SPI_SYNCBUSY_SWRST) {
	}

	// set bit transfer order
	dataOrder := 0
	if config.LSBFirst {
		dataOrder = 1
	}

	// Set SPI master
	spi.Bus.CTRLA.Set(uint32((sam.SERCOM_SPI_CTRLA_MODE_SPI_MASTER << sam.SERCOM_SPI_CTRLA_MODE_Pos) |
		(doPad << sam.SERCOM_SPI_CTRLA_DOPO_Pos) |
		(diPad << sam.SERCOM_SPI_CTRLA_DIPO_Pos) |
		(dataOrder << sam.SERCOM_SPI_CTRLA_DORD_Pos)))

	spi.Bus.CTRLB.SetBits((0 << sam.SERCOM_SPI_CTRLB_CHSIZE_Pos) | // 8bit char size
		sam.SERCOM_SPI_CTRLB_RXEN) // receive enable

	for spi.Bus.SYNCBUSY.HasBits(sam.SERCOM_SPI_SYNCBUSY_CTRLB) {
	}

	// set mode
	switch config.Mode {
	case 0:
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPI_CTRLA_CPHA)
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPI_CTRLA_CPOL)
	case 1:
		spi.Bus.CTRLA.SetBits(sam.SERCOM_SPI_CTRLA_CPHA)
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPI_CTRLA_CPOL)
	case 2:
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPI_CTRLA_CPHA)
		spi.Bus.CTRLA.SetBits(sam.SERCOM_SPI_CTRLA_CPOL)
	case 3:
		spi.Bus.CTRLA.SetBits(sam.SERCOM_SPI_CTRLA_CPHA | sam.SERCOM_SPI_CTRLA_CPOL)
	default: // to mode 0
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPI_CTRLA_CPHA)
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPI_CTRLA_CPOL)
	}

	// Set synch speed for SPI
	baudRate := (CPU_FREQUENCY / (2 * config.Frequency)) - 1
	spi.Bus.BAUD.Set(uint8(baudRate))

	// Enable SPI port.
	spi.Bus.CTRLA.SetBits(sam.SERCOM_SPI_CTRLA_ENABLE)
	for spi.Bus.SYNCBUSY.HasBits(sam.SERCOM_SPI_SYNCBUSY_ENABLE) {
	}
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi SPI) Transfer(w byte) (byte, error) {
	// write data
	spi.Bus.DATA.Set(uint32(w))

	// wait for receive
	for !spi.Bus.INTFLAG.HasBits(sam.SERCOM_SPI_INTFLAG_RXC) {
	}

	// return data
	return byte(spi.Bus.DATA.Get()), nil
}

// PWM
const period = 0xFFFF

// InitPWM initializes the PWM interface.
func InitPWM() {
	// turn on timer clocks used for PWM
	sam.PM.APBCMASK.SetBits(sam.PM_APBCMASK_TCC0_ | sam.PM_APBCMASK_TCC1_ | sam.PM_APBCMASK_TCC2_)

	// Use GCLK0 for TCC0/TCC1
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_TCC0_TCC1 << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	for sam.GCLK.STATUS.HasBits(sam.GCLK_STATUS_SYNCBUSY) {
	}

	// Use GCLK0 for TCC2/TC3
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_TCC2_TC3 << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	for sam.GCLK.STATUS.HasBits(sam.GCLK_STATUS_SYNCBUSY) {
	}
}

// Configure configures a PWM pin for output.
func (pwm PWM) Configure() {
	// figure out which TCCX timer for this pin
	timer := pwm.getTimer()

	// disable timer
	timer.CTRLA.ClearBits(sam.TCC_CTRLA_ENABLE)
	// Wait for synchronization
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_ENABLE) {
	}

	// Use "Normal PWM" (single-slope PWM)
	timer.WAVE.SetBits(sam.TCC_WAVE_WAVEGEN_NPWM)
	// Wait for synchronization
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_WAVE) {
	}

	// Set the period (the number to count to (TOP) before resetting timer)
	//TCC0->PER.reg = period;
	timer.PER.Set(period)
	// Wait for synchronization
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_PER) {
	}

	// Set pin as output
	sam.PORT.DIRSET0.Set(1 << uint8(pwm.Pin))
	// Set pin to low
	sam.PORT.OUTCLR0.Set(1 << uint8(pwm.Pin))

	// Enable the port multiplexer for pin
	pwm.setPinCfg(sam.PORT_PINCFG0_PMUXEN)

	// Connect TCCX timer to pin.
	// we normally use the F channel aka ALT
	pwmConfig := PinPWMAlt

	// in the case of PA6 or PA7 we have to use E channel
	if pwm.Pin == 6 || pwm.Pin == 7 {
		pwmConfig = PinPWM
	}

	if pwm.Pin&1 > 0 {
		// odd pin, so save the even pins
		val := pwm.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
		pwm.setPMux(val | uint8(pwmConfig<<sam.PORT_PMUX0_PMUXO_Pos))
	} else {
		// even pin, so save the odd pins
		val := pwm.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
		pwm.setPMux(val | uint8(pwmConfig<<sam.PORT_PMUX0_PMUXE_Pos))
	}
}

// Set turns on the duty cycle for a PWM pin using the provided value.
func (pwm PWM) Set(value uint16) {
	// figure out which TCCX timer for this pin
	timer := pwm.getTimer()

	// disable output
	timer.CTRLA.ClearBits(sam.TCC_CTRLA_ENABLE)

	// Wait for synchronization
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_ENABLE) {
	}

	// Set PWM signal to output duty cycle
	pwm.setChannel(uint32(value))

	// Wait for synchronization on all channels
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CC0 |
		sam.TCC_SYNCBUSY_CC1 |
		sam.TCC_SYNCBUSY_CC2 |
		sam.TCC_SYNCBUSY_CC3) {
	}

	// enable
	timer.CTRLA.SetBits(sam.TCC_CTRLA_ENABLE)
	// Wait for synchronization
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_ENABLE) {
	}
}

// getPMux returns the value for the correct PMUX register for this pin.
func (pwm PWM) getPMux() uint8 {
	return pwm.Pin.getPMux()
}

// setPMux sets the value for the correct PMUX register for this pin.
func (pwm PWM) setPMux(val uint8) {
	pwm.Pin.setPMux(val)
}

// getPinCfg returns the value for the correct PINCFG register for this pin.
func (pwm PWM) getPinCfg() uint8 {
	return pwm.Pin.getPinCfg()
}

// setPinCfg sets the value for the correct PINCFG register for this pin.
func (pwm PWM) setPinCfg(val uint8) {
	pwm.Pin.setPinCfg(val)
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
func (pwm PWM) setChannel(val uint32) {
	switch pwm.Pin {
	case 6:
		pwm.getTimer().CC0.Set(val)
	case 7:
		pwm.getTimer().CC1.Set(val)
	case 8:
		pwm.getTimer().CC0.Set(val)
	case 9:
		pwm.getTimer().CC1.Set(val)
	case 14:
		pwm.getTimer().CC0.Set(val)
	case 15:
		pwm.getTimer().CC1.Set(val)
	case 16:
		pwm.getTimer().CC2.Set(val)
	case 17:
		pwm.getTimer().CC3.Set(val)
	case 18:
		pwm.getTimer().CC2.Set(val)
	case 19:
		pwm.getTimer().CC3.Set(val)
	case 20:
		pwm.getTimer().CC2.Set(val)
	case 21:
		pwm.getTimer().CC3.Set(val)
	default:
		return // not supported on this pin
	}
}

// USBCDC is the USB CDC aka serial over USB interface on the SAMD21.
type USBCDC struct {
	Buffer *RingBuffer
}

// WriteByte writes a byte of data to the USB CDC interface.
func (usbcdc USBCDC) WriteByte(c byte) error {
	// Supposedly to handle problem with Windows USB serial ports?
	if usbLineInfo.lineState > 0 {
		// set the data
		udd_ep_in_cache_buffer[usb_CDC_ENDPOINT_IN][0] = c

		usbEndpointDescriptors[usb_CDC_ENDPOINT_IN].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[usb_CDC_ENDPOINT_IN]))))

		// clean multi packet size of bytes already sent
		usbEndpointDescriptors[usb_CDC_ENDPOINT_IN].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Mask << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

		// set count of bytes to be sent
		usbEndpointDescriptors[usb_CDC_ENDPOINT_IN].DeviceDescBank[1].PCKSIZE.SetBits((1 & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask) << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

		// clear transfer complete flag
		setEPINTFLAG(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_EPINTFLAG_TRCPT1)

		// send data by setting bank ready
		setEPSTATUSSET(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_EPSTATUSSET_BK1RDY)

		// wait for transfer to complete
		timeout := 3000
		for (getEPINTFLAG(usb_CDC_ENDPOINT_IN) & sam.USB_DEVICE_EPINTFLAG_TRCPT1) == 0 {
			timeout--
			if timeout == 0 {
				return errors.New("USBCDC write byte timeout")
			}
		}
	}

	return nil
}

func (usbcdc USBCDC) DTR() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_DTR) > 0
}

func (usbcdc USBCDC) RTS() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_RTS) > 0
}

const (
	// these are SAMD21 specific.
	usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos  = 0
	usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask = 0x3FFF

	usb_DEVICE_PCKSIZE_SIZE_Pos  = 28
	usb_DEVICE_PCKSIZE_SIZE_Mask = 0x7

	usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos  = 14
	usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Mask = 0x3FFF
)

var (
	usbEndpointDescriptors [8]usbDeviceDescriptor

	udd_ep_in_cache_buffer  [7][128]uint8
	udd_ep_out_cache_buffer [7][128]uint8

	isEndpointHalt        = false
	isRemoteWakeUpEnabled = false
	endPoints             = []uint32{usb_ENDPOINT_TYPE_CONTROL,
		(usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointIn),
		(usb_ENDPOINT_TYPE_BULK | usbEndpointOut),
		(usb_ENDPOINT_TYPE_BULK | usbEndpointIn)}

	usbConfiguration uint8
	usbSetInterface  uint8
	usbLineInfo      = cdcLineInfo{115200, 0x00, 0x00, 0x08, 0x00}
)

// Configure the USB CDC interface. The config is here for compatibility with the UART interface.
func (usbcdc USBCDC) Configure(config UARTConfig) {
	// reset USB interface
	sam.USB_DEVICE.CTRLA.SetBits(sam.USB_DEVICE_CTRLA_SWRST)
	for sam.USB_DEVICE.SYNCBUSY.HasBits(sam.USB_DEVICE_SYNCBUSY_SWRST) ||
		sam.USB_DEVICE.SYNCBUSY.HasBits(sam.USB_DEVICE_SYNCBUSY_ENABLE) {
	}

	sam.USB_DEVICE.DESCADD.Set(uint32(uintptr(unsafe.Pointer(&usbEndpointDescriptors))))

	// configure pins
	USBCDC_DM_PIN.Configure(PinConfig{Mode: PinCom})
	USBCDC_DP_PIN.Configure(PinConfig{Mode: PinCom})

	// performs pad calibration from store fuses
	handlePadCalibration()

	// run in standby
	sam.USB_DEVICE.CTRLA.SetBits(sam.USB_DEVICE_CTRLA_RUNSTDBY)

	// set full speed
	sam.USB_DEVICE.CTRLB.SetBits(sam.USB_DEVICE_CTRLB_SPDCONF_FS << sam.USB_DEVICE_CTRLB_SPDCONF_Pos)

	// attach
	sam.USB_DEVICE.CTRLB.ClearBits(sam.USB_DEVICE_CTRLB_DETACH)

	// enable interrupt for end of reset
	sam.USB_DEVICE.INTENSET.SetBits(sam.USB_DEVICE_INTENSET_EORST)

	// enable interrupt for start of frame
	sam.USB_DEVICE.INTENSET.SetBits(sam.USB_DEVICE_INTENSET_SOF)

	// enable USB
	sam.USB_DEVICE.CTRLA.SetBits(sam.USB_DEVICE_CTRLA_ENABLE)

	// enable IRQ
	arm.EnableIRQ(sam.IRQ_USB)
}

func handlePadCalibration() {
	// Load Pad Calibration data from non-volatile memory
	// This requires registers that are not included in the SVD file.
	// Modeled after defines from samd21g18a.h and nvmctrl.h:
	//
	// #define NVMCTRL_OTP4 0x00806020
	//
	// #define USB_FUSES_TRANSN_ADDR       (NVMCTRL_OTP4 + 4)
	// #define USB_FUSES_TRANSN_Pos        13           /**< \brief (NVMCTRL_OTP4) USB pad Transn calibration */
	// #define USB_FUSES_TRANSN_Msk        (0x1Fu << USB_FUSES_TRANSN_Pos)
	// #define USB_FUSES_TRANSN(value)     ((USB_FUSES_TRANSN_Msk & ((value) << USB_FUSES_TRANSN_Pos)))

	// #define USB_FUSES_TRANSP_ADDR       (NVMCTRL_OTP4 + 4)
	// #define USB_FUSES_TRANSP_Pos        18           /**< \brief (NVMCTRL_OTP4) USB pad Transp calibration */
	// #define USB_FUSES_TRANSP_Msk        (0x1Fu << USB_FUSES_TRANSP_Pos)
	// #define USB_FUSES_TRANSP(value)     ((USB_FUSES_TRANSP_Msk & ((value) << USB_FUSES_TRANSP_Pos)))

	// #define USB_FUSES_TRIM_ADDR         (NVMCTRL_OTP4 + 4)
	// #define USB_FUSES_TRIM_Pos          23           /**< \brief (NVMCTRL_OTP4) USB pad Trim calibration */
	// #define USB_FUSES_TRIM_Msk          (0x7u << USB_FUSES_TRIM_Pos)
	// #define USB_FUSES_TRIM(value)       ((USB_FUSES_TRIM_Msk & ((value) << USB_FUSES_TRIM_Pos)))
	//
	fuse := *(*uint32)(unsafe.Pointer(uintptr(0x00806020) + 4))
	calibTransN := uint16(fuse>>13) & uint16(0x1f)
	calibTransP := uint16(fuse>>18) & uint16(0x1f)
	calibTrim := uint16(fuse>>23) & uint16(0x7)

	if calibTransN == 0x1f {
		calibTransN = 5
	}
	sam.USB_DEVICE.PADCAL.SetBits(calibTransN << sam.USB_DEVICE_PADCAL_TRANSN_Pos)

	if calibTransP == 0x1f {
		calibTransP = 29
	}
	sam.USB_DEVICE.PADCAL.SetBits(calibTransP << sam.USB_DEVICE_PADCAL_TRANSP_Pos)

	if calibTrim == 0x7 {
		calibTransN = 3
	}
	sam.USB_DEVICE.PADCAL.SetBits(calibTrim << sam.USB_DEVICE_PADCAL_TRIM_Pos)
}

//go:export USB_IRQHandler
func handleUSB() {
	// reset all interrupt flags
	flags := sam.USB_DEVICE.INTFLAG.Get()
	sam.USB_DEVICE.INTFLAG.Set(flags)

	// End of reset
	if (flags & sam.USB_DEVICE_INTFLAG_EORST) > 0 {
		// Configure control endpoint
		initEndpoint(0, usb_ENDPOINT_TYPE_CONTROL)

		// Enable Setup-Received interrupt
		setEPINTENSET(0, sam.USB_DEVICE_EPINTENSET_RXSTP)

		usbConfiguration = 0

		// ack the End-Of-Reset interrupt
		sam.USB_DEVICE.INTFLAG.Set(sam.USB_DEVICE_INTFLAG_EORST)
	}

	// Start of frame
	if (flags & sam.USB_DEVICE_INTFLAG_SOF) > 0 {
		// if you want to blink LED showing traffic, this would be the place...
	}

	// Endpoint 0 Setup interrupt
	if getEPINTFLAG(0)&sam.USB_DEVICE_EPINTFLAG_RXSTP > 0 {
		// ack setup received
		setEPINTFLAG(0, sam.USB_DEVICE_EPINTFLAG_RXSTP)

		// parse setup
		setup := newUSBSetup(udd_ep_out_cache_buffer[0][:])

		// Clear the Bank 0 ready flag on Control OUT
		setEPSTATUSCLR(0, sam.USB_DEVICE_EPSTATUSCLR_BK0RDY)
		usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

		ok := false
		if (setup.bmRequestType & usb_REQUEST_TYPE) == usb_REQUEST_STANDARD {
			// Standard Requests
			ok = handleStandardSetup(setup)
		} else {
			// Class Interface Requests
			if setup.wIndex == usb_CDC_ACM_INTERFACE {
				ok = cdcSetup(setup)
			}
		}

		if ok {
			// set Bank1 ready
			setEPSTATUSSET(0, sam.USB_DEVICE_EPSTATUSSET_BK1RDY)
		} else {
			// Stall endpoint
			setEPSTATUSSET(0, sam.USB_DEVICE_EPINTFLAG_STALL1)
		}

		if getEPINTFLAG(0)&sam.USB_DEVICE_EPINTFLAG_STALL1 > 0 {
			// ack the stall
			setEPINTFLAG(0, sam.USB_DEVICE_EPINTFLAG_STALL1)

			// clear stall request
			setEPINTENCLR(0, sam.USB_DEVICE_EPINTENCLR_STALL1)
		}
	}

	// Now the actual transfer handlers, ignore endpoint number 0 (setup)
	var i uint32
	for i = 1; i < uint32(len(endPoints)); i++ {
		// Check if endpoint has a pending interrupt
		epFlags := getEPINTFLAG(i)
		if epFlags > 0 {
			switch i {
			case usb_CDC_ENDPOINT_OUT:
				if (epFlags & sam.USB_DEVICE_EPINTFLAG_TRCPT0) > 0 {
					handleEndpoint(i)
				}
				setEPINTFLAG(i, epFlags)
			case usb_CDC_ENDPOINT_IN, usb_CDC_ENDPOINT_ACM:
				// set bank ready
				setEPSTATUSCLR(i, sam.USB_DEVICE_EPSTATUSCLR_BK1RDY)

				// ack transfer complete
				setEPINTFLAG(i, sam.USB_DEVICE_EPINTFLAG_TRCPT1)
			}
		}
	}
}

func initEndpoint(ep, config uint32) {
	switch config {
	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointIn:
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_INTERRUPT + 1) << sam.USB_DEVICE_EPCFG_EPTYPE1_Pos))

	case usb_ENDPOINT_TYPE_BULK | usbEndpointOut:
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_BULK + 1) << sam.USB_DEVICE_EPCFG_EPTYPE0_Pos))

		// receive interrupts when current transfer complete
		setEPINTENSET(ep, sam.USB_DEVICE_EPINTENSET_TRCPT0)

		// set byte count to zero, we have not received anything yet
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

		// ready for next transfer
		setEPSTATUSCLR(ep, sam.USB_DEVICE_EPSTATUSCLR_BK0RDY)

	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointOut:
		// TODO: not really anything, seems like...

	case usb_ENDPOINT_TYPE_BULK | usbEndpointIn:
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_BULK + 1) << sam.USB_DEVICE_EPCFG_EPTYPE1_Pos))

		// NAK on endpoint IN, the bank is not yet filled in.
		setEPSTATUSCLR(ep, sam.USB_DEVICE_EPSTATUSCLR_BK1RDY)

	case usb_ENDPOINT_TYPE_CONTROL:
		// Control OUT
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, getEPCFG(ep)|((usb_ENDPOINT_TYPE_CONTROL+1)<<sam.USB_DEVICE_EPCFG_EPTYPE0_Pos))

		// Control IN
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, getEPCFG(ep)|((usb_ENDPOINT_TYPE_CONTROL+1)<<sam.USB_DEVICE_EPCFG_EPTYPE1_Pos))

		// Prepare OUT endpoint for receive
		// set multi packet size for expected number of receive bytes on control OUT
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(64 << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

		// set byte count to zero, we have not received anything yet
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

		// NAK on endpoint OUT to show we are ready to receive control data
		setEPSTATUSSET(ep, sam.USB_DEVICE_EPSTATUSSET_BK0RDY)
	}
}

func handleStandardSetup(setup usbSetup) bool {
	switch setup.bRequest {
	case usb_GET_STATUS:
		buf := []byte{0, 0}

		if setup.bmRequestType != 0 { // endpoint
			// TODO: actually check if the endpoint in question is currently halted
			if isEndpointHalt {
				buf[0] = 1
			}
		}

		sendUSBPacket(0, buf)
		return true

	case usb_CLEAR_FEATURE:
		if setup.wValueL == 1 { // DEVICEREMOTEWAKEUP
			isRemoteWakeUpEnabled = false
		} else if setup.wValueL == 0 { // ENDPOINTHALT
			isEndpointHalt = false
		}
		sendZlp(0)
		return true

	case usb_SET_FEATURE:
		if setup.wValueL == 1 { // DEVICEREMOTEWAKEUP
			isRemoteWakeUpEnabled = true
		} else if setup.wValueL == 0 { // ENDPOINTHALT
			isEndpointHalt = true
		}
		sendZlp(0)
		return true

	case usb_SET_ADDRESS:
		// set packet size 64 with auto Zlp after transfer
		usbEndpointDescriptors[0].DeviceDescBank[1].PCKSIZE.Set((epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos) |
			uint32(1<<31)) // autozlp

		// ack the transfer is complete from the request
		setEPINTFLAG(0, sam.USB_DEVICE_EPINTFLAG_TRCPT1)

		// set bank ready for data
		setEPSTATUSSET(0, sam.USB_DEVICE_EPSTATUSSET_BK1RDY)

		// wait for transfer to complete
		timeout := 3000
		for (getEPINTFLAG(0) & sam.USB_DEVICE_EPINTFLAG_TRCPT1) == 0 {
			timeout--
			if timeout == 0 {
				return true
			}
		}

		// last, set the device address to that requested by host
		sam.USB_DEVICE.DADD.SetBits(setup.wValueL)
		sam.USB_DEVICE.DADD.SetBits(sam.USB_DEVICE_DADD_ADDEN)

		return true

	case usb_GET_DESCRIPTOR:
		sendDescriptor(setup)
		return true

	case usb_SET_DESCRIPTOR:
		return false

	case usb_GET_CONFIGURATION:
		buff := []byte{usbConfiguration}
		sendUSBPacket(0, buff)
		return true

	case usb_SET_CONFIGURATION:
		if setup.bmRequestType&usb_REQUEST_RECIPIENT == usb_REQUEST_DEVICE {
			for i := 1; i < len(endPoints); i++ {
				initEndpoint(uint32(i), endPoints[i])
			}

			usbConfiguration = setup.wValueL

			// Enable interrupt for CDC control messages from host (OUT packet)
			setEPINTENSET(usb_CDC_ENDPOINT_ACM, sam.USB_DEVICE_EPINTENSET_TRCPT1)

			// Enable interrupt for CDC data messages from host
			setEPINTENSET(usb_CDC_ENDPOINT_OUT, sam.USB_DEVICE_EPINTENSET_TRCPT0)

			sendZlp(0)
			return true
		} else {
			return false
		}

	case usb_GET_INTERFACE:
		buff := []byte{usbSetInterface}
		sendUSBPacket(0, buff)
		return true

	case usb_SET_INTERFACE:
		usbSetInterface = setup.wValueL

		sendZlp(0)
		return true

	default:
		return true
	}
}

func cdcSetup(setup usbSetup) bool {
	if setup.bmRequestType == usb_REQUEST_DEVICETOHOST_CLASS_INTERFACE {
		if setup.bRequest == usb_CDC_GET_LINE_CODING {
			buf := bytes.NewBuffer(make([]byte, 0, 7))
			binary.Write(buf, binary.LittleEndian, usbLineInfo.dwDTERate)
			binary.Write(buf, binary.LittleEndian, usbLineInfo.bCharFormat)
			binary.Write(buf, binary.LittleEndian, usbLineInfo.bParityType)
			binary.Write(buf, binary.LittleEndian, usbLineInfo.bDataBits)

			sendUSBPacket(0, buf.Bytes())
			return true
		}
	}

	if setup.bmRequestType == usb_REQUEST_HOSTTODEVICE_CLASS_INTERFACE {
		if setup.bRequest == usb_CDC_SET_LINE_CODING {
			buf := bytes.NewBuffer(receiveUSBControlPacket())
			binary.Read(buf, binary.LittleEndian, &(usbLineInfo.dwDTERate))
			binary.Read(buf, binary.LittleEndian, &(usbLineInfo.bCharFormat))
			binary.Read(buf, binary.LittleEndian, &(usbLineInfo.bParityType))
			binary.Read(buf, binary.LittleEndian, &(usbLineInfo.bDataBits))
		}

		if setup.bRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			usbLineInfo.lineState = setup.wValueL
		}

		if setup.bRequest == usb_CDC_SET_LINE_CODING || setup.bRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			// auto-reset into the bootloader
			if usbLineInfo.dwDTERate == 1200 && (usbLineInfo.lineState&0x01) == 0 {
				// TODO: system reset
			} else {
				// TODO: cancel any reset
			}
		}

		if setup.bRequest == usb_CDC_SEND_BREAK {
			// TODO: something with this value?
			// breakValue = ((uint16_t)setup.wValueH << 8) | setup.wValueL;
			// return false;
		}
		return true
	}
	return false
}

func sendUSBPacket(ep uint32, data []byte) {
	copy(udd_ep_in_cache_buffer[ep][:], data)

	// Set endpoint address for sending data
	usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

	// clear multi-packet size which is total bytes already sent
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Mask << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

	// set byte count, which is total number of bytes to be sent
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(uint32((len(data) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask) << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos))
}

func receiveUSBControlPacket() []byte {
	// set ready to receive data
	setEPSTATUSCLR(0, sam.USB_DEVICE_EPSTATUSCLR_BK0RDY)

	// read the data
	bytesread := armRecvCtrlOUT(0)

	// return the data
	data := make([]byte, 0, bytesread)
	copy(data, udd_ep_out_cache_buffer[0][:bytesread])
	return data
}

func armRecvCtrlOUT(ep uint32) uint32 {
	// Set output address to receive data
	usbEndpointDescriptors[ep].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[ep]))))

	// set multi-packet size which is total expected number of bytes to receive.
	usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits((8 << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos) |
		uint32(epPacketSize(64)<<usb_DEVICE_PCKSIZE_SIZE_Pos))

	// clear byte count of bytes received so far.
	usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

	// clear ready state to start transfer
	setEPSTATUSCLR(ep, sam.USB_DEVICE_EPSTATUSCLR_BK0RDY)

	// Wait until OUT transfer is ready.
	timeout := 3000
	for (getEPSTATUS(ep) & sam.USB_DEVICE_EPSTATUS_BK0RDY) == 0 {
		timeout--
		if timeout == 0 {
			return 0
		}
	}

	// Wait until OUT transfer is completed.
	timeout = 3000
	for (getEPINTFLAG(ep) & sam.USB_DEVICE_EPINTFLAG_TRCPT0) == 0 {
		timeout--
		if timeout == 0 {
			return 0
		}
	}

	// return number of bytes received
	return (usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.Get() >>
		usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask
}

// sendDescriptor creates and sends the various USB descriptor types that
// can be requested by the host.
func sendDescriptor(setup usbSetup) {
	switch setup.wValueH {
	case usb_CONFIGURATION_DESCRIPTOR_TYPE:
		sendConfiguration(setup)
		return
	case usb_DEVICE_DESCRIPTOR_TYPE:
		if setup.wLength == 8 {
			// composite descriptor requested, so only send 8 bytes
			dd := NewDeviceDescriptor(0xEF, 0x02, 0x01, 64, usb_VID, usb_PID, 0x100, usb_IMANUFACTURER, usb_IPRODUCT, usb_ISERIAL, 1)
			sendUSBPacket(0, dd.Bytes()[:8])
		} else {
			// complete descriptor requested so send entire packet
			dd := NewDeviceDescriptor(0x00, 0x00, 0x00, 64, usb_VID, usb_PID, 0x100, usb_IMANUFACTURER, usb_IPRODUCT, usb_ISERIAL, 1)
			sendUSBPacket(0, dd.Bytes())
		}
		return

	case usb_STRING_DESCRIPTOR_TYPE:
		switch setup.wValueL {
		case 0:
			b := make([]byte, 4)
			b[0] = byte(usb_STRING_LANGUAGE[0] >> 8)
			b[1] = byte(usb_STRING_LANGUAGE[0] & 0xff)
			b[2] = byte(usb_STRING_LANGUAGE[1] >> 8)
			b[3] = byte(usb_STRING_LANGUAGE[1] & 0xff)
			sendUSBPacket(0, b)

		case usb_IPRODUCT:
			prod := []byte(usb_STRING_PRODUCT)
			b := make([]byte, len(prod)*2+2)
			b[0] = byte(len(prod)*2 + 2)
			b[1] = 0x03

			for i, val := range prod {
				b[i*2] = 0
				b[i*2+1] = val
			}

			sendUSBPacket(0, b)

		case usb_IMANUFACTURER:
			prod := []byte(usb_STRING_MANUFACTURER)
			b := make([]byte, len(prod)*2+2)
			b[0] = byte(len(prod)*2 + 2)
			b[1] = 0x03

			for i, val := range prod {
				b[i*2] = 0
				b[i*2+1] = val
			}

			sendUSBPacket(0, b)

		case usb_ISERIAL:
			// TODO: allow returning a product serial number
			sendZlp(0)
		}

		// send final zero length packet and return
		sendZlp(0)
		return
	}

	// do not know how to handle this message, so return zero
	sendZlp(0)
	return
}

// sendConfiguration creates and sends the configuration packet to the host.
func sendConfiguration(setup usbSetup) {
	if setup.wLength == 9 {
		sz := uint16(configDescriptorSize + cdcSize)
		config := NewConfigDescriptor(sz, 2)
		sendUSBPacket(0, config.Bytes())
	} else {
		iad := NewIADDescriptor(0, 2, usb_CDC_COMMUNICATION_INTERFACE_CLASS, usb_CDC_ABSTRACT_CONTROL_MODEL, 0)

		cif := NewInterfaceDescriptor(usb_CDC_ACM_INTERFACE, 1, usb_CDC_COMMUNICATION_INTERFACE_CLASS, usb_CDC_ABSTRACT_CONTROL_MODEL, 0)

		header := NewCDCCSInterfaceDescriptor(usb_CDC_HEADER, usb_CDC_V1_10&0xFF, (usb_CDC_V1_10>>8)&0x0FF)

		controlManagement := NewACMFunctionalDescriptor(usb_CDC_ABSTRACT_CONTROL_MANAGEMENT, 6)

		functionalDescriptor := NewCDCCSInterfaceDescriptor(usb_CDC_UNION, usb_CDC_ACM_INTERFACE, usb_CDC_DATA_INTERFACE)

		callManagement := NewCMFunctionalDescriptor(usb_CDC_CALL_MANAGEMENT, 1, 1)

		cifin := NewEndpointDescriptor((usb_CDC_ENDPOINT_ACM | usbEndpointIn), usb_ENDPOINT_TYPE_INTERRUPT, 0x10, 0x10)

		dif := NewInterfaceDescriptor(usb_CDC_DATA_INTERFACE, 2, usb_CDC_DATA_INTERFACE_CLASS, 0, 0)

		in := NewEndpointDescriptor((usb_CDC_ENDPOINT_OUT | usbEndpointOut), usb_ENDPOINT_TYPE_BULK, usbEndpointPacketSize, 0)

		out := NewEndpointDescriptor((usb_CDC_ENDPOINT_IN | usbEndpointIn), usb_ENDPOINT_TYPE_BULK, usbEndpointPacketSize, 0)

		cdc := NewCDCDescriptor(iad,
			cif,
			header,
			controlManagement,
			functionalDescriptor,
			callManagement,
			cifin,
			dif,
			in,
			out)

		sz := uint16(configDescriptorSize + cdcSize)
		config := NewConfigDescriptor(sz, 2)

		buf := make([]byte, 0, sz)
		buf = append(buf, config.Bytes()...)
		buf = append(buf, cdc.Bytes()...)

		sendUSBPacket(0, buf)
	}
}

func handleEndpoint(ep uint32) {
	// get data
	count := int((usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.Get() >>
		usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask)

	// move to ring buffer
	for i := 0; i < count; i++ {
		UART0.Receive(byte((udd_ep_out_cache_buffer[ep][i] & 0xFF)))
	}

	// set byte count to zero
	usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

	// set multi packet size to 64
	usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(64 << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

	// set ready for next data
	setEPSTATUSCLR(ep, sam.USB_DEVICE_EPSTATUSCLR_BK0RDY)
}

func sendZlp(ep uint32) {
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)
}

func epPacketSize(size uint16) uint32 {
	switch size {
	case 8:
		return 0
	case 16:
		return 1
	case 32:
		return 2
	case 64:
		return 3
	case 128:
		return 4
	case 256:
		return 5
	case 512:
		return 6
	case 1023:
		return 7
	default:
		return 0
	}
}

func getEPCFG(ep uint32) uint8 {
	switch ep {
	case 0:
		return sam.USB_DEVICE.EPCFG0.Get()
	case 1:
		return sam.USB_DEVICE.EPCFG1.Get()
	case 2:
		return sam.USB_DEVICE.EPCFG2.Get()
	case 3:
		return sam.USB_DEVICE.EPCFG3.Get()
	case 4:
		return sam.USB_DEVICE.EPCFG4.Get()
	case 5:
		return sam.USB_DEVICE.EPCFG5.Get()
	case 6:
		return sam.USB_DEVICE.EPCFG6.Get()
	case 7:
		return sam.USB_DEVICE.EPCFG7.Get()
	default:
		return 0
	}
}

func setEPCFG(ep uint32, val uint8) {
	switch ep {
	case 0:
		sam.USB_DEVICE.EPCFG0.Set(val)
	case 1:
		sam.USB_DEVICE.EPCFG1.Set(val)
	case 2:
		sam.USB_DEVICE.EPCFG2.Set(val)
	case 3:
		sam.USB_DEVICE.EPCFG3.Set(val)
	case 4:
		sam.USB_DEVICE.EPCFG4.Set(val)
	case 5:
		sam.USB_DEVICE.EPCFG5.Set(val)
	case 6:
		sam.USB_DEVICE.EPCFG6.Set(val)
	case 7:
		sam.USB_DEVICE.EPCFG7.Set(val)
	default:
		return
	}
}

func setEPSTATUSCLR(ep uint32, val uint8) {
	switch ep {
	case 0:
		sam.USB_DEVICE.EPSTATUSCLR0.Set(val)
	case 1:
		sam.USB_DEVICE.EPSTATUSCLR1.Set(val)
	case 2:
		sam.USB_DEVICE.EPSTATUSCLR2.Set(val)
	case 3:
		sam.USB_DEVICE.EPSTATUSCLR3.Set(val)
	case 4:
		sam.USB_DEVICE.EPSTATUSCLR4.Set(val)
	case 5:
		sam.USB_DEVICE.EPSTATUSCLR5.Set(val)
	case 6:
		sam.USB_DEVICE.EPSTATUSCLR6.Set(val)
	case 7:
		sam.USB_DEVICE.EPSTATUSCLR7.Set(val)
	default:
		return
	}
}

func setEPSTATUSSET(ep uint32, val uint8) {
	switch ep {
	case 0:
		sam.USB_DEVICE.EPSTATUSSET0.Set(val)
	case 1:
		sam.USB_DEVICE.EPSTATUSSET1.Set(val)
	case 2:
		sam.USB_DEVICE.EPSTATUSSET2.Set(val)
	case 3:
		sam.USB_DEVICE.EPSTATUSSET3.Set(val)
	case 4:
		sam.USB_DEVICE.EPSTATUSSET4.Set(val)
	case 5:
		sam.USB_DEVICE.EPSTATUSSET5.Set(val)
	case 6:
		sam.USB_DEVICE.EPSTATUSSET6.Set(val)
	case 7:
		sam.USB_DEVICE.EPSTATUSSET7.Set(val)
	default:
		return
	}
}

func getEPSTATUS(ep uint32) uint8 {
	switch ep {
	case 0:
		return sam.USB_DEVICE.EPSTATUS0.Get()
	case 1:
		return sam.USB_DEVICE.EPSTATUS1.Get()
	case 2:
		return sam.USB_DEVICE.EPSTATUS2.Get()
	case 3:
		return sam.USB_DEVICE.EPSTATUS3.Get()
	case 4:
		return sam.USB_DEVICE.EPSTATUS4.Get()
	case 5:
		return sam.USB_DEVICE.EPSTATUS5.Get()
	case 6:
		return sam.USB_DEVICE.EPSTATUS6.Get()
	case 7:
		return sam.USB_DEVICE.EPSTATUS7.Get()
	default:
		return 0
	}
}

func getEPINTFLAG(ep uint32) uint8 {
	switch ep {
	case 0:
		return sam.USB_DEVICE.EPINTFLAG0.Get()
	case 1:
		return sam.USB_DEVICE.EPINTFLAG1.Get()
	case 2:
		return sam.USB_DEVICE.EPINTFLAG2.Get()
	case 3:
		return sam.USB_DEVICE.EPINTFLAG3.Get()
	case 4:
		return sam.USB_DEVICE.EPINTFLAG4.Get()
	case 5:
		return sam.USB_DEVICE.EPINTFLAG5.Get()
	case 6:
		return sam.USB_DEVICE.EPINTFLAG6.Get()
	case 7:
		return sam.USB_DEVICE.EPINTFLAG7.Get()
	default:
		return 0
	}
}

func setEPINTFLAG(ep uint32, val uint8) {
	switch ep {
	case 0:
		sam.USB_DEVICE.EPINTFLAG0.Set(val)
	case 1:
		sam.USB_DEVICE.EPINTFLAG1.Set(val)
	case 2:
		sam.USB_DEVICE.EPINTFLAG2.Set(val)
	case 3:
		sam.USB_DEVICE.EPINTFLAG3.Set(val)
	case 4:
		sam.USB_DEVICE.EPINTFLAG4.Set(val)
	case 5:
		sam.USB_DEVICE.EPINTFLAG5.Set(val)
	case 6:
		sam.USB_DEVICE.EPINTFLAG6.Set(val)
	case 7:
		sam.USB_DEVICE.EPINTFLAG7.Set(val)
	default:
		return
	}
}

func setEPINTENCLR(ep uint32, val uint8) {
	switch ep {
	case 0:
		sam.USB_DEVICE.EPINTENCLR0.Set(val)
	case 1:
		sam.USB_DEVICE.EPINTENCLR1.Set(val)
	case 2:
		sam.USB_DEVICE.EPINTENCLR2.Set(val)
	case 3:
		sam.USB_DEVICE.EPINTENCLR3.Set(val)
	case 4:
		sam.USB_DEVICE.EPINTENCLR4.Set(val)
	case 5:
		sam.USB_DEVICE.EPINTENCLR5.Set(val)
	case 6:
		sam.USB_DEVICE.EPINTENCLR6.Set(val)
	case 7:
		sam.USB_DEVICE.EPINTENCLR7.Set(val)
	default:
		return
	}
}

func setEPINTENSET(ep uint32, val uint8) {
	switch ep {
	case 0:
		sam.USB_DEVICE.EPINTENSET0.Set(val)
	case 1:
		sam.USB_DEVICE.EPINTENSET1.Set(val)
	case 2:
		sam.USB_DEVICE.EPINTENSET2.Set(val)
	case 3:
		sam.USB_DEVICE.EPINTENSET3.Set(val)
	case 4:
		sam.USB_DEVICE.EPINTENSET4.Set(val)
	case 5:
		sam.USB_DEVICE.EPINTENSET5.Set(val)
	case 6:
		sam.USB_DEVICE.EPINTENSET6.Set(val)
	case 7:
		sam.USB_DEVICE.EPINTENSET7.Set(val)
	default:
		return
	}
}
