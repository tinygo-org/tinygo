package main

import (
	"machine"
	"time"
)

func main() {
	time.Sleep(10 * time.Second)
	println("beginning flash")

	// see if anything stored in flash to start with
	println("reading starting flash")
	data, _ := machine.FlashRead(0, 12)
	println(string(data))
	
	for {
		// println("erasing flash")
		// machine.FlashErase(0, 12)

		println("reading back flash")
		data, _ = machine.FlashRead(0, 5)
		println(string(data))

		println("writing flash")
		machine.FlashWrite(0, []byte("hello, world"))

		println("reading back flash")
		data, _ = machine.FlashRead(0, 12)
		println(string(data))

		// store something else
		machine.FlashWrite(0, []byte("happy world!"))

		println("reading back flash")
		data, _ = machine.FlashRead(0, 12)
		println(string(data))

		time.Sleep(time.Second)
	}
}
