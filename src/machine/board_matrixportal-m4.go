//go:build matrixportal_m4
// +build matrixportal_m4

package machine

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xF01669EF

// Digital pins
const (
	//    Pin   // Function          SERCOM  PWM  Interrupt
	//    ----  // ----------------  ------  ---  ---------
	D0  = PA01  // UART RX            1[1]   PWM  EXTI1
	D1  = PA00  // UART TX            1[0]   PWM  EXTI0
	D2  = PB22  // Button "Up"                    EXTI6
	D3  = PB23  // Button "Down"                  EXTI7
	D4  = PA23  // NeoPixel                       EXTI7
	D5  = PB31  // I2C SDA            5[1]        EXTI15
	D6  = PB30  // I2C SCL            5[0]        EXTI14
	D7  = PB00  // HUB75 R1                       EXTI0
	D8  = PB01  // HUB75 G1                       EXTI1
	D9  = PB02  // HUB75 B1                       EXTI2
	D10 = PB03  // HUB75 R2                       EXTI3
	D11 = PB04  // HUB75 G2                       EXTI4
	D12 = PB05  // HUB75 B2                       EXTI5
	D13 = PA14  // LED                       PWM  EXTI14
	D14 = PB06  // HUB75 CLK                      EXTI6
	D15 = PB14  // HUB75 LAT                      EXTI14
	D16 = PB12  // HUB75 OE                       EXTI12
	D17 = PB07  // HUB75 ADDR A                   EXTI7
	D18 = PB08  // HUB75 ADDR B                   EXTI8
	D19 = PB09  // HUB75 ADDR C                   EXTI9
	D20 = PB15  // HUB75 ADDR D                   EXTI15
	D21 = PB13  // HUB75 ADDR E                   EXTI13
	D22 = PA02  // ADC (A0)                       EXTI2
	D23 = PA05  // ADC (A1)                       EXTI5
	D24 = PA04  // ADC (A2)                  PWM  EXTI4
	D25 = PA06  // ADC (A3)                  PWM  EXTI6
	D26 = PA07  // ADC (A4)                       EXTI7
	D27 = PA12  // ESP32 UART RX      4[1]   PWM  EXTI12
	D28 = PA13  // ESP32 UART TX      4[0]   PWM  EXTI13
	D29 = PA20  // ESP32 GPIO0               PWM  EXTI4
	D30 = PA21  // ESP32 Reset               PWM  EXTI5
	D31 = PA22  // ESP32 Busy                PWM  EXTI6
	D32 = PA18  // ESP32 RTS                 PWM  EXTI2
	D33 = PB17  // ESP32 SPI CS              PWM  EXTI1
	D34 = PA16  // ESP32 SPI SCK      3[1]   PWM  EXTI0
	D35 = PA17  // ESP32 SPI SDI      3[0]   PWM  EXTI1
	D36 = PA19  // ESP32 SPI SDO      1[3]   PWM  EXTI3
	D37 = NoPin // USB Host enable
	D38 = PA24  // USB DM
	D39 = PA25  // USB DP
	D40 = PA03  // DAC/VREFP
	D41 = PB10  // Flash QSPI SCK
	D42 = PB11  // Flash QSPI CS
	D43 = PA08  // Flash QSPI I00
	D44 = PA09  // Flash QSPI IO1
	D45 = PA10  // Flash QSPI IO2
	D46 = PA11  // Flash QSPI IO3
	D47 = PA27  // LIS3DH IRQ                     EXTI11
	D48 = PA05  // SPI SCK            0[1]        EXTI5
	D49 = PA04  // SPI SDO            0[0]   PWM  EXTI4
	D50 = PA07  // SPI SDI            0[3]        EXTI7
)

// Analog pins
const (
	A0 = PA02 // ADC Channel 0
	A1 = PA05 // ADC Channel 5
	A2 = PA04 // ADC Channel 4
	A3 = PA06 // ADC Channel 6
	A4 = PA07 // ADC Channel 7
)

// LED pins
const (
	LED      = D13
	NEOPIXEL = D4
	WS2812   = D4
)

// Button pins
const (
	BUTTON_UP   = D2
	BUTTON_DOWN = D3
)

// UART pins
const (
	UART1_RX_PIN = D0 // SERCOM1[1]
	UART1_TX_PIN = D1 // SERCOM1[0]

	UART2_RX_PIN = D27 // SERCOM4[1] (ESP32 RX)
	UART2_TX_PIN = D28 // SERCOM4[0] (ESP32 TX)

	UART_RX_PIN = UART1_RX_PIN
	UART_TX_PIN = UART1_TX_PIN
)

// UART on the MatrixPortal M4
var (
	UART1 = &sercomUSART1
	UART2 = &sercomUSART4

	DefaultUART = UART1
)

// SPI pins
const (
	SPI0_SCK_PIN = D34 // SERCOM3[1] (ESP32 SCK)
	SPI0_SDO_PIN = D36 // SERCOM1[3] (ESP32 SDO)
	SPI0_SDI_PIN = D35 // SERCOM3[0] (ESP32 SDI)

	SPI1_SCK_PIN = D48 // SERCOM0[1]
	SPI1_SDO_PIN = D49 // SERCOM0[0]
	SPI1_SDI_PIN = D50 // SERCOM0[3]

	SPI_SCK_PIN = SPI0_SCK_PIN
	SPI_SDO_PIN = SPI0_SDO_PIN
	SPI_SDI_PIN = SPI0_SDI_PIN
)

// I2C pins
const (
	I2C0_SDA_PIN = D5 // SERCOM5[1]
	I2C0_SCL_PIN = D6 // SERCOM5[0]

	I2C_SDA_PIN = I2C0_SDA_PIN
	I2C_SCL_PIN = I2C0_SCL_PIN

	SDA_PIN = I2C_SDA_PIN // awkward naming required by machine_atsamd51.go
	SCL_PIN = I2C_SCL_PIN //
)

// I2C on the MatrixPortal M4
var (
	I2C0 = sercomI2CM5
)

// ESP32 pins
const (
	NINA_ACK    = D31
	NINA_GPIO0  = D29
	NINA_RESETN = D30

	NINA_RX  = UART2_RX_PIN
	NINA_TX  = UART2_TX_PIN
	NINA_RTS = D32

	NINA_CS  = D33
	NINA_SDO = SPI0_SDO_PIN
	NINA_SDI = SPI0_SDI_PIN
	NINA_SCK = SPI0_SCK_PIN
)

// SPI on the MatrixPortal M4
var (
	SPI0     = sercomSPIM3 // BUG: SDO on SERCOM1!
	NINA_SPI = SPI0

	SPI1 = sercomSPIM0
)

// HUB75 pins
const (
	HUB75_R1 = D7
	HUB75_G1 = D8
	HUB75_B1 = D9
	HUB75_R2 = D10
	HUB75_G2 = D11
	HUB75_B2 = D12

	HUB75_CLK    = D14
	HUB75_LAT    = D15
	HUB75_OE     = D16
	HUB75_ADDR_A = D17
	HUB75_ADDR_B = D18
	HUB75_ADDR_C = D19
	HUB75_ADDR_D = D20
	HUB75_ADDR_E = D21
)

// USB CDC pins (UART0)
const (
	USBCDC_DM_PIN = D38
	USBCDC_DP_PIN = D39

	UART0_RX_PIN = USBCDC_DM_PIN
	UART0_TX_PIN = USBCDC_DP_PIN
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Matrix Portal M4"
	usb_STRING_MANUFACTURER = "Adafruit Industries"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x80C9
)
