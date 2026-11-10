//go:build wasip3 && (custommalloc || gc.boehm)

package runtime

// mallocEnd is for the allocators that do not keep the map of allocations. They
// do not scan the stacks of the tasks that the host started.
//
//go:linkname mallocEnd internal/task.mallocEnd
func mallocEnd(addr uintptr) uintptr { return 0 }
