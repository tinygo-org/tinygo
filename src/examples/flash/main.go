package main

import "machine"

func main() {
	println("flash data start:", machine.FlashDataStart())
	println("flash data end:  ", machine.FlashDataEnd())
}
