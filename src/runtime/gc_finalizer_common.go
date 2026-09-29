//go:build gc.conservative || gc.precise || gc.boehm

package runtime

import "unsafe"

// finalizerGCThreshold starts pressure GC when registrations indicate external memory pressure.
// Larger tables use a proportional threshold. Zero disables this trigger.
const finalizerGCThreshold = 32
const finalizerGCDivisor = 2

// finalizerGCTrigger scales the threshold so scan work stays proportional to registrations.
// It uses finalizerGCThreshold as the minimum.
func finalizerGCTrigger(count uintptr) uintptr {
	if finalizerGCThreshold == 0 {
		return 0
	}
	if proportional := count / finalizerGCDivisor; proportional > finalizerGCThreshold {
		return proportional
	}
	return finalizerGCThreshold
}

// A func(*T) and a func(unsafe.Pointer) have the same TinyGo closure ABI.
func callFinalizer(objPtr unsafe.Pointer, fn interface{}) {
	fnBox := (*_interface)(unsafe.Pointer(&fn)).value
	f := *(*func(unsafe.Pointer))(fnBox)
	f(objPtr)
}
