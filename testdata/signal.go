package main

// Test POSIX signals.
// TODO: run `tinygo test os/signal` instead, once CGo errno return values are
// supported.

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)

	// Send the signal.
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)

	// Receive it directly, with nothing sleeping anywhere: under the threads
	// scheduler only the watcher can deliver it, so this hangs without the fix.
	if sig := <-c; sig == syscall.SIGUSR1 {
		println("got expected signal")
	} else {
		println("got signal:", sig.String())
	}

	// Stop notifying.
	// (This is just a smoke test, it's difficult to test the default behavior
	// in a unit test).
	signal.Ignore(syscall.SIGUSR1)

	signal.Stop(c)

	println("exiting signal program")
}
