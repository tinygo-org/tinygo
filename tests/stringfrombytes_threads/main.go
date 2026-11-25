package main

import (
	"sync"
	"sync/atomic"
)

const size = 4096

var (
	mu   sync.Mutex
	buf  = make([]byte, size)
	stop atomic.Bool
)

func fill(c byte) {
	for i := range buf {
		buf[i] = c
	}
}

// writer rewrites the whole buffer with one byte value at a time under mu.
func writer(done chan struct{}) {
	for c := byte(0); !stop.Load(); c++ {
		mu.Lock()
		fill('a' + c%26)
		mu.Unlock()
	}
	close(done)
}

// uniform reports whether every byte of s is equal. It runs without the lock.
//
//go:noinline
func uniform(s string) bool {
	for round := 0; round < 20; round++ {
		for i := 1; i < len(s); i++ {
			if s[i] != s[0] {
				return false
			}
		}
	}
	return true
}

// The conversion happens under the lock and the callee reads the private copy
// after unlocking. This is race free in Go, so the copy must not be removed.
func main() {
	done := make(chan struct{})
	go writer(done)
	torn := 0
	for i := 0; i < 20000; i++ {
		mu.Lock()
		s := string(buf)
		mu.Unlock()
		if !uniform(s) {
			torn++
		}
	}
	stop.Store(true)
	<-done
	if torn != 0 {
		println("FAIL torn strings:", torn)
		panic("string shared memory with a concurrently written buffer")
	}
	println("ok")
}
