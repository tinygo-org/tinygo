//go:build esp32_dev

package machine

// Reference 1:
// https://docs.espressif.com/projects/esp-idf/en/stable/esp32/_images/esp32-devkitC-v4-pinout.png

// Reference 2:
// This is the board I have, but it's slightly different, so for now I'm going to avoid pins that
// are missing/unlabeled
// https://ae01.alicdn.com/kf/Sa74c2ababf3640c5b77022481509c13bi.jpg

// Silkscreen Pins
const (
	VP  = GPIO36
	VN  = GPIO39
	CMD = GPIO11
	TX  = GPIO1
	RX  = GPIO3
	CLK = GPIO6
)

// Analog-to-Digital
const (
	// ADC(Bus)_(Channel)
	ADC1_0 = GPIO36
	//ADC1_1
	//ADC1_2
	ADC1_3 = GPIO39
	ADC1_4 = GPIO32
	ADC1_5 = GPIO33
	ADC1_6 = GPIO34
	ADC1_7 = GPIO35
	ADC1_8 = GPIO25
	//ADC1_9
	ADC2_0 = GPIO4
	ADC2_1 = GPIO0
	ADC2_2 = GPIO2
	ADC2_3 = GPIO15
	ADC2_4 = GPIO13
	ADC2_5 = GPIO12
	ADC2_6 = GPIO14
	ADC2_7 = GPIO27
	// ADC2_8
	ADC2_9 = GPIO26
)

// Digital-to-Analog
const (
	DAC_1 = GPIO25
	DAC_2 = GPIO26
)

// UART
const (
	UART0_TX_PIN = GPIO1
	UART0_RX_PIN = GPIO3
	// Unsure of these as don't appear on Ref 1 but do on Ref 2
	// UART1_TX_PIN = GPIO17
	// UART1_RX_PIN = GPIO16

	UART_TX_PIN = UART0_TX_PIN
	UART_RX_PIN = UART0_RX_PIN
)

// I2C
const (
	SDA_PIN = GPIO21
	SCL_PIN = GPIO22
)

// SPI
const (
	// Hardware
	SPI0_SCK_PIN = GPIO14
	SPI0_SDO_PIN = GPIO13
	SPI0_SDI_PIN = GPIO12
	// Virtual
	SPI1_SCK_PIN = GPIO18
	SPI1_SDO_PIN = GPIO23
	SPI1_SDI_PIN = GPIO19
)
