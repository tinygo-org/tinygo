package main

// This test case assumes the init functions are called in the order they are
// found.

func init() {
	println("init1")
}

func init() {
	println("init2")
}

func main() {
	println("main")
}
