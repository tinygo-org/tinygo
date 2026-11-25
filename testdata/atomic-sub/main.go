package main

import (
	"runtime"
	"time"
)

// AdjustTimeOffset does a 64-bit atomic add of a negated value, which is
// compiled to an atomic subtraction. This must link on MCUs without native
// 64-bit atomics.
func main() {
	t := time.Now()
	runtime.AdjustTimeOffset(-1 * int64(time.Since(t)))
}
