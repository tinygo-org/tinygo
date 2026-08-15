//go:build stm32f401

package machine

import (
	"device/stm32"
	"unsafe"
)

// InitADC initializes the registers needed for ADC1 on STM32F401.
func InitADC() {
	enableAltFuncClock(unsafe.Pointer(stm32.ADC1))

	stm32.ADC1.CR1.ClearBits(stm32.ADC_CR1_SCAN | stm32.ADC_CR1_RES_Msk)
	stm32.ADC1.CR1.SetBits(stm32.ADC_CR1_RES_TwelveBit)
	stm32.ADC1.CR2.ClearBits(stm32.ADC_CR2_CONT | stm32.ADC_CR2_ALIGN | stm32.ADC_CR2_EXTEN_Msk | stm32.ADC_CR2_EXTSEL_Msk)
	stm32.ADC1.CR2.SetBits(stm32.ADC_CR2_CONT_Single | stm32.ADC_CR2_ALIGN_Right)
	stm32.ADC1.SQR1.ClearBits(stm32.ADC_SQR1_L_Msk)
	stm32.ADC1.SQR1.SetBits(2 << stm32.ADC_SQR1_L_Pos)
	stm32.ADC1.CR2.SetBits(stm32.ADC_CR2_ADON)
}

// Configure configures an ADC pin on STM32F401.
// The F401 ADC SVD does not provide named Cycles84 constants per channel;
// the value 4 (binary 100) selects 84-cycle sample time in the 3-bit field.
func (a ADC) Configure(ADCConfig) {
	a.Pin.ConfigureAltFunc(PinConfig{Mode: PinInputAnalog}, 0)

	const cycles84 = 4
	ch := a.getChannel()
	if ch > 9 {
		stm32.ADC1.SMPR1.SetBits(cycles84 << ((ch - 10) * 3))
	} else {
		stm32.ADC1.SMPR2.SetBits(cycles84 << (ch * 3))
	}
}

// Get returns the current value of a ADC pin in the range 0..0xffff.
func (a ADC) Get() uint16 {
	ch := uint32(a.getChannel())
	stm32.ADC1.SQR3.SetBits(ch)
	stm32.ADC1.CR2.SetBits(stm32.ADC_CR2_SWSTART)
	for !stm32.ADC1.SR.HasBits(stm32.ADC_SR_EOC) {
	}
	result := uint16(stm32.ADC1.DR.Get()) << 4
	stm32.ADC1.SR.ClearBits(stm32.ADC_SR_EOC)
	stm32.ADC1.SQR3.ClearBits(ch)
	return result
}

func (a ADC) getChannel() uint8 {
	switch a.Pin {
	case PA0:
		return 0
	case PA1:
		return 1
	case PA2:
		return 2
	case PA3:
		return 3
	case PA4:
		return 4
	case PA5:
		return 5
	case PA6:
		return 6
	case PA7:
		return 7
	case PB0:
		return 8
	case PB1:
		return 9
	case PC0:
		return 10
	case PC1:
		return 11
	case PC2:
		return 12
	case PC3:
		return 13
	case PC4:
		return 14
	case PC5:
		return 15
	}
	return 0
}
