package main

// Interp must not write a whole pointer into the initializer of a global if
// the global type has no pointer-sized element at that position.

import "unsafe"

const ptrSize = unsafe.Sizeof(uintptr(0))

type word [ptrSize]byte

var x int

type byteArray struct {
	_ [0]uintptr
	b [2 * ptrSize]byte
}

var (
	split  [2]uintptr
	copied [2]uintptr
	float  float64
	bytes  byteArray
)

func init() {
	storeSplit(&split)
	copySplit(&copied)
	storeFloat(&float)
	storeBytes(&bytes)
}

// storeSplit stores a pointer across two uintptr elements. A byte array store
// does not tell LLVM it is aligned.
func storeSplit(a *[2]uintptr) {
	p := uintptr(unsafe.Pointer(&x))
	*(*word)(unsafe.Add(unsafe.Pointer(a), ptrSize/2)) = *(*word)(unsafe.Pointer(&p))
}

func copySplit(a *[2]uintptr) {
	p := uintptr(unsafe.Pointer(&x))
	dst := (*[2 * ptrSize]byte)(unsafe.Pointer(a))
	copy(dst[ptrSize/2:], (*word)(unsafe.Pointer(&p))[:])
}

func storeFloat(f *float64) {
	*(*uintptr)(unsafe.Pointer(f)) = uintptr(unsafe.Pointer(&x))
}

func storeBytes(b *byteArray) {
	*(*uintptr)(unsafe.Pointer(&b.b[ptrSize])) = uintptr(unsafe.Pointer(&x))
}

//go:noinline
func runtimeSplit() (a [2]uintptr) {
	storeSplit(&a)
	return
}

//go:noinline
func runtimeCopied() (a [2]uintptr) {
	copySplit(&a)
	return
}

//go:noinline
func runtimeFloat() (f float64) {
	storeFloat(&f)
	return
}

//go:noinline
func runtimeBytes() (b byteArray) {
	storeBytes(&b)
	return
}

func bits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

func main() {
	println("split:", split == runtimeSplit())
	println("copied:", copied == runtimeCopied())
	println("float:", bits(float) == bits(runtimeFloat()))
	println("bytes:", bytes.b == runtimeBytes().b)
}
