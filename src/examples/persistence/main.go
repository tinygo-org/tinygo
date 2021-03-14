package main

import (
	"time"
	"unsafe"

	"runtime/persistence"
)

func main() {
	p := persistence.NewRAM(uint32(unsafe.Sizeof(byte(0)) * 100))

	buf := []byte{0}

	println("\n*** ** * RESET * ** ***\n")
	println("Persistent Region Size: ", p.Size())

	for {
		time.Sleep(1 * time.Second)

		_, err := p.ReadAt(buf, 0)
		if err != nil {
			println("error: ", err)
		}

		println("value: ", buf[0])

		buf[0]++

		p.WriteAt(buf, 0)

	}
}
