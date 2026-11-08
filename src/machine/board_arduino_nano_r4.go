//go:build arduino_nano_r4

// Pin definitions are derived from ArduinoCore-renesas 1.6.0 sources:
//
// - variants/NANOR4/variant.cpp
// - variants/NANOR4/pinmux.inc
// - variants/NANOR4/pins_arduino.h

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
	A6 Pin = P0_04
	A7 Pin = P0_03
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
	LED = P2_04

	LED_RED   Pin = P4_09
	LED_GREEN Pin = P4_10
	LED_BLUE  Pin = P4_11
)

// ENABLE_3V3_PIN shuts down the 3V3 regulator when driven low.
const ENABLE_3V3_PIN Pin = P5_00

const (
	UART_TX_PIN Pin = UART1_TX_PIN
	UART_RX_PIN Pin = UART1_RX_PIN

	UART1_TX_PIN Pin = D1
	UART1_RX_PIN Pin = D0
)

const (
	SDA_PIN Pin = A4
	SCL_PIN Pin = A5

	// Qwiic connector
	SDA1_PIN Pin = P4_01
	SCL1_PIN Pin = P4_00
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

const AVCC_MEASURE_PIN Pin = A6

const (
	ADC_CH0_PIN  Pin = A1
	ADC_CH1_PIN  Pin = A2
	ADC_CH2_PIN  Pin = A3
	ADC_CH3_PIN  Pin = A7
	ADC_CH4_PIN  Pin = A6
	ADC_CH9_PIN  Pin = A0
	ADC_CH19_PIN Pin = D4
	ADC_CH20_PIN Pin = D5
	ADC_CH21_PIN Pin = A4
	ADC_CH22_PIN Pin = A5
)

// UART1 is on the D0/D1 header pins.
var (
	UART1       = &sci2UART
	DefaultUART = UART1
)

func init() {
	UART1.txPin, UART1.rxPin = UART1_TX_PIN, UART1_RX_PIN
}

// USB identifiers
const (
	usb_STRING_PRODUCT      = "Nano R4"
	usb_STRING_MANUFACTURER = "Arduino"
)

var (
	usb_VID uint16 = 0x2341
	usb_PID uint16 = 0x0074
)
