package main

// #error hello
// )))
import "C"

func main() {
}

// ERROR: # command-line-arguments
// ERROR: cgo.go:3:5: error: hello
// ERROR: cgo.go:4:4: error: expected identifier or '('
