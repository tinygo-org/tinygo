package main

var (
	scalar1 *byte
	scalar2 *int32
	scalar3 *int64
	scalar4 *float32

	array1 *[3]byte
	array2 *[71]byte
	array3 *[3]*byte

	struct1 *struct{}
	struct2 *struct {
		x int
		y int
	}
	struct3 *struct {
		x *byte
		y [60]uintptr
		z *byte
	}
	struct4 *struct {
		x *byte
		y [61]uintptr
	}

	slice1 []byte
	slice2 []*int
	slice3 [][]byte
)

func newScalar() {
	scalar1 = new(byte)
	scalar2 = new(int32)
	scalar3 = new(int64)
	scalar4 = new(float32)
}

func newArray() {
	array1 = new([3]byte)
	array2 = new([71]byte)
	array3 = new([3]*byte)
}

func newStruct() {
	struct1 = new(struct{})
	struct2 = new(struct {
		x int
		y int
	})
	struct3 = new(struct {
		x *byte
		y [60]uintptr
		z *byte
	})
	struct4 = new(struct {
		x *byte
		y [61]uintptr
	})
}

func newFuncValue() *func() {
	// On some platforms that use runtime.funcValue ("switch" style) function
	// values, a func value is allocated as having two pointer words while the
	// struct looks like {unsafe.Pointer; uintptr}. This is so that the interp
	// package won't get confused, see getPointerBitmap in compiler/llvm.go for
	// details.
	return new(func())
}

func makeSlice() {
	slice1 = make([]byte, 5)
	slice2 = make([]*int, 5)
	slice3 = make([][]byte, 5)
}

func makeInterface(v complex128) interface{} {
	return v // always stored in an allocation
}
