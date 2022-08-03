package volatile

// This file defines Register{8,16,32,64} types, which are convenience types for
// volatile register accesses.

// Special types that causes loads/stores to be volatile (necessary for
// memory-mapped registers).
type Register8 struct {
	Reg uint8
}

// Get returns the value in the register. It is the volatile equivalent of:
//
//	*r.Reg
//
//go:inline
func (r *Register8) Get() uint8 {
	return LoadUint8(&r.Reg)
}

// Set updates the register value. It is the volatile equivalent of:
//
//	*r.Reg = value
//
//go:inline
func (r *Register8) Set(value uint8) {
	StoreUint8(&r.Reg, value)
}

// SetBits reads the register, sets the given bits, and writes it back. It is
// the volatile equivalent of:
//
//	r.Reg |= value
//
//go:inline
func (r *Register8) SetBits(value uint8) {
	StoreUint8(&r.Reg, LoadUint8(&r.Reg)|value)
}

// ClearBits reads the register, clears the given bits, and writes it back. It
// is the volatile equivalent of:
//
//	r.Reg &^= value
//
//go:inline
func (r *Register8) ClearBits(value uint8) {
	StoreUint8(&r.Reg, LoadUint8(&r.Reg)&^value)
}

// HasBits reads the register and then checks to see if the passed bits are set. It
// is the volatile equivalent of:
//
//	(*r.Reg & value) > 0
//
//go:inline
func (r *Register8) HasBits(value uint8) bool {
	return (r.Get() & value) > 0
}

// ReplaceBits is a helper to simplify setting multiple bits high and/or low at
// once. It is the volatile equivalent of:
//
//	r.Reg = (r.Reg & ^(mask << pos)) | value << pos
//
// go:inline
func (r *Register8) ReplaceBits(value uint8, mask uint8, pos uint8) {
	StoreUint8(&r.Reg, LoadUint8(&r.Reg)&^(mask<<pos)|value<<pos)
}

type Register16 struct {
	Reg uint16
}

// Get returns the value in the register. It is the volatile equivalent of:
//
//	*r.Reg
//
//go:inline
func (r *Register16) Get() uint16 {
	return LoadUint16(&r.Reg)
}

// Set updates the register value. It is the volatile equivalent of:
//
//	*r.Reg = value
//
//go:inline
func (r *Register16) Set(value uint16) {
	StoreUint16(&r.Reg, value)
}

// SetBits reads the register, sets the given bits, and writes it back. It is
// the volatile equivalent of:
//
//	r.Reg |= value
//
//go:inline
func (r *Register16) SetBits(value uint16) {
	StoreUint16(&r.Reg, LoadUint16(&r.Reg)|value)
}

// ClearBits reads the register, clears the given bits, and writes it back. It
// is the volatile equivalent of:
//
//	r.Reg &^= value
//
//go:inline
func (r *Register16) ClearBits(value uint16) {
	StoreUint16(&r.Reg, LoadUint16(&r.Reg)&^value)
}

// HasBits reads the register and then checks to see if the passed bits are set. It
// is the volatile equivalent of:
//
//	(*r.Reg & value) > 0
//
//go:inline
func (r *Register16) HasBits(value uint16) bool {
	return (r.Get() & value) > 0
}

// ReplaceBits is a helper to simplify setting multiple bits high and/or low at
// once. It is the volatile equivalent of:
//
//	r.Reg = (r.Reg & ^(mask << pos)) | value << pos
//
// go:inline
func (r *Register16) ReplaceBits(value uint16, mask uint16, pos uint8) {
	StoreUint16(&r.Reg, LoadUint16(&r.Reg)&^(mask<<pos)|value<<pos)
}

type Register32 struct {
	Reg uint32
}

// Get returns the value in the register. It is the volatile equivalent of:
//
//	*r.Reg
//
//go:inline
func (r *Register32) Get() uint32 {
	return LoadUint32(&r.Reg)
}

// Set updates the register value. It is the volatile equivalent of:
//
//	*r.Reg = value
//
//go:inline
func (r *Register32) Set(value uint32) {
	StoreUint32(&r.Reg, value)
}

// SetBits reads the register, sets the given bits, and writes it back. It is
// the volatile equivalent of:
//
//	r.Reg |= value
//
//go:inline
func (r *Register32) SetBits(value uint32) {
	StoreUint32(&r.Reg, LoadUint32(&r.Reg)|value)
}

// ClearBits reads the register, clears the given bits, and writes it back. It
// is the volatile equivalent of:
//
//	r.Reg &^= value
//
//go:inline
func (r *Register32) ClearBits(value uint32) {
	StoreUint32(&r.Reg, LoadUint32(&r.Reg)&^value)
}

// HasBits reads the register and then checks to see if the passed bits are set. It
// is the volatile equivalent of:
//
//	(*r.Reg & value) > 0
//
//go:inline
func (r *Register32) HasBits(value uint32) bool {
	return (r.Get() & value) > 0
}

// ReplaceBits is a helper to simplify setting multiple bits high and/or low at
// once. It is the volatile equivalent of:
//
//	r.Reg = (r.Reg & ^(mask << pos)) | value << pos
//
// go:inline
func (r *Register32) ReplaceBits(value uint32, mask uint32, pos uint8) {
	StoreUint32(&r.Reg, LoadUint32(&r.Reg)&^(mask<<pos)|value<<pos)
}

type Register64 struct {
	Reg uint64
}

// Get returns the value in the register. It is the volatile equivalent of:
//
//	*r.Reg
//
//go:inline
func (r *Register64) Get() uint64 {
	return LoadUint64(&r.Reg)
}

// Set updates the register value. It is the volatile equivalent of:
//
//	*r.Reg = value
//
//go:inline
func (r *Register64) Set(value uint64) {
	StoreUint64(&r.Reg, value)
}

// SetBits reads the register, sets the given bits, and writes it back. It is
// the volatile equivalent of:
//
//	r.Reg |= value
//
//go:inline
func (r *Register64) SetBits(value uint64) {
	StoreUint64(&r.Reg, LoadUint64(&r.Reg)|value)
}

// ClearBits reads the register, clears the given bits, and writes it back. It
// is the volatile equivalent of:
//
//	r.Reg &^= value
//
//go:inline
func (r *Register64) ClearBits(value uint64) {
	StoreUint64(&r.Reg, LoadUint64(&r.Reg)&^value)
}

// HasBits reads the register and then checks to see if the passed bits are set. It
// is the volatile equivalent of:
//
//	(*r.Reg & value) > 0
//
//go:inline
func (r *Register64) HasBits(value uint64) bool {
	return (r.Get() & value) > 0
}

// ReplaceBits is a helper to simplify setting multiple bits high and/or low at
// once. It is the volatile equivalent of:
//
//	r.Reg = (r.Reg & ^(mask << pos)) | value << pos
//
// go:inline
func (r *Register64) ReplaceBits(value uint64, mask uint64, pos uint8) {
	StoreUint64(&r.Reg, LoadUint64(&r.Reg)&^(mask<<pos)|value<<pos)
}
