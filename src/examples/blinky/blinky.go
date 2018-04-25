
package main

import "runtime"

func main() {
	for {
		// TODO: enable/disable actual LED
		println("LED on")
		runtime.Sleep(runtime.Millisecond * 250)
		println("LED off")
		runtime.Sleep(runtime.Millisecond * 250)
	}
}
