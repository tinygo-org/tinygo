// A pointer that is live in a frame is not necessarily rooted by that frame.
//
// This is the loop shape from encoding/hex.TestDumper. The dumper reaches the
// loop header from the preheader and the error returned by Write reaches it
// from the latch, so SimplifyCFG merges the two runtime.trackPointer calls
// into one taking a phi. The single slot that results only roots whichever of
// the two the current iteration selected, so on the back edge the dumper is
// unrooted and the runtime.GC in Write is free to sweep it. scratch is an
// array of the same type, so a dumper freed while in use is handed back out
// and overwritten rather than left intact by luck.
//
// Three details keep the dumper reachable only through that slot, and removing
// any of them disarms the test rather than failing it: newDumper must stay out
// of line, or its own frame roots the dumper too; it must return an interface,
// so the value tracked is the extracted data pointer and not the allocation;
// and Write must allocate, so a collection can land inside it.
//
// This only fails at -opt=z, which is what TestBuild uses. At -opt=0, -opt=1,
// -opt=2 and -opt=s there is no signal at all, so a failure here is not
// something an optimisation level change fixes.
package main

import "runtime"

const magic = 0x5a5a5a5a

type writer interface {
	Write([]byte) (int, error)
}

type dumper struct {
	w     writer
	magic uint32
	buf   [4]byte
	used  int
}

var scratch [64]*dumper

//go:noinline
func newDumper(w writer) writer {
	return &dumper{w: w, magic: magic}
}

//go:noinline
func (d *dumper) Write(p []byte) (n int, err error) {
	runtime.GC()
	for i := range scratch {
		scratch[i] = &dumper{magic: 0xffffffff}
	}
	if d.magic != magic {
		println("magic", d.magic, "want", uint32(magic))
		panic("live object was collected")
	}
	for _, b := range p {
		d.buf[0] = b
		_, err = d.w.Write(d.buf[:1])
		if err != nil {
			return
		}
		d.used++
		n++
	}
	return
}

type collector struct {
	used int
}

func (c *collector) Write(p []byte) (int, error) {
	c.used += len(p)
	return len(p), nil
}

func main() {
	var input [40]byte
	c := &collector{}
	d := newDumper(c)
	for i := range input {
		d.Write(input[i : i+1])
	}
	if d.(*dumper).used != len(input) || c.used != len(input) {
		panic("wrong output")
	}
	println("ok")
}
