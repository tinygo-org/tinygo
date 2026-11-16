package main

// On AVR, uintptr has 16 bits, so the length 1 << 16 does not fit in a uintptr.

func indexBig(p *[1 << 16]byte, i uintptr) byte {
	return p[i]
}

func sliceBig(p *[1 << 16]byte) []byte {
	return p[:]
}

func convertBig(s []byte) *[1 << 16]byte {
	return (*[1 << 16]byte)(s)
}

// This array length fits in a 16-bit uintptr.
func indexSmall(p *[1<<16 - 1]byte, i uintptr) byte {
	return p[i]
}

func main() {
	println(indexBig(nil, 0), len(sliceBig(nil)), convertBig(nil) == nil, indexSmall(nil, 0))
}

// ERROR: # command-line-arguments
// ERROR: compiler-arraylen.go:6:10: array length 65536 does not fit in a 16-bit uintptr
// ERROR: compiler-arraylen.go:10:10: array length 65536 does not fit in a 16-bit uintptr
// ERROR: compiler-arraylen.go:14:25: array length 65536 does not fit in a 16-bit uintptr
