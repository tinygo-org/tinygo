package reflect

import (
	"unsafe"
)

// This stores a varint for each named type. Named types are identified by their
// name instead of by their type. The named types stored in this struct are
// non-basic types: pointer, struct, and channel.
//
//go:extern reflect.namedNonBasicTypesSidetable
var namedNonBasicTypesSidetable uintptr

//go:extern reflect.structTypesSidetable
var structTypesSidetable byte

//go:extern reflect.structNamesSidetable
var structNamesSidetable byte

//go:extern reflect.arrayTypesSidetable
var arrayTypesSidetable byte

// readStringSidetable reads a string from the given table (like
// structNamesSidetable) and returns this string. No heap allocation is
// necessary because it makes the string point directly to the raw bytes of the
// table.
func readStringSidetable(table unsafe.Pointer, index uintptr) string {
	nameLen, namePtr := readVarint(unsafe.Pointer(uintptr(table) + index))
	return *(*string)(unsafe.Pointer(&stringHeader{
		data: namePtr,
		len:  nameLen,
	}))
}

// readVarint decodes a varint as used in the encoding/binary package.
// It has an input pointer and returns the read varint and the pointer
// incremented to the next field in the data structure, just after the varint.
//
// Details:
// https://github.com/golang/go/blob/e37a1b1c/src/encoding/binary/varint.go#L7-L25
func readVarint(buf unsafe.Pointer) (uintptr, unsafe.Pointer) {
	var n uintptr
	shift := uintptr(0)
	for {
		// Read the next byte in the buffer.
		c := *(*byte)(buf)

		// Decode the bits from this byte and add them to the output number.
		n |= uintptr(c&0x7f) << shift
		shift += 7

		// Increment the buf pointer (pointer arithmetic!).
		buf = unsafe.Pointer(uintptr(buf) + 1)

		// Check whether this is the last byte of this varint. The upper bit
		// (msb) indicates whether any bytes follow.
		if c>>7 == 0 {
			return n, buf
		}
	}
}
