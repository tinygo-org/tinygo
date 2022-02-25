//go:build stm32f4
// +build stm32f4

package machine

import (
	"device/stm32"
	"unsafe"
)

// InitADC initializes the registers needed for ADC1.
func InitADC() {
	// Enable ADC clock
	enableAltFuncClock(unsafe.Pointer(stm32.ADC1))

	// stop scan, and clear scan resolution
	stm32.ADC1.CR1.ClearBits(stm32.ADC_CR1_SCAN | stm32.ADC_CR1_RES_Msk)

	// set conversion mode and resolution
	stm32.ADC1.CR1.SetBits(stm32.ADC_CR1_RES_TwelveBit)

	// clear CONT, ALIGN, EXTEN and EXTSEL bits from CR2
	stm32.ADC1.CR2.ClearBits(stm32.ADC_CR2_CONT | stm32.ADC_CR2_ALIGN | stm32.ADC_CR2_EXTEN_Msk | stm32.ADC_CR2_EXTSEL_Msk)

	// set CONT, ALIGN, EXTEN and EXTSEL bits from CR2
	stm32.ADC1.CR2.SetBits(stm32.ADC_CR2_CONT_Single | stm32.ADC_CR2_ALIGN_Right)

	stm32.ADC1.SQR1.ClearBits(stm32.ADC_SQR1_L_Msk)
	stm32.ADC1.SQR1.SetBits(2 << stm32.ADC_SQR1_L_Pos) // 2 means 3 conversions

	// enable
	stm32.ADC1.CR2.SetBits(stm32.ADC_CR2_ADON)

	return
}

// Configure configures an ADC pin to be able to read analog data.
func (a ADC) Configure(ADCConfig) {
	a.Pin.ConfigureAltFunc(PinConfig{Mode: PinInputAnalog}, 0)

	// set sample time
	ch := a.getChannel()
	if ch > 9 {
		stm32.ADC1.SMPR1.SetBits(stm32.ADC_SMPR1_SMP11_Cycles84 << (ch - 10) * stm32.ADC_SMPR1_SMP11_Pos)
	} else {
		stm32.ADC1.SMPR2.SetBits(stm32.ADC_SMPR2_SMP1_Cycles84 << (ch * stm32.ADC_SMPR2_SMP1_Pos))
	}

	return
}

// Get returns the current value of a ADC pin in the range 0..0xffff.
// TODO: DMA based implementation.
func (a ADC) Get() uint16 {
	// set rank
	ch := uint32(a.getChannel())
	stm32.ADC1.SQR3.SetBits(ch)

	// start conversion
	stm32.ADC1.CR2.SetBits(stm32.ADC_CR2_SWSTART)

	// wait for conversion to complete
	for !stm32.ADC1.SR.HasBits(stm32.ADC_SR_EOC) {
	}

	// read 12-bit result as 16 bit value
	result := uint16(stm32.ADC1.DR.Get()) << 4

	// clear flag
	stm32.ADC1.SR.ClearBits(stm32.ADC_SR_EOC)

	// clear rank
	stm32.ADC1.SMPR1.ClearBits(ch)

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
