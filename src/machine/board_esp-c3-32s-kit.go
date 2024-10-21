//go:build esp_c3_32s_kit

package machine

// See:
//  * https://www.waveshare.com/w/upload/8/8f/Esp32-c3s_specification.pdf
//  * https://www.waveshare.com/w/upload/4/46/Nodemcu-esp-c3-32s-kit-schematics.pdf

// Digital Pins
const (
	IO0  = GPIO0
	IO1  = GPIO1
	IO2  = GPIO2
	IO3  = GPIO3
	IO4  = GPIO4
	IO5  = GPIO5
	IO6  = GPIO6
	IO7  = GPIO7
	IO8  = GPIO8
	IO9  = GPIO9
	IO18 = GPIO18
	IO19 = GPIO19
)

const (
	LED_RED   = IO3
	LED_GREEN = IO4
	LED_BLUE  = IO5

	LED = LED_RED

	LED1 = LED_RED
	LED2 = LED_GREEN
)

// I2C pins
const (
	SDA_PIN = NoPin
	SCL_PIN = NoPin
)
