//go:build gc.boehm

// This is the Boehm-Demers-Weiser conservative garbage collector, integrated
// into TinyGo.
//
// Note that we use a special way of dealing with threads:
//   * All calls to the bdwgc library are serialized using locks.
//   * When the bdwgc library wants to push GC roots, all other threads that are
//     running are stopped.
//   * After returning from a bdwgc library call, the caller checks whether
//     other threads were stopped (meaning a GC cycle happened) and resumes the
//     world.
// This is not exactly the most efficient way to do this. We can likely speed
// things up by using bdwgc-native wrappers for starting/stopping threads (and
// also to resume the world while sweeping). Also, thread local allocation might
// help. But we don't do any of these right now, it is left as a possible future
// improvement.

package runtime

import (
	"internal/gclayout"
	"internal/reflectlite"
	"internal/task"
	"unsafe"
)

const needsStaticHeap = false

const boehmLayoutSizeBits = 4 + unsafe.Sizeof(uintptr(0))/4
const (
	boehmPtrFreeKind = 0
	boehmNormalKind  = 1
)

var (
	gcLock    task.PMutex
	gcMallocs uint64
)

type boehmFinalizer struct {
	next   *boehmFinalizer
	offset uintptr
	fn     interface{}
	ptr    unsafe.Pointer // keeps the object alive after Boehm dequeues its callback
}

var (
	finalizerPending       *boehmFinalizer
	numFinalizers          uintptr
	finalizersSinceGC      uintptr
	finalizerQueued        bool
	finalizerDraining      bool
	finalizerRunnerStarted bool
)

func initHeap() {
	libgc_init()

	// Call GC_set_push_other_roots(gcCallback) in C because of function
	// signature differences that do matter in WebAssembly.
	gcInit()
}

//export tinygo_runtime_bdwgc_init
func gcInit()

//export tinygo_runtime_bdwgc_callback
func gcCallback() {
	// Mark globals and all stacks, and stop the world if we're using threading.
	gcMarkReachable()
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
		runtimeFatal("goroutine stack not in a heap allocation?")
	}
	stackBottom := base + libgc_size(base)
	libgc_push_all_stack(sp, stackBottom)
}

//go:noinline
func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer {
	if size == 0 {
		return alloc_zero(size, layout)
	}

	gcLock.Lock()
	gcMallocs++
	var ptr unsafe.Pointer
	var needsZero bool
	switch layout {
	case gclayout.NoPtrs.AsPtr():
		// This object is entirely pointer free, for example make([]int, ...).
		// Make sure the GC knows this so it doesn't scan the object
		// unnecessarily to improve performance.
		ptr = libgc_malloc_kind(size, boehmPtrFreeKind)
		needsZero = true
	case gclayout.Conservative.AsPtr():
		// Stack storage does not have an ordinary repeating Go object layout.
		ptr = libgc_malloc_kind(size, boehmNormalKind)
	case gclayout.Pointer.AsPtr(), gclayout.PointerPair.AsPtr():
		// Conservative scanning is exact when every word is a pointer.
		ptr = libgc_malloc_kind(size, boehmNormalKind)
	default:
		elementWords := boehmLayoutElementWords(layout)
		pointerAlign := unsafe.Alignof(uintptr(0))
		if elementWords == 0 || elementWords > size/pointerAlign {
			// This should not happen for compiler-generated Go allocations.
			ptr = libgc_malloc_kind(size, boehmNormalKind)
			break
		}
		elementSize := elementWords * pointerAlign
		if size%elementSize != 0 {
			ptr = libgc_malloc_kind(size, boehmNormalKind)
			break
		}

		descriptor := libgc_make_descriptor(uintptr(layout))
		if descriptor == 0 {
			// The bridge returns zero if its cache or bitmap allocation fails.
			// It also rejects the no-pointer descriptor, which is handled above.
			ptr = libgc_malloc_kind(size, boehmNormalKind)
			break
		}
		elementCount := size / elementSize
		if elementCount == 1 {
			ptr = libgc_malloc_explicitly_typed(size, descriptor)
		} else {
			ptr = libgc_calloc_explicitly_typed(
				elementCount, elementSize, descriptor,
			)
		}
	}
	queued := boehmInvokeFinalizers()
	gcResumeWorld()
	gcLock.Unlock()
	if queued {
		wakeFinalizer()
	}
	if ptr == nil {
		runtimeFatal("gc: out of memory")
		return nil
	}
	if needsZero {
		memzero(ptr, size)
	}
	return ptr
}

func boehmLayoutElementWords(layout unsafe.Pointer) uintptr {
	value := uintptr(layout)
	if value&1 != 0 {
		return (value >> 1) & (1<<boehmLayoutSizeBits - 1)
	}
	return *(*uintptr)(layout)
}

func allocManual(size uintptr) unsafe.Pointer {
	if size == 0 {
		return alloc_zero(size, gclayout.NoPtrs.AsPtr())
	}

	gcLock.Lock()
	ptr := libgc_malloc_atomic_uncollectable(size)
	queued := boehmInvokeFinalizers()
	gcResumeWorld()
	gcLock.Unlock()
	if queued {
		wakeFinalizer()
	}
	if ptr == nil {
		runtimeFatal("gc: out of memory")
		return nil
	}
	memzero(ptr, size)
	return ptr
}

func free(ptr unsafe.Pointer) {
	gcLock.Lock()
	libgc_free(ptr)
	queued := boehmInvokeFinalizers()
	gcResumeWorld()
	gcLock.Unlock()
	if queued {
		wakeFinalizer()
	}
}

//go:noinline
func freeTaskStack(ptr uintptr) {
	free(unsafe.Pointer(ptr))
}

func GC() {
	gcLock.Lock()
	libgc_gcollect()
	finalizersSinceGC = 0
	queued := boehmInvokeFinalizers()
	gcResumeWorld()
	gcLock.Unlock()
	if queued {
		wakeFinalizer()
	}
}

// This should be stack-allocated, but we don't currently have a good way of
// ensuring that happens.
var gcMemStats libgc_prof_stats

func ReadMemStats(m *MemStats) {
	gcLock.Lock()

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
	m.NumGC = uint32(gcMemStats.gc_no)

	gcLock.Unlock()
}

func mallocs() uint64 {
	gcLock.Lock()
	mallocs := gcMallocs
	gcLock.Unlock()
	return mallocs
}

func setHeapEnd(newHeapEnd uintptr) {
	runtimeFatal("gc: did not expect setHeapEnd call")
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	value := reflectlite.ValueOf(obj)
	if value.Kind() != reflectlite.Pointer {
		runtimeFatal("runtime.SetFinalizer: first argument is not a pointer")
	}
	if finalizer != nil && reflectlite.ValueOf(finalizer).Kind() != reflectlite.Func {
		runtimeFatal("runtime.SetFinalizer: second argument is not a function")
	}
	objPtr := (*_interface)(unsafe.Pointer(&obj)).value
	if objPtr == nil {
		return
	}

	var entry *boehmFinalizer
	if finalizer != nil {
		entry = &boehmFinalizer{fn: finalizer}
	}
	gcLock.Lock()
	base := libgc_base(uintptr(objPtr))
	if base == 0 {
		gcLock.Unlock()
		return
	}
	if base != uintptr(objPtr) && !reflectlite.FinalizerAllowsInterior(value.RawType()) {
		runtimeFatal("runtime.SetFinalizer: pointer not at beginning of allocated block")
	}
	offset := uintptr(objPtr) - base
	old := libgc_register_finalizer(base, uintptr(unsafe.Pointer(entry)))
	if old == ^uintptr(0) {
		runtimeFatal("gc: cannot register finalizer")
	}
	head := (*boehmFinalizer)(unsafe.Pointer(old))
	replaced := false
	var prev *boehmFinalizer
	for n := head; n != nil; n = n.next {
		if n.offset == offset {
			if prev == nil {
				head = n.next
			} else {
				prev.next = n.next
			}
			*n = boehmFinalizer{}
			replaced = true
			if entry == nil {
				numFinalizers--
				if finalizersSinceGC != 0 {
					finalizersSinceGC--
				}
			}
			break
		}
		prev = n
	}
	if entry != nil {
		entry.offset = offset
		entry.next = head
		if !replaced {
			numFinalizers++
			finalizersSinceGC++
		}
	} else if head != nil {
		if libgc_register_finalizer(base, uintptr(unsafe.Pointer(head))) == ^uintptr(0) {
			runtimeFatal("gc: cannot register finalizer")
		}
		KeepAlive(head)
	}
	if entry != nil {
		initFinalizerScheduler()
	}
	gcResumeWorld()
	gcLock.Unlock()
	KeepAlive(obj)
}

//export tinygo_runtime_bdwgc_finalizer
func boehmQueueFinalizer(obj unsafe.Pointer, data unsafe.Pointer) {
	for n := (*boehmFinalizer)(data); n != nil; {
		next := n.next
		numFinalizers--
		n.ptr = unsafe.Add(obj, n.offset)
		n.next = finalizerPending
		finalizerPending = n
		finalizerQueued = true
		n = next
	}
}

// Call with gcLock held. Callbacks only enqueue Go work; they never run user code.
func boehmInvokeFinalizers() bool {
	if numFinalizers != 0 && libgc_should_invoke_finalizers() != 0 {
		libgc_invoke_finalizers()
	}
	queued := finalizerQueued
	finalizerQueued = false
	return queued
}

func finalizerPressureGC() bool {
	gcLock.Lock()
	trigger := finalizerGCTrigger(numFinalizers)
	if trigger == 0 || finalizersSinceGC < trigger {
		gcLock.Unlock()
		return false
	}
	libgc_gcollect()
	finalizersSinceGC = 0
	queued := boehmInvokeFinalizers()
	gcResumeWorld()
	gcLock.Unlock()
	if queued {
		wakeFinalizer()
	}
	return true
}

func wakeFinalizer() {
	if hasScheduler || hasParallelism {
		gcLock.Lock()
		spawn := finalizerPending != nil && !finalizerRunnerStarted
		if spawn {
			finalizerRunnerStarted = true
		}
		gcLock.Unlock()
		if spawn {
			spawnFinalizerRunner()
		}
	} else {
		drainFinalizers()
	}
}

func drainFinalizers() {
	if finalizerDraining {
		return
	}
	finalizerDraining = true
	for {
		gcLock.Lock()
		n := finalizerPending
		if n != nil {
			finalizerPending = n.next
		}
		gcLock.Unlock()
		if n == nil {
			break
		}
		ptr, fn := n.ptr, n.fn
		*n = boehmFinalizer{}
		callFinalizer(ptr, fn)
	}
	finalizerDraining = false
}

func finalizerRunner() {
	for {
		drainFinalizers()
		gcLock.Lock()
		if finalizerPending == nil {
			finalizerRunnerStarted = false
			gcLock.Unlock()
			return
		}
		gcLock.Unlock()
	}
}

//export tinygo_runtime_bdwgc_register_finalizer
func libgc_register_finalizer(uintptr, uintptr) uintptr

//export GC_should_invoke_finalizers
func libgc_should_invoke_finalizers() int32

//export GC_invoke_finalizers
func libgc_invoke_finalizers() int32

//export GC_init
func libgc_init()

//export GC_malloc_kind
func libgc_malloc_kind(uintptr, int32) unsafe.Pointer

//export GC_malloc_atomic_uncollectable
func libgc_malloc_atomic_uncollectable(uintptr) unsafe.Pointer

//export tinygo_runtime_bdwgc_make_descriptor
func libgc_make_descriptor(uintptr) uintptr

//export GC_calloc_explicitly_typed
func libgc_calloc_explicitly_typed(uintptr, uintptr, uintptr) unsafe.Pointer

//export GC_malloc_explicitly_typed
func libgc_malloc_explicitly_typed(uintptr, uintptr) unsafe.Pointer

//export GC_free
func libgc_free(unsafe.Pointer)

//export GC_base
func libgc_base(ptr uintptr) uintptr

//export GC_size
func libgc_size(ptr uintptr) uintptr

//export GC_push_all
func libgc_push_all(bottom, top uintptr)

//export GC_push_all_eager
func libgc_push_all_eager(bottom, top uintptr)

//export GC_push_all_stack
func libgc_push_all_stack(bottom, top uintptr)

//export GC_gcollect
func libgc_gcollect()

//export GC_get_prof_stats
func libgc_get_prof_stats(*libgc_prof_stats, uintptr) uintptr

//export GC_set_push_other_roots
func libgc_set_push_other_roots(unsafe.Pointer)

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
