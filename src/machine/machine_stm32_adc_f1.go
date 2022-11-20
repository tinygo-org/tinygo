//go:build stm32f103
// +build stm32f103

package machine

import (
	"device/stm32"
	"unsafe"
)

const (
	Cycles_1_5   = 0x0
	Cycles_7_5   = 0x1
	Cycles_13_5  = 0x2
	Cycles_28_5  = 0x3
	Cycles_41_5  = 0x4
	Cycles_55_5  = 0x5
	Cycles_71_5  = 0x6
	Cycles_239_5 = 0x7
)

// InitADC initializes the registers needed for ADC1.
func InitADC() {
	// Enable ADC clock
	enableAltFuncClock(unsafe.Pointer(stm32.ADC1))

	// set scan mode
	stm32.ADC1.CR1.SetBits(stm32.ADC_CR1_SCAN)

	// clear CONT, ALIGN, EXTRIG and EXTSEL bits from CR2
	stm32.ADC1.CR2.ClearBits(stm32.ADC_CR2_CONT | stm32.ADC_CR2_ALIGN | stm32.ADC_CR2_EXTTRIG_Msk | stm32.ADC_CR2_EXTSEL_Msk)

	stm32.ADC1.SQR1.ClearBits(stm32.ADC_SQR1_L_Msk)
	stm32.ADC1.SQR1.SetBits(2 << stm32.ADC_SQR1_L_Pos) // 2 means 3 conversions

	// enable
	stm32.ADC1.CR2.SetBits(stm32.ADC_CR2_ADON)

	return
}

// Configure configures an ADC pin to be able to read analog data.
func (a ADC) Configure(ADCConfig) {
	a.Pin.Configure(PinConfig{Mode: PinInputModeAnalog})

	// set sample time
	ch := a.getChannel()
	if ch > 9 {
		stm32.ADC1.SMPR1.SetBits(Cycles_28_5 << (ch - 10) * stm32.ADC_SMPR1_SMP11_Pos)
	} else {
		stm32.ADC1.SMPR2.SetBits(Cycles_28_5 << (ch * stm32.ADC_SMPR2_SMP1_Pos))
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

	// read result as 16 bit value
	result := uint16(stm32.ADC1.DR.Get()) << 4

	// clear flag
	stm32.ADC1.SR.ClearBits(stm32.ADC_SR_EOC)

	// clear rank
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
	}

	return 0
}
