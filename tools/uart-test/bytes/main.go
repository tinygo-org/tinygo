//go:build tinygo

package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/tinygo/tools/uart-test/common"
)

func pattern(k int) []byte {
	b := make([]byte, 256)
	for i := range b {
		switch k {
		case 0:
			b[i] = byte(i)
		case 1:
			b[i] = byte(255 - i)
		case 2:
			b[i] = 0x00
		default:
			if i%2 == 0 {
				b[i] = 0xAA
			} else {
				b[i] = 0x55
			}
		}
	}
	return b
}

func main() {
	common.WaitGo("bytes")
	println("@@GO bytes")
	for k := 0; k < 4; k++ {
		println("@@BIN", k)
		machine.Serial.Write(pattern(k))
		println()
	}
	println("@@END bytes")
	for {
		time.Sleep(time.Second)
	}
}
