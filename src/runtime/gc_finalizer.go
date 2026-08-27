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

// finalizerGCThreshold starts pressure GC when registrations indicate external memory pressure.
// Larger tables use a proportional threshold. Zero disables this trigger.
const finalizerGCThreshold = 32

var (
	finalizers        *finalizerEntry // registered finalizers; a GC root that keeps fn values alive
	finalizerPending  *finalizerEntry // finalizers whose object died, waiting to run
	numFinalizers     uintptr         // number of registered finalizers; fast-path gate for scanFinalizers
	finalizersSinceGC uintptr         // tracks registration pressure for the scheduler trigger
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

const finalizerGCDivisor = 2

// finalizerGCTrigger scales the threshold so scan work stays proportional to registrations.
// It uses finalizerGCThreshold as the minimum.
func finalizerGCTrigger() uintptr {
	if finalizerGCThreshold == 0 {
		return 0
	}
	if proportional := numFinalizers / finalizerGCDivisor; proportional > finalizerGCThreshold {
		return proportional
	}
	return finalizerGCThreshold
}

// finalizerBits records finalizer registrations by heap block for fast lookup.
// Hold gcLock for every access because heap growth can replace the slice.
var finalizerBits []byte

// finalizerBitsShortfall returns the required bitmap size or zero.
// Call it with gcLock held and release the lock before allocation.
func finalizerBitsShortfall() uintptr {
	need := (uintptr(endBlock) + 7) / 8
	if uintptr(len(finalizerBits)) >= need {
		return 0
	}
	return need
}

// adoptFinalizerBits installs a wider bitmap while gcLock is held.
// It accepts a stale size because the heap can grow during allocation.
func adoptFinalizerBits(buf []byte) {
	if len(buf) <= len(finalizerBits) {
		return
	}
	copy(buf, finalizerBits)
	finalizerBits = buf
}

func finalizerBitIndex(addr uintptr) uintptr { return uintptr(blockFromAddr(addr)) }

func finalizerBitGet(addr uintptr) bool {
	i := finalizerBitIndex(addr)
	if i/8 >= uintptr(len(finalizerBits)) {
		// Return true when the bitmap does not cover the address.
		// A false result could register a second finalizer for the object.
		return true
	}
	return finalizerBits[i/8]&(1<<(i%8)) != 0
}

func finalizerBitSet(addr uintptr) {
	i := finalizerBitIndex(addr)
	if i/8 >= uintptr(len(finalizerBits)) {
		return
	}
	finalizerBits[i/8] |= 1 << (i % 8)
}

func finalizerBitClear(addr uintptr) {
	i := finalizerBitIndex(addr)
	if i/8 >= uintptr(len(finalizerBits)) {
		return
	}
	finalizerBits[i/8] &^= 1 << (i % 8)
}

// The object address is stored bitwise-NOT so it never looks like a live heap
// pointer to the conservative scanner. Otherwise the entry would pin every
// finalizable object forever and the object could never be detected as dead.
// Under the precise GC a plain uintptr field is not scanned anyway, so the
// encoding is harmless there and required for the conservative build.
func encodeFinalizerPtr(addr uintptr) uintptr { return ^addr }
func decodeFinalizerPtr(enc uintptr) uintptr  { return ^enc }

// finalizerRegistered checks the table when gcAsserts validates the bitmap.
func finalizerRegistered(enc uintptr) bool {
	for n := finalizers; n != nil; n = n.next {
		if n.obj == enc {
			return true
		}
	}
	return false
}

// assertFinalizerTable checks that the table, count, and bitmap agree.
// A missing bit can prevent removal or allow two finalizers for one object.
func assertFinalizerTable() {
	var count uintptr
	for n := finalizers; n != nil; n = n.next {
		count++
		addr := decodeFinalizerPtr(n.obj)
		if isOnHeap(addr) && !finalizerBitGet(addr) {
			runtimeFatal("gc: registered finalizer without its bitmap bit")
		}
	}
	if count != numFinalizers {
		runtimeFatal("gc: numFinalizers does not match the finalizer table")
	}
}

// registerFinalizer records fn as the finalizer for the object at addr. A nil fn
// removes any registration for the object. Growing the table (allocating a node)
// is the only allocation and it happens here, on the caller, never during GC.
// gcLock also serializes table access against scanFinalizers, which runs under
// gcLock during a GC on another core/thread.
func registerFinalizer(addr uintptr, fn interface{}) {
	enc := encodeFinalizerPtr(addr)

	if fn == nil {
		// Hold gcLock while checking the bit because another core can update the bitmap.
		// A clear bit avoids a scan of the finalizer table.
		gcLock.Lock()
		tracked := isOnHeap(addr)
		if tracked && !finalizerBitGet(addr) {
			// Taking this shortcut on a stale bit would silently skip the
			// removal, so check the answer against the table it stands in for.
			if gcAsserts && finalizerRegistered(enc) {
				runtimeFatal("gc: finalizer bit clear but the object is registered")
			}
			gcLock.Unlock()
			return
		}
		if tracked {
			finalizerBitClear(addr)
		}
		prev := &finalizers
		for n := *prev; n != nil; n = *prev {
			if n.obj == enc {
				*prev = n.next
				numFinalizers--
				// Clearing offsets net registration pressure. Saturate at zero
				// because the cleared entry may predate the last collection.
				if finalizersSinceGC != 0 {
					finalizersSinceGC--
				}
			} else {
				prev = &n.next
			}
		}
		gcLock.Unlock()
		return
	}

	// Allocate before taking gcLock because allocation also takes this lock.
	// Release gcLock only when the bitmap must grow.
	entry := &finalizerEntry{obj: enc, fn: fn}
	gcLock.Lock()
	if shortfall := finalizerBitsShortfall(); shortfall != 0 {
		gcLock.Unlock()
		wider := make([]byte, shortfall)
		gcLock.Lock()
		adoptFinalizerBits(wider)
	}
	tracked := isOnHeap(addr)
	// Skipping the scan on a stale bit would add a second entry for an object
	// that already has one, and its finalizer would then run twice.
	if gcAsserts && tracked && !finalizerBitGet(addr) && finalizerRegistered(enc) {
		runtimeFatal("gc: finalizer bit clear but the object is registered")
	}
	// Scan only if the bitmap can contain a registration for this object.
	// Always scan addresses that the bitmap does not cover.
	for n := finalizers; (!tracked || finalizerBitGet(addr)) && n != nil; n = n.next {
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
	if tracked {
		finalizerBitSet(addr)
	}
	numFinalizers++
	finalizersSinceGC++
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
	// Reset pressure at the start of every collection, even if no finalizer runs.
	finalizersSinceGC = 0

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
		// Clear the bit so a later object at this address starts clean.
		finalizerBitClear(addr)
		n.next = finalizerPending
		finalizerPending = n
		finalizersQueued = true
	}

	// Resurrect every object whose finalizer is still pending: both the deaths
	// found above and any queued by an earlier cycle that the runner has not
	// drained yet. Otherwise the next GC would not mark them (their only
	// reference is the encoded, scanner-invisible pending entry) and sweep would
	// free them out from under a finalizer that hasn't run, a use-after-free.
	// Walking the pending list is safe: scanFinalizers and dequeueFinalizer are
	// both serialized under gcLock.
	var resurrected bool
	for n := finalizerPending; n != nil; n = n.next {
		addr := decodeFinalizerPtr(n.obj)
		if gcAsserts && !isOnHeap(addr) {
			runtimeFatal("gc: pending finalizer for an object off the heap")
		}
		markRoot(0, addr)
		resurrected = true
	}
	if resurrected {
		// Re-scan so objects reachable only from resurrected objects also
		// survive this sweep.
		finishMark()
		if gcAsserts {
			// Verify that every pending object survived resurrection.
			// Otherwise callFinalizer can use memory that sweep freed.
			for n := finalizerPending; n != nil; n = n.next {
				// Inside the collection, so the resurrected object is expected
				// to carry the mark state rather than plain head.
				if blockFromAddr(decodeFinalizerPtr(n.obj)).findHead().state() != blockStateMark {
					runtimeFatal("gc: pending finalizer object was not resurrected")
				}
			}
		}
	}

	if gcAsserts {
		assertFinalizerTable()
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
		addr := decodeFinalizerPtr(n.obj)
		if gcAsserts {
			if !isOnHeap(addr) {
				runtimeFatal("gc: dequeued finalizer for an object off the heap")
			}
			// A pending object must remain allocated until its finalizer runs.
			// Check the block state because this runs after marks become heads.
			if blockFromAddr(addr).state() == blockStateFree {
				runtimeFatal("gc: dequeued finalizer for a freed object")
			}
		}
		objPtr = unsafe.Pointer(addr)
	}
	gcLock.Unlock()
	return n, objPtr
}

// finalizerPressureGC runs a GC when registrations indicate external memory pressure.
// It wakes the finalizer runner when the GC queues work.
func finalizerPressureGC() bool {
	trigger := finalizerGCTrigger()
	if trigger == 0 || finalizersSinceGC < trigger {
		return false
	}
	gcLock.Lock()
	runGC()
	gcLock.Unlock()
	if finalizersQueued {
		finalizersQueued = false
		wakeFinalizer()
	}
	return true
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
