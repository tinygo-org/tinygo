//go:build tinygo

package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/tinygo/tools/uart-test/common"
)

func main() {
	common.WaitGo("echo")
	println("@@GO echo")
	for {
		if machine.Serial.Buffered() > 0 {
			b, _ := machine.Serial.ReadByte()
			machine.Serial.WriteByte(b)
		} else {
			time.Sleep(time.Millisecond)
		}
	}
}
