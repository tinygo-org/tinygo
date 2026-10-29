package main

import (
	"errors"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	case "mask":
		// Linux only. A caught SIGPIPE resets to the default across exec, an
		// ignored one stays ignored in children.
		reportMask()
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

func reportMask() {
	status, err := os.ReadFile("/proc/self/status")
	if err != nil {
		panic(err)
	}
	for _, line := range strings.Split(string(status), "\n") {
		name, value, _ := strings.Cut(line, ":")
		if name != "SigIgn" && name != "SigCgt" {
			continue
		}
		mask, err := strconv.ParseUint(strings.TrimSpace(value), 16, 64)
		if err != nil {
			panic(err)
		}
		set := mask&(1<<(syscall.SIGPIPE-1)) != 0
		os.Stderr.WriteString(name + " " + strconv.FormatBool(set) + "\n")
	}
}
