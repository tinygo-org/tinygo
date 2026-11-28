package main

// Scheduler/timer smoke test that builds and links against the internal/poll
// netpoller on linux/amd64: a pipe with a read deadline that expires, a late
// write waking a parked reader, and Close unblocking a parked read. net.Pipe
// is in-memory (no fd, no runtime_poll* calls), so the epoll path itself is
// not exercised here; it is exercised by programs using real fds under
// upstream os/net (see the stacked 0002+0001 evidence in the PR text).

import (
	"fmt"
	"net"
	"os"
	"time"
)

func check(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func main() {
	// net.Pipe: a read with a deadline that expires, then a late write.
	c1, c2 := net.Pipe()
	check(c1.SetReadDeadline(time.Now().Add(50 * time.Millisecond)))
	buf := make([]byte, 8)
	start := time.Now()
	_, err := c1.Read(buf)
	if err == nil {
		fmt.Println("pipe read unexpectedly succeeded")
		os.Exit(1)
	}
	if d := time.Since(start); d < 40*time.Millisecond {
		fmt.Println("pipe read deadline fired too early:", d)
		os.Exit(1)
	}
	fmt.Println("pipe deadline:", err)

	// A late write wakes a parked reader.
	go func() {
		time.Sleep(50 * time.Millisecond)
		_, err := c2.Write([]byte("hello"))
		check(err)
	}()
	check(c1.SetReadDeadline(time.Now().Add(2 * time.Second)))
	n, err := c1.Read(buf)
	check(err)
	fmt.Println("pipe wakeup:", n, string(buf[:n]))

	// Close unblocks a parked read.
	go func() {
		time.Sleep(50 * time.Millisecond)
		check(c2.Close())
	}()
	check(c1.SetReadDeadline(time.Now().Add(10 * time.Second)))
	_, err = c1.Read(buf)
	fmt.Println("pipe close:", err)
	if err == nil {
		fmt.Println("read after close unexpectedly succeeded")
		os.Exit(1)
	}

	// os.Exit must link wherever it is reached.
	fmt.Println("netpoll ok")
	os.Exit(0)
}
