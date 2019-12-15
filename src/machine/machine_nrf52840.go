// +build nrf52840

package machine

import (
	"device/nrf"
	"unsafe"
)

func CPUFrequency() uint32 {
	return 64000000
}

// Get peripheral and pin number for this GPIO pin.
func (p Pin) getPortPin() (*nrf.GPIO_Type, uint32) {
	if p >= 32 {
		return nrf.P1, uint32(p - 32)
	} else {
		return nrf.P0, uint32(p)
	}
}

func (uart UART) setPins(tx, rx Pin) {
	nrf.UART0.PSEL.TXD.Set(uint32(tx))
	nrf.UART0.PSEL.RXD.Set(uint32(rx))
}

func (i2c I2C) setPins(scl, sda Pin) {
	i2c.Bus.PSEL.SCL.Set(uint32(scl))
	i2c.Bus.PSEL.SDA.Set(uint32(sda))
}

// SPI
func (spi SPI) setPins(sck, mosi, miso Pin) {
	if sck == 0 {
		sck = SPI0_SCK_PIN
	}
	if mosi == 0 {
		mosi = SPI0_MOSI_PIN
	}
	if miso == 0 {
		miso = SPI0_MISO_PIN
	}
	spi.Bus.PSEL.SCK.Set(uint32(sck))
	spi.Bus.PSEL.MOSI.Set(uint32(mosi))
	spi.Bus.PSEL.MISO.Set(uint32(miso))
}

// InitADC initializes the registers needed for ADC.
func InitADC() {
	return // no specific setup on nrf52840 machine.
}

// Configure configures an ADC pin to be able to read analog data.
func (a ADC) Configure() {
	return // no pin specific setup on nrf52840 machine.
}

// Get returns the current value of a ADC pin in the range 0..0xffff.
func (a ADC) Get() uint16 {
	var pwmPin uint32
	var value int16

	switch a.Pin {
	case 2:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput0

	case 3:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput1

	case 4:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput2

	case 5:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput3

	case 28:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput4

	case 29:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput5

	case 30:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput6

	case 31:
		pwmPin = nrf.SAADC_CH_PSELP_PSELP_AnalogInput7

	default:
		return 0
	}

	nrf.SAADC.RESOLUTION.Set(nrf.SAADC_RESOLUTION_VAL_12bit)

	// Enable ADC.
	nrf.SAADC.ENABLE.Set(nrf.SAADC_ENABLE_ENABLE_Enabled << nrf.SAADC_ENABLE_ENABLE_Pos)
	for i := 0; i < 8; i++ {
		nrf.SAADC.CH[i].PSELN.Set(nrf.SAADC_CH_PSELP_PSELP_NC)
		nrf.SAADC.CH[i].PSELP.Set(nrf.SAADC_CH_PSELP_PSELP_NC)
	}

	// Configure ADC.
	nrf.SAADC.CH[0].CONFIG.Set(((nrf.SAADC_CH_CONFIG_RESP_Bypass << nrf.SAADC_CH_CONFIG_RESP_Pos) & nrf.SAADC_CH_CONFIG_RESP_Msk) |
		((nrf.SAADC_CH_CONFIG_RESP_Bypass << nrf.SAADC_CH_CONFIG_RESN_Pos) & nrf.SAADC_CH_CONFIG_RESN_Msk) |
		((nrf.SAADC_CH_CONFIG_GAIN_Gain1_5 << nrf.SAADC_CH_CONFIG_GAIN_Pos) & nrf.SAADC_CH_CONFIG_GAIN_Msk) |
		((nrf.SAADC_CH_CONFIG_REFSEL_Internal << nrf.SAADC_CH_CONFIG_REFSEL_Pos) & nrf.SAADC_CH_CONFIG_REFSEL_Msk) |
		((nrf.SAADC_CH_CONFIG_TACQ_3us << nrf.SAADC_CH_CONFIG_TACQ_Pos) & nrf.SAADC_CH_CONFIG_TACQ_Msk) |
		((nrf.SAADC_CH_CONFIG_MODE_SE << nrf.SAADC_CH_CONFIG_MODE_Pos) & nrf.SAADC_CH_CONFIG_MODE_Msk))

	// Set pin to read.
	nrf.SAADC.CH[0].PSELN.Set(pwmPin)
	nrf.SAADC.CH[0].PSELP.Set(pwmPin)

	// Destination for sample result.
	nrf.SAADC.RESULT.PTR.Set(uint32(uintptr(unsafe.Pointer(&value))))
	nrf.SAADC.RESULT.MAXCNT.Set(1) // One sample

	// Start tasks.
	nrf.SAADC.TASKS_START.Set(1)
	for nrf.SAADC.EVENTS_STARTED.Get() == 0 {
	}
	nrf.SAADC.EVENTS_STARTED.Set(0x00)

	// Start the sample task.
	nrf.SAADC.TASKS_SAMPLE.Set(1)

	// Wait until the sample task is done.
	for nrf.SAADC.EVENTS_END.Get() == 0 {
	}
	nrf.SAADC.EVENTS_END.Set(0x00)

	// Stop the ADC
	nrf.SAADC.TASKS_STOP.Set(1)
	for nrf.SAADC.EVENTS_STOPPED.Get() == 0 {
	}
	nrf.SAADC.EVENTS_STOPPED.Set(0)

	// Disable the ADC.
	nrf.SAADC.ENABLE.Set(nrf.SAADC_ENABLE_ENABLE_Disabled << nrf.SAADC_ENABLE_ENABLE_Pos)

	if value < 0 {
		value = 0
	}

	// Return 16-bit result from 12-bit value.
	return uint16(value << 4)
}

// PWM
var (
	pwmChannelPins     = [4]uint32{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF}
	pwms               = [4]*nrf.PWM_Type{nrf.PWM0, nrf.PWM1, nrf.PWM2, nrf.PWM3}
	pwmChannelSequence [4]uint16
)

// InitPWM initializes the registers needed for PWM.
func InitPWM() {
	return
}

// Configure configures a PWM pin for output.
func (pwm PWM) Configure() {
}

// Set turns on the duty cycle for a PWM pin using the provided value.
func (pwm PWM) Set(value uint16) {
	for i := 0; i < 4; i++ {
		if pwmChannelPins[i] == 0xFFFFFFFF || pwmChannelPins[i] == uint32(pwm.Pin) {
			pwmChannelPins[i] = uint32(pwm.Pin)
			pwmChannelSequence[i] = (value >> 2) | 0x8000 // set bit 15 to invert polarity

			p := pwms[i]

			p.PSEL.OUT[0].Set(uint32(pwm.Pin))
			p.PSEL.OUT[1].Set(uint32(pwm.Pin))
			p.PSEL.OUT[2].Set(uint32(pwm.Pin))
			p.PSEL.OUT[3].Set(uint32(pwm.Pin))
			p.ENABLE.Set(nrf.PWM_ENABLE_ENABLE_Enabled << nrf.PWM_ENABLE_ENABLE_Pos)
			p.PRESCALER.Set(nrf.PWM_PRESCALER_PRESCALER_DIV_2)
			p.MODE.Set(nrf.PWM_MODE_UPDOWN_Up)
			p.COUNTERTOP.Set(16384) // frequency
			p.LOOP.Set(0)
			p.DECODER.Set((nrf.PWM_DECODER_LOAD_Common << nrf.PWM_DECODER_LOAD_Pos) | (nrf.PWM_DECODER_MODE_RefreshCount << nrf.PWM_DECODER_MODE_Pos))
			p.SEQ[0].PTR.Set(uint32(uintptr(unsafe.Pointer(&pwmChannelSequence[i]))))
			p.SEQ[0].CNT.Set(1)
			p.SEQ[0].REFRESH.Set(1)
			p.SEQ[0].ENDDELAY.Set(0)
			p.TASKS_SEQSTART[0].Set(1)

			break
		}
	}
}
