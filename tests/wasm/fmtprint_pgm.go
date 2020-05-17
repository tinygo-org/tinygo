// +build tinygo

package main

import "fmt"

func main() {
	fmt.Println("test from fmtprint_pgm 1")
	fmt.Print("test from fmtprint_pgm 2\n")
	fmt.Print("test from fmtp")
	fmt.Print("rint_pgm 3\n")
	fmt.Printf("test from fmtprint_pgm %d\n", 4)
}
