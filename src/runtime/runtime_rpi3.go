// +build rpi3

package runtime

import _ "unsafe"

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
}

func preinit() {
	x = x + 1
}

//go:export main
func main() {
	preinit()
	initAll()
	callMain()
	abort()
}

func putchar(c byte) {
	x = c
}
