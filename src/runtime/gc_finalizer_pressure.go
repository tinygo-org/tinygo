//go:build (gc.conservative || gc.precise) && !runtime_nofinalizerpressure

package runtime

// finalizerGCThreshold starts pressure GC when registrations indicate external memory pressure.
// Larger tables use a proportional threshold. Zero disables this trigger.
const finalizerGCThreshold = 32
