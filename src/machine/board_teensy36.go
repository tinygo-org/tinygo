// +build nxp,mk66f18,teensy36

package machine

const (
	fCPU = 180000000 // core - 180MHz
	fCLK = 16000000  // oscillator - 16MHz
	fBUS = 60000000  // bus - 60MHz
)

// CPUFrequency returns the frequency of the ARM core clock (180MHz)
func CPUFrequency() uint32 { return fCPU }

// ClockFrequency returns the frequency of the external oscillator (16MHz)
func ClockFrequency() uint32 { return fCLK }

// LED on the Teensy
const LED = PC05

// digital IO
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
	D24 = PE26
	D25 = PA05
	D26 = PA14
	D27 = PA15
	D28 = PA16
	D29 = PB18
	D30 = PB19
	D31 = PB10
	D32 = PB11
	D33 = PE24
	D34 = PE25
	D35 = PC08
	D36 = PC09
	D37 = PC10
	D38 = PC11
	D39 = PA17
	D40 = PA28
	D41 = PA29
	D42 = PA26
	D43 = PB20
	D44 = PB22
	D45 = PB23
	D46 = PB21
	D47 = PD08
	D48 = PD09
	D49 = PB04
	D50 = PB05
	D51 = PD14
	D52 = PD13
	D53 = PD12
	D54 = PD15
	D55 = PD11
	D56 = PE10
	D57 = PE11
	D58 = PE00
	D59 = PE01
	D60 = PE02
	D61 = PE03
	D62 = PE04
	D63 = PE05
)

var (
	UART0RX0 = PinMux{D00, 3}
	UART0RX1 = PinMux{D21, 3}
	UART0RX2 = PinMux{D27, 3}
	UART0TX0 = PinMux{D01, 3}
	UART0TX1 = PinMux{D05, 3}
	UART0TX2 = PinMux{D26, 3}

	UART1RX0 = PinMux{D09, 3}
	UART1RX1 = PinMux{D59, 3}
	UART1TX0 = PinMux{D10, 3}
	UART1TX1 = PinMux{D58, 3}

	UART2RX0 = PinMux{D07, 3}
	UART2TX0 = PinMux{D08, 3}

	UART3RX0 = PinMux{D31, 3}
	UART3RX1 = PinMux{D63, 3}
	UART3TX0 = PinMux{D32, 3}
	UART3TX1 = PinMux{D62, 3}

	UART4RX0 = PinMux{D34, 3}
	UART4TX0 = PinMux{D33, 3}
)
