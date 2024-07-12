package main

import (
	_ "github.com/tinygo-org/tinygo/testdata/errors/multi-1"
	_ "github.com/tinygo-org/tinygo/testdata/errors/multi-2"
	_ "github.com/tinygo-org/tinygo/testdata/errors/multi-3"
)

func main() {
}

// ERROR: # github.com/tinygo-org/tinygo/testdata/errors/multi-2
// ERROR: multi-2/syntax.go:3:11: expected ')', found 'type'
// ERROR: # github.com/tinygo-org/tinygo/testdata/errors/multi-3
// ERROR: multi-3/syntax.go:3:11: expected ')', found 'import'
// ERROR: # github.com/tinygo-org/tinygo/testdata/errors/multi-1
// ERROR: multi-1/types.go:3:13: cannot use 3.14 (untyped float constant) as int value in variable declaration (truncated)
