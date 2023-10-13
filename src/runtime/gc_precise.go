//go:build gc.precise

// This implements the block-based GC as a partially precise GC. This means that
// for most heap allocations it is known which words contain a pointer and which
// don't. This should in theory make the GC faster (because it can skip
// non-pointer object) and have fewer false positives in a GC cycle. It does
// however use a bit more RAM to store the layout of each object.
//
// The pointer/non-pointer information for objects is stored in the first word
// of the object. It is described below but in essense it contains a bitstring
// of a particular size. This size does not indicate the size of the object:
// instead the allocated object is a multiple of the bitstring size. This is so
// that arrays and slices can store the size of the object efficiently. The
// bitstring indicates where the pointers are in the object (the bit is set when
// the value may be a pointer, and cleared when it certainly isn't a pointer).
// Some examples (assuming a 32-bit system for the moment):
//
// | object type | size | bitstring | note
// |-------------|------|-----------|------
// | int         | 1    | 0         | no pointers in this object
// | string      | 2    | 01        | {pointer, len} pair so there is one pointer
// | []int       | 3    | 001       | {pointer, len, cap}
// | [4]*int     | 1    | 1         | even though it contains 4 pointers, an array repeats so it can be stored with size=1
// | [30]byte    | 1    | 0         | there are no pointers so the layout is very simple
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

const preciseHeap = true

type gcObjectScanner struct {
	index      uintptr
	size       uintptr
	bitmap     uintptr
	bitmapAddr unsafe.Pointer
}

func newGCObjectScanner(block gcBlock) gcObjectScanner {
	if gcAsserts && block != block.findHead() {
		runtimePanic("gc: object scanner must start at head")
	}
	scanner := gcObjectScanner{}
	layout := *(*uintptr)(unsafe.Pointer(block.address()))
	if layout == 0 {
		// Unknown layout. Assume all words in the object could be pointers.
		// This layout value below corresponds to a slice of pointers like:
		//     make(*byte, N)
		scanner.size = 1
		scanner.bitmap = 1
	} else if layout&1 != 0 {
		// Layout is stored directly in the integer value.
		// Determine format of bitfields in the integer.
		const layoutBits = uint64(unsafe.Sizeof(layout) * 8)
		var sizeFieldBits uint64
		switch layoutBits { // note: this switch should be resolved at compile time
		case 16:
			sizeFieldBits = 4
		case 32:
			sizeFieldBits = 5
		case 64:
			sizeFieldBits = 6
		default:
			runtimePanic("unknown pointer size")
		}

		// Extract values from the bitfields.
		// See comment at the top of this file for more information.
		scanner.size = (layout >> 1) & (1<<sizeFieldBits - 1)
		scanner.bitmap = layout >> (1 + sizeFieldBits)
	} else {
		// Layout is stored separately in a global object.
		layoutAddr := unsafe.Pointer(layout)
		scanner.size = *(*uintptr)(layoutAddr)
		scanner.bitmapAddr = unsafe.Add(layoutAddr, unsafe.Sizeof(uintptr(0)))
	}
	return scanner
}

func (scanner *gcObjectScanner) pointerFree() bool {
	if scanner.bitmapAddr != nil {
		// While the format allows for large objects without pointers, this is
		// optimized by the compiler so if bitmapAddr is set, we know that there
		// are at least some pointers in the object.
		return false
	}
	// If the bitmap is zero, there are definitely no pointers in the object.
	return scanner.bitmap == 0
}

func (scanner *gcObjectScanner) nextIsPointer(word, parent, addrOfWord uintptr) bool {
	index := scanner.index
	scanner.index++
	if scanner.index == scanner.size {
		scanner.index = 0
	}

	if !isOnHeap(word) {
		// Definitely isn't a pointer.
		return false
	}

	// Might be a pointer. Now look at the object layout to know for sure.
	if scanner.bitmapAddr != nil {
		if (*(*uint8)(unsafe.Add(scanner.bitmapAddr, index/8))>>(index%8))&1 == 0 {
			return false
		}
		return true
	}
	if (scanner.bitmap>>index)&1 == 0 {
		// not a pointer!
		return false
	}

	// Probably a pointer.
	return true
}
