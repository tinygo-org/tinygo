//go:build stm32h7

package machine

import (
	"device/arm"
	"device/stm32"
)

// InitADC initializes the registers needed for ADC1 and ADC3.
func InitADC() {
	// 1. Enable ADC bus clocks
	stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_ADC12EN)
	stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_ADC3EN)

	// 2. Configure ADC clock mode (Async from kernel clock)
	// CCR is at offset 0x308 from ADC base.
	stm32.ADC12_Common.CR.ReplaceBits(0x0, 0x3, 16) // CKMODE = 00
	stm32.ADC3_Common.CR.ReplaceBits(0x0, 0x3, 16)  // CKMODE = 00

	// 3. Exit deep power-down mode
	stm32.ADC1.CR.ClearBits(stm32.ADC_CR_DEEPPWD)
	stm32.ADC3.CR.ClearBits(stm32.ADC_CR_DEEPPWD)

	// 4. Enable voltage regulators
	stm32.ADC1.CR.SetBits(stm32.ADC_CR_ADVREGEN)
	stm32.ADC3.CR.SetBits(stm32.ADC_CR_ADVREGEN)
	// Wait for T_ADCVREG_STUP (min 10us). The nop keeps the compiler from
	// eliminating the loop as free of side effects.
	for i := 0; i < 10000; i++ {
		arm.Asm("nop")
	}

	// Set BOOST[1:0]=0b11 for ADC kernel clock >25MHz (RM0433 §25.4.3 Table 121).
	// The kernel clock is 80MHz (PLL2_P), so BOOST must be enabled.
	stm32.ADC1.CR.ReplaceBits(0b11, 0x3, stm32.ADC_CR_BOOST_Pos)
	stm32.ADC3.CR.ReplaceBits(0b11, 0x3, stm32.ADC_CR_BOOST_Pos)

	// 5. Calibration
	// ADC1
	stm32.ADC1.CR.SetBits(stm32.ADC_CR_ADCAL | stm32.ADC_CR_ADCALLIN)
	for stm32.ADC1.CR.HasBits(stm32.ADC_CR_ADCAL) {
	}
	// ADC3
	stm32.ADC3.CR.SetBits(stm32.ADC_CR_ADCAL | stm32.ADC_CR_ADCALLIN)
	for stm32.ADC3.CR.HasBits(stm32.ADC_CR_ADCAL) {
	}

	// 6. Enable ADCs
	// ADC1
	stm32.ADC1.ISR.SetBits(stm32.ADC_ISR_ADRDY) // Clear ADRDY by writing 1
	stm32.ADC1.CR.SetBits(stm32.ADC_CR_ADEN)
	for !stm32.ADC1.ISR.HasBits(stm32.ADC_ISR_ADRDY) {
	}
	// ADC3
	stm32.ADC3.ISR.SetBits(stm32.ADC_ISR_ADRDY) // Clear ADRDY by writing 1
	stm32.ADC3.CR.SetBits(stm32.ADC_CR_ADEN)
	for !stm32.ADC3.ISR.HasBits(stm32.ADC_ISR_ADRDY) {
	}

	// 7. Configure resolution (16-bit)
	// RES[2:0] is at bits 4:2 in CFGR. 000: 16-bit.
	stm32.ADC1.CFGR.ReplaceBits(0x0, 0x7, 2)
	stm32.ADC3.CFGR.ReplaceBits(0x0, 0x7, 2)
}

// Configure configures an ADC pin to be able to read analog data.
func (a ADC) Configure(config ADCConfig) {
	a.Pin.Configure(PinConfig{Mode: PinInputAnalog})

	// Set sampling time.
	// H7 has SMPR1 (channels 0-9) and SMPR2 (channels 10-19).
	// Each channel has 3 bits.
	ch := a.getChannel()
	adc, _ := a.getPeripheral()
	const smpVal = 0x2 // 8.5 cycles
	if ch <= 9 {
		adc.SMPR1.ReplaceBits(uint32(smpVal), 0x7, uint8(ch)*3)
	} else {
		adc.SMPR2.ReplaceBits(uint32(smpVal), 0x7, uint8(ch-10)*3)
	}
}

// Get returns the current value of an ADC pin in the range 0..0xffff.
func (a ADC) Get() uint16 {
	ch := uint32(a.getChannel())
	adc, ok := a.getPeripheral()
	if !ok {
		return 0
	}

	// Select channel (PCSEL register)
	// Refer to RM0433 §25.4.12: Only one PCSELx bit must be set at a time.
	adc.PCSEL.Set(1 << ch)

	// Set rank 1 to channel
	// SQ1[4:0] at bits 10:6. L[3:0] at bits 3:0.
	adc.SQR1.ReplaceBits(ch, 0x1F, 6)
	adc.SQR1.ReplaceBits(0x0, 0xF, 0) // L=0 (1 conversion)

	// Start conversion
	adc.CR.SetBits(stm32.ADC_CR_ADSTART)

	// Wait for end of conversion
	for !adc.ISR.HasBits(stm32.ADC_ISR_EOC) {
	}

	// Read 16-bit result
	result := uint16(adc.DR.Get())

	// Clear EOC
	adc.ISR.SetBits(stm32.ADC_ISR_EOC)

	// Deselect channel
	adc.PCSEL.Set(0)

	return result
}

func (a ADC) getPeripheral() (*stm32.ADC_Type, bool) {
	switch a.Pin {
	case PF3, PF4, PF5, PF6, PF7, PF8, PF9, PF10:
		return stm32.ADC3, true
	default:
		// Assume ADC1 for PA/PB/PC pins
		return stm32.ADC1, true
	}
}

// getChannel returns the ADC channel number for a given GPIO pin.
// Mapping for STM32H743 per RM0433 and DS12110.
func (a ADC) getChannel() uint8 {
	switch a.Pin {
	case PA0:
		return 16
	case PA1:
		return 17
	case PA2:
		return 14
	case PA3:
		return 15
	case PA4:
		return 18
	case PA5:
		return 19
	case PA6:
		return 3
	case PA7:
		return 7
	case PB0:
		return 9
	case PB1:
		return 5
	case PC0:
		return 10
	case PC1:
		return 11
	case PC2:
		return 12
	case PC3:
		return 13
	case PC4:
		return 4
	case PC5:
		return 8
	case PF3:
		return 5
	case PF4:
		return 9
	case PF5:
		return 4
	case PF6:
		return 8
	case PF7:
		return 3
	case PF8:
		return 7
	case PF9:
		return 2
	case PF10:
		return 6
	}
	return 0
}
