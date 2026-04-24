package main

// This global has a default value. It should be overridable via
// -ldflags="-X main.someGlobal=value" just like an uninitialized global.
var someGlobal = "default"

func main() {
	println("someGlobal:", someGlobal)
}
