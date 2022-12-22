//go:build baremetal && serial.uart

package machine

// Serial is implemented via the default (usually the first) UART on the chip.
var Serial = DefaultUART

func InitSerial() {
	Serial.Configure(UARTConfig{})
}
