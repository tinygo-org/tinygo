package main

// #error hello
// )))
import "C"

func main() {
}

// TODO: this error should be relative to the current directory (so cgo.go
// instead of testdata/errors/cgo.go).

// ERROR: # command-line-arguments
// ERROR: testdata/errors/cgo.go:3:5: error: hello
// ERROR: testdata/errors/cgo.go:4:4: error: expected identifier or '('
