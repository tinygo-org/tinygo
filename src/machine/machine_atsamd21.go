// +build sam,atsamd21

// Peripheral abstraction layer for the atsamd21.
//
// Datasheet:
// http://ww1.microchip.com/downloads/en/DeviceDoc/SAMD21-Family-DataSheet-DS40001882D.pdf
//
package machine

import (
	"device/arm"
	"device/sam"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

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

type PinChange uint8

// Pin change interrupt constants for SetInterrupt.
const (
	PinRising  PinChange = sam.EIC_CONFIG_SENSE0_RISE
	PinFalling PinChange = sam.EIC_CONFIG_SENSE0_FALL
	PinToggle  PinChange = sam.EIC_CONFIG_SENSE0_BOTH
)

// Callbacks to be called for pins configured with SetInterrupt. Unfortunately,
// we also need to keep track of which interrupt channel is used by which pin,
// as the only alternative would be iterating through all pins.
//
// We're using the magic constant 16 here because the SAM D21 has 16 interrupt
// channels configurable for pins.
var (
	interruptPins [16]Pin // warning: the value is invalid when pinCallbacks[i] is not set!
	pinCallbacks  [16]func(Pin)
)

const (
	pinPadMapSERCOM0Pad0 byte = (0x10 << 1) | 0x00
	pinPadMapSERCOM1Pad0 byte = (0x20 << 1) | 0x00
	pinPadMapSERCOM2Pad0 byte = (0x30 << 1) | 0x00
	pinPadMapSERCOM3Pad0 byte = (0x40 << 1) | 0x00
	pinPadMapSERCOM4Pad0 byte = (0x50 << 1) | 0x00
	pinPadMapSERCOM5Pad0 byte = (0x60 << 1) | 0x00
	pinPadMapSERCOM0Pad2 byte = (0x10 << 1) | 0x10
	pinPadMapSERCOM1Pad2 byte = (0x20 << 1) | 0x10
	pinPadMapSERCOM2Pad2 byte = (0x30 << 1) | 0x10
	pinPadMapSERCOM3Pad2 byte = (0x40 << 1) | 0x10
	pinPadMapSERCOM4Pad2 byte = (0x50 << 1) | 0x10
	pinPadMapSERCOM5Pad2 byte = (0x60 << 1) | 0x10

	pinPadMapSERCOM0AltPad0 byte = (0x01 << 1) | 0x00
	pinPadMapSERCOM1AltPad0 byte = (0x02 << 1) | 0x00
	pinPadMapSERCOM2AltPad0 byte = (0x03 << 1) | 0x00
	pinPadMapSERCOM3AltPad0 byte = (0x04 << 1) | 0x00
	pinPadMapSERCOM4AltPad0 byte = (0x05 << 1) | 0x00
	pinPadMapSERCOM5AltPad0 byte = (0x06 << 1) | 0x00
	pinPadMapSERCOM0AltPad2 byte = (0x01 << 1) | 0x01
	pinPadMapSERCOM1AltPad2 byte = (0x02 << 1) | 0x01
	pinPadMapSERCOM2AltPad2 byte = (0x03 << 1) | 0x01
	pinPadMapSERCOM3AltPad2 byte = (0x04 << 1) | 0x01
	pinPadMapSERCOM4AltPad2 byte = (0x05 << 1) | 0x01
	pinPadMapSERCOM5AltPad2 byte = (0x06 << 1) | 0x01
)

// pinPadMapping lists which pins have which SERCOMs attached to them.
// The encoding is rather dense, with each byte encoding two pins and both
// SERCOM and SERCOM-ALT.
//
// Observations:
//   * There are six SERCOMs. Those SERCOM numbers can be encoded in 3 bits.
//   * Even pad numbers are always on even pins, and odd pad numbers are always on
//     odd pins.
//   * Pin pads come in pairs. If PA00 has pad 0, then PA01 has pad 1.
// With this information, we can encode SERCOM pin/pad numbers much more
// efficiently. First of all, due to pads coming in pairs, we can ignore half
// the pins: the information for an odd pin can be calculated easily from the
// preceding even pin. And second, if odd pads are always on odd pins and even
// pads on even pins, we can drop a single bit from the pad number.
//
// Each byte below is split in two nibbles. The 4 high bits are for SERCOM and
// the 4 low bits are for SERCOM-ALT. Of each nibble, the 3 high bits encode the
// SERCOM + 1 while the low bit encodes whether this is PAD0 or PAD2 (0 means
// PAD0, 1 means PAD2). It encodes SERCOM + 1 instead of just the SERCOM number,
// to make it easy to check whether a nibble is set at all.
var pinPadMapping = [32]byte{
	// page 21
	PA00 / 2: 0 | pinPadMapSERCOM1AltPad0,
	PB08 / 2: 0 | pinPadMapSERCOM4AltPad0,
	PA04 / 2: 0 | pinPadMapSERCOM0AltPad0,
	PA06 / 2: 0 | pinPadMapSERCOM0AltPad2,
	PA08 / 2: pinPadMapSERCOM0Pad0 | pinPadMapSERCOM2AltPad0,
	PA10 / 2: pinPadMapSERCOM0Pad2 | pinPadMapSERCOM2AltPad2,

	// page 22
	PB10 / 2: 0 | pinPadMapSERCOM4AltPad2,
	PB12 / 2: pinPadMapSERCOM4Pad0 | 0,
	PB14 / 2: pinPadMapSERCOM4Pad2 | 0,
	PA12 / 2: pinPadMapSERCOM2Pad0 | pinPadMapSERCOM4AltPad0,
	PA14 / 2: pinPadMapSERCOM2Pad2 | pinPadMapSERCOM4AltPad2,
	PA16 / 2: pinPadMapSERCOM1Pad0 | pinPadMapSERCOM3AltPad0,
	PA18 / 2: pinPadMapSERCOM1Pad2 | pinPadMapSERCOM3AltPad2,
	PB16 / 2: pinPadMapSERCOM5Pad0 | 0,
	PA20 / 2: pinPadMapSERCOM5Pad2 | pinPadMapSERCOM3AltPad2,
	PA22 / 2: pinPadMapSERCOM3Pad0 | pinPadMapSERCOM5AltPad0,
	PA24 / 2: pinPadMapSERCOM3Pad2 | pinPadMapSERCOM5AltPad2,

	// page 23
	PB22 / 2: 0 | pinPadMapSERCOM5AltPad2,
	PA30 / 2: 0 | pinPadMapSERCOM1AltPad2,
	PB30 / 2: 0 | pinPadMapSERCOM5AltPad0,
	PB00 / 2: 0 | pinPadMapSERCOM5AltPad2,
	PB02 / 2: 0 | pinPadMapSERCOM5AltPad0,
}

// findPinPadMapping looks up the pad number and the pinmode for a given pin,
// given a SERCOM number. The result can either be SERCOM, SERCOM-ALT, or "not
// found" (indicated by returning ok=false). The pad number is returned to
// calculate the DOPO/DIPO bitfields of the various serial peripherals.
func findPinPadMapping(sercom uint8, pin Pin) (pinMode PinMode, pad uint32, ok bool) {
	if int(pin)/2 >= len(pinPadMapping) {
		// This is probably NoPin, for which no mapping is available.
		return
	}

	nibbles := pinPadMapping[pin/2]
	upper := nibbles >> 4
	lower := nibbles & 0xf

	if upper != 0 {
		// SERCOM
		if (upper>>1)-1 == sercom {
			pinMode = PinSERCOM
			pad |= uint32((upper & 1) << 1)
			ok = true
		}
	}
	if lower != 0 {
		// SERCOM-ALT
		if (lower>>1)-1 == sercom {
			pinMode = PinSERCOMAlt
			pad |= uint32((lower & 1) << 1)
			ok = true
		}
	}

	if ok {
		// The lower bit of the pad is the same as the lower bit of the pin number.
		pad |= uint32(pin & 1)
	}
	return
}

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
//
// This call will replace a previously set callback on this pin. You can pass a
// nil func to unset the pin change interrupt. If you do so, the change
// parameter is ignored and can be set to any value (such as 0).
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) error {
	// Most pins follow a common pattern where the EXTINT value is the pin
	// number modulo 16. However, there are a few exceptions, as you can see
	// below.
	extint := uint8(0)
	switch p {
	case PA08:
		// Connected to NMI. This is not currently supported.
		return ErrInvalidInputPin
	case PA24:
		extint = 12
	case PA25:
		extint = 13
	case PA27:
		extint = 15
	case PA28:
		extint = 8
	case PA30:
		extint = 10
	case PA31:
		extint = 11
	default:
		// All other pins follow a normal pattern.
		extint = uint8(p) % 16
	}

	if callback == nil {
		// Disable this pin interrupt (if it was enabled).
		sam.EIC.INTENCLR.Set(1 << extint)
		if pinCallbacks[extint] != nil {
			pinCallbacks[extint] = nil
		}
		return nil
	}

	if pinCallbacks[extint] != nil {
		// The pin was already configured.
		// To properly re-configure a pin, unset it first and set a new
		// configuration.
		return ErrNoPinChangeChannel
	}
	pinCallbacks[extint] = callback
	interruptPins[extint] = p

	if sam.EIC.CTRL.Get() == 0 {
		// EIC peripheral has not yet been initialized. Initialize it now.

		// The EIC needs two clocks: CLK_EIC_APB and GCLK_EIC. CLK_EIC_APB is
		// enabled by default, so doesn't have to be re-enabled. The other is
		// required for detecting edges and must be enabled manually.
		sam.GCLK.CLKCTRL.Set(sam.GCLK_CLKCTRL_ID_EIC<<sam.GCLK_CLKCTRL_ID_Pos |
			sam.GCLK_CLKCTRL_GEN_GCLK0<<sam.GCLK_CLKCTRL_GEN_Pos |
			sam.GCLK_CLKCTRL_CLKEN)

		// should not be necessary (CLKCTRL is not synchronized)
		for sam.GCLK.STATUS.HasBits(sam.GCLK_STATUS_SYNCBUSY) {
		}

		sam.EIC.CTRL.Set(sam.EIC_CTRL_ENABLE)
		for sam.EIC.STATUS.HasBits(sam.EIC_STATUS_SYNCBUSY) {
		}
	}

	// Configure this pin. Set the 4 bits of the EIC.CONFIGx register to the
	// sense value (filter bit set to 0, sense bits set to the change value).
	addr := &sam.EIC.CONFIG0
	if extint >= 8 {
		addr = &sam.EIC.CONFIG1
	}
	pos := (extint % 8) * 4 // bit position in register
	addr.ReplaceBits(uint32(change), 0xf, pos)

	// Enable external interrupt for this pin.
	sam.EIC.INTENSET.Set(1 << extint)

	// Set the PMUXEN flag, while keeping the INEN and PULLEN flags (if they
	// were set before). This avoids clearing the pin pull mode while
	// configuring the pin interrupt.
	p.setPinCfg(sam.PORT_PINCFG0_PMUXEN | (p.getPinCfg() & (sam.PORT_PINCFG0_INEN | sam.PORT_PINCFG0_PULLEN)))
	if p&1 > 0 {
		// odd pin, so save the even pins
		val := p.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
		p.setPMux(val | (sam.PORT_PMUX0_PMUXO_A << sam.PORT_PMUX0_PMUXO_Pos))
	} else {
		// even pin, so save the odd pins
		val := p.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
		p.setPMux(val | (sam.PORT_PMUX0_PMUXE_A << sam.PORT_PMUX0_PMUXE_Pos))
	}

	interrupt.New(sam.IRQ_EIC, func(interrupt.Interrupt) {
		flags := sam.EIC.INTFLAG.Get()
		sam.EIC.INTFLAG.Set(flags)      // clear interrupt
		for i := uint(0); i < 16; i++ { // there are 16 channels
			if flags&(1<<i) != 0 {
				pinCallbacks[i](interruptPins[i])
			}
		}
	}).Enable()

	return nil
}

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
}

// Configure configures a ADC pin to be able to be used to read data.
func (a ADC) Configure(config ADCConfig) {

	// Wait for synchronization
	waitADCSync()

	var resolution uint32
	switch config.Resolution {
	case 8:
		resolution = sam.ADC_CTRLB_RESSEL_8BIT
	case 10:
		resolution = sam.ADC_CTRLB_RESSEL_10BIT
	case 12:
		resolution = sam.ADC_CTRLB_RESSEL_12BIT
	case 16:
		resolution = sam.ADC_CTRLB_RESSEL_16BIT
	default:
		resolution = sam.ADC_CTRLB_RESSEL_12BIT
	}
	// Divide Clock by 32 with 12 bits resolution as default
	sam.ADC.CTRLB.Set((sam.ADC_CTRLB_PRESCALER_DIV32 << sam.ADC_CTRLB_PRESCALER_Pos) |
		uint16(resolution<<sam.ADC_CTRLB_RESSEL_Pos))

	// Sampling Time Length
	sam.ADC.SAMPCTRL.Set(5)

	// Wait for synchronization
	waitADCSync()

	// Use internal ground
	sam.ADC.INPUTCTRL.Set(sam.ADC_INPUTCTRL_MUXNEG_GND << sam.ADC_INPUTCTRL_MUXNEG_Pos)

	// Averaging (see datasheet table in AVGCTRL register description)
	var samples uint32
	switch config.Samples {
	case 1:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_1
	case 2:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_2
	case 4:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_4
	case 8:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_8
	case 16:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_16
	case 32:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_32
	case 64:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_64
	case 128:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_128
	case 256:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_256
	case 512:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_512
	case 1024:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_1024
	default:
		samples = sam.ADC_AVGCTRL_SAMPLENUM_1
	}
	sam.ADC.AVGCTRL.Set(uint8(samples<<sam.ADC_AVGCTRL_SAMPLENUM_Pos) |
		(0x0 << sam.ADC_AVGCTRL_ADJRES_Pos))

	// TODO: use config.Reference to set AREF level

	// Analog Reference is AREF pin (3.3v)
	sam.ADC.INPUTCTRL.SetBits(sam.ADC_INPUTCTRL_GAIN_DIV2 << sam.ADC_INPUTCTRL_GAIN_Pos)

	// 1/2 VDDANA = 0.5 * 3V3 = 1.65V
	sam.ADC.REFCTRL.SetBits(sam.ADC_REFCTRL_REFSEL_INTVCC1 << sam.ADC_REFCTRL_REFSEL_Pos)

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

	// wait for first conversion to finish to fix same issue as
	// https://github.com/arduino/ArduinoCore-samd/issues/446
	for !sam.ADC.INTFLAG.HasBits(sam.ADC_INTFLAG_RESRDY) {
	}

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
	case PA03:
		return 1
	case PB04:
		return 12
	case PB05:
		return 13
	case PB06:
		return 14
	case PB07:
		return 15
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
	case PA08:
		return 16
	case PA09:
		return 17
	case PA10:
		return 18
	case PA11:
		return 19
	case PB00:
		return 8
	case PB01:
		return 9
	case PB02:
		return 10
	case PB03:
		return 11
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
	Buffer    *RingBuffer
	Bus       *sam.SERCOM_USART_Type
	SERCOM    uint8
	Interrupt interrupt.Interrupt
}

var (
	// UART0 is actually a USB CDC interface.
	UART0 = USBCDC{Buffer: NewRingBuffer()}
)

const (
	sampleRate16X = 16
	lsbFirst      = 1
)

// Configure the UART.
func (uart UART) Configure(config UARTConfig) error {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// Use default pins if pins are not set.
	if config.TX == 0 && config.RX == 0 {
		// use default pins
		config.TX = UART_TX_PIN
		config.RX = UART_RX_PIN
	}

	// Determine transmit pinout.
	txPinMode, txPad, ok := findPinPadMapping(uart.SERCOM, config.TX)
	if !ok {
		return ErrInvalidOutputPin
	}
	var txPinOut uint32
	// See table 25-9 of the datasheet (page 459) for how pads are mapped to
	// pinout values.
	switch txPad {
	case 0:
		txPinOut = 0
	case 2:
		txPinOut = 1
	default:
		// TODO: flow control (RTS/CTS)
		return ErrInvalidOutputPin
	}

	// Determine receive pinout.
	rxPinMode, rxPad, ok := findPinPadMapping(uart.SERCOM, config.RX)
	if !ok {
		return ErrInvalidInputPin
	}
	// As you can see in table 25-8 on page 459 of the datasheet, input pins
	// are mapped directly.
	rxPinOut := rxPad

	// configure pins
	config.TX.Configure(PinConfig{Mode: txPinMode})
	config.RX.Configure(PinConfig{Mode: rxPinMode})

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
	uart.Bus.CTRLA.SetBits((txPinOut << sam.SERCOM_USART_CTRLA_TXPO_Pos) |
		(rxPinOut << sam.SERCOM_USART_CTRLA_RXPO_Pos))

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
	uart.Interrupt.Enable()

	return nil
}

// SetBaudRate sets the communication speed for the UART.
func (uart UART) SetBaudRate(br uint32) {
	// Asynchronous fractional mode (Table 24-2 in datasheet)
	//   BAUD = fref / (sampleRateValue * fbaud)
	// (multiply by 8, to calculate fractional piece)
	// uint32_t baudTimes8 = (SystemCoreClock * 8) / (16 * baudrate);
	baud := (CPUFrequency() * 8) / (sampleRate16X * br)

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

// handleInterrupt should be called from the appropriate interrupt handler for
// this UART instance.
func (uart *UART) handleInterrupt(interrupt.Interrupt) {
	// should reset IRQ
	uart.Receive(byte((uart.Bus.DATA.Get() & 0xFF)))
	uart.Bus.INTFLAG.SetBits(sam.SERCOM_USART_INTFLAG_RXC)
}

// I2C on the SAMD21.
type I2C struct {
	Bus    *sam.SERCOM_I2CM_Type
	SERCOM uint8
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
func (i2c I2C) Configure(config I2CConfig) error {
	// Default I2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = 100e3 // default to 100kHz
	}
	if config.SDA == 0 && config.SCL == 0 {
		config.SDA = SDA_PIN
		config.SCL = SCL_PIN
	}

	sclPinMode, sclPad, ok := findPinPadMapping(i2c.SERCOM, config.SCL)
	if !ok || sclPad != 1 {
		// SCL must be on pad 1, according to section 27.5 of the datasheet.
		// Note: this is not an exhaustive test for I2C support on the pin: not
		// all pins support I2C.
		return ErrInvalidClockPin
	}
	sdaPinMode, sdaPad, ok := findPinPadMapping(i2c.SERCOM, config.SDA)
	if !ok || sdaPad != 0 {
		// SDA must be on pad 0, according to section 27.5 of the datasheet.
		// Note: this is not an exhaustive test for I2C support on the pin: not
		// all pins support I2C.
		return ErrInvalidDataPin
	}

	// reset SERCOM
	i2c.Bus.CTRLA.SetBits(sam.SERCOM_I2CM_CTRLA_SWRST)
	for i2c.Bus.CTRLA.HasBits(sam.SERCOM_I2CM_CTRLA_SWRST) ||
		i2c.Bus.SYNCBUSY.HasBits(sam.SERCOM_I2CM_SYNCBUSY_SWRST) {
	}

	// Set i2c controller mode
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
	config.SDA.Configure(PinConfig{Mode: sdaPinMode})
	config.SCL.Configure(PinConfig{Mode: sclPinMode})

	return nil
}

// SetBaudRate sets the communication speed for the I2C.
func (i2c I2C) SetBaudRate(br uint32) {
	// Synchronous arithmetic baudrate, via Arduino SAMD implementation:
	// SystemCoreClock / ( 2 * baudrate) - 5 - (((SystemCoreClock / 1000000) * WIRE_RISE_TIME_NANOSECONDS) / (2 * 1000));
	baud := CPUFrequency()/(2*br) - 5 - (((CPUFrequency() / 1000000) * riseTimeNanoseconds) / (2 * 1000))
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
				return errI2CWriteTimeout
			}
		}

		// ACK received (0: ACK, 1: NACK)
		if i2c.Bus.STATUS.HasBits(sam.SERCOM_I2CM_STATUS_RXNACK) {
			return errI2CAckExpected
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
			// If the peripheral NACKS the address, the MB bit will be set.
			// In that case, send a stop condition and return error.
			if i2c.Bus.INTFLAG.HasBits(sam.SERCOM_I2CM_INTFLAG_MB) {
				i2c.Bus.CTRLB.SetBits(wireCmdStop << sam.SERCOM_I2CM_CTRLB_CMD_Pos) // Stop condition
				return errI2CAckExpected
			}
		}

		// ACK received (0: ACK, 1: NACK)
		if i2c.Bus.STATUS.HasBits(sam.SERCOM_I2CM_STATUS_RXNACK) {
			return errI2CAckExpected
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
			return errI2CBusError
		}
		timeout--
		if timeout == 0 {
			return errI2CWriteTimeout
		}
	}

	if i2c.Bus.STATUS.HasBits(sam.SERCOM_I2CM_STATUS_RXNACK) {
		return errI2CAckExpected
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
			return errI2CBusReadyTimeout
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
			return errI2CSignalStopTimeout
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
			return errI2CSignalReadTimeout
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
	division_factor := CPUFrequency() / (config.AudioFrequency * uint32(config.DataFormat))

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
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_DATASIZE_8 << sam.I2S_SERCTRL_DATASIZE_Pos)

	case I2SDataFormat16bit:
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_DATASIZE_16 << sam.I2S_SERCTRL_DATASIZE_Pos)

	case I2SDataFormat24bit:
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_DATASIZE_24 << sam.I2S_SERCTRL_DATASIZE_Pos)

	case I2SDataFormat32bit:
	case I2SDataFormatDefault:
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_DATASIZE_32 << sam.I2S_SERCTRL_DATASIZE_Pos)
	}

	// set serializer slot adjustment
	if config.Standard == I2SStandardLSB {
		// adjust right
		i2s.Bus.SERCTRL1.ClearBits(sam.I2S_SERCTRL_SLOTADJ)

		// transfer LSB first
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_BITREV)
	} else {
		// adjust left
		i2s.Bus.SERCTRL1.SetBits(sam.I2S_SERCTRL_SLOTADJ)
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
	Bus    *sam.SERCOM_SPI_Type
	SERCOM uint8
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

	// set default frequency
	if config.Frequency == 0 {
		config.Frequency = 4000000
	}

	// Determine the input pinout (for SDI).
	SDIPinMode, SDIPad, ok := findPinPadMapping(spi.SERCOM, config.SDI)
	if !ok {
		return ErrInvalidInputPin
	}
	dataInPinout := SDIPad // mapped directly

	// Determine the output pinout (for SDO/SCK).
	// See table 26-7 on page 494 of the datasheet.
	var dataOutPinout uint32
	sckPinMode, sckPad, ok := findPinPadMapping(spi.SERCOM, config.SCK)
	if !ok {
		return ErrInvalidOutputPin
	}
	SDOPinMode, SDOPad, ok := findPinPadMapping(spi.SERCOM, config.SDO)
	if !ok {
		return ErrInvalidOutputPin
	}
	switch sckPad {
	case 1:
		switch SDOPad {
		case 0:
			dataOutPinout = 0x0
		case 3:
			dataOutPinout = 0x2
		default:
			return ErrInvalidOutputPin
		}
	case 3:
		switch SDOPad {
		case 2:
			dataOutPinout = 0x1
		case 0:
			dataOutPinout = 0x3
		default:
			return ErrInvalidOutputPin
		}
	default:
		return ErrInvalidOutputPin
	}

	// Disable SPI port.
	spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPI_CTRLA_ENABLE)
	for spi.Bus.SYNCBUSY.HasBits(sam.SERCOM_SPI_SYNCBUSY_ENABLE) {
	}

	// enable pins
	config.SCK.Configure(PinConfig{Mode: sckPinMode})
	config.SDO.Configure(PinConfig{Mode: SDOPinMode})
	config.SDI.Configure(PinConfig{Mode: SDIPinMode})

	// reset SERCOM
	spi.Bus.CTRLA.SetBits(sam.SERCOM_SPI_CTRLA_SWRST)
	for spi.Bus.CTRLA.HasBits(sam.SERCOM_SPI_CTRLA_SWRST) ||
		spi.Bus.SYNCBUSY.HasBits(sam.SERCOM_SPI_SYNCBUSY_SWRST) {
	}

	// set bit transfer order
	dataOrder := uint32(0)
	if config.LSBFirst {
		dataOrder = 1
	}

	// Set SPI mode to controller
	spi.Bus.CTRLA.Set((sam.SERCOM_SPI_CTRLA_MODE_SPI_MASTER << sam.SERCOM_SPI_CTRLA_MODE_Pos) |
		(dataOutPinout << sam.SERCOM_SPI_CTRLA_DOPO_Pos) |
		(dataInPinout << sam.SERCOM_SPI_CTRLA_DIPO_Pos) |
		(dataOrder << sam.SERCOM_SPI_CTRLA_DORD_Pos))

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
	baudRate := CPUFrequency() / (2 * config.Frequency)
	if baudRate > 0 {
		baudRate--
	}
	spi.Bus.BAUD.Set(uint8(baudRate))

	// Enable SPI port.
	spi.Bus.CTRLA.SetBits(sam.SERCOM_SPI_CTRLA_ENABLE)
	for spi.Bus.SYNCBUSY.HasBits(sam.SERCOM_SPI_SYNCBUSY_ENABLE) {
	}

	return nil
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
func (pwm PWM) Configure() error {
	// figure out which TCCX timer for this pin
	timer := pwm.getTimer()
	if timer == nil {
		return ErrInvalidOutputPin
	}

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

	return nil
}

// Set turns on the duty cycle for a PWM pin using the provided value.
func (pwm PWM) Set(value uint16) {
	// figure out which TCCX timer for this pin
	timer := pwm.getTimer()
	if timer == nil {
		// The Configure call above cannot have succeeded, so simply ignore this
		// error.
		return
	}

	// disable output
	timer.CTRLA.ClearBits(sam.TCC_CTRLA_ENABLE)

	// Wait for synchronization
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_ENABLE) {
	}

	// Set PWM signal to output duty cycle
	pwm.setChannel(timer, uint32(value))

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
func (pwm PWM) setChannel(timer *sam.TCC_Type, val uint32) {
	switch pwm.Pin {
	case 6:
		timer.CC0.Set(val)
	case 7:
		timer.CC1.Set(val)
	case 8:
		timer.CC0.Set(val)
	case 9:
		timer.CC1.Set(val)
	case 14:
		timer.CC0.Set(val)
	case 15:
		timer.CC1.Set(val)
	case 16:
		timer.CC2.Set(val)
	case 17:
		timer.CC3.Set(val)
	case 18:
		timer.CC2.Set(val)
	case 19:
		timer.CC3.Set(val)
	case 20:
		timer.CC2.Set(val)
	case 21:
		timer.CC3.Set(val)
	default:
		return // not supported on this pin
	}
}

// USBCDC is the USB CDC aka serial over USB interface on the SAMD21.
type USBCDC struct {
	Buffer            *RingBuffer
	TxIdx             volatile.Register8
	waitTxc           bool
	waitTxcRetryCount uint8
	sent              bool
}

const (
	usbcdcTxSizeMask          uint8 = 0x3F
	usbcdcTxBankMask          uint8 = ^usbcdcTxSizeMask
	usbcdcTxBank1st           uint8 = 0x00
	usbcdcTxBank2nd           uint8 = usbcdcTxSizeMask + 1
	usbcdcTxMaxRetriesAllowed uint8 = 5
)

// Flush flushes buffered data.
func (usbcdc *USBCDC) Flush() error {
	if usbLineInfo.lineState > 0 {
		idx := usbcdc.TxIdx.Get()
		sz := idx & usbcdcTxSizeMask
		bk := idx & usbcdcTxBankMask
		if 0 < sz {

			if usbcdc.waitTxc {
				// waiting for the next flush(), because the transmission is not complete
				return nil
			}
			usbcdc.waitTxc = true
			usbcdc.waitTxcRetryCount = 0

			// set the data
			usbEndpointDescriptors[usb_CDC_ENDPOINT_IN].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[usb_CDC_ENDPOINT_IN][bk]))))
			if bk == usbcdcTxBank1st {
				usbcdc.TxIdx.Set(usbcdcTxBank2nd)
			} else {
				usbcdc.TxIdx.Set(usbcdcTxBank1st)
			}

			// clean multi packet size of bytes already sent
			usbEndpointDescriptors[usb_CDC_ENDPOINT_IN].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Mask << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

			// set count of bytes to be sent
			usbEndpointDescriptors[usb_CDC_ENDPOINT_IN].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)
			usbEndpointDescriptors[usb_CDC_ENDPOINT_IN].DeviceDescBank[1].PCKSIZE.SetBits((uint32(sz) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask) << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

			// clear transfer complete flag
			setEPINTFLAG(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_EPINTFLAG_TRCPT1)

			// send data by setting bank ready
			setEPSTATUSSET(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_EPSTATUSSET_BK1RDY)
			UART0.sent = true
		}
	}
	return nil
}

// WriteByte writes a byte of data to the USB CDC interface.
func (usbcdc *USBCDC) WriteByte(c byte) error {
	// Supposedly to handle problem with Windows USB serial ports?
	if usbLineInfo.lineState > 0 {
		ok := false
		for {
			mask := interrupt.Disable()

			idx := UART0.TxIdx.Get()
			if (idx & usbcdcTxSizeMask) < usbcdcTxSizeMask {
				udd_ep_in_cache_buffer[usb_CDC_ENDPOINT_IN][idx] = c
				UART0.TxIdx.Set(idx + 1)
				ok = true
			}

			interrupt.Restore(mask)

			if ok {
				break
			} else if usbcdcTxMaxRetriesAllowed < UART0.waitTxcRetryCount {
				mask := interrupt.Disable()
				UART0.waitTxc = false
				UART0.waitTxcRetryCount = 0
				usbcdc.TxIdx.Set(0)
				usbLineInfo.lineState = 0
				interrupt.Restore(mask)
				break
			} else {
				mask := interrupt.Disable()
				if UART0.sent {
					if UART0.waitTxc {
						if (getEPINTFLAG(usb_CDC_ENDPOINT_IN) & sam.USB_DEVICE_EPINTFLAG_TRCPT1) != 0 {
							setEPSTATUSCLR(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_EPSTATUSCLR_BK1RDY)
							setEPINTFLAG(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_EPINTFLAG_TRCPT1)
							UART0.waitTxc = false
							UART0.Flush()
						}
					} else {
						UART0.Flush()
					}
				}
				interrupt.Restore(mask)
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
	intr := interrupt.New(sam.IRQ_USB, handleUSB)
	intr.Enable()
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

func handleUSB(intr interrupt.Interrupt) {
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
		if (epFlags&sam.USB_DEVICE_EPINTFLAG_TRCPT0) > 0 ||
			(epFlags&sam.USB_DEVICE_EPINTFLAG_TRCPT1) > 0 {
			switch i {
			case usb_CDC_ENDPOINT_OUT:
				handleEndpoint(i)
				setEPINTFLAG(i, epFlags)
			case usb_CDC_ENDPOINT_IN, usb_CDC_ENDPOINT_ACM:
				setEPSTATUSCLR(i, sam.USB_DEVICE_EPSTATUSCLR_BK1RDY)
				setEPINTFLAG(i, sam.USB_DEVICE_EPINTFLAG_TRCPT1)

				if i == usb_CDC_ENDPOINT_IN {
					UART0.waitTxc = false
				}
			}
		}

		if i == usb_CDC_ENDPOINT_IN && UART0.waitTxc {
			UART0.waitTxcRetryCount++
		}
	}

	UART0.Flush()
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
		sendZlp()
		return true

	case usb_SET_FEATURE:
		if setup.wValueL == 1 { // DEVICEREMOTEWAKEUP
			isRemoteWakeUpEnabled = true
		} else if setup.wValueL == 0 { // ENDPOINTHALT
			isEndpointHalt = true
		}
		sendZlp()
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

			sendZlp()
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

		sendZlp()
		return true

	default:
		return true
	}
}

func cdcSetup(setup usbSetup) bool {
	if setup.bmRequestType == usb_REQUEST_DEVICETOHOST_CLASS_INTERFACE {
		if setup.bRequest == usb_CDC_GET_LINE_CODING {
			b := make([]byte, 7)
			b[0] = byte(usbLineInfo.dwDTERate)
			b[1] = byte(usbLineInfo.dwDTERate >> 8)
			b[2] = byte(usbLineInfo.dwDTERate >> 16)
			b[3] = byte(usbLineInfo.dwDTERate >> 24)
			b[4] = byte(usbLineInfo.bCharFormat)
			b[5] = byte(usbLineInfo.bParityType)
			b[6] = byte(usbLineInfo.bDataBits)

			sendUSBPacket(0, b)
			return true
		}
	}

	if setup.bmRequestType == usb_REQUEST_HOSTTODEVICE_CLASS_INTERFACE {
		if setup.bRequest == usb_CDC_SET_LINE_CODING {
			b := receiveUSBControlPacket()
			usbLineInfo.dwDTERate = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
			usbLineInfo.bCharFormat = b[4]
			usbLineInfo.bParityType = b[5]
			usbLineInfo.bDataBits = b[6]
		}

		if setup.bRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			usbLineInfo.lineState = setup.wValueL
		}

		if setup.bRequest == usb_CDC_SET_LINE_CODING || setup.bRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			// auto-reset into the bootloader
			if usbLineInfo.dwDTERate == 1200 && usbLineInfo.lineState&usb_CDC_LINESTATE_DTR == 0 {
				ResetProcessor()
			} else {
				// TODO: cancel any reset
			}
			sendZlp()
		}

		if setup.bRequest == usb_CDC_SEND_BREAK {
			// TODO: something with this value?
			// breakValue = ((uint16_t)setup.wValueH << 8) | setup.wValueL;
			// return false;
			sendZlp()
		}
		return true
	}
	return false
}

//go:noinline
func sendUSBPacket(ep uint32, data []byte) {
	copy(udd_ep_in_cache_buffer[ep][:], data)

	// Set endpoint address for sending data
	usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

	// clear multi-packet size which is total bytes already sent
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Mask << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

	// set byte count, which is total number of bytes to be sent
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)
	usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(uint32((len(data) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask) << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos))
}

func receiveUSBControlPacket() []byte {
	// address
	usbEndpointDescriptors[0].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[0]))))

	// set byte count to zero
	usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

	// set ready for next data
	setEPSTATUSCLR(0, sam.USB_DEVICE_EPSTATUSCLR_BK0RDY)

	// Wait until OUT transfer is ready.
	timeout := 300000
	for (getEPSTATUS(0) & sam.USB_DEVICE_EPSTATUS_BK0RDY) == 0 {
		timeout--
		if timeout == 0 {
			return []byte{}
		}
	}

	// Wait until OUT transfer is completed.
	timeout = 300000
	for (getEPINTFLAG(0) & sam.USB_DEVICE_EPINTFLAG_TRCPT0) == 0 {
		timeout--
		if timeout == 0 {
			return []byte{}
		}
	}

	// get data
	bytesread := uint32((usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.Get() >>
		usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask)

	data := make([]byte, bytesread)
	copy(data, udd_ep_out_cache_buffer[0][:])

	return data
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

func sendZlp() {
	usbEndpointDescriptors[0].DeviceDescBank[1].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)
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

// ResetProcessor should perform a system reset in preperation
// to switch to the bootloader to flash new firmware.
func ResetProcessor() {
	arm.DisableInterrupts()

	// Perform magic reset into bootloader, as mentioned in
	// https://github.com/arduino/ArduinoCore-samd/issues/197
	*(*uint32)(unsafe.Pointer(uintptr(0x20007FFC))) = RESET_MAGIC_VALUE

	arm.SystemReset()
}

// DAC on the SAMD21.
type DAC struct {
}

var (
	DAC0 = DAC{}
)

// DACConfig placeholder for future expansion.
type DACConfig struct {
}

// Configure the DAC.
// output pin must already be configured.
func (dac DAC) Configure(config DACConfig) {
	// Turn on clock for DAC
	sam.PM.APBCMASK.SetBits(sam.PM_APBCMASK_DAC_)

	// Use Generic Clock Generator 0 as source for DAC.
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_DAC << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	waitForSync()

	// reset DAC
	sam.DAC.CTRLA.Set(sam.DAC_CTRLA_SWRST)
	syncDAC()

	// wait for reset complete
	for sam.DAC.CTRLA.HasBits(sam.DAC_CTRLA_SWRST) {
	}

	// enable
	sam.DAC.CTRLB.Set(sam.DAC_CTRLB_EOEN | sam.DAC_CTRLB_REFSEL_AVCC)
	sam.DAC.CTRLA.Set(sam.DAC_CTRLA_ENABLE)
}

// Set writes a single 16-bit value to the DAC.
// Since the ATSAMD21 only has a 10-bit DAC, the passed-in value will be scaled down.
func (dac DAC) Set(value uint16) error {
	sam.DAC.DATA.Set(value >> 6)
	syncDAC()
	return nil
}

func syncDAC() {
	for sam.DAC.STATUS.HasBits(sam.DAC_STATUS_SYNCBUSY) {
	}
}
