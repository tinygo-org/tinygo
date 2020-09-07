// +build esp32_wroom_32

package machine

// Blue LED on the ESP32-WROOM-32 module.
const LED = Pin(2)

// I2C pins
const (
	SDA_PIN = 21
	SCL_PIN = 22
)

// I2C on the ESP32-WROOM-32.
var (
	I2C0 = I2C{}
)
