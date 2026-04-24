package main

// This global can be changed using -ldflags="-X main.someGlobal=value".
var someGlobal string

func main() {
	println("someGlobal:", someGlobal)
}
