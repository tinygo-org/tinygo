// +build sam,atsamd51

// Peripheral abstraction layer for the atsamd51.
//
// Datasheet:
// http://ww1.microchip.com/downloads/en/DeviceDoc/60001507C.pdf
//
package machine

import (
	"device/arm"
	"device/sam"
	"errors"
	"unsafe"
)

const CPU_FREQUENCY = 120000000

type PinMode uint8

const (
	PinAnalog        PinMode = 1
	PinSERCOM        PinMode = 2
	PinSERCOMAlt     PinMode = 3
	PinTimer         PinMode = 4
	PinTimerAlt      PinMode = 5
	PinTCCPDEC       PinMode = 6
	PinCom           PinMode = 7
	PinSDHC          PinMode = 8
	PinI2S           PinMode = 9
	PinPCC           PinMode = 10
	PinGMAC          PinMode = 11
	PinACCLK         PinMode = 12
	PinCCL           PinMode = 13
	PinDigital       PinMode = 14
	PinInput         PinMode = 15
	PinInputPullup   PinMode = 16
	PinOutput        PinMode = 17
	PinPWME          PinMode = PinTimer
	PinPWMF          PinMode = PinTimerAlt
	PinPWMG          PinMode = PinTCCPDEC
	PinInputPulldown PinMode = 18
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

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskSet() (*uint32, uint32) {
	group, pin_in_group := p.getPinGrouping()
	return &sam.PORT.GROUP[group].OUTSET.Reg, 1 << pin_in_group
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskClear() (*uint32, uint32) {
	group, pin_in_group := p.getPinGrouping()
	return &sam.PORT.GROUP[group].OUTCLR.Reg, 1 << pin_in_group
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	group, pin_in_group := p.getPinGrouping()
	if high {
		sam.PORT.GROUP[group].OUTSET.Set(1 << pin_in_group)
	} else {
		sam.PORT.GROUP[group].OUTCLR.Set(1 << pin_in_group)
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	group, pin_in_group := p.getPinGrouping()
	return (sam.PORT.GROUP[group].IN.Get()>>pin_in_group)&1 > 0
}

// Toggle switches an output pin from low to high or from high to low.
// Warning: only use this on an output pin!
func (p Pin) Toggle() {
	group, pin_in_group := p.getPinGrouping()
	sam.PORT.GROUP[group].OUTTGL.Set(1 << pin_in_group)
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	group, pin_in_group := p.getPinGrouping()
	switch config.Mode {
	case PinOutput:
		sam.PORT.GROUP[group].DIRSET.Set(1 << pin_in_group)
		// output is also set to input enable so pin can read back its own value
		p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN)

	case PinInput:
		sam.PORT.GROUP[group].DIRCLR.Set(1 << pin_in_group)
		p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN)

	case PinInputPulldown:
		sam.PORT.GROUP[group].DIRCLR.Set(1 << pin_in_group)
		sam.PORT.GROUP[group].OUTCLR.Set(1 << pin_in_group)
		p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN | sam.PORT_GROUP_PINCFG_PULLEN)

	case PinInputPullup:
		sam.PORT.GROUP[group].DIRCLR.Set(1 << pin_in_group)
		sam.PORT.GROUP[group].OUTSET.Set(1 << pin_in_group)
		p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN | sam.PORT_GROUP_PINCFG_PULLEN)

	case PinSERCOM:
		if p&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXE_Msk
			p.setPMux(val | (uint8(PinSERCOM) << sam.PORT_GROUP_PMUX_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXO_Msk
			p.setPMux(val | (uint8(PinSERCOM) << sam.PORT_GROUP_PMUX_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_GROUP_PINCFG_PMUXEN | sam.PORT_GROUP_PINCFG_DRVSTR | sam.PORT_GROUP_PINCFG_INEN)

	case PinSERCOMAlt:
		if p&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXE_Msk
			p.setPMux(val | (uint8(PinSERCOMAlt) << sam.PORT_GROUP_PMUX_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXO_Msk
			p.setPMux(val | (uint8(PinSERCOMAlt) << sam.PORT_GROUP_PMUX_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_GROUP_PINCFG_PMUXEN | sam.PORT_GROUP_PINCFG_DRVSTR)

	case PinCom:
		if p&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXE_Msk
			p.setPMux(val | (uint8(PinCom) << sam.PORT_GROUP_PMUX_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXO_Msk
			p.setPMux(val | (uint8(PinCom) << sam.PORT_GROUP_PMUX_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_GROUP_PINCFG_PMUXEN)
	case PinAnalog:
		if p&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXE_Msk
			p.setPMux(val | (uint8(PinAnalog) << sam.PORT_GROUP_PMUX_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXO_Msk
			p.setPMux(val | (uint8(PinAnalog) << sam.PORT_GROUP_PMUX_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_GROUP_PINCFG_PMUXEN | sam.PORT_GROUP_PINCFG_DRVSTR)
	}
}

// getPMux returns the value for the correct PMUX register for this pin.
func (p Pin) getPMux() uint8 {
	group, pin_in_group := p.getPinGrouping()
	return sam.PORT.GROUP[group].PMUX[pin_in_group>>1].Get()
}

// setPMux sets the value for the correct PMUX register for this pin.
func (p Pin) setPMux(val uint8) {
	group, pin_in_group := p.getPinGrouping()
	sam.PORT.GROUP[group].PMUX[pin_in_group>>1].Set(val)
}

// getPinCfg returns the value for the correct PINCFG register for this pin.
func (p Pin) getPinCfg() uint8 {
	group, pin_in_group := p.getPinGrouping()
	return sam.PORT.GROUP[group].PINCFG[pin_in_group].Get()
}

// setPinCfg sets the value for the correct PINCFG register for this pin.
func (p Pin) setPinCfg(val uint8) {
	group, pin_in_group := p.getPinGrouping()
	sam.PORT.GROUP[group].PINCFG[pin_in_group].Set(val)
}

// getPinGrouping calculates the gpio group and pin id from the pin number.
// Pins are split into groups of 32, and each group has its own set of
// control registers.
func (p Pin) getPinGrouping() (uint8, uint8) {
	group := uint8(p) >> 5
	pin_in_group := uint8(p) & 0x1f
	return group, pin_in_group
}

// InitADC initializes the ADC.
func InitADC() {
	// ADC Bias Calibration
	// NVMCTRL_SW0 0x00800080
	// #define ADC0_FUSES_BIASCOMP_ADDR    NVMCTRL_SW0
	// #define ADC0_FUSES_BIASCOMP_Pos     2            /**< \brief (NVMCTRL_SW0) ADC Comparator Scaling */
	// #define ADC0_FUSES_BIASCOMP_Msk     (_Ul(0x7) << ADC0_FUSES_BIASCOMP_Pos)
	// #define ADC0_FUSES_BIASCOMP(value)  (ADC0_FUSES_BIASCOMP_Msk & ((value) << ADC0_FUSES_BIASCOMP_Pos))

	// #define ADC0_FUSES_BIASR2R_ADDR     NVMCTRL_SW0
	// #define ADC0_FUSES_BIASR2R_Pos      8            /**< \brief (NVMCTRL_SW0) ADC Bias R2R ampli scaling */
	// #define ADC0_FUSES_BIASR2R_Msk      (_Ul(0x7) << ADC0_FUSES_BIASR2R_Pos)
	// #define ADC0_FUSES_BIASR2R(value)   (ADC0_FUSES_BIASR2R_Msk & ((value) << ADC0_FUSES_BIASR2R_Pos))

	// #define ADC0_FUSES_BIASREFBUF_ADDR  NVMCTRL_SW0
	// #define ADC0_FUSES_BIASREFBUF_Pos   5            /**< \brief (NVMCTRL_SW0) ADC Bias Reference Buffer Scaling */
	// #define ADC0_FUSES_BIASREFBUF_Msk   (_Ul(0x7) << ADC0_FUSES_BIASREFBUF_Pos)
	// #define ADC0_FUSES_BIASREFBUF(value) (ADC0_FUSES_BIASREFBUF_Msk & ((value) << ADC0_FUSES_BIASREFBUF_Pos))

	// #define ADC1_FUSES_BIASCOMP_ADDR    NVMCTRL_SW0
	// #define ADC1_FUSES_BIASCOMP_Pos     16           /**< \brief (NVMCTRL_SW0) ADC Comparator Scaling */
	// #define ADC1_FUSES_BIASCOMP_Msk     (_Ul(0x7) << ADC1_FUSES_BIASCOMP_Pos)
	// #define ADC1_FUSES_BIASCOMP(value)  (ADC1_FUSES_BIASCOMP_Msk & ((value) << ADC1_FUSES_BIASCOMP_Pos))

	// #define ADC1_FUSES_BIASR2R_ADDR     NVMCTRL_SW0
	// #define ADC1_FUSES_BIASR2R_Pos      22           /**< \brief (NVMCTRL_SW0) ADC Bias R2R ampli scaling */
	// #define ADC1_FUSES_BIASR2R_Msk      (_Ul(0x7) << ADC1_FUSES_BIASR2R_Pos)
	// #define ADC1_FUSES_BIASR2R(value)   (ADC1_FUSES_BIASR2R_Msk & ((value) << ADC1_FUSES_BIASR2R_Pos))

	// #define ADC1_FUSES_BIASREFBUF_ADDR  NVMCTRL_SW0
	// #define ADC1_FUSES_BIASREFBUF_Pos   19           /**< \brief (NVMCTRL_SW0) ADC Bias Reference Buffer Scaling */
	// #define ADC1_FUSES_BIASREFBUF_Msk   (_Ul(0x7) << ADC1_FUSES_BIASREFBUF_Pos)
	// #define ADC1_FUSES_BIASREFBUF(value) (ADC1_FUSES_BIASREFBUF_Msk & ((value) << ADC1_FUSES_BIASREFBUF_Pos))

	adcFuse := *(*uint32)(unsafe.Pointer(uintptr(0x00800080)))

	// uint32_t biascomp = (*((uint32_t *)ADC0_FUSES_BIASCOMP_ADDR) & ADC0_FUSES_BIASCOMP_Msk) >> ADC0_FUSES_BIASCOMP_Pos;
	biascomp := (adcFuse & uint32(0x7<<2)) //>> 2

	// uint32_t biasr2r = (*((uint32_t *)ADC0_FUSES_BIASR2R_ADDR) & ADC0_FUSES_BIASR2R_Msk) >> ADC0_FUSES_BIASR2R_Pos;
	biasr2r := (adcFuse & uint32(0x7<<8)) //>> 8

	// uint32_t biasref = (*((uint32_t *)ADC0_FUSES_BIASREFBUF_ADDR) & ADC0_FUSES_BIASREFBUF_Msk) >> ADC0_FUSES_BIASREFBUF_Pos;
	biasref := (adcFuse & uint32(0x7<<5)) //>> 5

	// calibrate ADC0
	sam.ADC0.CALIB.Set(uint16(biascomp | biasr2r | biasref))

	// biascomp = (*((uint32_t *)ADC1_FUSES_BIASCOMP_ADDR) & ADC1_FUSES_BIASCOMP_Msk) >> ADC1_FUSES_BIASCOMP_Pos;
	biascomp = (adcFuse & uint32(0x7<<16)) //>> 16

	// biasr2r = (*((uint32_t *)ADC1_FUSES_BIASR2R_ADDR) & ADC1_FUSES_BIASR2R_Msk) >> ADC1_FUSES_BIASR2R_Pos;
	biasr2r = (adcFuse & uint32(0x7<<22)) //>> 22

	// biasref = (*((uint32_t *)ADC1_FUSES_BIASREFBUF_ADDR) & ADC1_FUSES_BIASREFBUF_Msk) >> ADC1_FUSES_BIASREFBUF_Pos;
	biasref = (adcFuse & uint32(0x7<<19)) //>> 19

	// calibrate ADC1
	sam.ADC1.CALIB.Set(uint16((biascomp | biasr2r | biasref) >> 16))

	sam.ADC0.CTRLA.SetBits(sam.ADC_CTRLA_PRESCALER_DIV32 << sam.ADC_CTRLA_PRESCALER_Pos)
	// adcs[i]->CTRLB.bit.RESSEL = ADC_CTRLB_RESSEL_10BIT_Val;
	sam.ADC0.CTRLB.SetBits(sam.ADC_CTRLB_RESSEL_10BIT << sam.ADC_CTRLB_RESSEL_Pos)

	// wait for sync
	for sam.ADC0.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_CTRLB) {
	}

	// sampling Time Length
	sam.ADC0.SAMPCTRL.Set(5)

	// wait for sync
	for sam.ADC0.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_SAMPCTRL) {
	}

	// No Negative input (Internal Ground)
	sam.ADC0.INPUTCTRL.Set(sam.ADC_INPUTCTRL_MUXNEG_GND << sam.ADC_INPUTCTRL_MUXNEG_Pos)

	// wait for sync
	for sam.ADC0.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_INPUTCTRL) {
	}

	// Averaging (see datasheet table in AVGCTRL register description)
	// 1 sample only (no oversampling nor averaging), adjusting result by 0
	sam.ADC0.AVGCTRL.Set(sam.ADC_AVGCTRL_SAMPLENUM_1 | (0 << sam.ADC_AVGCTRL_ADJRES_Pos))

	// wait for sync
	for sam.ADC0.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_AVGCTRL) {
	}

	// same for ADC1, as for ADC0
	sam.ADC1.CTRLA.SetBits(sam.ADC_CTRLA_PRESCALER_DIV32 << sam.ADC_CTRLA_PRESCALER_Pos)
	sam.ADC1.CTRLB.SetBits(sam.ADC_CTRLB_RESSEL_10BIT << sam.ADC_CTRLB_RESSEL_Pos)
	for sam.ADC1.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_CTRLB) {
	}
	sam.ADC1.SAMPCTRL.Set(5)
	for sam.ADC1.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_SAMPCTRL) {
	}
	sam.ADC1.INPUTCTRL.Set(sam.ADC_INPUTCTRL_MUXNEG_GND << sam.ADC_INPUTCTRL_MUXNEG_Pos)
	for sam.ADC1.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_INPUTCTRL) {
	}
	sam.ADC1.AVGCTRL.Set(sam.ADC_AVGCTRL_SAMPLENUM_1 | (0 << sam.ADC_AVGCTRL_ADJRES_Pos))
	for sam.ADC1.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_AVGCTRL) {
	}

	// wait for sync
	for sam.ADC0.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_REFCTRL) {
	}
	for sam.ADC1.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_REFCTRL) {
	}

	// default is 3V3 reference voltage
	sam.ADC0.REFCTRL.SetBits(sam.ADC_REFCTRL_REFSEL_INTVCC1)
	sam.ADC1.REFCTRL.SetBits(sam.ADC_REFCTRL_REFSEL_INTVCC1)
}

// Configure configures a ADCPin to be able to be used to read data.
func (a ADC) Configure() {
	a.Pin.Configure(PinConfig{Mode: PinAnalog})
}

// Get returns the current value of a ADC pin, in the range 0..0xffff.
func (a ADC) Get() uint16 {
	bus := a.getADCBus()
	ch := a.getADCChannel()

	for bus.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_INPUTCTRL) {
	}

	// Selection for the positive ADC input channel
	bus.INPUTCTRL.SetBits((uint16(ch) & sam.ADC_INPUTCTRL_MUXPOS_Msk) << sam.ADC_INPUTCTRL_MUXPOS_Pos)
	for bus.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_ENABLE) {
	}

	// Enable ADC
	bus.CTRLA.SetBits(sam.ADC_CTRLA_ENABLE)
	for bus.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_ENABLE) {
	}

	// Start conversion
	bus.SWTRIG.SetBits(sam.ADC_SWTRIG_START)
	for !bus.INTFLAG.HasBits(sam.ADC_INTFLAG_RESRDY) {
	}

	// Clear the Data Ready flag
	bus.INTFLAG.SetBits(sam.ADC_INTFLAG_RESRDY)
	for bus.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_ENABLE) {
	}

	// Start conversion again, since first conversion after reference voltage changed is invalid.
	bus.SWTRIG.SetBits(sam.ADC_SWTRIG_START)

	// Waiting for conversion to complete
	for !bus.INTFLAG.HasBits(sam.ADC_INTFLAG_RESRDY) {
	}
	val := bus.RESULT.Get()

	// Disable ADC
	for bus.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_ENABLE) {
	}
	bus.CTRLA.ClearBits(sam.ADC_CTRLA_ENABLE)
	for bus.SYNCBUSY.HasBits(sam.ADC_SYNCBUSY_ENABLE) {
	}

	return uint16(val) << 4 // scales from 12 to 16-bit result
}

func (a ADC) getADCBus() *sam.ADC_Type {
	return sam.ADC0
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

// UART on the SAMD51.
type UART struct {
	Buffer *RingBuffer
	Bus    *sam.SERCOM_USART_INT_Type
	Mode   PinMode
}

var (
	// UART0 is actually a USB CDC interface.
	UART0 = USBCDC{Buffer: NewRingBuffer()}

	// The first hardware serial port on the SAMD51. Uses the SERCOM3 interface.
	UART1 = UART{Bus: sam.SERCOM3_USART_INT,
		Buffer: NewRingBuffer(),
		Mode:   PinSERCOMAlt,
	}
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
	case PA04:
		txpad = sercomTXPad0
	case PA10:
		txpad = sercomTXPad2
	case PA18:
		txpad = sercomTXPad2
	case PA16:
		txpad = sercomTXPad0
	default:
		panic("Invalid TX pin for UART")
	}

	switch config.RX {
	case PA07:
		rxpad = sercomRXPad3
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
	default:
		panic("Invalid RX pin for UART")
	}

	// configure pins
	config.TX.Configure(PinConfig{Mode: uart.Mode})
	config.RX.Configure(PinConfig{Mode: uart.Mode})

	// reset SERCOM0
	uart.Bus.CTRLA.SetBits(sam.SERCOM_USART_INT_CTRLA_SWRST)
	for uart.Bus.CTRLA.HasBits(sam.SERCOM_USART_INT_CTRLA_SWRST) ||
		uart.Bus.SYNCBUSY.HasBits(sam.SERCOM_USART_INT_SYNCBUSY_SWRST) {
	}

	// set UART mode/sample rate
	// SERCOM_USART_CTRLA_MODE(mode) |
	// SERCOM_USART_CTRLA_SAMPR(sampleRate);
	// sam.SERCOM_USART_CTRLA_MODE_USART_INT_CLK = 1?
	uart.Bus.CTRLA.Set((1 << sam.SERCOM_USART_INT_CTRLA_MODE_Pos) |
		(1 << sam.SERCOM_USART_INT_CTRLA_SAMPR_Pos)) // sample rate of 16x

	// Set baud rate
	uart.SetBaudRate(config.BaudRate)

	// setup UART frame
	// SERCOM_USART_CTRLA_FORM( (parityMode == SERCOM_NO_PARITY ? 0 : 1) ) |
	// dataOrder << SERCOM_USART_CTRLA_DORD_Pos;
	uart.Bus.CTRLA.SetBits((0 << sam.SERCOM_USART_INT_CTRLA_FORM_Pos) | // no parity
		(lsbFirst << sam.SERCOM_USART_INT_CTRLA_DORD_Pos)) // data order

	// set UART stop bits/parity
	// SERCOM_USART_CTRLB_CHSIZE(charSize) |
	// 	nbStopBits << SERCOM_USART_CTRLB_SBMODE_Pos |
	// 	(parityMode == SERCOM_NO_PARITY ? 0 : parityMode) << SERCOM_USART_CTRLB_PMODE_Pos; //If no parity use default value
	uart.Bus.CTRLB.SetBits((0 << sam.SERCOM_USART_INT_CTRLB_CHSIZE_Pos) | // 8 bits is 0
		(0 << sam.SERCOM_USART_INT_CTRLB_SBMODE_Pos) | // 1 stop bit is zero
		(0 << sam.SERCOM_USART_INT_CTRLB_PMODE_Pos)) // no parity

	// set UART pads. This is not same as pins...
	//  SERCOM_USART_CTRLA_TXPO(txPad) |
	//   SERCOM_USART_CTRLA_RXPO(rxPad);
	uart.Bus.CTRLA.SetBits(uint32((txpad << sam.SERCOM_USART_INT_CTRLA_TXPO_Pos) |
		(rxpad << sam.SERCOM_USART_INT_CTRLA_RXPO_Pos)))

	// Enable Transceiver and Receiver
	//sercom->USART.CTRLB.reg |= SERCOM_USART_CTRLB_TXEN | SERCOM_USART_CTRLB_RXEN ;
	uart.Bus.CTRLB.SetBits(sam.SERCOM_USART_INT_CTRLB_TXEN | sam.SERCOM_USART_INT_CTRLB_RXEN)

	// Enable USART1 port.
	// sercom->USART.CTRLA.bit.ENABLE = 0x1u;
	uart.Bus.CTRLA.SetBits(sam.SERCOM_USART_INT_CTRLA_ENABLE)
	for uart.Bus.SYNCBUSY.HasBits(sam.SERCOM_USART_INT_SYNCBUSY_ENABLE) {
	}

	// setup interrupt on receive
	uart.Bus.INTENSET.Set(sam.SERCOM_USART_INT_INTENSET_RXC)

	// Enable RX IRQ. Currently assumes SERCOM3.
	arm.EnableIRQ(sam.IRQ_SERCOM3_0)
	arm.EnableIRQ(sam.IRQ_SERCOM3_1)
	arm.EnableIRQ(sam.IRQ_SERCOM3_2)
	arm.EnableIRQ(sam.IRQ_SERCOM3_OTHER)
}

// SetBaudRate sets the communication speed for the UART.
func (uart UART) SetBaudRate(br uint32) {
	// Asynchronous fractional mode (Table 24-2 in datasheet)
	//   BAUD = fref / (sampleRateValue * fbaud)
	// (multiply by 8, to calculate fractional piece)
	// uint32_t baudTimes8 = (SystemCoreClock * 8) / (16 * baudrate);
	baud := (SERCOM_FREQ_REF * 8) / (sampleRate16X * br)

	// sercom->USART.BAUD.FRAC.FP   = (baudTimes8 % 8);
	// sercom->USART.BAUD.FRAC.BAUD = (baudTimes8 / 8);
	uart.Bus.BAUD.Set(uint16(((baud % 8) << sam.SERCOM_USART_INT_BAUD_FRAC_MODE_FP_Pos) |
		((baud / 8) << sam.SERCOM_USART_INT_BAUD_FRAC_MODE_BAUD_Pos)))
}

// WriteByte writes a byte of data to the UART.
func (uart UART) WriteByte(c byte) error {
	// wait until ready to receive
	for !uart.Bus.INTFLAG.HasBits(sam.SERCOM_USART_INT_INTFLAG_DRE) {
	}
	uart.Bus.DATA.Set(uint32(c))
	return nil
}

//go:export SERCOM3_0_IRQHandler
func handleSERCOM3_0() {
	handleUART1()
}

//go:export SERCOM3_1_IRQHandler
func handleSERCOM3_1() {
	handleUART1()
}

//go:export SERCOM3_2_IRQHandler
func handleSERCOM3_2() {
	handleUART1()
}

//go:export SERCOM3_OTHER_IRQHandler
func handleSERCOM3_OTHER() {
	handleUART1()
}

func handleUART1() {
	// should reset IRQ
	UART1.Receive(byte((UART1.Bus.DATA.Get() & 0xFF)))
	UART1.Bus.INTFLAG.SetBits(sam.SERCOM_USART_INT_INTFLAG_RXC)
}

// I2C on the SAMD51.
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
	// SERCOM_FREQ_REF is always reference frequency on SAMD51 regardless of CPU speed.
	SERCOM_FREQ_REF = 48000000

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
	// sam.SERCOM_I2CM_CTRLA_MODE_I2C_MASTER = 5?
	i2c.Bus.CTRLA.Set(5 << sam.SERCOM_I2CM_CTRLA_MODE_Pos) // |

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
	// Synchronous arithmetic baudrate, via Adafruit SAMD51 implementation:
	// sercom->I2CM.BAUD.bit.BAUD = SERCOM_FREQ_REF / ( 2 * baudrate) - 1 ;
	baud := SERCOM_FREQ_REF/(2*br) - 1
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

// SPI
type SPI struct {
	Bus         *sam.SERCOM_SPIM_Type
	SCK         Pin
	MOSI        Pin
	MISO        Pin
	DOpad       int
	DIpad       int
	SCKPinMode  PinMode
	MOSIPinMode PinMode
	MISOPinMode PinMode
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
	doPad := spi.DOpad
	diPad := spi.DIpad

	// set default frequency
	if config.Frequency == 0 {
		config.Frequency = 4000000
	}

	// Disable SPI port.
	spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPIM_CTRLA_ENABLE)
	for spi.Bus.SYNCBUSY.HasBits(sam.SERCOM_SPIM_SYNCBUSY_ENABLE) {
	}

	// enable pins
	if spi.SCKPinMode == 0 {
		spi.SCKPinMode = PinSERCOMAlt
	}
	if spi.MOSIPinMode == 0 {
		spi.MOSIPinMode = PinSERCOMAlt
	}
	if spi.MISOPinMode == 0 {
		spi.MISOPinMode = PinSERCOMAlt
	}

	spi.SCK.Configure(PinConfig{Mode: spi.SCKPinMode})
	spi.MOSI.Configure(PinConfig{Mode: spi.MOSIPinMode})
	spi.MISO.Configure(PinConfig{Mode: spi.MISOPinMode})

	// reset SERCOM
	spi.Bus.CTRLA.SetBits(sam.SERCOM_SPIM_CTRLA_SWRST)
	for spi.Bus.CTRLA.HasBits(sam.SERCOM_SPIM_CTRLA_SWRST) ||
		spi.Bus.SYNCBUSY.HasBits(sam.SERCOM_SPIM_SYNCBUSY_SWRST) {
	}

	// set bit transfer order
	dataOrder := 0
	if config.LSBFirst {
		dataOrder = 1
	}

	// Set SPI master
	// SERCOM_SPIM_CTRLA_MODE_SPI_MASTER = 3
	spi.Bus.CTRLA.Set(uint32((3 << sam.SERCOM_SPIM_CTRLA_MODE_Pos) |
		(doPad << sam.SERCOM_SPIM_CTRLA_DOPO_Pos) |
		(diPad << sam.SERCOM_SPIM_CTRLA_DIPO_Pos) |
		(dataOrder << sam.SERCOM_SPIM_CTRLA_DORD_Pos)))

	spi.Bus.CTRLB.SetBits((0 << sam.SERCOM_SPIM_CTRLB_CHSIZE_Pos) | // 8bit char size
		sam.SERCOM_SPIM_CTRLB_RXEN) // receive enable

	for spi.Bus.SYNCBUSY.HasBits(sam.SERCOM_SPIM_SYNCBUSY_CTRLB) {
	}

	// set mode
	switch config.Mode {
	case 0:
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPIM_CTRLA_CPHA)
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPIM_CTRLA_CPOL)
	case 1:
		spi.Bus.CTRLA.SetBits(sam.SERCOM_SPIM_CTRLA_CPHA)
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPIM_CTRLA_CPOL)
	case 2:
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPIM_CTRLA_CPHA)
		spi.Bus.CTRLA.SetBits(sam.SERCOM_SPIM_CTRLA_CPOL)
	case 3:
		spi.Bus.CTRLA.SetBits(sam.SERCOM_SPIM_CTRLA_CPHA | sam.SERCOM_SPIM_CTRLA_CPOL)
	default: // to mode 0
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPIM_CTRLA_CPHA)
		spi.Bus.CTRLA.ClearBits(sam.SERCOM_SPIM_CTRLA_CPOL)
	}

	// Set synch speed for SPI
	baudRate := SERCOM_FREQ_REF / (2 * config.Frequency)
	spi.Bus.BAUD.Set(uint8(baudRate))

	// Enable SPI port.
	spi.Bus.CTRLA.SetBits(sam.SERCOM_SPIM_CTRLA_ENABLE)
	for spi.Bus.SYNCBUSY.HasBits(sam.SERCOM_SPIM_SYNCBUSY_ENABLE) {
	}
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi SPI) Transfer(w byte) (byte, error) {
	// write data
	spi.Bus.DATA.Set(uint32(w))

	// wait for receive
	for !spi.Bus.INTFLAG.HasBits(sam.SERCOM_SPIM_INTFLAG_RXC) {
	}

	// return data
	return byte(spi.Bus.DATA.Get()), nil
}

// PWM
const period = 0xFFFF

// InitPWM initializes the PWM interface.
func InitPWM() {
	// turn on timer clocks used for PWM
	sam.MCLK.APBBMASK.SetBits(sam.MCLK_APBBMASK_TCC0_ | sam.MCLK_APBBMASK_TCC1_)
	sam.MCLK.APBCMASK.SetBits(sam.MCLK_APBCMASK_TCC2_)

	//use clock generator 0
	sam.GCLK.PCHCTRL[25].Set((sam.GCLK_PCHCTRL_GEN_GCLK0 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
	sam.GCLK.PCHCTRL[29].Set((sam.GCLK_PCHCTRL_GEN_GCLK0 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
}

// Configure configures a PWM pin for output.
func (pwm PWM) Configure() {
	// Set pin as output
	sam.PORT.GROUP[0].DIRSET.Set(1 << uint8(pwm.Pin))
	// Set pin to low
	sam.PORT.GROUP[0].OUTCLR.Set(1 << uint8(pwm.Pin))

	// Enable the port multiplexer for pin
	pwm.setPinCfg(sam.PORT_GROUP_PINCFG_PMUXEN)

	// Connect timer/mux to pin.
	pwmConfig := pwm.getMux()

	if pwm.Pin&1 > 0 {
		// odd pin, so save the even pins
		val := pwm.getPMux() & sam.PORT_GROUP_PMUX_PMUXE_Msk
		pwm.setPMux(val | uint8(pwmConfig<<sam.PORT_GROUP_PMUX_PMUXO_Pos))
	} else {
		// even pin, so save the odd pins
		val := pwm.getPMux() & sam.PORT_GROUP_PMUX_PMUXO_Msk
		pwm.setPMux(val | uint8(pwmConfig<<sam.PORT_GROUP_PMUX_PMUXE_Pos))
	}

	// figure out which TCCX timer for this pin
	timer := pwm.getTimer()

	// disable timer
	timer.CTRLA.ClearBits(sam.TCC_CTRLA_ENABLE)
	// Wait for synchronization
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_ENABLE) {
	}

	// Set prescaler to 1/256
	// TCCx->CTRLA.reg = TCC_CTRLA_PRESCALER_DIV256 | TCC_CTRLA_PRESCSYNC_GCLK;
	timer.CTRLA.SetBits(sam.TCC_CTRLA_PRESCALER_DIV256 | sam.TCC_CTRLA_PRESCSYNC_GCLK)

	// Use "Normal PWM" (single-slope PWM)
	timer.WAVE.SetBits(sam.TCC_WAVE_WAVEGEN_NPWM)
	// Wait for synchronization
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_WAVE) {
	}

	// while (TCCx->SYNCBUSY.bit.CC0 || TCCx->SYNCBUSY.bit.CC1);
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CC0) ||
		timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CC1) {
	}

	// Set the initial value
	// TCCx->CC[tcChannel].reg = (uint32_t) value;
	pwm.setChannel(0)

	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CC0) ||
		timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CC1) {
	}

	// Set the period (the number to count to (TOP) before resetting timer)
	//TCC0->PER.reg = period;
	timer.PER.Set(period)
	// Wait for synchronization
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_PER) {
	}

	// enable timer
	timer.CTRLA.SetBits(sam.TCC_CTRLA_ENABLE)
	// Wait for synchronization
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_ENABLE) {
	}
}

// Set turns on the duty cycle for a PWM pin using the provided value.
func (pwm PWM) Set(value uint16) {
	// figure out which TCCX timer for this pin
	timer := pwm.getTimer()

	// Wait for synchronization
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CTRLB) {
	}
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CC0) ||
		timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CC1) {
	}

	// TCCx->CCBUF[tcChannel].reg = (uint32_t) value;
	pwm.setChannelBuffer(uint32(value))

	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CC0) ||
		timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CC1) {
	}

	// TCCx->CTRLBCLR.bit.LUPD = 1;
	timer.CTRLBCLR.SetBits(sam.TCC_CTRLBCLR_LUPD)
	for timer.SYNCBUSY.HasBits(sam.TCC_SYNCBUSY_CTRLB) {
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
	case PA16:
		return sam.TCC1
	case PA17:
		return sam.TCC1
	case PA14:
		return sam.TCC2
	case PA15:
		return sam.TCC2
	case PA18:
		return sam.TCC1
	case PA19:
		return sam.TCC1
	case PA20:
		return sam.TCC0
	case PA21:
		return sam.TCC0
	case PA23:
		return sam.TCC0
	case PA22:
		return sam.TCC0
	default:
		return nil // not supported on this pin
	}
}

// setChannel sets the value for the correct channel for PWM on this pin
func (pwm PWM) setChannel(val uint32) {
	switch pwm.Pin {
	case PA16:
		pwm.getTimer().CC[0].Set(val)
	case PA17:
		pwm.getTimer().CC[1].Set(val)
	case PA14:
		pwm.getTimer().CC[0].Set(val)
	case PA15:
		pwm.getTimer().CC[1].Set(val)
	case PA18:
		pwm.getTimer().CC[2].Set(val)
	case PA19:
		pwm.getTimer().CC[3].Set(val)
	case PA20:
		pwm.getTimer().CC[0].Set(val)
	case PA21:
		pwm.getTimer().CC[1].Set(val)
	case PA23:
		pwm.getTimer().CC[3].Set(val)
	case PA22:
		pwm.getTimer().CC[2].Set(val)
	default:
		return // not supported on this pin
	}
}

// setChannelBuffer sets the value for the correct channel buffer for PWM on this pin
func (pwm PWM) setChannelBuffer(val uint32) {
	switch pwm.Pin {
	case PA16:
		pwm.getTimer().CCBUF[0].Set(val)
	case PA17:
		pwm.getTimer().CCBUF[1].Set(val)
	case PA14:
		pwm.getTimer().CCBUF[0].Set(val)
	case PA15:
		pwm.getTimer().CCBUF[1].Set(val)
	case PA18:
		pwm.getTimer().CCBUF[2].Set(val)
	case PA19:
		pwm.getTimer().CCBUF[3].Set(val)
	case PA20:
		pwm.getTimer().CCBUF[0].Set(val)
	case PA21:
		pwm.getTimer().CCBUF[1].Set(val)
	case PA23:
		pwm.getTimer().CCBUF[3].Set(val)
	case PA22:
		pwm.getTimer().CCBUF[2].Set(val)
	default:
		return // not supported on this pin
	}
}

// getMux returns the pin mode mux to be used for PWM on this pin.
func (pwm PWM) getMux() PinMode {
	switch pwm.Pin {
	case PA16:
		return PinPWMF
	case PA17:
		return PinPWMF
	case PA14:
		return PinPWMF
	case PA15:
		return PinPWMF
	case PA18:
		return PinPWMF
	case PA19:
		return PinPWMF
	case PA20:
		return PinPWMG
	case PA21:
		return PinPWMG
	case PA23:
		return PinPWMG
	case PA22:
		return PinPWMG
	default:
		return 0 // not supported on this pin
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
		setEPINTFLAG(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)

		// send data by setting bank ready
		setEPSTATUSSET(usb_CDC_ENDPOINT_IN, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)

		// wait for transfer to complete
		timeout := 3000
		for (getEPINTFLAG(usb_CDC_ENDPOINT_IN) & sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1) == 0 {
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
	// these are SAMD51 specific.
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
	arm.EnableIRQ(sam.IRQ_USB_OTHER)
	arm.EnableIRQ(sam.IRQ_USB_SOF_HSOF)
	arm.EnableIRQ(sam.IRQ_USB_TRCPT0)
	arm.EnableIRQ(sam.IRQ_USB_TRCPT1)
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

//go:export USB_OTHER_IRQHandler
func handleUSBOther() {
	handleUSBIRQ()
}

//go:export USB_SOF_HSOF_IRQHandler
func handleUSBSOFHSOF() {
	handleUSBIRQ()
}

//go:export USB_TRCPT0_IRQHandler
func handleUSBTRCPT0() {
	handleUSBIRQ()
}

//go:export USB_TRCPT1_IRQHandler
func handleUSBTRCPT1() {
	handleUSBIRQ()
}

func handleUSBIRQ() {
	// reset all interrupt flags
	flags := sam.USB_DEVICE.INTFLAG.Get()
	sam.USB_DEVICE.INTFLAG.Set(flags)

	// End of reset
	if (flags & sam.USB_DEVICE_INTFLAG_EORST) > 0 {
		// Configure control endpoint
		initEndpoint(0, usb_ENDPOINT_TYPE_CONTROL)

		// Enable Setup-Received interrupt
		setEPINTENSET(0, sam.USB_DEVICE_ENDPOINT_EPINTENSET_RXSTP)

		usbConfiguration = 0

		// ack the End-Of-Reset interrupt
		sam.USB_DEVICE.INTFLAG.Set(sam.USB_DEVICE_INTFLAG_EORST)
	}

	// Start of frame
	if (flags & sam.USB_DEVICE_INTFLAG_SOF) > 0 {
		// if you want to blink LED showing traffic, this would be the place...
	}

	// Endpoint 0 Setup interrupt
	if getEPINTFLAG(0)&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_RXSTP > 0 {
		// ack setup received
		setEPINTFLAG(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_RXSTP)

		// parse setup
		setup := newUSBSetup(udd_ep_out_cache_buffer[0][:])

		// Clear the Bank 0 ready flag on Control OUT
		setEPSTATUSCLR(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)
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
			setEPSTATUSSET(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)
		} else {
			// Stall endpoint
			setEPSTATUSSET(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1)
		}

		if getEPINTFLAG(0)&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1 > 0 {
			// ack the stall
			setEPINTFLAG(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_STALL1)

			// clear stall request
			setEPINTENCLR(0, sam.USB_DEVICE_ENDPOINT_EPINTENCLR_STALL1)
		}
	}

	// Now the actual transfer handlers, ignore endpoint number 0 (setup)
	var i uint32
	for i = 1; i < uint32(len(endPoints)); i++ {
		// Check if endpoint has a pending interrupt
		epFlags := getEPINTFLAG(i)
		if (epFlags&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT0) > 0 ||
			(epFlags&sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1) > 0 {
			switch i {
			case usb_CDC_ENDPOINT_OUT:
				handleEndpoint(i)
				setEPINTFLAG(i, epFlags)
			case usb_CDC_ENDPOINT_IN, usb_CDC_ENDPOINT_ACM:
				setEPSTATUSCLR(i, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK1RDY)
				setEPINTFLAG(i, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)
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
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_INTERRUPT + 1) << sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos))

	case usb_ENDPOINT_TYPE_BULK | usbEndpointOut:
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_BULK + 1) << sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE0_Pos))

		// receive interrupts when current transfer complete
		setEPINTENSET(ep, sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT0)

		// set byte count to zero, we have not received anything yet
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

		// ready for next transfer
		setEPSTATUSCLR(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)

	case usb_ENDPOINT_TYPE_INTERRUPT | usbEndpointOut:
		// TODO: not really anything, seems like...

	case usb_ENDPOINT_TYPE_BULK | usbEndpointIn:
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, ((usb_ENDPOINT_TYPE_BULK + 1) << sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos))

		// NAK on endpoint IN, the bank is not yet filled in.
		setEPSTATUSCLR(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK1RDY)

	case usb_ENDPOINT_TYPE_CONTROL:
		// Control OUT
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, getEPCFG(ep)|((usb_ENDPOINT_TYPE_CONTROL+1)<<sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE0_Pos))

		// Control IN
		// set packet size
		usbEndpointDescriptors[ep].DeviceDescBank[1].PCKSIZE.SetBits(epPacketSize(64) << usb_DEVICE_PCKSIZE_SIZE_Pos)

		// set data buffer address
		usbEndpointDescriptors[ep].DeviceDescBank[1].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_in_cache_buffer[ep]))))

		// set endpoint type
		setEPCFG(ep, getEPCFG(ep)|((usb_ENDPOINT_TYPE_CONTROL+1)<<sam.USB_DEVICE_ENDPOINT_EPCFG_EPTYPE1_Pos))

		// Prepare OUT endpoint for receive
		// set multi packet size for expected number of receive bytes on control OUT
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.SetBits(64 << usb_DEVICE_PCKSIZE_MULTI_PACKET_SIZE_Pos)

		// set byte count to zero, we have not received anything yet
		usbEndpointDescriptors[ep].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

		// NAK on endpoint OUT to show we are ready to receive control data
		setEPSTATUSSET(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK0RDY)
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
		setEPINTFLAG(0, sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1)

		// set bank ready for data
		setEPSTATUSSET(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSSET_BK1RDY)

		// wait for transfer to complete
		timeout := 3000
		for (getEPINTFLAG(0) & sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1) == 0 {
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
			setEPINTENSET(usb_CDC_ENDPOINT_ACM, sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT1)

			// Enable interrupt for CDC data messages from host
			setEPINTENSET(usb_CDC_ENDPOINT_OUT, sam.USB_DEVICE_ENDPOINT_EPINTENSET_TRCPT0)

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
			sendZlp(0)
		}

		if setup.bRequest == usb_CDC_SEND_BREAK {
			// TODO: something with this value?
			// breakValue = ((uint16_t)setup.wValueH << 8) | setup.wValueL;
			// return false;
			sendZlp(0)
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
	// address
	usbEndpointDescriptors[0].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[0]))))

	// set byte count to zero
	usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

	// set ready for next data
	setEPSTATUSCLR(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)

	// Wait until OUT transfer is ready.
	timeout := 300000
	for (getEPSTATUS(0) & sam.USB_DEVICE_ENDPOINT_EPSTATUS_BK0RDY) == 0 {
		timeout--
		if timeout == 0 {
			return []byte{}
		}
	}

	// Wait until OUT transfer is completed.
	timeout = 300000
	for (getEPINTFLAG(0) & sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT1) == 0 {
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

		out := NewEndpointDescriptor((usb_CDC_ENDPOINT_OUT | usbEndpointOut), usb_ENDPOINT_TYPE_BULK, usbEndpointPacketSize, 0)

		in := NewEndpointDescriptor((usb_CDC_ENDPOINT_IN | usbEndpointIn), usb_ENDPOINT_TYPE_BULK, usbEndpointPacketSize, 0)

		cdc := NewCDCDescriptor(iad,
			cif,
			header,
			controlManagement,
			functionalDescriptor,
			callManagement,
			cifin,
			dif,
			out,
			in)

		sz := uint16(configDescriptorSize + cdcSize)
		config := NewConfigDescriptor(sz, 2)

		buf := make([]byte, 0)
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
	setEPSTATUSCLR(ep, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)
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
	return sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPCFG.Get()
}

func setEPCFG(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPCFG.Set(val)
}

func setEPSTATUSCLR(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPSTATUSCLR.Set(val)
}

func setEPSTATUSSET(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPSTATUSSET.Set(val)
}

func getEPSTATUS(ep uint32) uint8 {
	return sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPSTATUS.Get()
}

func getEPINTFLAG(ep uint32) uint8 {
	return sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTFLAG.Get()
}

func setEPINTFLAG(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTFLAG.Set(val)
}

func setEPINTENCLR(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTENCLR.Set(val)
}

func setEPINTENSET(ep uint32, val uint8) {
	sam.USB_DEVICE.DEVICE_ENDPOINT[ep].EPINTENSET.Set(val)
}

// ResetProcessor should perform a system reset in preperation
// to switch to the bootloader to flash new firmware.
func ResetProcessor() {
	arm.DisableInterrupts()

	// Perform magic reset into bootloader, as mentioned in
	// https://github.com/arduino/ArduinoCore-samd/issues/197
	*(*uint32)(unsafe.Pointer(uintptr(0x20000000 + 0x00030000 - 4))) = RESET_MAGIC_VALUE

	arm.SystemReset()
}
