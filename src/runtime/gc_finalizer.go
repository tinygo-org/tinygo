//go:build gc.conservative || gc.precise

package runtime

// This file implements a minimal runtime.SetFinalizer for the block-based
// garbage collector. It supports the common, contract-correct case only:
//
//   - SetFinalizer(ptr, func(ptrType)) registers a finalizer that runs once,
//     after the object becomes unreachable.
//   - SetFinalizer(ptr, nil) clears any finalizer for the object.
//
// It intentionally does not implement full Go finalizer semantics (ordering
// guarantees, cycles, AddCleanup, ...). The whole feature is zero-cost when no
// finalizer is ever registered: the table stays empty, scanFinalizers returns
// immediately, and no background goroutine is spawned.

import (
	"internal/task"
	"unsafe"
)

// finalizerEntry is one registered finalizer. The same node type is reused for
// the pending queue: when an object dies, its entry is spliced out of the
// registered list and into the pending list with pure pointer operations, so no
// allocation happens during a GC cycle.
type finalizerEntry struct {
	next *finalizerEntry
	// obj is the object address stored bitwise-NOT (see encodeFinalizerPtr).
	obj uintptr
	// fn is the finalizer func value. It is kept alive because the registered
	// list (a package global) is a GC root, so the boxed closure and any
	// captured state survive until the finalizer runs.
	fn interface{}
}

var (
	finalizers        *finalizerEntry // registered finalizers; a GC root that keeps fn values alive
	finalizerPending  *finalizerEntry // finalizers whose object died, waiting to run
	numFinalizers     uintptr         // number of registered finalizers; fast-path gate for scanFinalizers
	finalizersQueued  bool            // set when scanFinalizers queued at least one finalizer to run
	finalizerFutex    task.Futex      // wakes the finalizerRunner goroutine after a GC queues work
	finalizerDraining bool            // guards against re-entrant inline draining (scheduler.none)

	// finalizerRunnerStarted records whether the background finalizerRunner
	// goroutine has been spawned yet. The runner is spawned lazily, on the first
	// SetFinalizer, so builds that never register a finalizer let the linker DCE
	// the runner and drain machinery. Read/written only under gcLock, so no
	// atomics are needed. Unused under scheduler.none (spawnFinalizerRunner is a
	// no-op there, and the linker drops the flag).
	finalizerRunnerStarted bool
)

// The object address is stored bitwise-NOT so it never looks like a live heap
// pointer to the conservative scanner. Otherwise the entry would pin every
// finalizable object forever and the object could never be detected as dead.
// Under the precise GC a plain uintptr field is not scanned anyway, so the
// encoding is harmless there and required for the conservative build.
func encodeFinalizerPtr(addr uintptr) uintptr { return ^addr }
func decodeFinalizerPtr(enc uintptr) uintptr  { return ^enc }

// registerFinalizer records fn as the finalizer for the object at addr. A nil fn
// removes any registration for the object. Growing the table (allocating a node)
// is the only allocation and it happens here, on the caller, never during GC.
// gcLock also serializes table access against scanFinalizers, which runs under
// gcLock during a GC on another core/thread.
func registerFinalizer(addr uintptr, fn interface{}) {
	enc := encodeFinalizerPtr(addr)

	if fn == nil {
		// Clear: remove every registration for this object.
		gcLock.Lock()
		prev := &finalizers
		for n := *prev; n != nil; n = *prev {
			if n.obj == enc {
				*prev = n.next
				numFinalizers--
			} else {
				prev = &n.next
			}
		}
		gcLock.Unlock()
		return
	}

	// Register or replace. The allocation happens before gcLock is taken,
	// because alloc acquires gcLock itself.
	entry := &finalizerEntry{obj: enc, fn: fn}
	gcLock.Lock()
	for n := finalizers; n != nil; n = n.next {
		if n.obj == enc {
			// Replace the finalizer for an already-registered object, so it
			// still runs only once (Go SetFinalizer replace semantics).
			n.fn = fn
			// A finalizer is registered, so make sure the runner exists. The
			// flag is serialized by gcLock; the spawn itself allocates, so it
			// must run after the lock is released.
			spawn := !finalizerRunnerStarted
			finalizerRunnerStarted = true
			gcLock.Unlock()
			if spawn {
				spawnFinalizerRunner()
			}
			return
		}
	}
	entry.next = finalizers
	finalizers = entry
	numFinalizers++
	// A finalizer is registered, so make sure the runner exists. The flag is
	// serialized by gcLock; the spawn itself allocates, so it must run after the
	// lock is released.
	spawn := !finalizerRunnerStarted
	finalizerRunnerStarted = true
	gcLock.Unlock()
	if spawn {
		spawnFinalizerRunner()
	}
}

// scanFinalizers detects finalizable objects that became unreachable in the
// current GC cycle and queues their finalizers. It must be called under gcLock,
// after marking is complete and before sweep frees anything.
func scanFinalizers() {
	// Nothing registered and nothing waiting to run: fast path.
	if numFinalizers == 0 && finalizerPending == nil {
		return
	}

	// Detect newly-unreachable objects and move their finalizers to the pending
	// queue.
	prev := &finalizers
	for n := *prev; n != nil; n = *prev {
		addr := decodeFinalizerPtr(n.obj)
		if !isOnHeap(addr) {
			// Not a heap object we can track; keep it registered.
			prev = &n.next
			continue
		}
		if blockFromAddr(addr).findHead().state() == blockStateMark {
			// Still reachable; keep the finalizer for a later cycle.
			prev = &n.next
			continue
		}

		// The object is unreachable. Splice its entry out of the registered list
		// and into the pending queue (alloc-free), so its finalizer runs once.
		*prev = n.next
		numFinalizers--
		n.next = finalizerPending
		finalizerPending = n
		finalizersQueued = true
	}

	// Resurrect every object whose finalizer is still pending: both the deaths
	// found above and any queued by an earlier cycle that the runner has not
	// drained yet. Otherwise the next GC would not mark them (their only
	// reference is the encoded, scanner-invisible pending entry) and sweep would
	// free them out from under a finalizer that hasn't run — a use-after-free.
	// Walking the pending list is safe: scanFinalizers and dequeueFinalizer are
	// both serialized under gcLock.
	var resurrected bool
	for n := finalizerPending; n != nil; n = n.next {
		markRoot(0, decodeFinalizerPtr(n.obj))
		resurrected = true
	}
	if resurrected {
		// Re-scan so objects reachable only from resurrected objects also
		// survive this sweep.
		finishMark()
	}
}

// callFinalizer invokes a finalizer func value on the given object pointer.
func callFinalizer(objPtr unsafe.Pointer, fn interface{}) {
	// SetFinalizer already validated that fn is a func. A finalizer is
	// contractually func(ptrType), and func(*T) and func(unsafe.Pointer) are
	// ABI-identical in TinyGo (one pointer arg + trailing context, no result).
	// reflect.Value.Call is unimplemented, so reinterpret the boxed closure and
	// call it via the same closure-ABI indirect call the runtime uses elsewhere.
	fnBox := (*_interface)(unsafe.Pointer(&fn)).value
	f := *(*func(unsafe.Pointer))(fnBox)
	f(objPtr)
}

// drainFinalizers runs every queued finalizer, with gcLock released so the
// finalizers may allocate.
func drainFinalizers() {
	if finalizerDraining {
		// Re-entered from a finalizer that triggered a GC (only possible with
		// scheduler.none, which drains inline). Let the outer loop handle any
		// newly queued finalizers.
		return
	}
	finalizerDraining = true
	for {
		n, objPtr := dequeueFinalizer()
		if n == nil {
			break
		}
		callFinalizer(objPtr, n.fn)
	}
	finalizerDraining = false
}

// dequeueFinalizer pops the next pending finalizer. The pending list is shared
// with scanFinalizers (which runs under gcLock), so the pop is guarded by the
// same lock; the finalizer itself runs afterwards with the lock released.
//
// It also decodes the real object pointer while still holding gcLock and returns
// it. Once the entry leaves finalizerPending it is no longer in the kept-alive
// set, and the only remaining references are the encoded n.obj (invisible to the
// conservative scanner) and n.fn (which for a non-capturing finalizer does not
// reference the object). Materializing the pointer under the lock puts it on the
// caller's stack as a real GC root before any concurrent stop-the-world GC can
// run, so the object cannot be swept out from under callFinalizer.
func dequeueFinalizer() (*finalizerEntry, unsafe.Pointer) {
	gcLock.Lock()
	n := finalizerPending
	var objPtr unsafe.Pointer
	if n != nil {
		finalizerPending = n.next
		objPtr = unsafe.Pointer(decodeFinalizerPtr(n.obj))
	}
	gcLock.Unlock()
	return n, objPtr
}

// wakeFinalizer is called after a GC (with gcLock already released) that queued
// finalizers. On schedulers with goroutines it wakes the finalizerRunner; on
// scheduler.none it drains inline.
func wakeFinalizer() {
	if hasScheduler || hasParallelism {
		// A finalizerRunner exists. Bump the futex before waking so a runner
		// caught between draining and waiting doesn't miss this wakeup.
		finalizerFutex.Add(1)
		finalizerFutex.Wake()
	} else {
		// scheduler.none: no goroutines, so drain inline. Finalizers must not
		// block here; this is safe because gcLock has already been released.
		drainFinalizers()
	}
}

// finalizerRunner is the background goroutine that runs finalizers off the
// allocating goroutine's stack. It drains all pending finalizers, then blocks on
// the futex until the next GC queues more. It is spawned lazily by
// spawnFinalizerRunner on the first SetFinalizer, so builds that never register a
// finalizer let the linker eliminate it and the drain machinery entirely.
func finalizerRunner() {
	for {
		// Sample the futex before draining. A wake that lands after we drain but
		// before Wait then leaves the counter changed, so Wait returns at once
		// instead of losing the wakeup (at worst one harmless spurious re-drain).
		val := finalizerFutex.Load()
		drainFinalizers()
		finalizerFutex.Wait(val)
	}
}
