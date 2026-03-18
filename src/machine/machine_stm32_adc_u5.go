//go:build stm32u5

package machine

import (
	"device/stm32"
	"unsafe"
)

// InitADC initializes the registers needed for ADC1.
func InitADC() {
	// Enable ADC bus clock.
	enableAltFuncClock(unsafe.Pointer(stm32.ADC1))

	// Declare VDDA analog supply valid. Without ASV the ADC LDO cannot start.
	stm32.PWR.SVMCR.SetBits(stm32.PWR_SVMCR_ASV)

	// Set ADC12_Common CCR: synchronous clock HCLK/1.
	// ADC12_Common is mapped as ADC_Type; CCR at offset 0x08 = .CR field.
	stm32.ADC12_Common.CR.Set(1 << 16) // CKMODE=01

	// Exit deep power-down mode.
	stm32.ADC1.CR.ClearBits(stm32.ADC_CR_DEEPPWD)

	// Enable internal voltage regulator and wait for LDO ready.
	stm32.ADC1.CR.SetBits(stm32.ADC_CR_ADVREGEN)
	for !stm32.ADC1.ISR.HasBits(stm32.ADC_ISR_LDORDY) {
	}

	// Calibrate ADC (single-ended).
	stm32.ADC1.CR.SetBits(stm32.ADC_CR_ADCAL)
	for stm32.ADC1.CR.HasBits(stm32.ADC_CR_ADCAL) {
	}

	// 12-bit resolution, overwrite DR on overrun.
	stm32.ADC1.CFGR1.Set(stm32.ADC_CFGR1_RES_TwelveBit<<stm32.ADC_CFGR1_RES_Pos |
		stm32.ADC_CFGR1_OVRMOD)

	// Pre-select all channels (PCSEL must be written while ADEN=0).
	stm32.ADC1.PCSEL.Set(0xFFFFF)

	// Clear ADRDY flag, enable ADC, wait for ready.
	stm32.ADC1.ISR.Set(stm32.ADC_ISR_ADRDY)
	stm32.ADC1.CR.SetBits(stm32.ADC_CR_ADEN)
	for !stm32.ADC1.ISR.HasBits(stm32.ADC_ISR_ADRDY) {
	}
}

// Configure configures an ADC pin to be able to read analog data.
func (a ADC) Configure(config ADCConfig) {
	a.Pin.Configure(PinConfig{Mode: PinInputAnalog})

	ch := a.getChannel()

	// Set sampling time for this channel.
	// Channels 0-9 use SMPR1, channels 10-19 use SMPR2.
	// Each channel uses 3 bits. Default to 36.5 cycles for reliable reads.
	const smpVal = 0x4 // 36.5 ADC clock cycles
	if ch <= 9 {
		pos := uint8(ch) * 3
		stm32.ADC1.SMPR1.ReplaceBits(smpVal, 0x7, pos)
	} else {
		pos := uint8(ch-10) * 3
		stm32.ADC1.SMPR2.ReplaceBits(smpVal, 0x7, pos)
	}
}

// Get returns the current value of an ADC pin in the range 0..0xffff.
func (a ADC) Get() uint16 {
	ch := uint32(a.getChannel())

	// PCSEL already set for all channels in InitADC (must be written while ADEN=0).

	// Set up regular sequence: 1 conversion, channel in SQ1
	// Note: ADC_SQR1_SQ1_Pos (0xC) in the device header is wrong for the 14-bit ADC1/ADC2.
	// RM0456 shows SQ1 at bits [10:6] = position 6.
	stm32.ADC1.SQR1.Set(ch << 6) // L=0 means 1 conversion

	// Start conversion
	stm32.ADC1.CR.SetBits(stm32.ADC_CR_ADSTART)

	// Wait for end of conversion
	for stm32.ADC1.ISR.Get()&stm32.ADC_ISR_EOC == 0 {
	}

	// Read 12-bit result and scale to 16-bit
	result := uint16(stm32.ADC1.DR.Get()) << 4

	return result
}

// getChannel returns the ADC1 channel number for a given GPIO pin.
// STM32U585 ADC1 channel mapping (from datasheet Table 24):
//
//	PC0=CH1, PC1=CH2, PC2=CH3, PC3=CH4
//	PA0=CH5, PA1=CH6, PA2=CH7, PA3=CH8
//	PA4=CH9, PA5=CH10, PA6=CH11, PA7=CH12
//	PB0=CH13, PB1=CH14
func (a ADC) getChannel() uint8 {
	switch a.Pin {
	case PC0:
		return 1
	case PC1:
		return 2
	case PC2:
		return 3
	case PC3:
		return 4
	case PA0:
		return 5
	case PA1:
		return 6
	case PA2:
		return 7
	case PA3:
		return 8
	case PA4:
		return 9
	case PA5:
		return 10
	case PA6:
		return 11
	case PA7:
		return 12
	case PB0:
		return 13
	case PB1:
		return 14
	}
	return 0
}
