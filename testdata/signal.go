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
	// A signal must reach the channel while the program does nothing else. The
	// receive below is the only thing that runs, so no timer and no sleep can
	// carry the delivery.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	report(<-c)
	signal.Stop(c)

	// The same again, with a goroutine that reads the channel while the main
	// goroutine sleeps.
	c2 := make(chan os.Signal, 1)
	signal.Notify(c2, syscall.SIGUSR1)

	// Wait for signals to arrive.
	go func() {
		for sig := range c2 {
			report(sig)
		}
	}()

	// Send the signal.
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)

	time.Sleep(time.Millisecond * 100)

	// Stop notifying.
	// (This is just a smoke test, it's difficult to test the default behavior
	// in a unit test).
	signal.Ignore(syscall.SIGUSR1)

	signal.Stop(c2)

	println("exiting signal program")
}

func report(sig os.Signal) {
	if sig == syscall.SIGUSR1 {
		println("got expected signal")
	} else {
		println("got signal:", sig.String())
	}
}
