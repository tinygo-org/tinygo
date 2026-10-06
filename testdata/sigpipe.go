package main

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
)

// The test runs this with SIGPIPE ignored and stdout a pipe without a reader.
func main() {
	switch os.Args[1] {
	case "stdout":
		os.Stdout.Write([]byte("x"))
	case "notify":
		signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)
		_, err := os.Stdout.Write([]byte("x"))
		report(err)
	case "reset":
		signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)
		signal.Reset(syscall.SIGPIPE)
		r, w, err := os.Pipe()
		if err != nil {
			panic(err)
		}
		r.Close()
		_, err = w.Write([]byte("x"))
		report(err)
	}
	os.Stderr.WriteString("still running\n")
}

func report(err error) {
	if errors.Is(err, syscall.EPIPE) {
		os.Stderr.WriteString("EPIPE\n")
	} else {
		os.Stderr.WriteString("unexpected error\n")
	}
}
