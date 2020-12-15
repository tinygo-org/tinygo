// +build nrf52

package machine

import (
	"device/nrf"
	"unsafe"
)

// Hardware pins
const (
	P0_00 Pin = 0
	P0_01 Pin = 1
	P0_02 Pin = 2
	P0_03 Pin = 3
	P0_04 Pin = 4
	P0_05 Pin = 5
	P0_06 Pin = 6
	P0_07 Pin = 7
	P0_08 Pin = 8
	P0_09 Pin = 9
	P0_10 Pin = 10
	P0_11 Pin = 11
	P0_12 Pin = 12
	P0_13 Pin = 13
	P0_14 Pin = 14
	P0_15 Pin = 15
	P0_16 Pin = 16
	P0_17 Pin = 17
	P0_18 Pin = 18
	P0_19 Pin = 19
	P0_20 Pin = 20
	P0_21 Pin = 21
	P0_22 Pin = 22
	P0_23 Pin = 23
	P0_24 Pin = 24
	P0_25 Pin = 25
	P0_26 Pin = 26
	P0_27 Pin = 27
	P0_28 Pin = 28
	P0_29 Pin = 29
	P0_30 Pin = 30
	P0_31 Pin = 31
)

var (
	UART0 = NRF_UART0
)

// Get peripheral and pin number for this GPIO pin.
func (p Pin) getPortPin() (*nrf.GPIO_Type, uint32) {
	return nrf.P0, uint32(p)
}

func (uart UART) setPins(tx, rx Pin) {
	nrf.UART0.PSELTXD.Set(uint32(tx))
	nrf.UART0.PSELRXD.Set(uint32(rx))
}

func (i2c I2C) setPins(scl, sda Pin) {
	i2c.Bus.PSELSCL.Set(uint32(scl))
	i2c.Bus.PSELSDA.Set(uint32(sda))
}

// SPI
func (spi SPI) setPins(sck, sdo, sdi Pin) {
	if sck == 0 {
		sck = SPI0_SCK_PIN
	}
	if sdo == 0 {
		sdo = SPI0_SDO_PIN
	}
	if sdi == 0 {
		sdi = SPI0_SDI_PIN
	}
	spi.Bus.PSEL.SCK.Set(uint32(sck))
	spi.Bus.PSEL.MOSI.Set(uint32(sdo))
	spi.Bus.PSEL.MISO.Set(uint32(sdi))
}

// InitADC initializes the registers needed for ADC.
func InitADC() {
	return // no specific setup on nrf52 machine.
}

// Configure configures an ADC pin to be able to read analog data.
func (a ADC) Configure(ADCConfig) {
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
	pwmChannelPins     = [3]uint32{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF}
	pwms               = [3]*nrf.PWM_Type{nrf.PWM0, nrf.PWM1, nrf.PWM2}
	pwmChannelSequence [3]uint16
)
