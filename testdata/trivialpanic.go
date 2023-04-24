package main

var n *int

func main() {
	println(*n) // this will panic
}
