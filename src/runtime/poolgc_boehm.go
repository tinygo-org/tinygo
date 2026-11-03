//go:build gc.boehm

package runtime

//export GC_get_gc_no
func libgc_get_gc_no() uintptr

// poolGCCount is the number of collections so far, for sync.Pool.
func poolGCCount() uint32 { return uint32(libgc_get_gc_no()) }
