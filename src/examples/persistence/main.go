package main

import (
	"time"
)

//go:section .persist
var buffer [32]byte

func main() {
	println("\n*** ** * RESET * ** ***\n")

	for {
		time.Sleep(1 * time.Second)

		println("value: ", buffer[0])
		buffer[0]++
	}
}
