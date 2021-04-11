package runtime

// This file implements some data types that may be useful for some
// implementations of func values.

import (
	"unsafe"
)

// funcValue is the underlying type of func values, depending on which func
// value representation was used.
type funcValue struct {
	context unsafe.Pointer // function context, for closures and bound methods
	id      uintptr        // ptrtoint of *funcValueWithSignature before lowering, opaque index (non-0) after lowering
}

// funcValueWithSignature is used before the func lowering pass.
type funcValueWithSignature struct {
	funcPtr   uintptr // ptrtoint of the actual function pointer
	signature *uint8  // external *i8 with a name identifying the function signature
}

// getFuncPtr is a dummy function that may be used if the func lowering pass is
// not used. It is generally too slow but may be a useful fallback to debug the
// func lowering pass.
func getFuncPtr(val funcValue, signature *uint8) uintptr {
	return (*funcValueWithSignature)(unsafe.Pointer(val.id)).funcPtr
}
