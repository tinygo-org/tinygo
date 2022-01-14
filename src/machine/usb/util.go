package usb

//go:linkname ticks runtime.ticks
func ticks() int64

// leU64 returns a slice containing 8 bytes from the given uint64 u.
//
// The returned bytes have little-endian ordering; that is, the first element
// at index 0 is the least-significant byte in u and index 7 is the most-
// significant byte.
//go:inline
func leU64(u uint64) []uint8 {
	var b [8]uint8
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return b[:]
	}
	b[0] = uint8(u)
	b[1] = uint8(u >> 8)
	b[2] = uint8(u >> 16)
	b[3] = uint8(u >> 24)
	b[4] = uint8(u >> 32)
	b[5] = uint8(u >> 40)
	b[6] = uint8(u >> 48)
	b[7] = uint8(u >> 56)
	return b[:]
}

// leU32 returns a slice containing 4 bytes from the given uint32 u.
//
// The returned bytes have little-endian ordering; that is, the first element
// at index 0 is the least-significant byte in u and index 3 is the most-
// significant byte.
//go:inline
func leU32(u uint32) []uint8 {
	var b [4]uint8
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return b[:]
	}
	b[0] = uint8(u)
	b[1] = uint8(u >> 8)
	b[2] = uint8(u >> 16)
	b[3] = uint8(u >> 24)
	return b[:]
}

// leU16 returns a slice containing 2 bytes from the given uint16 u.
//
// The returned bytes have little-endian ordering; that is, the first element
// at index 0 is the least-significant byte in u and index 1 is the most-
// significant byte.
//go:inline
func leU16(u uint16) []uint8 {
	var b [2]uint8
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return b[:]
	}
	b[0] = uint8(u)
	b[1] = uint8(u >> 8)
	return b[:]
}

// beU64 returns a slice containing 8 bytes from the given uint64 u.
//
// The returned bytes have big-endian ordering; that is, the first element at
// index 0 is the most-significant byte in u and index 7 is the least-
// significant byte.
//go:inline
func beU64(u uint64) []uint8 {
	var b [8]uint8
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return b[:]
	}
	b[7] = uint8(u)
	b[6] = uint8(u >> 8)
	b[5] = uint8(u >> 16)
	b[4] = uint8(u >> 24)
	b[3] = uint8(u >> 32)
	b[2] = uint8(u >> 40)
	b[1] = uint8(u >> 48)
	b[0] = uint8(u >> 56)
	return b[:]
}

// beU32 returns a slice containing 4 bytes from the given uint32 u.
//
// The returned bytes have big-endian ordering; that is, the first element at
// index 0 is the most-significant byte in u and index 3 is the least-
// significant byte.
//go:inline
func beU32(u uint32) []uint8 {
	var b [4]uint8
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return b[:]
	}
	b[3] = uint8(u)
	b[2] = uint8(u >> 8)
	b[1] = uint8(u >> 16)
	b[0] = uint8(u >> 24)
	return b[:]
}

// beU16 returns a slice containing 2 bytes from the given uint16 u.
//
// The returned bytes have big-endian ordering; that is, the first element at
// index 0 is the most-significant byte in u and index 1 is the least-
// significant byte.
//go:inline
func beU16(u uint16) []uint8 {
	var b [2]uint8
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return b[:]
	}
	b[1] = uint8(u)
	b[0] = uint8(u >> 8)
	return b[:]
}

// revU64 returns the given uint64 u with bytes in the reverse order.
//go:inline
func revU64(u uint64) uint64 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return 0
	}
	return ((u & 0x00000000000000FF) << 56) |
		((u & 0x000000000000FF00) << 40) |
		((u & 0x0000000000FF0000) << 24) |
		((u & 0x00000000FF000000) << 8) |
		((u & 0x000000FF00000000) >> 8) |
		((u & 0x0000FF0000000000) >> 24) |
		((u & 0x00FF000000000000) >> 40) |
		((u & 0xFF00000000000000) >> 56)
}

// revU32 returns the given uint32 u with bytes in the reverse order.
//go:inline
func revU32(u uint32) uint32 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return 0
	}
	return ((u & 0x000000FF) << 24) | ((u & 0x0000FF00) << 8) |
		((u & 0x00FF0000) >> 8) | ((u & 0xFF000000) >> 24)
}

// revU16 returns the given uint16 u with bytes in the reverse order.
//go:inline
func revU16(u uint16) uint16 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return 0
	}
	return ((u & 0x00FF) << 8) | ((u & 0xFF00) >> 8)
}

// packU64 returns a uint64 constructed by concatenating the bytes in slice b.
//
// The least-significant byte in the returned value is the first element at
// index 0 in b and the most significant byte is index 7, if given. If fewer
// than 8 elements are given in b, the corresponding bytes in the returned value
// are all 0.
//go:inline
func packU64(b []uint8) (u uint64) {
	for i := 0; i < 8 && i < len(b); i++ {
		u |= uint64(b[i]) << (i * 8)
	}
	return
}

// packU32 returns a uint32 constructed by concatenating the bytes in slice b.
//
// The least-significant byte in the returned value is the first element at
// index 0 in b and the most significant byte is index 3, if given. If fewer
// than 4 elements are given in b, the corresponding bytes in the returned value
// are all 0.
//go:inline
func packU32(b []uint8) (u uint32) {
	for i := 0; i < 4 && i < len(b); i++ {
		u |= uint32(b[i]) << (i * 8)
	}
	return
}

// packU16 returns a uint16 constructed by concatenating the bytes in slice b.
//
// The least-significant byte in the returned value is the first element at
// index 0 in b and the most significant byte is index 1, if given. If fewer
// than 2 elements are given in b, the corresponding bytes in the returned value
// are all 0.
//go:inline
func packU16(b []uint8) (u uint16) {
	for i := 0; i < 2 && i < len(b); i++ {
		u |= uint16(b[i]) << (i * 8)
	}
	return
}

// msU8 returns the most-significant byte of u.
//go:inline
func msU8(u uint16) uint8 { return uint8(u >> 8) }

// lsU8 returns the least-significant byte of u.
//go:inline
func lsU8(u uint16) uint8 { return uint8(u) }

// cycles converts the given number of microseconds to CPU cycles for a CPU with
// given frequency.
//go:inline
func cycles(microsec, cpuFreqHz uint32) uint32 {
	return uint32((uint64(microsec) * uint64(cpuFreqHz)) / 1000000)
}

//go:inline
func unpackEndpoint(address uint8) (number, direction uint8) {
	return (address & descEndptAddrNumberMsk) >> descEndptAddrNumberPos,
		(address & descEndptAddrDirectionMsk) >> descEndptAddrDirectionPos
}

//go:inline
func rxEndpoint(number uint8) uint8 {
	return (number & descEndptAddrNumberMsk) | descEndptAddrDirectionOut
}

//go:inline
func txEndpoint(number uint8) uint8 {
	return (number & descEndptAddrNumberMsk) | descEndptAddrDirectionIn
}

//go:inline
func endpointIndex(address uint8) uint8 {
	return ((address & descEndptAddrNumberMsk) << 1) |
		((address & descEndptAddrDirectionMsk) >> descEndptAddrDirectionPos)
}

//go:inline
func indexEndpoint(index uint8) uint8 {
	return ((index >> 1) & descEndptAddrNumberMsk) |
		((index & 0x1) << descEndptAddrDirectionPos)
}

// wrap computes the index into a circular buffer of length mod by walking
// forward n elements if n is positive, or reverse n elements if n is negative.
// For example, both wrap(42, 10) and wrap(-308, 10) return 2.
//go:inline
func wrap(n, mod int) int {
	if mod <= 0 || n == mod {
		return 0
	}
	if n < 0 {
		if -n < mod {
			return mod + n
		}
		return mod - (-n % mod)
	}
	if n < mod {
		return n
	}
	return n % mod
}

// The following buffLo and buffHi are helper methods for slice definitions from
// potentially zero-length arrays (depending on compile-time constants).
//
// For example, if we have an array containing a 5-element buffer for three
// instances of some device class (15 total elements), partitioned as follows,
// then we compute the indices for instance 2 as usual:
//
//    Index:   01234 56789 ABCDE
//    Array:  [  1  |  2  |  3  ]
//
//       Lo:  (n-1) * size   =>   (2-1) * 5   =>   5
//       Hi:    (n) * size   =>     (2) * 5   =>   10 (0xA)
//
// However, if we have specified (via const definition) that 0 instances of some
// device class be allocated, then the associated device class buffer arrays
// will all be zero-length arrays, and the arithmetic to compute the slice
// indices used above will result in out-of-bounds indices:
//
//    Index:
//    Array:  []
//
//       Lo:  (n-1) * size   =>   (2-1) * 5   =>   5           [Error!]
//       Hi:    (n) * size   =>     (2) * 5   =>   10 (0xA)    [Error!]
//
//
// I couldn't figure out a straight-forward way to resolve these slice indices
// using only arithmetic, so I've resorted to simple conditionals. If the number
// of instances for some given class is zero (count=0), defined via compile-time
// constant, then just use the empty slice range [0:0].

// buffLo returns the starting array slice index for the n'th region of size
// elements from an array containing count regions of size elements.
// Regions are specified using a 1-based index (n > 0). Returns 0 if any given
// argument equals 0.
func buffLo(n, count, size uint16) uint16 {
	if 0 == n || 0 == count || 0 == size {
		return 0
	}
	return (n - 1) * size
}

// buffHi returns the ending array slice index for the n'th region of size
// elements from an array containing count regions of size elements.
// Regions are specified using a 1-based index (n > 0). Returns 0 if any given
// argument equals 0.
func buffHi(n, count, size uint16) uint16 {
	if 0 == n || 0 == count || 0 == size {
		return 0
	}
	return n * size
}
