package main

/*
#cgo CFLAGS: -Iinclude
int value(void);
static int inlineValue(void) { return 1; }
*/
import "C"

import (
	_ "embed"
	"runtime"

	"example.com/dependency/subpackage"
)

//go:embed data.txt
var data string

func main() {
	println(C.value(), C.inlineValue(), dependency.Value(), dependency.HeaderPath(), data, runtime.GOROOT())
}
