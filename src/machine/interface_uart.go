//go:build !gameboyadvance && !attiny85
// +build !gameboyadvance,!attiny85

package machine

import "io"

func init() {
	type uarter interface {
		io.Writer
		io.Reader

		Buffered() int
	}

	var _ uarter = &UART{}
}
