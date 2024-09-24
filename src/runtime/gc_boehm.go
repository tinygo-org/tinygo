//go:build gc.boehm

package runtime

import (
	"unsafe"
)

// Layout value for a pointer free object.
const pointerFreeLayout = 0x3

const needsStaticHeap = false

// zeroSizedAlloc is just a sentinel that gets returned when allocating 0 bytes.
var zeroSizedAlloc uint8

func initHeap() {
	libgc_init()

	gcInit()
}

//export tinygo_runtime_bdwgc_init
func gcInit()

//export tinygo_runtime_bdwgc_callback
func gcCallback() {
	// Mark the system stack and (if we're on a goroutine stack) also the
	// current goroutine stack.
	markStack()

	findGlobals(func(start, end uintptr) {
		libgc_push_all(start, end)
	})
}

func markRoots(start, end uintptr) {
	libgc_push_all(start, end)
}

func markCurrentGoroutineStack(sp uintptr) {
	// Only mark the area of the stack that is currently in use.
	// (This doesn't work for other goroutines, but at least it doesn't keep
	// more pointers alive than needed on the current stack).
	base := libgc_base(sp)
	if base == 0 { // && asserts
		runtimePanic("goroutine stack not in a heap allocation?")
	}
	stackBottom := base + libgc_size(base)
	libgc_push_all_stack(sp, stackBottom)
}

//go:noinline
func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer {
	if size == 0 {
		return unsafe.Pointer(&zeroSizedAlloc)
	}

	var ptr unsafe.Pointer
	if uintptr(layout) == pointerFreeLayout {
		// This object is entirely pointer free, for example make([]int, ...).
		// Make sure the GC knows this so it doesn't scan the object
		// unnecessarily to improve performance.
		ptr = libgc_malloc_atomic(size)
		// Memory returned from libgc_malloc_atomic has not been zeroed so we
		// have to do that manually.
		memzero(ptr, size)
	} else {
		// TODO: bdwgc supports typed allocations, which could be useful to
		// implement a mostly-precise GC.
		ptr = libgc_malloc(size)
		// Memory returned from libgc_malloc has already been zeroed, so nothing
		// to do here.
	}
	if ptr == nil {
		runtimePanic("gc: out of memory")
	}

	return ptr
}

func free(ptr unsafe.Pointer) {
	libgc_free(ptr)
}

func GC() {
	libgc_gcollect()
}

// This should be stack-allocated, but we don't currently have a good way of
// ensuring that happens.
var gcMemStats libgc_prof_stats

func ReadMemStats(m *MemStats) {
	libgc_get_prof_stats(&gcMemStats, unsafe.Sizeof(gcMemStats))

	// Fill in MemStats as well as we can, given the information that bdwgc
	// provides to us.
	m.HeapIdle = uint64(gcMemStats.free_bytes_full - gcMemStats.unmapped_bytes)
	m.HeapInuse = uint64(gcMemStats.heapsize_full - gcMemStats.unmapped_bytes)
	m.HeapReleased = uint64(gcMemStats.unmapped_bytes)
	m.HeapSys = uint64(m.HeapInuse + m.HeapIdle)
	m.GCSys = 0 // not provided by bdwgc
	m.TotalAlloc = uint64(gcMemStats.allocd_bytes_before_gc + gcMemStats.bytes_allocd_since_gc)
	m.Mallocs = 0 // not provided by bdwgc
	m.Frees = 0   // not provided by bdwgc
	m.Sys = uint64(gcMemStats.obtained_from_os_bytes)
}

func setHeapEnd(newHeapEnd uintptr) {
	runtimePanic("gc: did not expect setHeapEnd call")
}

//export GC_init
func libgc_init()

//export GC_malloc
func libgc_malloc(uintptr) unsafe.Pointer

//export GC_malloc_atomic
func libgc_malloc_atomic(uintptr) unsafe.Pointer

//export GC_free
func libgc_free(unsafe.Pointer)

//export GC_base
func libgc_base(ptr uintptr) uintptr

//export GC_size
func libgc_size(ptr uintptr) uintptr

//export GC_push_all
func libgc_push_all(bottom, top uintptr)

//export GC_push_all_stack
func libgc_push_all_stack(bottom, top uintptr)

//export GC_gcollect
func libgc_gcollect()

//export GC_get_prof_stats
func libgc_get_prof_stats(*libgc_prof_stats, uintptr) uintptr

type libgc_prof_stats struct {
	heapsize_full             uintptr
	free_bytes_full           uintptr
	unmapped_bytes            uintptr
	bytes_allocd_since_gc     uintptr
	allocd_bytes_before_gc    uintptr
	non_gc_bytes              uintptr
	gc_no                     uintptr
	markers_m1                uintptr
	bytes_reclaimed_since_gc  uintptr
	reclaimed_bytes_before_gc uintptr
	expl_freed_bytes_since_gc uintptr
	obtained_from_os_bytes    uintptr
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// Unimplemented.
	// The GC *does* support finalization, so this could be added relatively
	// easily I think.
}
