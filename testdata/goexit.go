package main

import "runtime"

func main() {
	defer func() {
		panic("panic after Goexit")
	}()
	runtime.Goexit()
}
