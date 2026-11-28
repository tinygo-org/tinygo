package main

import (
	"fmt"
	"os"
	"time"
	_ "unsafe"
)

//go:linkname pollWait runtime.runtime_netpoll_wait
func pollWait(fd uint32, mode uint8, timeout uint64) uint32

func main() {
	if os.Args[1] == "ready" {
		fd := os.Stdin.PollFD()
		if err := fd.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			panic(err)
		}
		var p [1]byte
		n, err := fd.Read(p[:])
		fmt.Printf("ready: %d %v %s\n", n, err, p[:n])
		if err := fd.SetReadDeadline(time.Now().Add(-time.Millisecond)); err != nil {
			panic(err)
		}
		n, err = fd.Read(p[:])
		fmt.Printf("expired: %d %v\n", n, err)
		return
	}
	if os.Args[1] != "timeout" {
		panic("unknown poll test")
	}
	start := time.Now()
	if errno := pollWait(0, 1, uint64(20*time.Millisecond)); errno != 0 {
		fmt.Println("poll error:", errno)
		return
	}
	if time.Since(start) < 10*time.Millisecond {
		fmt.Println("early wake")
		return
	}
	fmt.Println("timed out")
	fmt.Println("bad fd:", pollWait(999, 1, uint64(time.Millisecond)))
}
