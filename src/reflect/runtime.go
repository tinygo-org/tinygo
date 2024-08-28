package reflect

// This file defines functions that logically belong in the runtime but that
// depend on the internals of the reflect package and are therefore better
// implemented directly in the reflect package.

import (
	"unsafe"
)

// Return whether the concrete type implements the asserted method set.
// This function is called both from the compiler and from the reflect package.
//
//go:linkname typeImplementsMethodSet runtime.typeImplementsMethodSet
func typeImplementsMethodSet(concreteType, assertedMethodSet unsafe.Pointer) bool {
	if concreteType == nil {
		// This can happen when doing a type assert on a nil interface value.
		// According to the Go language specification, a type assert on a nil
		// interface always fails.
		return false
	}

	const ptrSize = unsafe.Sizeof((*byte)(nil))
	itfNumMethod := *(*uintptr)(assertedMethodSet)
	if itfNumMethod == 0 {
		// While the compiler shouldn't emit such type asserts, the reflect
		// package may do so.
		// This is actually needed for correctness: for example a value of type
		// int may be assigned to a variable of type interface{}.
		return true
	}

	// Pull the method set out of the concrete type.
	var methods *methodSet
	metaByte := *(*uint8)(concreteType)
	if metaByte&flagNamed != 0 {
		concreteType := (*namedType)(concreteType)
		methods = &concreteType.methods
	} else if metaByte&kindMask == uint8(Struct) {
		concreteType := (*structType)(concreteType)
		methods = &concreteType.methods
	} else if metaByte&kindMask == uint8(Pointer) {
		concreteType := (*ptrType)(concreteType)
		methods = &concreteType.methods
	} else {
		// Other types don't have a method set.
		return false
	}
	concreteTypePtr := unsafe.Pointer(&methods.methods)
	concreteTypeEnd := unsafe.Add(concreteTypePtr, uintptr(methods.length)*ptrSize)

	// Iterate over each method in the interface method set, and check whether
	// the method exists in the method set of the concrete type.
	// Both method sets are sorted in the same way, so we can use a simple loop
	// to check for a match.
	assertedTypePtr := unsafe.Add(assertedMethodSet, ptrSize)
	assertedTypeEnd := unsafe.Add(assertedTypePtr, itfNumMethod*ptrSize)
	for assertedTypePtr != assertedTypeEnd {
		assertedMethod := *(*unsafe.Pointer)(assertedTypePtr)

		// Search for the same method in the concrete type.
		for {
			if concreteTypePtr == concreteTypeEnd {
				// Reached the end of the concrete type method set.
				// The method wasn't found, so the type assert failed.
				return false
			}
			concreteMethod := *(*unsafe.Pointer)(concreteTypePtr)
			concreteTypePtr = unsafe.Add(concreteTypePtr, ptrSize)
			if concreteMethod == assertedMethod {
				// Found the method in the concrete type. Continue with the
				// next in the interface method set.
				break
			}
		}

		// The method was found. Continue with the next.
		assertedTypePtr = unsafe.Add(assertedTypePtr, ptrSize)
	}

	// All methods in the interface were found.
	return true
}
