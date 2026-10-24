//go:build tinygo

package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/tinygo/tools/uart-test/common"
)

func main() {
	common.WaitGo("burst")
	println("@@GO burst")

	// Do not read while the host sends a burst bigger than the RX buffer.
	time.Sleep(1500 * time.Millisecond)

	var got []byte
	for machine.Serial.Buffered() > 0 {
		b, _ := machine.Serial.ReadByte()
		got = append(got, b)
	}
	println("@@BUF", len(got))
	println("@@DATA", common.Hex(got))

	// The UART must keep receiving after an overrun.
	println("@@RECOVER")
	rec := common.ReadN(16, 5*time.Second)
	println("@@REC", common.Hex(rec))
	println("@@END burst")
	for {
		time.Sleep(time.Second)
	}
}
