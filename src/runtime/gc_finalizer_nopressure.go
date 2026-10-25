//go:build (gc.conservative || gc.precise) && runtime_nofinalizerpressure

package runtime

// finalizerGCThreshold is zero with the build tag runtime_nofinalizerpressure:
// finalizer registrations do not start a collection, and finalizers run after
// the collections the heap itself needs. For programs that register many
// finalizers and go idle often, such as syscall/js programs, where every
// js.Value registers one and every await is an idle.
const finalizerGCThreshold = 0
