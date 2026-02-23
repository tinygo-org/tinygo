//go:build stm32g0

package machine

import (
	"device/stm32"
	"unsafe"
)

// ADC sampling time constants for STM32G0
const (
	ADC_SMPR_1_5   = 0x0 // 1.5 ADC clock cycles
	ADC_SMPR_3_5   = 0x1 // 3.5 ADC clock cycles
	ADC_SMPR_7_5   = 0x2 // 7.5 ADC clock cycles
	ADC_SMPR_12_5  = 0x3 // 12.5 ADC clock cycles
	ADC_SMPR_19_5  = 0x4 // 19.5 ADC clock cycles
	ADC_SMPR_39_5  = 0x5 // 39.5 ADC clock cycles
	ADC_SMPR_79_5  = 0x6 // 79.5 ADC clock cycles
	ADC_SMPR_160_5 = 0x7 // 160.5 ADC clock cycles
)

// InitADC initializes the registers needed for ADC.
func InitADC() {
	// Enable ADC clock
	enableAltFuncClock(unsafe.Pointer(stm32.ADC))

	// Ensure ADC is disabled before configuration
	if stm32.ADC.GetCR_ADEN() != 0 {
		// Clear ADEN by setting ADDIS
		stm32.ADC.SetCR_ADDIS(1)
		// Wait for ADC to be disabled
		for stm32.ADC.GetCR_ADEN() != 0 {
		}
	}

	// Enable ADC voltage regulator
	stm32.ADC.SetCR_ADVREGEN(1)

	// Wait for ADC voltage regulator startup time (20us at max)
	// Using simple busy loop - approximately 1280 cycles at 64MHz = 20us
	for i := 0; i < 1280; i++ {
		// nop
	}

	// Configure ADC:
	// - 12-bit resolution (RES = 0b00)
	// - Right alignment (ALIGN = 0)
	// - Single conversion mode (CONT = 0)
	// - Software trigger (EXTEN = 0b00)
	stm32.ADC.CFGR1.Set(0)

	// Set clock mode to synchronous with PCLK/2
	stm32.ADC.SetCFGR2_CKMODE(0x1) // PCLK/2

	// Set sample time to 12.5 cycles for all channels using SMP1
	stm32.ADC.SetSMPR_SMP1(ADC_SMPR_12_5)

	// Calibrate ADC
	stm32.ADC.SetCR_ADCAL(1)
	for stm32.ADC.GetCR_ADCAL() != 0 {
	}

	// Clear ADRDY by writing 1
	stm32.ADC.SetISR_ADRDY(1)

	// Enable ADC
	stm32.ADC.SetCR_ADEN(1)

	// Wait until ADC is ready
	for stm32.ADC.GetISR_ADRDY() == 0 {
	}
}

// Configure configures an ADC pin to be able to read analog data.
func (a ADC) Configure(config ADCConfig) {
	// Configure pin as analog input
	a.Pin.Configure(PinConfig{Mode: PinInputAnalog})

	// Set sampling time based on config
	// Use SMP2 and set SMPSEL bit for this channel to select SMP2
	ch := a.getChannel()
	if ch <= 18 {
		// Select sampling time based on config (using SMP2 for per-channel control)
		// Map microseconds to sample cycles (at ~32MHz ADC clock after /2 prescaler)
		// Each cycle = 1/32MHz = 31.25ns
		var smpTime int
		switch {
		case config.SampleTime == 0:
			smpTime = ADC_SMPR_79_5 // Default to 79.5 cycles for good accuracy
		case config.SampleTime <= 1:
			smpTime = ADC_SMPR_1_5
		case config.SampleTime <= 2:
			smpTime = ADC_SMPR_3_5
		case config.SampleTime <= 3:
			smpTime = ADC_SMPR_7_5
		case config.SampleTime <= 4:
			smpTime = ADC_SMPR_12_5
		case config.SampleTime <= 5:
			smpTime = ADC_SMPR_19_5
		case config.SampleTime <= 10:
			smpTime = ADC_SMPR_39_5
		case config.SampleTime <= 20:
			smpTime = ADC_SMPR_79_5
		default:
			smpTime = ADC_SMPR_160_5
		}
		stm32.ADC.SetSMPR_SMP2(uint32(smpTime))

		// Set SMPSEL bit for this channel to use SMP2
		stm32.ADC.SMPR.SetBits(1 << (8 + ch))
	}
}

// Get returns the current value of a ADC pin in the range 0..0xffff.
func (a ADC) Get() uint16 {
	ch := a.getChannel()

	// Wait until channel configuration is ready if needed
	// (CCRDY indicates when CHSELR changes are applied)
	for stm32.ADC.GetISR_CCRDY() != 0 {
		stm32.ADC.SetISR_CCRDY(1) // Clear by writing 1
	}

	// Select the channel to convert using CHSELR
	// CHSELR uses a bitfield where bit N = 1 enables channel N
	stm32.ADC.CHSELR0.Set(1 << ch)

	// Wait for channel configuration ready
	for stm32.ADC.GetISR_CCRDY() == 0 {
	}
	stm32.ADC.SetISR_CCRDY(1) // Clear flag

	// Start conversion
	stm32.ADC.SetCR_ADSTART(1)

	// Wait for end of conversion
	for stm32.ADC.GetISR_EOC() == 0 {
	}

	// Read the 12-bit result and scale to 16-bit
	result := uint16(stm32.ADC.GetDR_DATA()) << 4

	return result
}

// getChannel returns the ADC channel number for a given pin.
// STM32G0B1 ADC channel mapping:
// PA0-PA7: CH0-CH7
// PB0-PB2: CH8-CH10
// PB10-PB12: CH11-CH13 (some variants)
// PC4-PC5: CH17-CH18 (some variants)
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
	case PB2:
		return 10
	case PB10:
		return 11
	case PB11:
		return 12
	case PB12:
		return 13
	case PC4:
		return 17
	case PC5:
		return 18
	}
	return 0
}
