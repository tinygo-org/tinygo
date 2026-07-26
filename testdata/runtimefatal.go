package main

import "sync"

func main() {
	defer func() {
		println("recovered:", recover())
	}()

	var mutex sync.Mutex
	mutex.Unlock()
}
