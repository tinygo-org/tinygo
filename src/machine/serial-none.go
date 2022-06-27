//go:build baremetal && serial.none
// +build baremetal,serial.none

package machine

var Serial NullSerial

func InitSerial() {
	Serial = NullSerial{}
}
