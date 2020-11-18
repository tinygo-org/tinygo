// +build max32620

// Peripheral abstraction layer for the max32620.
//
// Datasheet:
// https://datasheets.maximintegrated.com/en/ds/MAX32620-MAX32621.pdf
//
package machine

func CPUFrequency() uint32 {
	return 96000000
}

// Hardware pins (Primary / Secondary functions)
const (
	P0_0 Pin = 0  // UART0A_RX UART0B_TX
	P0_1 Pin = 1  // UART0A_TX UART0B_RX
	P0_2 Pin = 2  // UART0A_CTS UART0B_RTS
	P0_3 Pin = 3  // UART0A_RTS UART0B_CTS
	P0_4 Pin = 4  // SPIM0_SCK
	P0_5 Pin = 5  // SPIM0_MOSI/SDIO0
	P0_6 Pin = 6  // SPIM0_MISO/SDIO1
	P0_7 Pin = 7  // SPIM0_SS0
	P1_0 Pin = 8  // SPIM1_SCK SPIX_SCK
	P1_1 Pin = 9  // SPIM1_MOSI/SDIO0 SPIX_SDIO0
	P1_2 Pin = 10 // SPIM1_MISO/SDIO1 SPIX_SDIO1
	P1_3 Pin = 11 // SPIM1_SS0 SPIX_SS
	P1_4 Pin = 12 // SPIM1_SDIO2 SPIX_SDIO2
	P1_5 Pin = 13 // SPIM1_SDIO3 SPIX_SDIO3
	P1_6 Pin = 14 // I2CM0/SA_SDA
	P1_7 Pin = 15 // I2CM0/SA_SCL
	P2_0 Pin = 16 // UART1A_RX UART1B_TX
	P2_1 Pin = 17 // UART1A_TX UART1B_RX
	P2_2 Pin = 18 // UART1A_CTS UART1B_RTS
	P2_3 Pin = 19 // UART1A_RTS UART1B_CTS
	P2_4 Pin = 20 // SPIM2A_SCK
	P2_5 Pin = 21 // SPIM2A_MOSI/SDIO0
	P2_6 Pin = 22 // SPIM2A_MISO/SDIO1
	P2_7 Pin = 23 // SPIM2A_SS0
	P3_0 Pin = 24 // UART2A_RX UART2B_TX
	P3_1 Pin = 25 // UART2A_TX UART2B_RX
	P3_2 Pin = 26 // UART2A_CTS UART2B_RTS
	P3_3 Pin = 27 // UART2A_RTS UART2B_CTS
	P3_4 Pin = 28 // I2CM1/SB_SDA SPIM2A_SS1
	P3_5 Pin = 29 // I2CM1/SB_SCL SPIM2A_SS2
	P3_6 Pin = 30 // SPIM1_SS1 SPIX_SS1
	P3_7 Pin = 31 // SPIM1_SS2 SPIX_SS2
	P4_0 Pin = 32 // OWM_I/O SPIM2A_SR0
	P4_1 Pin = 33 // OWM_PUPEN SPIM2A_SR1
	P4_2 Pin = 34 // SPIM0_SDIO2
	P4_3 Pin = 35 // SPIM0_SDIO3
	P4_4 Pin = 36 // SPIM0_SS1
	P4_5 Pin = 37 // SPIM0_SS2
	P4_6 Pin = 38 // SPIM0_SS3
	P4_7 Pin = 39 // SPIM0_SS4
	P5_0 Pin = 40 // Reserved SPIM2B_SCK
	P5_1 Pin = 41 // Reserved SPIM2B_MOSI/SDIO0
	P5_2 Pin = 42 // Reserved SPIM2B_MISO/SDIO1
	P5_3 Pin = 43 // Reserved SPIM2B_SS0 UART3A_RX UART3B_TX
	P5_4 Pin = 44 // Reserved SPIM2B_SDIO2 UART3A_TX UART3B_RX
	P5_5 Pin = 45 // Reserved SPIM2B_SDIO3 UART3A_CTS UART3B_RTS
	P5_6 Pin = 46 // Reserved SPIM2B_SR UART3A_RTS UART3B_CTS
	P5_7 Pin = 47 // I2CM2/SC_SDA SPIM2B_SS1
	P6_0 Pin = 48 // I2CM2/SC_SCL SPIM2B_SS2

	AIN_0 Pin = 49
	AIN_1 Pin = 50
	AIN_2 Pin = 51
	AIN_3 Pin = 52
)

type PinMode uint8

func (p Pin) Set(high bool) {
}

func (p Pin) Get() bool {
	return false
}
