//go:build gc.precise

// This implements the block-based GC as a partially precise GC. This means that
// for most heap allocations it is known which words contain a pointer and which
// don't. This should in theory make the GC faster (because it can skip
// non-pointer object) and have fewer false positives in a GC cycle. It does
// however use a bit more RAM to store the layout of each object.
//
// The pointer/non-pointer information for objects is stored in the first word
// of the object. It is described below but in essence it contains a bitstring
// of a particular size. This size does not indicate the size of the object:
// instead the allocated object is a multiple of the bitstring size. This is so
// that arrays and slices can store the size of the object efficiently. The
// bitstring indicates where the pointers are in the object (the bit is set when
// the value may be a pointer, and cleared when it certainly isn't a pointer).
// Some examples (assuming a 32-bit system for the moment):
//
// | object type | size | bitstring | note
// |-------------|------|-----------|------
// | int         | 1    |   0       | no pointers in this object
// | string      | 2    |  01       | {pointer, len} pair so there is one pointer
// | []int       | 3    | 001       | {pointer, len, cap}
// | [4]*int     | 1    |   1       | even though it contains 4 pointers, an array repeats so it can be stored with size=1
// | [30]byte    | 1    |   0       | there are no pointers so the layout is very simple
//
// The garbage collector scans objects by starting at the first word value in
// the object. If the least significant bit of the bitstring is clear, it is
// skipped (it's not a pointer). If the bit is set, it is treated as if it could
// be a pointer. The garbage collector continues by scanning further words in
// the object and checking them against the corresponding bit in the bitstring.
// Once it reaches the end of the bitstring, it wraps around (for arrays,
// slices, strings, etc).
//
// The layout as passed to the runtime.alloc function and stored in the object
// is a pointer-sized value. If the least significant bit of the value is set,
// the bitstring is contained directly inside the value, of the form
// pppp_pppp_ppps_sss1.
//   * The 'p' bits indicate which parts of the object are a pointer.
//   * The 's' bits indicate the size of the object. In this case, there are 11
//     pointer bits so four bits are enough for the size (0-15).
//   * The lowest bit is always set to distinguish this value from a pointer.
// This example is for a 16-bit architecture. For example, 32-bit architectures
// use a layout format of pppppppp_pppppppp_pppppppp_ppsssss1 (26 bits for
// pointer/non-pointer information, 5 size bits, and one bit that's always set).
//
// For larger objects that don't fit in an uintptr, the layout value is a
// pointer to a global with a format as follows:
//     struct {
//         size uintptr
//         bits [...]uint8
//     }
// The 'size' field is the number of bits in the bitstring. The 'bits' field is
// a byte array that contains the bitstring itself, in little endian form. The
// length of the bits array is ceil(size/8).

package runtime

import "unsafe"

const sizeFieldBits = 4 + (unsafe.Sizeof(uintptr(0)) / 4)

// parseGCLayout stores the layout information passed to alloc into a gcLayout value.
func parseGCLayout(layout unsafe.Pointer) gcLayout {
	return gcLayout(layout)
}

// gcLayout tracks pointer locations in a heap object.
type gcLayout uintptr

func (layout gcLayout) pointerFree() bool {
	return layout&1 != 0 && layout>>(sizeFieldBits+1) == 0
}

// scan an object with this element layout.
// The starting address must be valid and pointer-aligned.
// The length is rounded down to a multiple of the element size.
func (layout gcLayout) scan(start, len uintptr) {
	switch {
	case layout == 0:
		// This is an unknown layout.
		// Scan conservatively.
		// NOTE: This is *NOT* equivalent to a slice of pointers on AVR.
		scanConservative(start, len)

	case layout&1 != 0:
		// The layout is stored directly in the integer value.
		// Extract the bitfields.
		size := uintptr(layout>>1) & (1<<sizeFieldBits - 1)
		mask := uintptr(layout) >> (1 + sizeFieldBits)

		// Scan with the extracted mask.
		scanSimple(start, len, size*unsafe.Alignof(start), mask)

	default:
		// The layout is stored seperately in a global object.
		// Extract the size and bitmap.
		layoutAddr := unsafe.Pointer(layout)
		size := *(*uintptr)(layoutAddr)
		bitmapPtr := unsafe.Add(layoutAddr, unsafe.Sizeof(uintptr(0)))
		bitmapLen := (size + 7) / 8
		bitmap := unsafe.Slice((*byte)(bitmapPtr), bitmapLen)

		// Scan with the bitmap.
		scanComplex(start, len, size*unsafe.Alignof(start), bitmap)
	}
}

// scanSimple scans an object with an integer bitmask of pointer locations.
// The starting address must be valid and pointer-aligned.
func scanSimple(start, len, size, mask uintptr) {
	for len >= size {
		// Scan this element.
		scanWithMask(start, mask)

		// Move to the next element.
		start += size
		len -= size
	}
}

// scanComplex scans an object with a bitmap of pointer locations.
// The starting address must be valid and pointer-aligned.
func scanComplex(start, len, size uintptr, bitmap []byte) {
	for len >= size {
		// Scan this element.
		for i, mask := range bitmap {
			addr := start + 8*unsafe.Alignof(start)*uintptr(i)
			scanWithMask(addr, uintptr(mask))
		}

		// Move to the next element.
		start += size
		len -= size
	}
}

// scanWithMask scans a portion of an object with a mask of pointer locations.
// The address must be valid and pointer-aligned.
func scanWithMask(addr, mask uintptr) {
	// TODO: use ctz when available
	for mask != 0 {
		if mask&1 != 0 {
			// Load and mark this pointer.
			root := *(*uintptr)(unsafe.Pointer(addr))
			markRoot(addr, root)
		}

		// Move to the next offset.
		mask >>= 1
		addr += unsafe.Alignof(addr)
	}
}
