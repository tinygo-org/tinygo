package main

import (
	"os"
	"os/signal"
	"time"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	time.Sleep(10 * time.Millisecond)
	println("done")
}
