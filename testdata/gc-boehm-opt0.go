package main

import "runtime"

func main() {
	runtime.GC()
	println("ok")
}
