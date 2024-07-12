package main

import _ "github.com/tinygo-org/tinygo/testdata/errors/invaliddep"

func main() {
}

// ERROR: invaliddep{{[\\/]}}invaliddep.go:1:1: expected 'package', found ppackage
