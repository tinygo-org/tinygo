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

// finalizerGCThreshold is the minimum net registration pressure accumulated
// since the last collection before the scheduler runs one at its idle point.
// finalizerGCTrigger raises this floor proportionally for larger tables.
// A registered finalizer almost always guards an external resource, most
// importantly a syscall/js bridge-table slot (js.Value or js.Func), that costs
// only a few bytes of Go heap but pins a whole JS object and its slot. Without
// this, a long-lived instance with a large resident heap defers GC (and thus
// finalizer draining) until the Go heap itself fills, which for a bursty,
// mostly-idle workload may be never, so the external resources accumulate
// without bound. Coupling a GC to finalizer-registration pressure bounds that
// accumulation.
//
// This is a compile-time policy constant, in the spirit of Go's forcegcperiod.
// The trigger only fires at the scheduler's idle point (a drained run queue) and
// each firing resets the count (see scanFinalizers), so it is throttled to that
// point rather than firing once per this-many registrations: a setup phase that
// registers many long-lived finalizers pays at most one extra collection at the
// first idle point after it, not one per threshold, and that collection just
// marks still-live data during otherwise-idle time without freeing anything
// early. Keeping it a const also lets the compiler constant-fold the check and,
// with zero, drop the pressure path entirely. Zero disables the trigger.
const finalizerGCThreshold = 32

var (
	finalizers        *finalizerEntry // registered finalizers; a GC root that keeps fn values alive
	finalizerPending  *finalizerEntry // finalizers whose object died, waiting to run
	numFinalizers     uintptr         // number of registered finalizers; fast-path gate for scanFinalizers
	finalizersSinceGC uintptr         // net registration pressure since the last GC; drives the scheduler idle-point trigger
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

// finalizerGCDivisor scales the registration trigger with the size of the
// table: the next collection is due after roughly numFinalizers/this much net
// pressure, never less than finalizerGCThreshold.
const finalizerGCDivisor = 2

// finalizerGCTrigger returns how much net registration pressure is needed to run
// the next collection. Each collection scans the whole table, which
// costs O(numFinalizers), so a trigger that stays constant while the table grows
// makes N registrations cost O(N^2) in scanning alone. Scaling the trigger with
// the table keeps the amortized scan cost per registration constant, the same
// reasoning behind Go's proportional GOGC pacing: collect when the tracked set
// has grown by a fraction of itself, not by a fixed count.
//
// The floor keeps the original behaviour for small tables, where a proportional
// trigger would fire too rarely to be useful.
func finalizerGCTrigger() uintptr {
	if finalizerGCThreshold == 0 {
		return 0
	}
	if proportional := numFinalizers / finalizerGCDivisor; proportional > finalizerGCThreshold {
		return proportional
	}
	return finalizerGCThreshold
}

// finalizerBits records, one bit per heap block, whether the object starting at
// that block already has a registered finalizer. It answers the "is this object
// already registered?" question that SetFinalizer's replace semantics require
// without walking the table, so the common case (a fresh object, which is every
// syscall/js value) never scans anything.
//
// This mirrors what upstream Go gets from its per-span specials plus the
// arena-level "span has specials" bitmap: a constant-time way to skip objects
// that have nothing registered.
//
// The bitmap is allocated on the first registration and grown with the heap, so
// a program that never registers a finalizer keeps the whole feature dead.
//
// Every access goes through gcLock, including the reads. The slice header itself
// is replaced when the heap grows, so an unlocked reader on a parallel scheduler
// (cores, threads) could observe a stale bit, or tear the header and index the
// old, shorter buffer with the new length.
var finalizerBits []byte

// finalizerBitsShortfall returns the bitmap length needed to cover the current
// heap, or zero if the current bitmap already covers it. It must be called under
// gcLock: that is what makes reading finalizerBits and endBlock safe against a
// concurrent adoptFinalizerBits on another core. The caller then allocates with
// the lock released (allocating takes gcLock) and installs the result with
// adoptFinalizerBits.
func finalizerBitsShortfall() uintptr {
	need := (uintptr(endBlock) + 7) / 8
	if uintptr(len(finalizerBits)) >= need {
		return 0
	}
	return need
}

// adoptFinalizerBits installs a wider bitmap under gcLock, carrying the old bits
// over. A nil or already-obsolete buffer is ignored, which is what makes it safe
// for the heap to have grown again (or another core to have installed its own
// wider bitmap) while the caller was allocating with the lock released. A buffer
// that covers less than the current heap is still an improvement: addresses past
// its end just keep answering conservatively in finalizerBitGet.
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
		// The bitmap does not describe this address yet (the heap grew since it
		// was sized). Answer conservatively: a spurious "maybe" only costs one
		// scan, while a wrong "no" would let a second entry be registered for an
		// object that already has one, and its finalizer would run twice.
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

// finalizerRegistered reports whether the table holds an entry for enc. This is
// the linear answer the registration bitmap exists to avoid, so it is only used
// under gcAsserts, to check the bitmap against the table it summarizes.
func finalizerRegistered(enc uintptr) bool {
	for n := finalizers; n != nil; n = n.next {
		if n.obj == enc {
			return true
		}
	}
	return false
}

// assertFinalizerTable verifies the bookkeeping that the table, the counter and
// the registration bitmap have to agree on. A registered entry without its bit
// is the dangerous direction: the clear and replace paths trust a clear bit to
// mean "nothing registered" and skip the table walk, so a missing bit turns
// SetFinalizer(obj, nil) into a silent no-op and lets a re-registration add a
// second entry, which runs the finalizer twice. Only called under gcAsserts.
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
		// Clear: remove every registration for this object. The bit proves in
		// one test that there is nothing to remove, but only while gcLock is
		// held: a registration on another core may be setting that same bit (and
		// replacing the bitmap) right now, and a stale read of zero would skip
		// the removal and leave the finalizer registered on a live object.
		// Holding the lock for the check costs nothing extra, because the removal
		// below needs it anyway; what the bit saves is the O(numFinalizers) walk.
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

	// Register or replace. Allocating acquires gcLock, so the entry is allocated
	// before the lock is taken and a wider bitmap is allocated by dropping the
	// lock for just that call. Only a heap that outgrew the bitmap pays that
	// round trip; the common case holds the lock once, and adoptFinalizerBits
	// tolerates the heap having grown again (or another core having installed a
	// wider bitmap) while this one was allocating.
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
	// Only an object whose bit is set can already be in the table, so a fresh
	// object skips the scan entirely. An address the bitmap cannot describe
	// (not on the heap) always scans, as before.
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
	finalizersSinceGC++ // pressure signal for the proactive GC trigger at the scheduler's idle point
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
	// A collection is running now, so reset the registration-pressure counter
	// that drives the proactive idle-point trigger, regardless of whether any
	// finalizer is registered or fires this cycle.
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
		// The object is gone; clear its bit so a later object reusing the
		// address starts clean.
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
			// Every pending object must have survived the resurrection above.
			// One that did not is about to be swept while its finalizer is
			// still queued, which is a use-after-free in callFinalizer.
			for n := finalizerPending; n != nil; n = n.next {
				// Inside the collection, so the resurrected object is expected
				// to carry the mark state rather than plain head.
				if blockFromAddr(decodeFinalizerPtr(n.obj)).state() != blockStateMark {
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
			// scanFinalizers resurrects everything still pending, so sweep must
			// have left the object allocated. A freed block here means
			// callFinalizer is about to run on memory that is back in the free
			// list. The mark bit is not the thing to check: this runs outside a
			// collection, where unmark has already turned mark back into head.
			if blockFromAddr(addr).state() == blockStateFree {
				runtimeFatal("gc: dequeued finalizer for a freed object")
			}
		}
		objPtr = unsafe.Pointer(addr)
	}
	gcLock.Unlock()
	return n, objPtr
}

// finalizerPressureGC collects when net registration pressure since the last GC
// reaches finalizerGCTrigger, then hands any freshly-queued finalizers to the
// runner. It reports whether it collected. A registered
// finalizer almost always guards an external resource whose Go-heap cost is tiny
// (a few bytes) relative to what it pins, so the registration count is a proxy
// for external memory pressure that the heap-size GC trigger cannot see.
//
// It is installed as the cooperative scheduler's idle hook by the first
// SetFinalizer (see spawnFinalizerRunner) and called only from the scheduler's
// drained-runqueue point, where no goroutine is running on its own stack. That
// reclaims a completed run of goroutines' now-dead values in a single pass,
// rather than forcing a collection synchronously inside alloc while an
// operation's values are still live, which would scale GC frequency with
// allocation churn and waste most collections on still-live values.
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
