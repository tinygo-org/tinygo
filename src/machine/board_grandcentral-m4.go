//go:build grandcentral_m4

package machine

// Digital pins
const (
	//  = Pin     Alt. Function        SERCOM   PWM Timer   Interrupt
	//   ------  -------------------- -------- ----------- -----------
	D0  = PB25 // UART1 RX              0[1]                 EXTI9
	D1  = PB24 // UART1 TX              0[0]                 EXTI8
	D2  = PC18 //                                TCC0[2]     EXTI2
	D3  = PC19 //                                TCC0[3]     EXTI3
	D4  = PC20 //                                TCC0[4]     EXTI4
	D5  = PC21 //                                TCC0[5]     EXTI5
	D6  = PD20 //                                TCC1[0]     EXTI10
	D7  = PD21 //                                TCC1[1]     EXTI11
	D8  = PB18 //                                TCC1[0]     EXTI2
	D9  = PB02 //                                TC6[0]      EXTI3
	D10 = PB22 //                                TC7[0]      EXTI6
	D11 = PB23 //                                            EXTI7
	D12 = PB00 //                                TC7[0]      EXTI0
	D13 = PB01 // On-board LED                   TC7[1]      EXTI1
	D14 = PB16 // UART4 TX, I2S0 SCK    5[0]     TC6[0]      EXTI0
	D15 = PB17 // UART4 RX, I2S0 MCK    5[1]                 EXTI1
	D16 = PC22 // UART3 TX              1[0]                 EXTI6
	D17 = PC23 // UART3 RX              1[1]                 EXTI6
	D18 = PB12 // UART2 TX              4[0]     TCC3[0]     EXTI12
	D19 = PB13 // UART2 RX              4[1]     TCC3[1]     EXTI13
	D20 = PB20 // I2C0 SDA              3[0]                 EXTI4
	D21 = PB21 // I2C0 SCL              3[1]                 EXTI5
	D22 = PD12 //                                            EXTI7
	D23 = PA15 //                                TCC2[1]     EXTI15
	D24 = PC17 // I2C1 SCL              6[1]     TCC0[1]     EXTI1
	D25 = PC16 // I2C1 SDA              6[0]     TCC0[0]     EXTI0
	D26 = PA12 // PCC DEN1                       TC2[0]      EXTI12
	D27 = PA13 // PCC DEN2                       TC2[1]      EXTI13
	D28 = PA14 // PCC CLK                        TCC2[0]     EXTI14
	D29 = PB19 // PCC XCLK                                   EXTI3
	D30 = PA23 // PCC D7                         TC4[1]      EXTI7
	D31 = PA22 // PCC D6, I2S0 SDI               TC4[0]      EXTI6
	D32 = PA21 // PCC D5, I2S0 SDO                           EXTI5
	D33 = PA20 // PCC D4, I2S0 FS                            EXTI4
	D34 = PA19 // PCC D3                         TC3[1]      EXTI3
	D35 = PA18 // PCC D2                         TC3[0]      EXTI2
	D36 = PA17 // PCC D1                                     EXTI1
	D37 = PA16 // PCC D0                                     EXTI0
	D38 = PB15 // PCC D9                         TCC4[1]     EXTI15
	D39 = PB14 // PCC D8                         TCC4[0]     EXTI14
	D40 = PC13 // PCC D11                                    EXTI13
	D41 = PC12 // PCC D10                                    EXTI12
	D42 = PC15 // PCC D13                                    EXTI15
	D43 = PC14 // PCC D12                                    EXTI14
	D44 = PC11 //                                            EXTI11
	D45 = PC10 //                                            EXTI10
	D46 = PC06 //                                            EXTI6
	D47 = PC07 //                                            EXTI5
	D48 = PC04 //                                            EXTI4
	D49 = PC05 //                                            EXTI5
	D50 = PD11 // SPI0 SDI              7[3]                 EXTI11
	D51 = PD08 // SPI0 SDO              7[0]                 EXTI8
	D52 = PD09 // SPI0 SCK              7[1]                 EXTI9
	D53 = PD10 // SPI0 CS                                    EXTI10
	D54 = PB05 // ADC1 (A8)                                  EXTI5
	D55 = PB06 // ADC1 (A9)                                  EXTI6
	D56 = PB07 // ADC1 (A10)                                 EXTI7
	D57 = PB08 // ADC1 (A11)                                 EXTI8
	D58 = PB09 // ADC1 (A12)                                 EXTI9
	D59 = PA04 // ADC0 (A13)                     TC0[0]      EXTI4
	D60 = PA06 // ADC0 (A14)                     TC1[0]      EXTI6
	D61 = PA07 // ADC0 (A15)                     TC1[1]      EXTI7
	D62 = PB20 // I2C0 SDA              3[0]     TCC1[2]     EXTI4
	D63 = PB21 // I2C0 SCL              3[1]     TCC1[3]     EXTI5
	D64 = PD11 // SPI0 SDI              7[3]                 EXTI6
	D65 = PD08 // SPI0 SDO              7[0]                 EXTI3
	D66 = PD09 // SPI0 SCK              7[1]                 EXTI4
	D67 = PA02 // ADC0 (A0), DAC0                            EXTI2
	D68 = PA05 // ADC0 (A1), DAC1                            EXTI5
	D69 = PB03 // ADC0 (A2)                      TC6[1]      EXTI3
	D70 = PC00 // ADC1 (A3)                                  EXTI0
	D71 = PC01 // ADC1 (A4)                                  EXTI1
	D72 = PC02 // ADC1 (A5)                                  EXTI2
	D73 = PC03 // ADC1 (A6)                                  EXTI3
	D74 = PB04 // ADC1 (A7)                                  EXTI4
	D75 = PC31 // UART RX LED
	D76 = PC30 // UART TX LED
	D77 = PA27 // USB HOST EN
	D78 = PA24 // USB DM                                     EXTI8
	D79 = PA25 // USB DP                                     EXTI9
	D80 = PB29 // SD/SPI1 SDI           2[3]
	D81 = PB27 // SD/SPI1 SCK           2[1]
	D82 = PB26 // SD/SPI1 SDO           2[0]
	D83 = PB28 // SD/SPI1 CS
	D84 = PA03 // AREF                                       EXTI3
	D85 = PA02 // DAC0                                       EXTI2
	D86 = PA05 // DAC1                                       EXTI5
	D87 = PB01 // On-board LED (D13)             TC7[1]      EXTI1
	D88 = PC24 // On-board NeoPixel
	D89 = PB10 // QSPI SCK                                   EXTI10
	D90 = PB11 // QSPI CS                                    EXTI11
	D91 = PA08 // QSPI ID0                                   EXTI(NMI)
	D92 = PA09 // QSPI ID1                                   EXTI9
	D93 = PA10 // QSPI ID2                                   EXTI10
	D94 = PA11 // QSPI ID3                                   EXTI11
	D95 = PB31 // SD Detect                                  EXTI15
	D96 = PB30 // SWO                                        EXTI14
)

// Analog pins
const (
	A0  = D67 // (PA02) ADC0 ch. 0,
	A1  = D68 // (PA05) ADC0 ch. 5,
	A2  = D69 // (PB03) ADC0 ch. 15
	A3  = D70 // (PC00) ADC1 ch. 10
	A4  = D71 // (PC01) ADC1 ch. 11
	A5  = D72 // (PC02) ADC1 ch. 4
	A6  = D73 // (PC03) ADC1 ch. 5
	A7  = D74 // (PB04) ADC1 ch. 6
	A8  = D54 // (PB05) ADC1 ch. 7
	A9  = D55 // (PB06) ADC1 ch. 8
	A10 = D56 // (PB07) ADC1 ch. 9
	A11 = D57 // (PB08) ADC1 ch. 0
	A12 = D58 // (PB09) ADC1 ch. 1
	A13 = D59 // (PA04) ADC0 ch. 4
	A14 = D60 // (PA06) ADC0 ch. 6
	A15 = D61 // (PA07) ADC0 ch. 7

	AREF = D84 // (PA03)
)

// LED pins
const (
	LED_PIN         = D13 // (PB01), also on D87
	UART_RX_LED_PIN = D75 // (PC31)
	UART_TX_LED_PIN = D76 // (PC30)
	NEOPIXEL_PIN    = D88 // (PC24)

	// aliases used by examples and drivers
	LED      = LED_PIN
	LED_RX   = UART_RX_LED_PIN
	LED_TX   = UART_TX_LED_PIN
	NEOPIXEL = NEOPIXEL_PIN
	WS2812   = NEOPIXEL_PIN
)

// UART pins
const (
	UART1_RX_PIN = D0 // (PB25)
	UART1_TX_PIN = D1 // (PB24)

	UART2_RX_PIN = D19 // (PB13)
	UART2_TX_PIN = D18 // (PB12)

	UART3_RX_PIN = D17 // (PC23)
	UART3_TX_PIN = D16 // (PC22)

	UART4_RX_PIN = D15 // (PB17)
	UART4_TX_PIN = D14 // (PB16)

	UART_RX_PIN = UART1_RX_PIN // default pins
	UART_TX_PIN = UART1_TX_PIN //
)

// UART on the Grand Central M4
var (
	UART1 = &sercomUSART0
	UART2 = &sercomUSART4
	UART3 = &sercomUSART1
	UART4 = &sercomUSART5

	DefaultUART = UART1
)

// SPI pins
const (
	SPI0_SCK_PIN = D66 // (PD09), also on D52
	SPI0_SDO_PIN = D65 // (PD08), also on D51
	SPI0_SDI_PIN = D64 // (PD11), also on D50
	SPI0_CS_PIN  = D53 // (PD10)

	SPI1_SCK_PIN = D81 // (PB27)
	SPI1_SDO_PIN = D82 // (PB26)
	SPI1_SDI_PIN = D80 // (PB29)

	SPI_SCK_PIN = SPI0_SCK_PIN // default pins
	SPI_SDO_PIN = SPI0_SDO_PIN //
	SPI_SDI_PIN = SPI0_SDI_PIN //
	SPI_CS_PIN  = SPI0_CS_PIN  //
)

// SPI on the Grand Central M4
var (
	SPI0 = sercomSPIM7
	SPI1 = sercomSPIM2 // SD card
)

// I2C pins
const (
	I2C0_SDA_PIN = D62 // (PB20), also on D20
	I2C0_SCL_PIN = D63 // (PB21), also on D21

	I2C1_SDA_PIN = D25 // (PC16)
	I2C1_SCL_PIN = D24 // (PC17)

	I2C_SDA_PIN = I2C0_SDA_PIN // default pins
	I2C_SCL_PIN = I2C0_SCL_PIN //

	SDA_PIN = I2C_SDA_PIN // unconventional pin names
	SCL_PIN = I2C_SCL_PIN //  (required by machine_atsamd51.go)
)

// I2C on the Grand Central M4
var (
	I2C0 = sercomI2CM3
	I2C1 = sercomI2CM6
)

// I2S pins
const (
	I2S0_SCK_PIN = D14 // (PB16)
	I2S0_MCK_PIN = D15 // (PB17)
	I2S0_FS_PIN  = D33 // (PA20)
	I2S0_SDO_PIN = D32 // (PA21)
	I2S0_SDI_PIN = D31 // (PA22)

	I2S_SCK_PIN = I2S0_SCK_PIN // default pins
	I2S_WS_PIN  = I2S0_FS_PIN  //
	I2S_SD_PIN  = I2S0_SDO_PIN //
)

// SD card pins
const (
	SD0_SCK_PIN = D81 // (PB27)
	SD0_SDO_PIN = D82 // (PB26)
	SD0_SDI_PIN = D80 // (PB29)
	SD0_CS_PIN  = D83 // (PB28)
	SD0_DET_PIN = D95 // (PB31)

	SDCARD_SCK_PIN = SD0_SCK_PIN // default pins
	SDCARD_SDO_PIN = SD0_SDO_PIN //
	SDCARD_SDI_PIN = SD0_SDI_PIN //
	SDCARD_CS_PIN  = SD0_CS_PIN  //
	SDCARD_DET_PIN = SD0_DET_PIN //
)

// Other peripheral constants
const (
	resetMagicValue = 0xF01669EF // Used to reset into bootloader
)

// USB CDC pins
const (
	USBCDC_HOSTEN_PIN = D77 // (PA27) host enable
	USBCDC_DM_PIN     = D78 // (PA24) D-
	USBCDC_DP_PIN     = D79 // (PA25) D+
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Adafruit Grand Central M4"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x8031
)
