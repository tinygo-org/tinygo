package main

import _ "github.com/tinygo-org/tinygo/testdata/errors/importcycle"

func main() {
}

// ERROR: package command-line-arguments
// ERROR: 	imports github.com/tinygo-org/tinygo/testdata/errors/importcycle
// ERROR: 	imports github.com/tinygo-org/tinygo/testdata/errors/importcycle: import cycle not allowed
