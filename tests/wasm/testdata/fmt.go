package main

import "fmt"

func main() {
	var _ fmt.Stringer
	println("did not panic")
}
