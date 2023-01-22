//go:build sam && atsamd21

// Peripheral abstraction layer for the atsamd21.
//
// Datasheet:
// http://ww1.microchip.com/downloads/en/DeviceDoc/SAMD21-Family-DataSheet-DS40001882D.pdf
package machine

import (
	"bytes"
	"device/arm"
	"device/sam"
	"encoding/binary"
	"errors"
	"runtime/interrupt"
	"unsafe"
)

const deviceName = sam.Device

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
	PinTCC           PinMode = PinTimer
	PinTCCAlt        PinMode = PinTimerAlt
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
//   - There are six SERCOMs. Those SERCOM numbers can be encoded in 3 bits.
//   - Even pad numbers are always on even pins, and odd pad numbers are always on
//     odd pins.
//   - Pin pads come in pairs. If PA00 has pad 0, then PA01 has pad 1.
//
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

	// scales to 16-bit result
	switch (sam.ADC.CTRLB.Get() & sam.ADC_CTRLB_RESSEL_Msk) >> sam.ADC_CTRLB_RESSEL_Pos {
	case sam.ADC_CTRLB_RESSEL_8BIT:
		val = val << 8
	case sam.ADC_CTRLB_RESSEL_10BIT:
		val = val << 6
	case sam.ADC_CTRLB_RESSEL_16BIT:
		val = val << 4
	case sam.ADC_CTRLB_RESSEL_12BIT:
		val = val << 4
	}
	return val
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

const (
	sampleRate16X = 16
	lsbFirst      = 1
)

// Configure the UART.
func (uart *UART) Configure(config UARTConfig) error {
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
func (uart *UART) SetBaudRate(br uint32) {
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
func (uart *UART) writeByte(c byte) error {
	// wait until ready to receive
	for !uart.Bus.INTFLAG.HasBits(sam.SERCOM_USART_INTFLAG_DRE) {
	}
	uart.Bus.DATA.Set(uint16(c))
	return nil
}

func (uart *UART) flush() {}

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
func (i2c *I2C) Configure(config I2CConfig) error {
	// Default I2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = 100 * KHz
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

// SetBaudRate sets the communication speed for I2C.
func (i2c *I2C) SetBaudRate(br uint32) error {
	// Synchronous arithmetic baudrate, via Arduino SAMD implementation:
	// SystemCoreClock / ( 2 * baudrate) - 5 - (((SystemCoreClock / 1000000) * WIRE_RISE_TIME_NANOSECONDS) / (2 * 1000));
	baud := CPUFrequency()/(2*br) - 5 - (((CPUFrequency() / 1000000) * riseTimeNanoseconds) / (2 * 1000))
	i2c.Bus.BAUD.Set(baud)
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
func (i2c *I2C) WriteByte(data byte) error {
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
func (i2c *I2C) sendAddress(address uint16, write bool) error {
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

func (i2c *I2C) signalStop() error {
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

func (i2c *I2C) signalRead() error {
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

func (i2c *I2C) readByte() byte {
	for !i2c.Bus.INTFLAG.HasBits(sam.SERCOM_I2CM_INTFLAG_SB) {
	}
	return byte(i2c.Bus.DATA.Get())
}

// I2S on the SAMD21.

// I2S
type I2S struct {
	Bus *sam.I2S_Type
}

var I2S0 = I2S{Bus: sam.I2S}

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
	switch {
	case w == nil:
		// read only, so write zero and read a result.
		spi.rx(r)
	case r == nil:
		// write only
		spi.tx(w)

	default:
		// write/read
		if len(w) != len(r) {
			return ErrTxInvalidSliceSize
		}

		spi.txrx(w, r)
	}

	return nil
}

func (spi SPI) tx(tx []byte) {
	for i := 0; i < len(tx); i++ {
		for !spi.Bus.INTFLAG.HasBits(sam.SERCOM_SPI_INTFLAG_DRE) {
		}
		spi.Bus.DATA.Set(uint32(tx[i]))
	}
	for !spi.Bus.INTFLAG.HasBits(sam.SERCOM_SPI_INTFLAG_TXC) {
	}

	// read to clear RXC register
	for spi.Bus.INTFLAG.HasBits(sam.SERCOM_SPI_INTFLAG_RXC) {
		spi.Bus.DATA.Get()
	}
}

func (spi SPI) rx(rx []byte) {
	spi.Bus.DATA.Set(0)
	for !spi.Bus.INTFLAG.HasBits(sam.SERCOM_SPI_INTFLAG_DRE) {
	}

	for i := 1; i < len(rx); i++ {
		spi.Bus.DATA.Set(0)
		for !spi.Bus.INTFLAG.HasBits(sam.SERCOM_SPI_INTFLAG_RXC) {
		}
		rx[i-1] = byte(spi.Bus.DATA.Get())
	}
	for !spi.Bus.INTFLAG.HasBits(sam.SERCOM_SPI_INTFLAG_RXC) {
	}
	rx[len(rx)-1] = byte(spi.Bus.DATA.Get())
}

func (spi SPI) txrx(tx, rx []byte) {
	spi.Bus.DATA.Set(uint32(tx[0]))
	for !spi.Bus.INTFLAG.HasBits(sam.SERCOM_SPI_INTFLAG_DRE) {
	}

	for i := 1; i < len(rx); i++ {
		spi.Bus.DATA.Set(uint32(tx[i]))
		for !spi.Bus.INTFLAG.HasBits(sam.SERCOM_SPI_INTFLAG_RXC) {
		}
		rx[i-1] = byte(spi.Bus.DATA.Get())
	}
	for !spi.Bus.INTFLAG.HasBits(sam.SERCOM_SPI_INTFLAG_RXC) {
	}
	rx[len(rx)-1] = byte(spi.Bus.DATA.Get())
}

// TCC is one timer/counter peripheral, which consists of a counter and multiple
// output channels (that can be connected to actual pins). You can set the
// frequency using SetPeriod, but only for all the channels in this TCC
// peripheral at once.
type TCC sam.TCC_Type

// The SAM D21 has three TCC peripherals, which have PWM as one feature.
var (
	TCC0 = (*TCC)(sam.TCC0)
	TCC1 = (*TCC)(sam.TCC1)
	TCC2 = (*TCC)(sam.TCC2)
)

//go:inline
func (tcc *TCC) timer() *sam.TCC_Type {
	return (*sam.TCC_Type)(tcc)
}

// Configure enables and configures this TCC.
func (tcc *TCC) Configure(config PWMConfig) error {
	// Enable the clock source for this timer.
	switch tcc.timer() {
	case sam.TCC0:
		sam.PM.APBCMASK.SetBits(sam.PM_APBCMASK_TCC0_)
		// Use GCLK0 for TCC0/TCC1
		sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_TCC0_TCC1 << sam.GCLK_CLKCTRL_ID_Pos) |
			(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
			sam.GCLK_CLKCTRL_CLKEN)
		for sam.GCLK.STATUS.HasBits(sam.GCLK_STATUS_SYNCBUSY) {
		}
	case sam.TCC1:
		sam.PM.APBCMASK.SetBits(sam.PM_APBCMASK_TCC1_)
		// Use GCLK0 for TCC0/TCC1
		sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_TCC0_TCC1 << sam.GCLK_CLKCTRL_ID_Pos) |
			(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
			sam.GCLK_CLKCTRL_CLKEN)
		for sam.GCLK.STATUS.HasBits(sam.GCLK_STATUS_SYNCBUSY) {
		}
	case sam.TCC2:
		sam.PM.APBCMASK.SetBits(sam.PM_APBCMASK_TCC2_)
		// Use GCLK0 for TCC2/TC3
		sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_TCC2_TC3 << sam.GCLK_CLKCTRL_ID_Pos) |
			(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
			sam.GCLK_CLKCTRL_CLKEN)
		for sam.GCLK.STATUS.HasBits(sam.GCLK_STATUS_SYNCBUSY) {
		}
	}

	// Disable timer (if it was enabled). This is necessary because
	// tcc.setPeriod may want to change the prescaler bits in CTRLA, which is
	// only allowed when the TCC is disabled.
	tcc.timer().CTRLA.ClearBits(sam.TCC_CTRLA_ENABLE)

	// Use "Normal PWM" (single-slope PWM)
	tcc.timer().WAVE.Set(sam.TCC_WAVE_WAVEGEN_NPWM)

	// Wait for synchronization of all changed registers.
	for tcc.timer().SYNCBUSY.Get() != 0 {
	}

	// Set the period and prescaler.
	err := tcc.setPeriod(config.Period, true)

	// Enable the timer.
	tcc.timer().CTRLA.SetBits(sam.TCC_CTRLA_ENABLE)

	// Wait for synchronization of all changed registers.
	for tcc.timer().SYNCBUSY.Get() != 0 {
	}

	// Return any error that might have occured in the tcc.setPeriod call.
	return err
}

// SetPeriod updates the period of this TCC peripheral.
// To set a particular frequency, use the following formula:
//
//	period = 1e9 / frequency
//
// If you use a period of 0, a period that works well for LEDs will be picked.
//
// SetPeriod will not change the prescaler, but also won't change the current
// value in any of the channels. This means that you may need to update the
// value for the particular channel.
//
// Note that you cannot pick any arbitrary period after the TCC peripheral has
// been configured. If you want to switch between frequencies, pick the lowest
// frequency (longest period) once when calling Configure and adjust the
// frequency here as needed.
func (tcc *TCC) SetPeriod(period uint64) error {
	err := tcc.setPeriod(period, false)
	if err == nil {
		if tcc.Counter() >= tcc.Top() {
			// When setting the timer to a shorter period, there is a chance
			// that it passes the counter value and thus goes all the way to MAX
			// before wrapping back to zero.
			// To avoid this, reset the counter back to 0.
			tcc.timer().COUNT.Set(0)
		}
	}
	return err
}

// setPeriod sets the period of this TCC, possibly updating the prescaler as
// well. The prescaler can only modified when the TCC is disabled, that is, in
// the Configure function.
func (tcc *TCC) setPeriod(period uint64, updatePrescaler bool) error {
	var top uint64
	if period == 0 {
		// Make sure the TOP value is at 0xffff (enough for a 16-bit timer).
		top = 0xffff
	} else {
		// The formula below calculates the following formula, optimized:
		//     period * (48e6 / 1e9)
		// This assumes that the chip is running at the (default) 48MHz speed.
		top = period * 6 / 125
	}

	maxTop := uint64(0xffffff)
	if tcc.timer() == sam.TCC2 {
		// TCC2 is a 16-bit timer, not a 24-bit timer.
		maxTop = 0xffff
	}

	if updatePrescaler {
		// This function was called during Configure(), with the timer disabled.
		// Note that updating the prescaler can only happen while the peripheral
		// is disabled.
		var prescaler uint32
		switch {
		case top <= maxTop:
			prescaler = sam.TCC_CTRLA_PRESCALER_DIV1
		case top/2 <= maxTop:
			prescaler = sam.TCC_CTRLA_PRESCALER_DIV2
			top = top / 2
		case top/4 <= maxTop:
			prescaler = sam.TCC_CTRLA_PRESCALER_DIV4
			top = top / 4
		case top/8 <= maxTop:
			prescaler = sam.TCC_CTRLA_PRESCALER_DIV8
			top = top / 8
		case top/16 <= maxTop:
			prescaler = sam.TCC_CTRLA_PRESCALER_DIV16
			top = top / 16
		case top/64 <= maxTop:
			prescaler = sam.TCC_CTRLA_PRESCALER_DIV64
			top = top / 64
		case top/256 <= maxTop:
			prescaler = sam.TCC_CTRLA_PRESCALER_DIV256
			top = top / 256
		case top/1024 <= maxTop:
			prescaler = sam.TCC_CTRLA_PRESCALER_DIV1024
			top = top / 1024
		default:
			return ErrPWMPeriodTooLong
		}
		tcc.timer().CTRLA.Set((tcc.timer().CTRLA.Get() &^ sam.TCC_CTRLA_PRESCALER_Msk) | (prescaler << sam.TCC_CTRLA_PRESCALER_Pos))
	} else {
		// Do not update the prescaler, but use the already-configured
		// prescaler. This is the normal SetPeriod case, where the prescaler
		// must not be changed.
		prescaler := (tcc.timer().CTRLA.Get() & sam.TCC_CTRLA_PRESCALER_Msk) >> sam.TCC_CTRLA_PRESCALER_Pos
		switch prescaler {
		case sam.TCC_CTRLA_PRESCALER_DIV1:
			top /= 1 // no-op
		case sam.TCC_CTRLA_PRESCALER_DIV2:
			top /= 2
		case sam.TCC_CTRLA_PRESCALER_DIV4:
			top /= 4
		case sam.TCC_CTRLA_PRESCALER_DIV8:
			top /= 8
		case sam.TCC_CTRLA_PRESCALER_DIV16:
			top /= 16
		case sam.TCC_CTRLA_PRESCALER_DIV64:
			top /= 64
		case sam.TCC_CTRLA_PRESCALER_DIV256:
			top /= 256
		case sam.TCC_CTRLA_PRESCALER_DIV1024:
			top /= 1024
		default:
			// unreachable
		}
		if top > maxTop {
			return ErrPWMPeriodTooLong
		}
	}

	// Set the period (the counter top).
	tcc.timer().PER.Set(uint32(top) - 1)

	// Wait for synchronization of CTRLA.PRESCALER and PER registers.
	for tcc.timer().SYNCBUSY.Get() != 0 {
	}

	return nil
}

// Top returns the current counter top, for use in duty cycle calculation. It
// will only change with a call to Configure or SetPeriod, otherwise it is
// constant.
//
// The value returned here is hardware dependent. In general, it's best to treat
// it as an opaque value that can be divided by some number and passed to Set
// (see Set documentation for more information).
func (tcc *TCC) Top() uint32 {
	return tcc.timer().PER.Get() + 1
}

// Counter returns the current counter value of the timer in this TCC
// peripheral. It may be useful for debugging.
func (tcc *TCC) Counter() uint32 {
	tcc.timer().CTRLBSET.Set(sam.TCC_CTRLBSET_CMD_READSYNC << sam.TCC_CTRLBSET_CMD_Pos)
	for tcc.timer().SYNCBUSY.Get() != 0 {
	}
	return tcc.timer().COUNT.Get()
}

// Some constans to make pinTimerMapping below easier to read.
const (
	pinTCC0     = 1
	pinTCC1     = 2
	pinTCC2     = 3
	pinTimerCh0 = 0 << 3
	pinTimerCh2 = 1 << 3
	pinTCC0Ch0  = pinTCC0 | pinTimerCh0
	pinTCC0Ch2  = pinTCC0 | pinTimerCh2
	pinTCC1Ch0  = pinTCC1 | pinTimerCh0
	pinTCC1Ch2  = pinTCC1 | pinTimerCh2
	pinTCC2Ch0  = pinTCC2 | pinTimerCh0
)

// Mapping from pin number to TCC peripheral and channel using a special
// encoding. Note that only TCC0-TCC2 are included, not TC3 and up.
// Every byte is split in two nibbles where the low nibble describes PinTCC and
// the high nibble describes PinTCCAlt. Within a nibble, there is one bit that
// indicates Ch0/Ch1 or Ch2/Ch3, and three other bits that contain the TCC
// peripheral number plus one (to distinguish between TCC0Ch0 and 0).
//
// The encoding can be so compact because all pins are configured in pairs, so
// if you know PA00 you can infer the configuration of PA01. And only channel 0
// or 2 need to be included (taking up just one bit), because channel 0 and 2
// are only ever used on odd pins and channel 1 and 3 on even pins, again using
// the pin pair pattern to reduce the amount of information needed to be stored.
//
// Datasheet: https://cdn.sparkfun.com/datasheets/Dev/Arduino/Boards/Atmel-42181-SAM-D21_Datasheet.pdf
var pinTimerMapping = [...]uint8{
	// page 21
	PA00 / 2: pinTCC2Ch0 | 0,
	PA04 / 2: pinTCC0Ch0 | 0,
	PA06 / 2: pinTCC1Ch0 | 0,
	PA08 / 2: pinTCC0Ch0 | pinTCC1Ch2<<4,
	PA10 / 2: pinTCC1Ch0 | pinTCC0Ch2<<4,
	// page 22
	PB10 / 2: 0 | pinTCC0Ch0<<4,
	PB12 / 2: 0 | pinTCC0Ch2<<4,
	PA12 / 2: pinTCC2Ch0 | pinTCC0Ch2<<4,
	PA14 / 2: 0 | pinTCC0Ch0<<4,
	PA16 / 2: pinTCC2Ch0 | pinTCC0Ch2<<4,
	PA18 / 2: 0 | pinTCC0Ch2<<4,
	PB16 / 2: 0 | pinTCC0Ch0<<4,
	PA20 / 2: 0 | pinTCC0Ch2<<4,
	PA22 / 2: 0 | pinTCC0Ch0<<4,
	PA24 / 2: 0 | pinTCC1Ch2<<4,
	// page 23
	PA30 / 2: 0 | pinTCC1Ch0<<4,
	PB30 / 2: pinTCC0Ch0 | pinTCC1Ch2<<4,
}

// findPinPadMapping returns the pin mode (PinTCC or PinTCCAlt) and the channel
// number for a given timer and pin. A zero PinMode is returned if no mapping
// could be found.
func findPinTimerMapping(timer uint8, pin Pin) (PinMode, uint8) {
	mapping := pinTimerMapping[pin/2]
	// evenChannel below indicates the channel 0 or 2, for the even part of the
	// pin pair. The next pin will also have the next channel (1 or 3).
	if mapping&0x07 == timer+1 {
		// PWM output is on peripheral function E.
		evenChannel := ((mapping >> 3) & 1) * 2
		return PinTCC, evenChannel + uint8(pin&1)
	}
	if (mapping&0x70)>>4 == timer+1 {
		// PWM output is on peripheral function F.
		evenChannel := ((mapping >> 7) & 1) * 2
		return PinTCCAlt, evenChannel + uint8(pin&1)
	}
	return 0, 0
}

// Channel returns a PWM channel for the given pin. Note that one channel may be
// shared between multiple pins, and so will have the same duty cycle. If this
// is not desirable, look for a different TCC peripheral or consider using a
// different pin.
func (tcc *TCC) Channel(pin Pin) (uint8, error) {
	var pinMode PinMode
	var channel uint8
	switch tcc.timer() {
	case sam.TCC0:
		pinMode, channel = findPinTimerMapping(0, pin)
	case sam.TCC1:
		pinMode, channel = findPinTimerMapping(1, pin)
	case sam.TCC2:
		pinMode, channel = findPinTimerMapping(2, pin)
	}

	if pinMode == 0 {
		// No pin could be found.
		return 0, ErrInvalidOutputPin
	}

	// Enable the port multiplexer for pin
	pin.setPinCfg(sam.PORT_PINCFG0_PMUXEN)

	if pin&1 > 0 {
		// odd pin, so save the even pins
		val := pin.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
		pin.setPMux(val | uint8(pinMode<<sam.PORT_PMUX0_PMUXO_Pos))
	} else {
		// even pin, so save the odd pins
		val := pin.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
		pin.setPMux(val | uint8(pinMode<<sam.PORT_PMUX0_PMUXE_Pos))
	}
	return channel, nil
}

// SetInverting sets whether to invert the output of this channel.
// Without inverting, a 25% duty cycle would mean the output is high for 25% of
// the time and low for the rest. Inverting flips the output as if a NOT gate
// was placed at the output, meaning that the output would be 25% low and 75%
// high with a duty cycle of 25%.
func (tcc *TCC) SetInverting(channel uint8, inverting bool) {
	if inverting {
		tcc.timer().WAVE.SetBits(1 << (sam.TCC_WAVE_POL0_Pos + channel))
	} else {
		tcc.timer().WAVE.ClearBits(1 << (sam.TCC_WAVE_POL0_Pos + channel))
	}

	// Wait for synchronization of the WAVE register.
	for tcc.timer().SYNCBUSY.Get() != 0 {
	}
}

// Set updates the channel value. This is used to control the channel duty
// cycle, in other words the fraction of time the channel output is high (or low
// when inverted). For example, to set it to a 25% duty cycle, use:
//
//	tcc.Set(channel, tcc.Top() / 4)
//
// tcc.Set(channel, 0) will set the output to low and tcc.Set(channel,
// tcc.Top()) will set the output to high, assuming the output isn't inverted.
func (tcc *TCC) Set(channel uint8, value uint32) {
	// Set PWM signal to output duty cycle
	switch channel {
	case 0:
		tcc.timer().CC0.Set(value)
	case 1:
		tcc.timer().CC1.Set(value)
	case 2:
		tcc.timer().CC2.Set(value)
	case 3:
		tcc.timer().CC3.Set(value)
	default:
		// invalid PWM channel, ignore.
	}

	// Wait for synchronization on all channels (or anything in this peripheral,
	// really).
	for tcc.timer().SYNCBUSY.Get() != 0 {
	}
}

// EnterBootloader should perform a system reset in preperation
// to switch to the bootloader to flash new firmware.
func EnterBootloader() {
	arm.DisableInterrupts()

	// Perform magic reset into bootloader, as mentioned in
	// https://github.com/arduino/ArduinoCore-samd/issues/197
	*(*uint32)(unsafe.Pointer(uintptr(0x20007FFC))) = resetMagicValue

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

// Flash related code
const memoryStart = 0x0

// compile-time check for ensuring we fulfill BlockDevice interface
var _ BlockDevice = flashBlockDevice{}

var Flash flashBlockDevice

type flashBlockDevice struct {
	initComplete bool
}

// ReadAt reads the given number of bytes from the block device.
func (f flashBlockDevice) ReadAt(p []byte, off int64) (n int, err error) {
	if FlashDataStart()+uintptr(off)+uintptr(len(p)) > FlashDataEnd() {
		return 0, errFlashCannotReadPastEOF
	}

	f.ensureInitComplete()

	waitWhileFlashBusy()

	data := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(FlashDataStart()), uintptr(off))), len(p))
	copy(p, data)

	return len(p), nil
}

// WriteAt writes the given number of bytes to the block device.
// Only word (32 bits) length data can be programmed.
// See Atmel-42181G–SAM-D21_Datasheet–09/2015 page 359.
// If the length of p is not long enough it will be padded with 0xFF bytes.
// This method assumes that the destination is already erased.
func (f flashBlockDevice) WriteAt(p []byte, off int64) (n int, err error) {
	if FlashDataStart()+uintptr(off)+uintptr(len(p)) > FlashDataEnd() {
		return 0, errFlashCannotWritePastEOF
	}

	f.ensureInitComplete()

	address := FlashDataStart() + uintptr(off)
	padded := f.pad(p)

	waitWhileFlashBusy()

	for j := 0; j < len(padded); j += int(f.WriteBlockSize()) {
		// write word
		*(*uint32)(unsafe.Pointer(address)) = binary.LittleEndian.Uint32(padded[j : j+int(f.WriteBlockSize())])

		sam.NVMCTRL.SetADDR(uint32(address >> 1))
		sam.NVMCTRL.CTRLA.Set(sam.NVMCTRL_CTRLA_CMD_WP | (sam.NVMCTRL_CTRLA_CMDEX_KEY << sam.NVMCTRL_CTRLA_CMDEX_Pos))

		waitWhileFlashBusy()

		if err := checkFlashError(); err != nil {
			return j, err
		}

		address += uintptr(f.WriteBlockSize())
	}

	return len(padded), nil
}

// Size returns the number of bytes in this block device.
func (f flashBlockDevice) Size() int64 {
	return int64(FlashDataEnd() - FlashDataStart())
}

const writeBlockSize = 4

// WriteBlockSize returns the block size in which data can be written to
// memory. It can be used by a client to optimize writes, non-aligned writes
// should always work correctly.
func (f flashBlockDevice) WriteBlockSize() int64 {
	return writeBlockSize
}

const eraseBlockSizeValue = 256

func eraseBlockSize() int64 {
	return eraseBlockSizeValue
}

// EraseBlockSize returns the smallest erasable area on this particular chip
// in bytes. This is used for the block size in EraseBlocks.
func (f flashBlockDevice) EraseBlockSize() int64 {
	return eraseBlockSize()
}

// EraseBlocks erases the given number of blocks. An implementation may
// transparently coalesce ranges of blocks into larger bundles if the chip
// supports this. The start and len parameters are in block numbers, use
// EraseBlockSize to map addresses to blocks.
func (f flashBlockDevice) EraseBlocks(start, len int64) error {
	f.ensureInitComplete()

	address := FlashDataStart() + uintptr(start*f.EraseBlockSize())
	waitWhileFlashBusy()

	for i := start; i < start+len; i++ {
		sam.NVMCTRL.SetADDR(uint32(address >> 1))
		sam.NVMCTRL.CTRLA.Set(sam.NVMCTRL_CTRLA_CMD_ER | (sam.NVMCTRL_CTRLA_CMDEX_KEY << sam.NVMCTRL_CTRLA_CMDEX_Pos))

		waitWhileFlashBusy()

		if err := checkFlashError(); err != nil {
			return err
		}

		address += uintptr(f.EraseBlockSize())
	}

	return nil
}

// pad data if needed so it is long enough for correct byte alignment on writes.
func (f flashBlockDevice) pad(p []byte) []byte {
	overflow := int64(len(p)) % f.WriteBlockSize()
	if overflow == 0 {
		return p
	}

	padding := bytes.Repeat([]byte{0xff}, int(f.WriteBlockSize()-overflow))
	return append(p, padding...)
}

func (f flashBlockDevice) ensureInitComplete() {
	if f.initComplete {
		return
	}

	sam.NVMCTRL.SetCTRLB_READMODE(sam.NVMCTRL_CTRLB_READMODE_NO_MISS_PENALTY)
	sam.NVMCTRL.SetCTRLB_SLEEPPRM(sam.NVMCTRL_CTRLB_SLEEPPRM_WAKEONACCESS)

	waitWhileFlashBusy()

	f.initComplete = true
}

func waitWhileFlashBusy() {
	for sam.NVMCTRL.GetINTFLAG_READY() != sam.NVMCTRL_INTFLAG_READY {
	}
}

var (
	errFlashPROGE = errors.New("errFlashPROGE")
	errFlashLOCKE = errors.New("errFlashLOCKE")
	errFlashNVME  = errors.New("errFlashNVME")
)

func checkFlashError() error {
	switch {
	case sam.NVMCTRL.GetSTATUS_PROGE() != 0:
		return errFlashPROGE
	case sam.NVMCTRL.GetSTATUS_LOCKE() != 0:
		return errFlashLOCKE
	case sam.NVMCTRL.GetSTATUS_NVME() != 0:
		return errFlashNVME
	}

	return nil
}
