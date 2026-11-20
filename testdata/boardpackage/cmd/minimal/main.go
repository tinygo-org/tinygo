package main

import "machine"

func main() {
	println("hello")
	_ = machine.CPUFrequency()
	_ = machine.Watchdog
}
