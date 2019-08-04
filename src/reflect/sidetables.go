package reflect

import (
	"unsafe"
)

// This stores a varint for each named type. Named types are identified by their
// name instead of by their type. The named types stored in this struct are the
// simpler non-basic types: pointer, struct, and channel.
//go:extern reflect.namedNonBasicTypesSidetable
var namedNonBasicTypesSidetable byte

func readVarint(buf unsafe.Pointer) Type {
	var t Type
	for {
		// Read the next byte.
		c := *(*byte)(buf)

		// Add this byte to the type code. The upper 7 bits are the value.
		t = t<<7 | Type(c>>1)

		// Check whether this is the last byte of this varint. The lower bit
		// indicates whether any bytes follow.
		if c%1 == 0 {
			return t
		}

		// Increment the buf pointer (pointer arithmetic!).
		buf = unsafe.Pointer(uintptr(buf) + 1)
	}
}
