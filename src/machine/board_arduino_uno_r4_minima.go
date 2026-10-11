//go:build arduino_uno_r4_minima

// Pin definitions are derived from ArduinoCore-renesas 1.6.0 sources:
//
// - variants/MINIMA/variant.cpp
// - variants/MINIMA/pinmux.inc

package machine

const (
	D0  Pin = P3_01
	D1  Pin = P3_02
	D2  Pin = P1_05
	D3  Pin = P1_04
	D4  Pin = P1_03
	D5  Pin = P1_02
	D6  Pin = P1_06
	D7  Pin = P1_07
	D8  Pin = P3_04
	D9  Pin = P3_03
	D10 Pin = P1_12
	D11 Pin = P1_09
	D12 Pin = P1_10
	D13 Pin = P1_11

	D14 Pin = A0
	D15 Pin = A1
)

const (
	A0 Pin = P0_14
	A1 Pin = P0_00
	A2 Pin = P0_01
	A3 Pin = P0_02
	A4 Pin = P1_01
	A5 Pin = P1_00
)

const (
	PWM_D3_PIN  Pin = D3
	PWM_D5_PIN  Pin = D5
	PWM_D6_PIN  Pin = D6
	PWM_D9_PIN  Pin = D9
	PWM_D10_PIN Pin = D10
	PWM_D11_PIN Pin = D11
)

const (
	SWDIO_PIN Pin = P1_08
	SWCLK_PIN Pin = P3_00
)

const (
	LED = D13

	LED_TX Pin = P0_12
	LED_RX Pin = P0_13
)

const (
	UART_TX_PIN Pin = UART1_TX_PIN
	UART_RX_PIN Pin = UART1_RX_PIN

	UART1_TX_PIN    Pin = D1
	UART1_RX_PIN    Pin = D0
	UART_SWD_TX_PIN Pin = P5_01
	UART_SWD_RX_PIN Pin = P5_02
)

const (
	SDA_PIN Pin = A4
	SCL_PIN Pin = A5
)

const (
	SPI_SDO_PIN Pin = D11
	SPI_SDI_PIN Pin = D12
	SPI_SCK_PIN Pin = D13
	SPI_CS_PIN  Pin = D10
)

const (
	CAN_TX_PIN Pin = D4
	CAN_RX_PIN Pin = D5
)

const DAC0_PIN Pin = A0

const AVCC_MEASURE_PIN Pin = P5_00

const (
	ADC_CH0_PIN  Pin = A1
	ADC_CH1_PIN  Pin = A2
	ADC_CH2_PIN  Pin = A3
	ADC_CH9_PIN  Pin = A0
	ADC_CH16_PIN Pin = AVCC_MEASURE_PIN
	ADC_CH17_PIN Pin = UART_SWD_TX_PIN
	ADC_CH18_PIN Pin = UART_SWD_RX_PIN
	ADC_CH19_PIN Pin = D4
	ADC_CH20_PIN Pin = D5
	ADC_CH21_PIN Pin = A4
	ADC_CH22_PIN Pin = A5
)
