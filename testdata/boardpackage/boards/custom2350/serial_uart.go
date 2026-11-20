//go:build serial.uart

package custom2350

import "machine"

func setBoardSerial(u *machine.UART) {
	machine.Serial = u
}
