package persistence

import (
	"errors"
	"unsafe"
)

var errOverflow = errors.New("Overflow of persistent storage")

//go:extern _persist_start
var persistStartSymbol [0]byte

//go:extern _persist_end
var persistEndSymbol [0]byte

//go:extern _persist_ptr
var reserved uintptr

func GetRAM() Region {
	startAddr := unsafe.Pointer(&persistStartSymbol)
	endAddr := uintptr(unsafe.Pointer(&persistEndSymbol))

	return &ram{
		start:  startAddr,
		length: int(endAddr - uintptr(startAddr)),
	}
}

type ram struct {
	start  unsafe.Pointer
	length int
}

// Size gets the size of the persistence area (in bytes)
func (r *ram) Size() int64 {
	return int64(r.length)
}

// ReadAt reads from the persistence area, it is signature
// compatible with Go's `io.ReaderAt` interface
func (r *ram) ReadAt(dest []byte, offset int64) (n int, err error) {
	if offset < 0 || len(dest)+int(offset) > r.length {
		return 0, errOverflow
	}

	p := uintptr(r.start) + uintptr(offset)
	for i := 0; i < len(dest); i++ {
		dest[i] = *(*byte)(unsafe.Pointer(p))
		p++
	}

	return len(dest), nil
}

// WriterAt reads from the persistence area, it is signature
// compatible with Go's `io.WriterAt` interface
func (r *ram) WriteAt(src []byte, offset int64) (n int, err error) {
	if offset < 0 || len(src)+int(offset) > r.length {
		return 0, errOverflow
	}

	p := uintptr(r.start) + uintptr(offset)
	for i := 0; i < len(src); i++ {
		*(*byte)(unsafe.Pointer(p)) = src[i]
		p++
	}

	return len(src), nil
}

// SubAllocation creates a new persistence object that
// represents a sub-allocation of the main persistence
// area
func (r *ram) SubAllocation(offset int, len int) (p Region, err error) {
	if offset+len > r.length {
		return nil, errOverflow
	}

	return &ram{
		start:  unsafe.Pointer(uintptr(r.start) + uintptr(offset)),
		length: len,
	}, nil
}
