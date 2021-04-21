package usb

// leU64 returns a slice containing 8 bytes from the given uint64 u.
//
// The returned bytes have little-endian ordering; that is, the first element
// at index 0 is the least-significant byte in u and index 7 is the most-
// significant byte.
//go:inline
func leU64(u uint64) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0, 0, 0, 0, 0, 0, 0}
	}
	return []uint8{
		uint8(u), uint8(u >> 8), uint8(u >> 16), uint8(u >> 24),
		uint8(u >> 32), uint8(u >> 40), uint8(u >> 48), uint8(u >> 56),
	}
}

// leU32 returns a slice containing 4 bytes from the given uint32 u.
//
// The returned bytes have little-endian ordering; that is, the first element
// at index 0 is the least-significant byte in u and index 3 is the most-
// significant byte.
//go:inline
func leU32(u uint32) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0, 0, 0}
	}
	return []uint8{
		uint8(u), uint8(u >> 8), uint8(u >> 16), uint8(u >> 24),
	}
}

// leU16 returns a slice containing 2 bytes from the given uint16 u.
//
// The returned bytes have little-endian ordering; that is, the first element
// at index 0 is the least-significant byte in u and index 1 is the most-
// significant byte.
//go:inline
func leU16(u uint16) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0}
	}
	return []uint8{
		uint8(u), uint8(u >> 8),
	}
}

// beU64 returns a slice containing 8 bytes from the given uint64 u.
//
// The returned bytes have big-endian ordering; that is, the first element at
// index 0 is the most-significant byte in u and index 7 is the least-
// significant byte.
//go:inline
func beU64(u uint64) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0, 0, 0, 0, 0, 0, 0}
	}
	return []uint8{
		uint8(u >> 56), uint8(u >> 48), uint8(u >> 40), uint8(u >> 32),
		uint8(u >> 24), uint8(u >> 16), uint8(u >> 8), uint8(u),
	}
}

// beU32 returns a slice containing 4 bytes from the given uint32 u.
//
// The returned bytes have big-endian ordering; that is, the first element at
// index 0 is the most-significant byte in u and index 3 is the least-
// significant byte.
//go:inline
func beU32(u uint32) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0, 0, 0}
	}
	return []uint8{
		uint8(u >> 24), uint8(u >> 16), uint8(u >> 8), uint8(u),
	}
}

// beU16 returns a slice containing 2 bytes from the given uint16 u.
//
// The returned bytes have big-endian ordering; that is, the first element at
// index 0 is the most-significant byte in u and index 1 is the least-
// significant byte.
//go:inline
func beU16(u uint16) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0}
	}
	return []uint8{
		uint8(u >> 8), uint8(u),
	}
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
