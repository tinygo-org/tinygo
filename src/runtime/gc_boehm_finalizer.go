//go:build gc.boehm

package runtime

// This file implements runtime.SetFinalizer with bdwgc finalization.
// It has the same limits as the block GC version in gc_finalizer.go.
// The code that runs the finalizers is a copy of the code in that file.

import (
	"internal/reflectlite"
	"internal/task"
	"unsafe"
)

// finalizerGCThreshold starts pressure GC when registrations indicate external memory pressure.
// Larger tables use a proportional threshold. Zero disables this trigger.
const finalizerGCThreshold = 32

const finalizerGCDivisor = 2

// finalizerEntry is the client data that bdwgc keeps for one finalizer.
// The first three fields match struct finalizer_entry in gc_boehm.c.
type finalizerEntry struct {
	next *finalizerEntry
	// obj is nil until the object is unreachable, so it does not keep it alive.
	obj unsafe.Pointer
	// offset is the distance from the start of the allocation to the pointer.
	offset uintptr
	fn     interface{}
}

var (
	finalizerPending  *finalizerEntry // finalizers that wait to run, a GC root
	numFinalizers     uintptr         // number of registered finalizers
	finalizersSinceGC uintptr         // tracks registration pressure for the scheduler trigger
	finalizerGCNo     uintptr         // collection count when finalizersSinceGC was reset
	finalizerFutex    task.Futex      // wakes the finalizerRunner goroutine
	finalizerDraining bool            // guards against re-entrant inline draining (scheduler.none)

	// Read and written only under gcLock.
	finalizerRunnerStarted bool
	finalizersQueued       bool // set by bdwgc when a collection queued a finalizer

	// The first SetFinalizer installs this so unused code can be removed.
	// It is written under gcLock and is set before finalizersQueued can be.
	finalizerCollect func()
)

func SetFinalizer(obj interface{}, finalizer interface{}) {
	if reflectlite.ValueOf(obj).Kind() != reflectlite.Pointer {
		runtimeFatal("runtime.SetFinalizer: first argument is not a pointer")
	}
	if finalizer != nil && reflectlite.ValueOf(finalizer).Kind() != reflectlite.Func {
		runtimeFatal("runtime.SetFinalizer: second argument is not a function")
	}

	objPtr := (*_interface)(unsafe.Pointer(&obj)).value
	if objPtr == nil {
		return
	}

	// Allocate before taking gcLock because allocation also takes this lock.
	var entry *finalizerEntry
	if finalizer != nil {
		entry = &finalizerEntry{fn: finalizer}
	}

	gcLock.Lock()
	base := libgc_base(uintptr(objPtr))
	if base == 0 {
		// The object is not on the heap, so it is never collected.
		gcLock.Unlock()
		return
	}
	resetFinalizerPressure()
	spawn := entry != nil && !finalizerRunnerStarted
	if spawn {
		// Enable the queue first so bdwgc never runs a finalizer by itself.
		finalizerRunnerStarted = true
		finalizerCollect = collectFinalizers
		libgc_enable_finalizers(
			uintptr(unsafe.Pointer(&finalizerPending)),
			uintptr(unsafe.Pointer(&finalizersQueued)),
		)
	}
	if entry != nil {
		entry.offset = uintptr(objPtr) - base
	}
	old := libgc_register_finalizer(base, uintptr(unsafe.Pointer(entry)))
	gcResumeWorld()
	switch {
	case entry != nil && old == 0:
		numFinalizers++
		finalizersSinceGC++
	case entry == nil && old != 0:
		numFinalizers--
		if finalizersSinceGC != 0 {
			finalizersSinceGC--
		}
	}
	queued := finalizersQueued
	gcLock.Unlock()

	if spawn {
		spawnFinalizerRunner()
	}
	if queued {
		collectFinalizers()
	}
}

// resetFinalizerPressure clears the pressure count after a collection.
// Call it with gcLock held.
func resetFinalizerPressure() {
	if no := libgc_get_gc_no(); no != finalizerGCNo {
		finalizerGCNo = no
		finalizersSinceGC = 0
	}
}

// finalizerGCTrigger scales the threshold so GC work stays proportional to registrations.
func finalizerGCTrigger() uintptr {
	if finalizerGCThreshold == 0 {
		return 0
	}
	if proportional := numFinalizers / finalizerGCDivisor; proportional > finalizerGCThreshold {
		return proportional
	}
	return finalizerGCThreshold
}

// finalizerPressureGC runs a GC when registrations indicate external memory pressure.
func finalizerPressureGC() bool {
	trigger := finalizerGCTrigger()
	if trigger == 0 {
		return false
	}
	gcLock.Lock()
	resetFinalizerPressure()
	pressure := finalizersSinceGC
	gcLock.Unlock()
	if pressure < trigger {
		return false
	}
	GC()
	return true
}

// collectFinalizers moves the finalizers that bdwgc queued to finalizerPending.
// The C code in gc_boehm.c fills the list. Call it without gcLock.
func collectFinalizers() {
	gcLock.Lock()
	finalizersQueued = false
	libgc_invoke_finalizers()
	gcResumeWorld()
	gcLock.Unlock()
	wakeFinalizer()
}

func dequeueFinalizer() (fn interface{}, objPtr unsafe.Pointer, ok bool) {
	gcLock.Lock()
	n := finalizerPending
	if n != nil {
		finalizerPending = n.next
		n.next = nil
		numFinalizers--
		fn, objPtr, ok = n.fn, n.obj, true
	}
	gcLock.Unlock()
	return fn, objPtr, ok
}

// callFinalizer invokes a finalizer func value on the given object pointer.
// See callFinalizer in gc_finalizer.go for the reason this cast is valid.
func callFinalizer(objPtr unsafe.Pointer, fn interface{}) {
	fnBox := (*_interface)(unsafe.Pointer(&fn)).value
	f := *(*func(unsafe.Pointer))(fnBox)
	f(objPtr)
}

// drainFinalizers runs every queued finalizer, with gcLock released so the
// finalizers may allocate.
func drainFinalizers() {
	if finalizerDraining {
		// A finalizer caused a GC under scheduler.none. The outer loop continues.
		return
	}
	finalizerDraining = true
	for {
		fn, objPtr, ok := dequeueFinalizer()
		if !ok {
			break
		}
		callFinalizer(objPtr, fn)
	}
	finalizerDraining = false
}

// wakeFinalizer wakes the finalizerRunner, or drains inline under scheduler.none.
// Call it without gcLock.
func wakeFinalizer() {
	if hasScheduler || hasParallelism {
		// Change the futex first so a runner that is about to wait sees it.
		finalizerFutex.Add(1)
		finalizerFutex.Wake()
	} else {
		drainFinalizers()
	}
}

// finalizerRunner is the goroutine that runs finalizers.
// The first SetFinalizer starts it.
func finalizerRunner() {
	for {
		// Read the futex first so a wake during the drain is not lost.
		val := finalizerFutex.Load()
		drainFinalizers()
		finalizerFutex.Wait(val)
	}
}

//export tinygo_runtime_bdwgc_enable_finalizers
func libgc_enable_finalizers(pending, queued uintptr)

// The compiler assumes that pointer parameters of exported functions do not
// escape, so the entry is passed as an integer to keep it on the heap.
//
//export tinygo_runtime_bdwgc_register_finalizer
func libgc_register_finalizer(obj, entry uintptr) uintptr

//export GC_invoke_finalizers
func libgc_invoke_finalizers() int32

//export GC_get_gc_no
func libgc_get_gc_no() uintptr
