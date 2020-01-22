// +build nxp,mk66f18

package volatile

import "unsafe"

const registerBase = 0x40000000
const registerEnd = 0x40100000
const bitbandBase = 0x42000000
const ptrBytes = unsafe.Sizeof(uintptr(0))

//go:inline
func bitbandAddress(reg uintptr, bit uintptr) uintptr {
	if bit > ptrBytes*8 {
		panic("invalid bit position")
	}
	if reg < registerBase || reg >= registerEnd {
		panic("register is out of range")
	}
	return (reg-registerBase)*ptrBytes*8 + bit*ptrBytes + bitbandBase
}

// Special types that causes loads/stores to be volatile (necessary for
// memory-mapped registers).
type BitRegister struct {
	Reg uint32
}

// Get returns the of the mapped register bit. It is the volatile equivalent of:
//
//     *r.Reg
//
//go:inline
func (r *BitRegister) Get() bool {
	return LoadUint32(&r.Reg) != 0
}

// Set sets the mapped register bit. It is the volatile equivalent of:
//
//     *r.Reg = 1
//
//go:inline
func (r *BitRegister) Set() {
	StoreUint32(&r.Reg, 1)
}

// Clear clears the mapped register bit. It is the volatile equivalent of:
//
//     *r.Reg = 0
//
//go:inline
func (r *BitRegister) Clear() {
	StoreUint32(&r.Reg, 0)
}

// Bit maps bit N of register R to the corresponding bitband address. Bit panics
// if R is not an AIPS or GPIO register or if N is out of range (greater than
// the number of bits in a register minus one).
//
// go:inline
func (r *Register8) Bit(bit uintptr) *BitRegister {
	ptr := bitbandAddress(uintptr(unsafe.Pointer(&r.Reg)), bit)
	return (*BitRegister)(unsafe.Pointer(ptr))
}

// Bit maps bit N of register R to the corresponding bitband address. Bit panics
// if R is not an AIPS or GPIO register or if N is out of range (greater than
// the number of bits in a register minus one).
//
// go:inline
func (r *Register16) Bit(bit uintptr) *BitRegister {
	ptr := bitbandAddress(uintptr(unsafe.Pointer(&r.Reg)), bit)
	return (*BitRegister)(unsafe.Pointer(ptr))
}

// Bit maps bit N of register R to the corresponding bitband address. Bit panics
// if R is not an AIPS or GPIO register or if N is out of range (greater than
// the number of bits in a register minus one).
//
// go:inline
func (r *Register32) Bit(bit uintptr) *BitRegister {
	ptr := bitbandAddress(uintptr(unsafe.Pointer(&r.Reg)), bit)
	return (*BitRegister)(unsafe.Pointer(ptr))
}
