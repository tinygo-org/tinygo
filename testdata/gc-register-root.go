// The only references to these objects are in callee-saved registers across
// runtime.GC. The goroutine stack must be scanned while scanCurrentStack still
// has those registers pushed. This failed on cortex-m-qemu at -opt=0 and -opt=z.
package main

import "runtime"

type obj struct {
	magic uint32
	pad   [25]uint32
}

var churn [32]*obj

//go:noinline
func many() {
	a, b, c, d := &obj{magic: 1}, &obj{magic: 2}, &obj{magic: 3}, &obj{magic: 4}
	e, f, g, h := &obj{magic: 5}, &obj{magic: 6}, &obj{magic: 7}, &obj{magic: 8}
	runtime.GC()
	for i := range churn {
		churn[i] = &obj{magic: 0xffffffff}
	}
	ps := [...]*obj{a, b, c, d, e, f, g, h}
	for i, p := range ps {
		if p.magic != uint32(i+1) {
			println("object", i, "magic", p.magic)
			panic("live object collected")
		}
	}
}

func main() {
	many()
	println("ok")
}
