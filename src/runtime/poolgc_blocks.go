//go:build gc.conservative || gc.precise

package runtime

// poolGCCount is the number of collections so far, for sync.Pool.
func poolGCCount() uint32 { return gcNumGC }
