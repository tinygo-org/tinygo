//go:build baremetal && serial.uart && boardpackage

package machine

// Serial must be assigned by boardInit1 from zero-value globals before initAll.
// Leaving Serial nil drops writes silently and spins getchar on Gosched.
var Serial *UART

func InitSerial() {}
