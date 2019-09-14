// +build teensy36

package machine

// TODO: create device
import "device/xyz"

// GPIO Pins
const (
    D0 = 0 // MOSI1
    D1 = 1 // MISO1
    D2 = 2 //
    D3 = 3 // CAN0TX SCL2
    D4 = 4 // CAN0RX SDA2
    D5 = 5 //
    D6 = 6 //
    D7 = 7 // RX3
    D8 = 8 // TX3
    D9 = 9 // RX2 CS0
    D10 = 10 // TX2 CS0
    D11 = 11 // MOSI0
    D12 = 12 // MISO0
    D13 = 13 // SCK0
    D14 = 14 //
    D15 = 15 // CS0
    D16 = 16 //
    D17 = 17 //
    D18 = 18 // SDA0
    D19 = 19 // SCL0
    D20 = 20 // CS0
    D21 = 21 // CS0
    D22 = 22 //
    D23 = 23 //
    D24 = 24 //
    D25 = 25 //
    D26 = 26 //
    D27 = 27 //
    D28 = 28 //
    D29 = 29 //
    D30 = 30 //
    D31 = 31 //
    D32 = 32 //
    D33 = 33 //
    D34 = 34 //
    D35 = 35 //
    D36 = 36 //
    D37 = 37 //
    D38 = 38 //
    D39 = 39 //
)

// Analog pins
const (
    A0 = 14 // PWM
    A1 = 15 // CS0
    A2 = 16 // PWM
    A3 = 17 // PWM
    A4 = 18 // SDA0
    A5 = 19 // SCL0
    A6 = 20 // PWM CS0
    A7 = 21 // PWM CS0
    A8 = 22 // PWM
    A9 = 23 // PWM
    A10 = 64 //
    A11 = 65 //
    A12 = 31 // RX4 CS1
    A13 = 32 // TX4 SCK1
    A14 = 33 // CAN1TX TX5
    A15 = 34 // CAN1RX RX5
    A16 = 35 // PWM
    A17 = 36 // PWM
    A18 = 37 // PWM SCL1
    A19 = 38 // PWM SDA1
    A20 = 39 //
    A21 = 66 //
    A22 = 67 //
    A23 = 49 //
    A24 = 50 //
    A25 = 68 //
    A26 = 69 //
)


// TODO: port the following functions
    // #define analogInputToDigitalPin(p) (((p) <= 9) ? (p) + 14 : (((p) >= 12 && (p) <= 20) ? (p) + 19 : (((p) == 23 || (p) == 24) ? (p) + 26 : -1)))
    // #define digitalPinHasPWM(p) (((p) >= 2 && (p) <= 10) || (p) == 14 || (p) == 16 || (p) == 17 || ((p) >= 20 && (p) <= 23) || (p) == 29 || (p) == 30 || ((p) >= 35 && (p) <= 38))
    // #define digitalPinToInterrupt(p)  ((p) < NUM_DIGITAL_PINS ? (p) : -1)

const (
	LED_BUILTIN = 13
)

// Serial1 pins
const (
	UART1_RX_PIN = 0
	UART1_TX_PIN = 1
)

// Serial2 pins
const (
	UART2_RX_PIN = 9
	UART2_TX_PIN = 10
)

// Serial3 pins
const (
	UART3_RX_PIN = 7
	UART3_TX_PIN = 8
)

// Serial4 pins
const (
	UART4_RX_PIN = 31
	UART4_TX_PIN = 32
)

// Serial5 pins
const (
	UART5_RX_PIN = 34
	UART5_TX_PIN = 33
)

// TODO: create a handler like this
// UART1 on the Feather M0.
// var (
// 	UART1 = UART{Bus: sam.SERCOM1_USART,
// 		Buffer: NewRingBuffer(),
// 		Mode:   PinSERCOM,
// 		IRQVal: sam.IRQ_SERCOM1,
// 	}
// )
//
// go:export SERCOM1_IRQHandler
// func handleUART1() {
// defaultUART1Handler()
// }


// I2C0 pins
const (
	SCL0_PIN = 19
	SDA0_PIN = 18

)

// I2C1 pins
const (
	SCL1_PIN = 38
	SDA1_PIN = 37

)

// I2C2 pins
const (
	SCL2_PIN = 3
	SDA2_PIN = 4

)

// TODO: create handler like this
// var (
//	I2C0 = I2C{Bus: sam.SERCOM3_I2CM,
//		SDA:     SDA_PIN,
//		SCL:     SCL_PIN,
//		PinMode: PinSERCOM}
// )


// SPI0 pins
const (
    SS0_PIN = 10
    MOSI0_PIN = 11
    MISO0_PIN = 12
    SCK0_PIN = 13
)

// SPI1 pins
const (
    SS1_PIN = 31
    MOSI1_PIN = 0
    MISO1_PIN = 1
    SCK1_PIN = 32
)


// TODO: create handler like this
// var (
// 	SPI0 = SPI{Bus: sam.SERCOM4_SPI}
// )
