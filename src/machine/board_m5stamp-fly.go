//go:build m5stamp_fly

// Pin mappings for the M5Stack StampFly, a small quadcopter built on the Stamp-S3A module.
// https://docs.m5stack.com/en/app/StampFly_v1.1 and https://github.com/m5stack/M5StampFly

package machine

// Digital pins
const (
	G0  = GPIO0
	G1  = GPIO1
	G2  = GPIO2
	G3  = GPIO3
	G4  = GPIO4
	G5  = GPIO5
	G6  = GPIO6
	G7  = GPIO7
	G8  = GPIO8
	G9  = GPIO9
	G10 = GPIO10
	G12 = GPIO12
	G13 = GPIO13
	G14 = GPIO14
	G15 = GPIO15
	G21 = GPIO21
	G39 = GPIO39
	G40 = GPIO40
	G41 = GPIO41
	G42 = GPIO42
	G43 = GPIO43
	G44 = GPIO44
	G46 = GPIO46
)

// Button
const (
	BUTTON = GPIO0
)

// UART pins. They are shared with the SPI bus, so the console uses USB.
const (
	UART_RX_PIN = GPIO44
	UART_TX_PIN = GPIO43
)

// Internal I2C bus with the BMM150, BMP280, two VL53L3CX and the INA3221.
const (
	SDA_PIN = GPIO3
	SCL_PIN = GPIO4
)

// Grove port
const (
	GROVE_SDA = GPIO13
	GROVE_SCL = GPIO15
	GROVE_G1  = GPIO1
	GROVE_G2  = GPIO2
)

// SPI bus with the BMI270 IMU and the PMW3901 optical flow sensor.
const (
	SPI1_SCK_PIN  = GPIO44
	SPI1_MISO_PIN = GPIO43
	SPI1_MOSI_PIN = GPIO14
	SPI1_CS_PIN   = NoPin

	SPI2_SCK_PIN  = NoPin
	SPI2_MISO_PIN = NoPin
	SPI2_MOSI_PIN = NoPin
	SPI2_CS_PIN   = NoPin
)

// Sensor pins
const (
	BMI270_CS        = GPIO46
	PMW3901_CS       = GPIO12
	TOF_BOTTOM_INT   = GPIO6
	TOF_BOTTOM_XSHUT = GPIO7
	TOF_FRONT_INT    = GPIO8
	TOF_FRONT_XSHUT  = GPIO9
)

// Motor PWM pins, seen from above with the front ToF sensor forward.
const (
	MOTOR_FRONT_LEFT  = GPIO5
	MOTOR_FRONT_RIGHT = GPIO42
	MOTOR_REAR_LEFT   = GPIO10
	MOTOR_REAR_RIGHT  = GPIO41
)

// Buzzer
const (
	BUZZER = GPIO40
)

// Onboard LEDs
const (
	// WS2812 is the data pin for the addressable RGB LED on the StampFly board.
	WS2812 = GPIO39

	// WS2812_STAMP is the data pin for the addressable RGB LED on the Stamp-S3A module.
	WS2812_STAMP = GPIO21
)
