// +build nrf52

package machine

import (
	"device/nrf"
	"unsafe"
)

// Get peripheral and pin number for this GPIO pin.
func (p GPIO) getPortPin() (*nrf.GPIO_Type, uint8) {
	return nrf.P0, p.Pin
}

// InitADC initializes the registers needed for ADC.
func InitADC() {
	return // no specific setup on nrf52 machine.
}

// Configure configures an ADC pin to be able to read analog data.
func (a ADC) Configure() {
	return // no pin specific setup on nrf52 machine.
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

	nrf.SAADC.RESOLUTION = nrf.SAADC_RESOLUTION_VAL_12bit

	// Enable ADC.
	nrf.SAADC.ENABLE = (nrf.SAADC_ENABLE_ENABLE_Enabled << nrf.SAADC_ENABLE_ENABLE_Pos)
	for i := 0; i < 8; i++ {
		nrf.SAADC.CH[i].PSELN = nrf.SAADC_CH_PSELP_PSELP_NC
		nrf.SAADC.CH[i].PSELP = nrf.SAADC_CH_PSELP_PSELP_NC
	}

	// Configure ADC.
	nrf.SAADC.CH[0].CONFIG = ((nrf.SAADC_CH_CONFIG_RESP_Bypass << nrf.SAADC_CH_CONFIG_RESP_Pos) & nrf.SAADC_CH_CONFIG_RESP_Msk) |
		((nrf.SAADC_CH_CONFIG_RESP_Bypass << nrf.SAADC_CH_CONFIG_RESN_Pos) & nrf.SAADC_CH_CONFIG_RESN_Msk) |
		((nrf.SAADC_CH_CONFIG_GAIN_Gain1_5 << nrf.SAADC_CH_CONFIG_GAIN_Pos) & nrf.SAADC_CH_CONFIG_GAIN_Msk) |
		((nrf.SAADC_CH_CONFIG_REFSEL_Internal << nrf.SAADC_CH_CONFIG_REFSEL_Pos) & nrf.SAADC_CH_CONFIG_REFSEL_Msk) |
		((nrf.SAADC_CH_CONFIG_TACQ_3us << nrf.SAADC_CH_CONFIG_TACQ_Pos) & nrf.SAADC_CH_CONFIG_TACQ_Msk) |
		((nrf.SAADC_CH_CONFIG_MODE_SE << nrf.SAADC_CH_CONFIG_MODE_Pos) & nrf.SAADC_CH_CONFIG_MODE_Msk)

	// Set pin to read.
	nrf.SAADC.CH[0].PSELN = nrf.RegValue(pwmPin)
	nrf.SAADC.CH[0].PSELP = nrf.RegValue(pwmPin)

	// Destination for sample result.
	nrf.SAADC.RESULT.PTR = nrf.RegValue(uintptr(unsafe.Pointer(&value)))
	nrf.SAADC.RESULT.MAXCNT = 1 // One sample

	// Start tasks.
	nrf.SAADC.TASKS_START = 1
	for nrf.SAADC.EVENTS_STARTED == 0 {
	}
	nrf.SAADC.EVENTS_STARTED = 0x00

	// Start the sample task.
	nrf.SAADC.TASKS_SAMPLE = 1

	// Wait until the sample task is done.
	for nrf.SAADC.EVENTS_END == 0 {
	}
	nrf.SAADC.EVENTS_END = 0x00

	// Stop the ADC
	nrf.SAADC.TASKS_STOP = 1
	for nrf.SAADC.EVENTS_STOPPED == 0 {
	}
	nrf.SAADC.EVENTS_STOPPED = 0

	// Disable the ADC.
	nrf.SAADC.ENABLE = (nrf.SAADC_ENABLE_ENABLE_Disabled << nrf.SAADC_ENABLE_ENABLE_Pos)

	if value < 0 {
		value = 0
	}

	// Return 16-bit result from 12-bit value.
	return uint16(value << 4)
}

// PWM
var (
	pwmChannelPins     = [3]uint32{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF}
	pwms               = [3]*nrf.PWM_Type{nrf.PWM0, nrf.PWM1, nrf.PWM2}
	pwmChannelSequence [3]uint16
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
	for i := 0; i < 3; i++ {
		if pwmChannelPins[i] == 0xFFFFFFFF || pwmChannelPins[i] == uint32(pwm.Pin) {
			pwmChannelPins[i] = uint32(pwm.Pin)
			pwmChannelSequence[i] = (value >> 2) | 0x8000 // set bit 15 to invert polarity

			p := pwms[i]

			p.PSEL.OUT[0] = nrf.RegValue(pwm.Pin)
			p.PSEL.OUT[1] = nrf.RegValue(pwm.Pin)
			p.PSEL.OUT[2] = nrf.RegValue(pwm.Pin)
			p.PSEL.OUT[3] = nrf.RegValue(pwm.Pin)
			p.ENABLE = (nrf.PWM_ENABLE_ENABLE_Enabled << nrf.PWM_ENABLE_ENABLE_Pos)
			p.PRESCALER = nrf.PWM_PRESCALER_PRESCALER_DIV_2
			p.MODE = nrf.PWM_MODE_UPDOWN_Up
			p.COUNTERTOP = 16384 // frequency
			p.LOOP = 0
			p.DECODER = (nrf.PWM_DECODER_LOAD_Common << nrf.PWM_DECODER_LOAD_Pos) | (nrf.PWM_DECODER_MODE_RefreshCount << nrf.PWM_DECODER_MODE_Pos)
			p.SEQ[0].PTR = nrf.RegValue(uintptr(unsafe.Pointer(&pwmChannelSequence[i])))
			p.SEQ[0].CNT = 1
			p.SEQ[0].REFRESH = 1
			p.SEQ[0].ENDDELAY = 0
			p.TASKS_SEQSTART[0] = 1

			break
		}
	}
}
