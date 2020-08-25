// +build nxp,mkl26z4,teensylc

package machine

const (
	fCPU = 48000000 // core - 48MHz
	fCLK = 16000000 // oscillator - 16MHz
	fPLL = 96000000 // PLL - 96MHz
	fBUS = 24000000 // bus - 24MHz
	fMEM = 24000000 // RAM - 24MHz
)

// CPUFrequency returns the frequency of the ARM core clock (48MHz)
func CPUFrequency() uint32 { return fCPU }

// ClockFrequency returns the frequency of the external oscillator (16MHz)
func ClockFrequency() uint32 { return fCLK }

// LED on the Teensy
const LED = D13

// Digital IO
const (
	D00 = PB16
	D01 = PB17
	D02 = PD00
	D03 = PA12
	D04 = PA13
	D05 = PD07
	D06 = PD04
	D07 = PD02
	D08 = PD03
	D09 = PC03
	D10 = PC04
	D11 = PC06
	D12 = PC07
	D13 = PC05
	D14 = PD01
	D15 = PC00
	D16 = PB00
	D17 = PB01
	D18 = PB03
	D19 = PB02
	D20 = PD05
	D21 = PD06
	D22 = PC01
	D23 = PC02
	D24 = PA05
	D25 = PB19
	D26 = PE01
)

// Analog IO
const (
	A00 = D14
	A01 = D15
	A02 = D16
	A03 = D17
	A04 = D18
	A05 = D19
	A06 = D20
	A07 = D21
	A08 = D22
	A09 = D23
	A10 = D24
	A11 = D25
	A12 = D26
)

var (
	UART0RX0 = PinMux{D00, 3}
	UART0RX1 = PinMux{D21, 3}
	UART0RX2 = PinMux{D03, 2}
	UART0RX3 = PinMux{D25, 4}
	UART0TX0 = PinMux{D01, 3}
	UART0TX1 = PinMux{D05, 3}
	UART0TX2 = PinMux{D04, 2}
	UART0TX3 = PinMux{D24, 4}

	UART1RX0 = PinMux{D09, 3}
	UART1TX0 = PinMux{D10, 3}

	UART2RX0 = PinMux{D07, 3}
	UART2RX1 = PinMux{D06, 3}
	UART2TX0 = PinMux{D08, 3}
	UART2TX1 = PinMux{D20, 3}
)
