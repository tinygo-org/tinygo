//go:build !serial.uart

package nousb2040

import "machine"

func setBoardSerial(u *machine.UART) {}
