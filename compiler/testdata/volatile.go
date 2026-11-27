package main

import (
	"runtime/volatile"
	"unsafe"
)

func load8(p *uint8) uint8 {
	return volatile.LoadUint8(p)
}

func load32(p *uint32) uint32 {
	return volatile.LoadUint32(p)
}

func store8(p *uint8, v uint8) {
	volatile.StoreUint8(p, v)
}

func store32(p *uint32, v uint32) {
	volatile.StoreUint32(p, v)
}

type peripheral struct {
	CTRL   volatile.Register32
	STATUS volatile.Register8
}

func regGet(p *peripheral) uint32 {
	return p.CTRL.Get()
}

func regSet(p *peripheral, v uint32) {
	p.CTRL.Set(v)
}

func regSetBits(p *peripheral, v uint32) {
	p.CTRL.SetBits(v)
}

func regClearBits(p *peripheral, v uint32) {
	p.CTRL.ClearBits(v)
}

func regHasBits(p *peripheral, v uint8) bool {
	return p.STATUS.HasBits(v)
}

func regReplaceBits(p *peripheral, v, mask uint32, pos uint8) {
	p.CTRL.ReplaceBits(v, mask, pos)
}

func regPointerSet(r *volatile.Register32, v uint32) {
	r.Set(v)
}

func regMethodValue(p *peripheral) func(uint32) {
	return p.CTRL.Set
}

func regPointerMethodValue(r *volatile.Register32) func(uint32) {
	return r.Set
}

func regCallMethodValue(p *peripheral, v uint32) {
	f := p.CTRL.Set
	f(v)
}

type uint32Ptr = *uint32

func load32Alias(p uint32Ptr) uint32 {
	return volatile.LoadUint32(p)
}

type register32Ptr = *volatile.Register32

func regGetAlias(r register32Ptr) uint32 {
	return r.Get()
}

func load32Unsafe(p unsafe.Pointer) uint32 {
	return volatile.LoadUint32((*uint32)(p))
}

func store32Unsafe(p unsafe.Pointer, v uint32) {
	volatile.StoreUint32((*uint32)(p), v)
}

func regGetUnsafe(p unsafe.Pointer) uint32 {
	return (*volatile.Register32)(p).Get()
}
