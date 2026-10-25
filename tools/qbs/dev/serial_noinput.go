//go:build tinygo && qbs_noinput

package dev

// Serial has no input and writes through the runtime. Some targets, such as
// the QEMU ones, cannot receive input and have no machine package to use.
var Serial serial

type serial struct{}

func (serial) Configure(baud uint32) {}

func (serial) Buffered() int {
	return 0
}

func (serial) ReadByte() (byte, error) {
	return 0, errNoInput
}

func (serial) Write(b []byte) (int, error) {
	print(string(b))
	return len(b), nil
}

func (serial) WriteByte(c byte) error {
	print(string([]byte{c}))
	return nil
}

type noInputError struct{}

func (noInputError) Error() string {
	return "no input on this target"
}

var errNoInput error = noInputError{}
