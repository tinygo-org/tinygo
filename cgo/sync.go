package cgo

import (
	"sync"
	"unsafe"
)

// #include <stdlib.h>
import "C"

// refMap is a convenient way to store opaque references that can be passed to
// C. It is useful if an API uses function pointers and you cannot pass a Go
// pointer but only a C pointer.
type refMap struct {
	refs map[unsafe.Pointer]interface{}
	lock sync.Mutex
}

// Put stores a value in the map. It can later be retrieved using Get. It must
// be removed using Remove to avoid memory leaks.
func (m *refMap) Put(v interface{}) unsafe.Pointer {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.refs == nil {
		m.refs = make(map[unsafe.Pointer]interface{}, 1)
	}
	ref := C.malloc(1)
	m.refs[ref] = v
	return ref
}

// Get returns a stored value previously inserted with Put. Use the same
// reference as you got from Put.
func (m *refMap) Get(ref unsafe.Pointer) interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.refs[ref]
}

// Remove deletes a single reference from the map.
func (m *refMap) Remove(ref unsafe.Pointer) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.refs, ref)
	C.free(ref)
}
