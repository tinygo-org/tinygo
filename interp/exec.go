package interp

import (
	"encoding/binary"
	"errors"
	"math/bits"

	"tinygo.org/x/go-llvm"
)

type execState struct {
	cp constParser

	// stack is the current value stack (all produced values in use by running functions).
	stack []value

	// sp is the index of the start of the running function's stack.
	sp uint

	// oldMem is the state of old memory.
	oldMem []memSave

	// version is the memory version of the current function.
	version uint64

	// nextVersion is the next memory version.
	nextVersion uint64

	// pc is the index of the next instruction to execute.
	// A pc beyond the end of the function is treated as a return.
	pc uint

	// curFn is the runnable for the currently running function.
	curFn *runnable

	// rt is used to build the runtime function.
	rt builder

	// escapeStack is a stack of escaped objects.
	// This can be used to restore the state of "escaped" on revert.
	escapeStack []*memObj

	// escBuf is an internal map used temporarily by escapeValue.
	// It is only here so that it can be reused.
	escBuf map[*memObj]struct{}
}

func (s *execState) locals() []value {
	return s.stack[s.sp:]
}

func (s *execState) invalidateEscaped(dbg llvm.Metadata) error {
	// TODO: make this more efficient
	if debug {
		println("invalidate escaped globals")
	}
	for i := 0; i < len(s.escapeStack); i++ {
		err := s.invalidate(s.escapeStack[i], dbg)
		if err != nil {
			return err
		}
	}

	return nil
}

// invalidate flushes and invalidates the contents of the object.
func (s *execState) invalidate(obj *memObj, dbg llvm.Metadata) error {
	skip, err := s.flushOrInvalidatePreCheck(obj)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}
	err = s.doFlush(obj, dbg)
	if err != nil {
		return err
	}
	if obj.data.hasAny() {
		if debug {
			println("invalidate", obj.String())
		}
		version := s.version
		data, err := obj.data.clobber(0, obj.size, version)
		if err != nil {
			return err
		}
		if obj.version < s.version {
			s.oldMem = append(s.oldMem, memSave{
				obj:  obj,
				tree: obj.data,
			})
		}
		obj.data = data
		obj.version = version
		if obj.data.hasAny() {
			println(obj.data.String())
			panic("has some")
		}
	}
	return nil
}

func (s *execState) flushEscaped(dbg llvm.Metadata) error {
	// TODO: make this more efficient
	if debug {
		println("flush escaped globals")
	}
	for i := 0; i < len(s.escapeStack); i++ {
		err := s.flush(s.escapeStack[i], dbg)
		if err != nil {
			return err
		}
	}

	return nil
}

// flush the contents of the object.
// This blocks future modifications to the intializer of the object.
func (s *execState) flush(obj *memObj, dbg llvm.Metadata) error {
	skip, err := s.flushOrInvalidatePreCheck(obj)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}
	return s.doFlush(obj, dbg)
}

func (s *execState) doFlush(obj *memObj, dbg llvm.Metadata) error {
	version := s.version
	if obj.version < s.version {
		s.oldMem = append(s.oldMem, memSave{
			obj:  obj,
			tree: obj.data,
		})
	}
	if obj.data.hasPending() {
		if debug {
			println("flush", obj.String())
		}
		if !obj.data.hasAny() {
			panic("pending nothing")
		}
		data, err := obj.data.flush(s, obj, 0, 0, obj.size, version, dbg)
		if err != nil {
			return err
		}
		obj.data = data
		obj.version = version
	}
	// TODO: do not do this every time.
	data, err := obj.data.blockInit(0, obj.size, version)
	if err != nil {
		return err
	}
	obj.data = data
	obj.version = version
	return nil
}

func (s *execState) flushOrInvalidatePreCheck(obj *memObj) (skip bool, err error) {
	if obj.isFunc || obj.isConst {
		return true, nil
	}
	err = obj.parseInit(&s.cp)
	if err != nil {
		if isRuntimeOrRevert(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (br *memBranch) flush(to *execState, obj *memObj, off, idx, size uint64, version uint64, dbg llvm.Metadata) (memTreeNode, error) {
	br = br.modify(version)
	end := idx + size
	for i := idx; i < end; i = (i | ((1 << br.shift) - 1)) + 1 {
		if br.pending&(1<<((i>>br.shift)%8)) == 0 {
			// There is nothing to flush.
			continue
		}

		// Flush the contents of the child.
		j := end
		if j > (i|((1<<br.shift)-1))+1 {
			j = (i | ((1 << br.shift) - 1)) + 1
		}
		k := i >> br.shift
		c, err := br.sub[k].flush(to, obj, off+(i-idx), i&((1<<br.shift)-1), j-i, version, dbg)
		if err != nil {
			return nil, err
		}
		if !c.hasPending() {
			br.pending &^= 1 << idx
		}
		br.sub[k] = c
	}
	return br, nil
}

func (l *memLeaf) flush(to *execState, obj *memObj, off, idx, size uint64, version uint64, dbg llvm.Metadata) (memTreeNode, error) {
	if assert && idx+size > 64 {
		panic("flush out of bounds")
	}
	var mask uint64 = ((1 << size) - 1) << idx
	mask &= l.pending
	starts := (mask &^ (mask << 1)) |
		(mask & (l.metaHigh ^ (l.metaHigh << 1))) |
		(mask & (l.metaLow ^ (l.metaLow << 1))) |
		(mask & l.specialStarts)
	optimalStoreScale := bits.Len64(obj.ptrTy.bytes()) - 1 // TODO: this is a rough guess
	for starts != 0 {
		i := uint64(bits.TrailingZeros64(starts))
		starts &^= 1 << i
		size := uint64(bits.TrailingZeros64(((starts | ^mask) >> i) | (1 << (64 - i))))
		switch (((l.metaHigh >> i) & 1) << 1) | ((l.metaLow >> i) & 1) {
		case 0b10:
			// Store a special.
			start := bits.Len64(l.specialStarts&((1<<(i+1))-1)) - 1
			v := l.specials[bits.OnesCount64(l.specialStarts&((1<<i)-1))]
			v = slice(v, i-uint64(start), 8*uint64(size))
			to.escape(v)
			to.rt.insertInst(&storeInst{
				to:         obj.ptr(off + i),
				v:          v,
				alignScale: uint(bits.TrailingZeros64((1 << obj.alignScale) | (off + i))),
				order:      llvm.AtomicOrderingNotAtomic,
				dbg:        dbg,
				init:       l.noInit&(((1<<size)-1)<<i) == 0,
			})
		case 0b01:
			// Store undef.
			for size != 0 {
				storeScale := bits.Len64(size) - 1
				if storeScale > optimalStoreScale {
					storeScale = optimalStoreScale
				}
				alignScale := bits.TrailingZeros64((1 << obj.alignScale) | (off + i))
				if storeScale > alignScale {
					storeScale = alignScale
				}
				storeSize := 1 << storeScale
				to.rt.insertInst(&storeInst{
					to:         obj.ptr(off + i),
					v:          undefValue(iType(8 * storeSize)),
					alignScale: uint(alignScale),
					order:      llvm.AtomicOrderingNotAtomic,
					dbg:        dbg,
					init:       l.noInit&(((1<<storeSize)-1)<<i) == 0,
				})
				i += uint64(storeSize)
				size -= uint64(storeSize)
			}
		case 0b11:
			// Store bytes.
			for size != 0 {
				storeScale := bits.Len64(size) - 1
				if storeScale > optimalStoreScale {
					storeScale = optimalStoreScale
				}
				alignScale := bits.TrailingZeros64((1 << obj.alignScale) | (off + i))
				if storeScale > alignScale {
					storeScale = alignScale
				}
				storeSize := 1 << storeScale
				var buf [8]byte
				copy(buf[:], l.data[i:])
				to.rt.insertInst(&storeInst{
					to:         obj.ptr(off + i),
					v:          smallIntValue(iType(8*storeSize), binary.LittleEndian.Uint64(buf[:])),
					alignScale: uint(alignScale),
					order:      llvm.AtomicOrderingNotAtomic,
					dbg:        dbg,
					init:       l.noInit&(((1<<storeSize)-1)<<i) == 0,
				})
				i += uint64(storeSize)
				size -= uint64(storeSize)
			}
		default:
			return nil, errors.New("invalid pending value")
		}
	}
	if mask != 0 {
		l = l.modify(version)
		l.pending &^= mask
	}
	return l, nil
}

func (s *execState) escape(vals ...value) {
	m := s.escBuf
	defer func(m map[*memObj]struct{}) {
		for k := range m {
			delete(m, k)
		}
	}(m)
	if assert && len(m) != 0 {
		panic("escape buffer is not empty")
	}
	start := len(s.escapeStack)
	for _, v := range vals {
		v.aliases(m)
	}
	for obj := range m {
		if obj.escaped {
			continue
		}
		if debug {
			println("escape", obj.String())
		}
		s.escapeStack = append(s.escapeStack, obj)
	}
	s.finishEscape(start)
}

func (s *execState) finishEscape(i int) {
	stack := s.escapeStack
	sortObjects(stack[i:])
	for ; i < len(stack); i++ {
		by := s.escapeStack[i]
		for _, obj := range by.esc {
			if obj.escaped {
				continue
			}
			if debug {
				println("transitive escape of", obj.String(), "by", by.String())
			}
			stack = append(stack, obj)
		}
	}
	s.escapeStack = stack
}

func sortObjects(list []*memObj) {
	for i := range list {
		for list[(i-1)/2].id < list[i].id {
			list[(i-1)/2], list[i] = list[i], list[(i-1)/2]
			i = (i - 1) / 2
		}
	}
	for len(list) > 0 {
		list[0], list[len(list)-1] = list[len(list)-1], list[0]
		list = list[:len(list)-1]
		i := 0
		for {
			max := i
			if l := 2*i + 1; l < len(list) && list[l].id > list[max].id {
				max = l
			}
			if r := 2*i + 2; r < len(list) && list[r].id > list[max].id {
				max = r
			}
			if max == i {
				break
			}
			list[i], list[max] = list[max], list[i]
			i = max
		}
	}
}

type memSave struct {
	obj     *memObj
	tree    memTreeNode
	version uint64
}

var errRuntime = errors.New("operation must be processed at runtime")

var errUnknownBranch = errRevert{errors.New("unknown branch destination")}

var errUB = errors.New("encountered undefined behavior")

var errMaxStackSize = errors.New("recursed beyond max stack size")

func isRuntimeOrRevert(err error) bool {
	_, rev := err.(errRevert)
	return rev || err == errRuntime
}

type errRevert struct {
	err error
}

func (err errRevert) Error() string {
	return "revert: " + err.err.Error()
}

func (err errRevert) Unwrap() error {
	return err.err
}

func todo(desc string) error {
	return errRevert{errTODO{desc}}
}

type errTODO struct {
	desc string
}

func (err errTODO) Error() string {
	return "TODO: " + err.desc
}

type typeError struct {
	expected, got typ
}

func (err typeError) Error() string {
	var expected string
	if err.expected != nil {
		expected = err.expected.String()
	} else {
		expected = "void"
	}
	var got string
	if err.got != nil {
		got = err.got.String()
	} else {
		got = "void"
	}
	return "expected type " + expected + " but got " + got
}

/*
type backtrace struct {
	err   error
	stack []instruction
}

func (b backtrace) Error() string {
	strs := make([]string, len(b.stack))
	for i, inst := range b.stack {
		strs[i] = "\tfrom: " + inst.String()
	}
	return b.err.Error() + "\n" + strings.Join(strs, "\n")
}

func (b backtrace) Unwrap() error {
	return b.err
}
*/
