//go:build arduino_mega1280

package machine

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 16000000
}

const (
	A0  Pin = PF0
	A1  Pin = PF1
	A2  Pin = PF2
	A3  Pin = PF3
	A4  Pin = PF4
	A5  Pin = PF5
	A6  Pin = PF6
	A7  Pin = PF7
	A8  Pin = PK0
	A9  Pin = PK1
	A10 Pin = PK2
	A11 Pin = PK3
	A12 Pin = PK4
	A13 Pin = PK5
	A14 Pin = PK6
	A15 Pin = PK7

	// Analog Input
	ADC0  Pin = PF0
	ADC1  Pin = PF1
	ADC2  Pin = PF2
	ADC3  Pin = PF3
	ADC4  Pin = PF4
	ADC5  Pin = PF5
	ADC6  Pin = PF6
	ADC7  Pin = PF7
	ADC8  Pin = PK0
	ADC9  Pin = PK1
	ADC10 Pin = PK2
	ADC11 Pin = PK3
	ADC12 Pin = PK4
	ADC13 Pin = PK5
	ADC14 Pin = PK6
	ADC15 Pin = PK7

	// Digital pins
	D0  Pin = PE0
	D1  Pin = PE1
	D2  Pin = PE4
	D3  Pin = PE5
	D4  Pin = PG5
	D5  Pin = PE3
	D6  Pin = PH3
	D7  Pin = PH4
	D8  Pin = PH5
	D9  Pin = PH6
	D10 Pin = PB4
	D11 Pin = PB5
	D12 Pin = PB6
	D13 Pin = PB7
	D14 Pin = PJ1
	D15 Pin = PJ0
	D16 Pin = PH1
	D17 Pin = PH0
	D18 Pin = PD3
	D19 Pin = PD2
	D20 Pin = PD1
	D21 Pin = PD0
	D22 Pin = PA0
	D23 Pin = PA1
	D24 Pin = PA2
	D25 Pin = PA3
	D26 Pin = PA4
	D27 Pin = PA5
	D28 Pin = PA6
	D29 Pin = PA7
	D30 Pin = PC7
	D31 Pin = PC6
	D32 Pin = PC5
	D33 Pin = PC4
	D34 Pin = PC3
	D35 Pin = PC2
	D36 Pin = PC1
	D37 Pin = PC0
	D38 Pin = PD7
	D39 Pin = PG2
	D40 Pin = PG1
	D41 Pin = PG0
	D42 Pin = PL7
	D43 Pin = PL6
	D44 Pin = PL5
	D45 Pin = PL4
	D46 Pin = PL3
	D47 Pin = PL2
	D48 Pin = PL1
	D49 Pin = PL0
	D50 Pin = PB3
	D51 Pin = PB2
	D52 Pin = PB1
	D53 Pin = PB0

	AREF Pin = NoPin
	LED  Pin = PB7
)
