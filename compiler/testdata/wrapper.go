package main

import (
	"io"

	"github.com/tinygo-org/tinygo/compiler/testdata/os"
	"github.com/tinygo-org/tinygo/compiler/testdata/os/exec"
)

var (
	stdin io.ReadCloser = new(os.File)
	cmd                 = new(exec.Cmd)
)

func main() {
	wc, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	if err := wc.Close(); err != nil {
		panic(err)
	}
}
