//go:build !serial.uart

package custom2040

import "machine"

func setBoardSerial(u *machine.UART) {}
