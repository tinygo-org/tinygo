package machine

import "errors"

var errNoByte = errors.New("machine: no byte read")

// UARTConfig is a struct with which a UART (or similar object) can be
// configured. The baud rate is usually respected, but TX and RX may be ignored
// depending on the chip and the type of object.
type UARTConfig struct {
	BaudRate uint32
	TX       Pin
	RX       Pin
}

// NullSerial is a serial version of /dev/null (or null router): it drops
// everything that is written to it.
type NullSerial struct {
}

// Configure does nothing: the null serial has no configuration.
func (ns NullSerial) Configure(config UARTConfig) error {
	return nil
}

// WriteByte is a no-op: the null serial doesn't write bytes.
func (ns NullSerial) WriteByte(b byte) error {
	return nil
}

// ReadByte always returns an error because there aren't any bytes to read.
func (ns NullSerial) ReadByte() (byte, error) {
	return 0, errNoByte
}

// Buffered returns how many bytes are buffered in the UART. It always returns 0
// as there are no bytes to read.
func (ns NullSerial) Buffered() int {
	return 0
}

// Write is a no-op: none of the data is being written and it will not return an
// error.
func (ns NullSerial) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type Serialer interface {
	WriteByte(c byte) error
	Write(data []byte) (n int, err error)
	Configure(config UARTConfig) error
	Buffered() int
	ReadByte() (byte, error)
	DTR() bool
	RTS() bool
}
