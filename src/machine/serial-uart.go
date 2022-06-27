//go:build baremetal && serial.uart
// +build baremetal,serial.uart

package machine

var Serial *UART

func InitSerial() {
	Serial = DefaultUART
}
