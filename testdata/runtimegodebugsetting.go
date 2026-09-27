//go:debug tarinsecurepath=0

package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"syscall"
)

func main() {
	var data bytes.Buffer
	writer := tar.NewWriter(&data)
	if err := writer.WriteHeader(&tar.Header{Name: "../test"}); err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}
	printState := func() {
		_, err := tar.NewReader(bytes.NewReader(data.Bytes())).Next()
		if err != nil && err != tar.ErrInsecurePath {
			panic(err)
		}
		fmt.Println(err == tar.ErrInsecurePath)
	}

	printState()
	if err := syscall.Setenv("GODEBUG", "tarinsecurepath=1"); err != nil {
		panic(err)
	}
	printState()
	if err := syscall.Unsetenv("GODEBUG"); err != nil {
		panic(err)
	}
	printState()
	if err := syscall.Setenv("GODEBUG", "tarinsecurepath=1"); err != nil {
		panic(err)
	}
	syscall.Clearenv()
	printState()
}
