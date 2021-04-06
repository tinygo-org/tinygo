package main

// These globals can be changed using -ldflags="-X main.someGlobal=value".
// At the moment, only globals without an initializer can be replaced this way.
var someGlobal string

func main() {
	println("someGlobal:", someGlobal)
}
