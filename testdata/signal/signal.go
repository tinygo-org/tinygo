package main

// Test POSIX signals.
// TODO: run `tinygo test os/signal` instead, once CGo errno return values are
// supported.

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)

	// Wait for signals to arrive.
	go func() {
		for sig := range c {
			if sig == syscall.SIGUSR1 {
				println("got expected signal")
			} else {
				println("got signal:", sig.String())
			}
		}
	}()
	// test signal_enable

	// test signal_disable

	// test signal_ignore
	signal.Ignore(syscall.SIGUSR1)
	if signal.Ignored(syscall.SIGUSR1) {
		println("SIGUSR1 is ignored")
	} else {
		println("SIGUSR1 is not ignored")
	}

	// test signal_ignore SIGKILL and SIGSTOP that cannot be caught or ignored
	signal.Ignore()

	// send the signal.
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	time.Sleep(time.Millisecond * 100)
	println("exiting signal program")
}
