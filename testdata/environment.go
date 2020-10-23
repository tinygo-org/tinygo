package main

import "os"

func main() {
	println("args:", len(os.Args))
	for _, arg := range os.Args {
		println("arg:", arg)
	}
}
