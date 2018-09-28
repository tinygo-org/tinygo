package main

import "time"

func main() {
	println("main 1")
	go sub()
	time.Sleep(1 * time.Millisecond)
	println("main 2")
	time.Sleep(2 * time.Millisecond)
	println("main 3")
}

func sub() {
	println("sub 1")
	time.Sleep(2 * time.Millisecond)
	println("sub 2")
}
