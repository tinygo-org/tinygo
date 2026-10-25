//go:build tinygo && baremetal

package dev

import "machine"

// Serial is the default serial port of the board.
var Serial serial

type serial struct{}

// Configure sets the baud rate.
func (serial) Configure(baud uint32) {
	machine.Serial.Configure(machine.UARTConfig{BaudRate: baud})
}

func (serial) Buffered() int {
	return machine.Serial.Buffered()
}

func (serial) ReadByte() (byte, error) {
	return machine.Serial.ReadByte()
}

func (serial) Write(b []byte) (int, error) {
	return machine.Serial.Write(b)
}

func (serial) WriteByte(c byte) error {
	return machine.Serial.WriteByte(c)
}
