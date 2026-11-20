//go:build baremetal && serial.uart && boardpackage

package machine

// Serial is populated by boardInit1() before runtime.initAll runs.
var Serial *UART

func InitSerial() {}
