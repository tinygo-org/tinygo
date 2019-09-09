// +build rpi3

package runtime

import dev "device/rpi3"

//const GOOS = "linux"
const tickMicros = int64(1)
const asyncScheduler = false

type timeUnit int64

var z timeUnit
var x byte

//go:export sleepticks sleepticks
func sleepTicks(n timeUnit) {
	dev.WaitMuSec(uint32(n))
}

func ticks() timeUnit {
	return timeUnit(dev.SysTimer())
}

func preinit() {

	dev.UART0Init()
	//MinuUARTInit() if you prefer the MiniUart

	// XXX what should we do for a bootloader?
	heapStart := 0x90000
	heapEnd = 0xAFFF8
	heapptr = uintptr(heapStart)
	globalsStart = 0xB0000
	globalsEnd = 0xB0FF8

}

//go:export main
func main() {
	preinit()
	primaryMain()
}

// main on non bootloader (the bootloader itself uses this path)
func primaryMain() {
	initAll()
	callMain()
	abort()
}

//go:export mainFromBootloader
func mainFromBootloader() {
	heapStart := dev.ReadRegister("x2")
	heapEnd = heapStart + 0x8000
	heapptr = uintptr(heapStart)
	//globalsStart = 0xB0000
	//globalsEnd = 0xB0FF8
	// XXX FIXME XXX
	// need to do something with time
	//time:=dev.ReadRegister("x3")
	primaryMain()
}

func putchar(c byte) {
	//MiniUARTSend(c) if you prefer the mini uart
	dev.UART0Send(c)
}

// just send to device code which ends up calling WFE
func abort() {
	dev.Abort()
}
