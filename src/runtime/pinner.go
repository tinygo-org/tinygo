package runtime

import (
	"internal/reflectlite"
	"unsafe"
)

// A Pinner is a set of Go objects each pinned to a fixed location in memory.
// The [Pinner.Pin] method pins one object, while [Pinner.Unpin] unpins all
// pinned objects.
//
// The zero value of Pinner is ready to use.
type Pinner struct {
	refs []unsafe.Pointer
}

// Pin pins a Go object, preventing it from being moved or freed by the garbage
// collector until the [Pinner.Unpin] method has been called.
//
// The argument must be a pointer of any type or an [unsafe.Pointer].
func (p *Pinner) Pin(pointer any) {
	v := reflectlite.ValueOf(pointer)
	if v.RawType() == nil {
		runtimePanic(plainError("runtime.Pinner: argument is nil"))
	}
	if kind := v.Kind(); kind != reflectlite.Pointer && kind != reflectlite.UnsafePointer {
		runtimePanic(plainError("runtime.Pinner: argument is not a pointer: " + v.RawType().String()))
	}
	if ptr := v.UnsafePointer(); ptr != nil {
		p.refs = append(p.refs, ptr)
	}
}

// Unpin unpins all pinned objects of the [Pinner].
// It is safe to reuse a Pinner after calling Unpin.
func (p *Pinner) Unpin() {
	clear(p.refs)
	p.refs = p.refs[:0]
}
