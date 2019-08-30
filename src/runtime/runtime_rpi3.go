// +build rpi3

package runtime

import _ "unsafe"
import dev "device/rpi3"

//const GOOS = "linux"
const tickMicros = int64(1)
const asyncScheduler = false

type timeUnit int64

var z timeUnit
var x byte

//go:export sleepticks sleepticks
func sleepTicks(n timeUnit) {
	//sleep this long
	z = n
}

func ticks() timeUnit {
	return timeUnit(0) //current time in tickts
}

func abort() {
	dev.UARTPuts("program aborted")
	for {

	}
}

func preinit() {
	dev.UARTInit()
}

//go:export main
func main() {
	preinit()
	initAll()
	callMain()
	abort()
}

func putchar(c byte) {
	dev.UARTSend(c)
}
