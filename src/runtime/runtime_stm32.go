// +build stm32

package runtime

import (
	"device/arm"
)

type timeUnit int64

const tickMicros = 1 // TODO

//go:export Reset_Handler
func main() {
	preinit()
	initAll()
	mainWrapper()
	abort()
}

func putchar(c byte) {
	// TODO
}

func sleepTicks(d timeUnit) {
	// TODO: use a real timer here
	for i := 0; i < int(d/535); i++ {
		arm.Asm("")
	}
}

func ticks() timeUnit {
	return 0 // TODO
}
