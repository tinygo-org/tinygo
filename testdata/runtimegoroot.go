package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	fmt.Println(runtime.GOROOT())
	os.Setenv("GOROOT", "/changed")
	fmt.Println(runtime.GOROOT())
}
