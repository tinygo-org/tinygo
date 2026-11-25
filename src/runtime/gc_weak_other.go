//go:build !gc.boehm && !gc.conservative && !gc.precise

package runtime

import "unsafe"

//go:linkname registerWeakPointer weak.runtime_registerWeakPointer
func registerWeakPointer(ptr unsafe.Pointer) unsafe.Pointer {
	return ptr
}

//go:linkname makeStrongFromWeak weak.runtime_makeStrongFromWeak
func makeStrongFromWeak(ptr unsafe.Pointer) unsafe.Pointer {
	return ptr
}
