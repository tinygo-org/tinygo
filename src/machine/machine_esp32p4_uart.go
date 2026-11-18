//go:build esp32p4 && serial.uart

package machine

// FlushSerial is a no-op for UART serial, which writes synchronously.
func FlushSerial() {}
