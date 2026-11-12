//go:build linux || darwin

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "pipe" {
		r, w, err := os.Pipe()
		if err != nil {
			panic(err)
		}
		done := make(chan error, 1)
		go func() {
			var buf [1]byte
			_, err := r.Read(buf[:])
			done <- err
		}()
		time.Sleep(100 * time.Millisecond)
		go func() {
			time.Sleep(2 * time.Second)
			w.Close()
		}()
		start := time.Now()
		closeErr := r.Close()
		err = <-done
		fmt.Printf("pipe close=%v read=%v elapsed=%v\n", closeErr, err, time.Since(start))
		return
	}
	cmd := exec.Command("/bin/sh", "-c", "sleep 2 &")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.WaitDelay = 100 * time.Millisecond
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)
	fmt.Printf("WaitDelay=%v elapsed=%v error=%v\n", cmd.WaitDelay, elapsed, err)
	if !errors.Is(err, exec.ErrWaitDelay) || elapsed >= time.Second {
		os.Exit(1)
	}
}
