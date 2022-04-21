package interp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
	"strings"

	"tinygo.org/x/go-llvm"
)

// makeZeroMem creates a memory tree initialized with zero bytes.
// This matches the behavior of the 'runtime.alloc' function.
func makeZeroMem(size uint64) memTreeNode {
	return zeroMemInitTable.make(size)
}

var zeroMemInitTable = memInitTable{
	leaf: memLeaf{
		metaHigh: ^uint64(0),
		metaLow:  ^uint64(0),
		noInit:   ^uint64(0),
	},
}

func init() {
	zeroMemInitTable.init()
}

// makeInitializedMem makes a memory tree node initialized by a value.
// The resulting node is equal in size to the value (with padding).
// Initializing stores are blocked on padding.
func makeInitializedMem(v value) (memTreeNode, error) {
	size := v.typ().bytes()
	mem := undefMemInitTable.make(size)
	mem, err := mem.store(v, 0, memVersionInit)
	if err != nil {
		return nil, err
	}
	mem, err = mem.markStored(0, size, memVersionInit)
	if err != nil {
		return nil, err
	}
	mem, err = blockPaddingInit(mem, v.typ(), 0, memVersionInit)
	if err != nil {
		return nil, err
	}
	if mem.hasPending() {
		return nil, errors.New("has pending")
	}
	return mem, nil
}

// blockPaddingInit blocks initializing stores to padding.
func blockPaddingInit(mem memTreeNode, t typ, off uint64, version uint64) (memTreeNode, error) {
	switch t := t.(type) {
	case nonAggTyp:
		// These types do not contain "padding" in the way that structs do.

	case arrType:
		// Apply element padding.
		elemSize := t.of.bytes()
		for i := uint32(0); i < t.n; i++ {
			var err error
			mem, err = blockPaddingInit(mem, t.of, off+(elemSize*uint64(i)), version)
			if err != nil {
				return nil, err
			}
		}

	case *structType:
		// Apply padding between and within fields.
		var i uint64
		size := t.size
		j := 0
		fields := t.fields
		for i < size {
			if j < len(fields) && fields[j].offset == i {
				fty := fields[j].ty
				var err error
				mem, err = blockPaddingInit(mem, fty, off+fields[j].offset, version)
				if err != nil {
					return nil, err
				}
				i += fty.bytes()
				j++
			} else {
				padEnd := size
				if j < len(fields) {
					padEnd = fields[j].offset
				}
				var err error
				mem, err = mem.blockInit(off+i, padEnd-i, version)
				if err != nil {
					return nil, err
				}
				i = padEnd
			}
		}

	default:
		return nil, todo("block padding init")
	}

	return mem, nil
}

var undefMemInitTable = memInitTable{
	leaf: memLeaf{
		metaLow: ^uint64(0),
	},
}

func init() {
	undefMemInitTable.init()
}

// makeUndefMem creates a non-initializable memory tree initialized with undef.
// This matches the behavior of the 'alloca' instruction.
func makeUndefMem(size uint64) memTreeNode {
	return undefNoInitMemInitTable.make(size)
}

var undefNoInitMemInitTable = memInitTable{
	leaf: memLeaf{
		metaLow: ^uint64(0),
		noInit:  ^uint64(0),
	},
}

func init() {
	undefNoInitMemInitTable.init()
}

// makeUnknownMem creates a memory tree initialized with "unknown" bytes.
// Stores to this will not be marked as initializing.
// This can be used to back a global with contents that cannot be understood.
func makeUnknownMem(size, version uint64) memTreeNode {
	return unknownMemInitTable.make(size)
}

var unknownMemInitTable = memInitTable{
	leaf: memLeaf{
		noInit: ^uint64(0),
	},
}

type memInitTable struct {
	br   [20]memBranch
	leaf memLeaf
}

func (tbl *memInitTable) init() {
	var anything, pending uint8
	if tbl.leaf.hasAny() {
		anything = ^uint8(0)
	}
	if tbl.leaf.hasPending() {
		pending = ^uint8(0)
	}
	var child memTreeNode = &tbl.leaf
	for i := range tbl.br {
		tbl.br[i].anything = anything
		tbl.br[i].pending = pending
		tbl.br[i].shift = 6 + uint8(3*i)
		sub := &tbl.br[i].sub
		for i := range sub {
			sub[i] = child
		}
		child = &tbl.br[i]
	}
	go func() {
		_ = *tbl
	}()
}

func (tbl *memInitTable) make(size uint64) memTreeNode {
	if size <= 64 {
		l := &tbl.leaf
		if size != 64 {
			l = l.modify(memVersionInit)
			err := l.doClobber(-(1 << size))
			if err != nil {
				panic(err)
			}
		}
		return l
	} else {
		br := &tbl.br[(bits.Len64(size-1)-6)/3]
		brSize := uint64(8) << br.shift
		if size != brSize {
			br = br.modify(memVersionInit)
			err := br.doClobber(size, brSize-size, memVersionInit)
			if err != nil {
				panic(err)
			}
		}
		return br
	}
}

type memTreeNode interface {
	// load a value of a given type from an offset.
	// This returns errRuntime if the value is not statically known.
	load(typ typ, off uint64) (value, error)

	// store a value of a given type to an offset.
	store(v value, off uint64, version uint64) (memTreeNode, error)

	// fillUndef fills a section of the value with undef.
	fillUndef(off, size uint64, version uint64) (memTreeNode, error)

	// memSet fills a section with a specified byte value.
	memSet(off, size uint64, v byte, version uint64) (memTreeNode, error)

	// copyTo copies data from this node to the destination node.
	// This returns the modified destination node.
	copyTo(dst memTreeNode, dstOff, srcOff, n uint64, version uint64) (memTreeNode, error)

	// copyFrom copies a section of the contents of src to this node.
	copyFrom(src *memLeaf, dstOff, srcOff, n uint64, version uint64) (memTreeNode, error)

	// clobber all values within a range.
	// After this, the range is placed in an "unknown" state.
	clobber(off, size uint64, version uint64) (memTreeNode, error)

	// markStored marks the range to indicate that it has already been stored.
	markStored(off, size uint64, version uint64) (memTreeNode, error)

	// blockInit blocks intitializing stores on the range.
	blockInit(off, size uint64, version uint64) (memTreeNode, error)

	// canInit checks if an initializing store can be done at the provided range.
	canInit(off, size uint64) (bool, error)

	// hasAny checks whether there are any known values within this node.
	hasAny() bool

	// hadPending checks whether there are any values waiting to be flushed.
	hasPending() bool

	flush(to *execState, obj *memObj, off, idx, size uint64, version uint64, dbg llvm.Metadata) (memTreeNode, error)

	fmt.Stringer
}

const (
	memVersionZero uint64 = iota
	memVersionInit
	memVersionStart
)

type memBranch struct {
	sub               [8]memTreeNode
	version           uint64
	shift             uint8
	anything, pending uint8
}

func (br *memBranch) load(to typ, off uint64) (value, error) {
	size := to.bytes()
	if size == 0 {
		return to.zero(), nil
	}
	if off+size > 8<<br.shift {
		return value{}, errors.New("load out of bounds")
	}

	if (off+size-1)>>br.shift == off>>br.shift {
		// Load directly from a child.
		return br.sub[off>>br.shift].load(to, off&((1<<br.shift)-1))
	}

	switch ty := to.(type) {
	case nonAggTyp:
		// Split into smaller integer loads.
		var tmp []value
		end := off + size
		for i := off; i < end; i = (i | ((1 << br.shift) - 1)) + 1 {
			j := end
			if j > (i|((1<<br.shift)-1))+1 {
				j = (i | ((1 << br.shift) - 1)) + 1
			}
			v, err := br.sub[i>>br.shift].load(iType(8*(j-i)), i&((1<<br.shift)-1))
			if err != nil {
				return value{}, err
			}
			tmp = append(tmp, v)
		}
		return cast(ty, cat(tmp)), nil

	case arrType:
		// Load array elements seperately.
		elems := make([]value, ty.n)
		elemSize := ty.of.bytes()
		for i := range elems {
			v, err := br.load(ty.of, off+elemSize*uint64(i))
			if err != nil {
				return value{}, err
			}
			elems[i] = v
		}
		return arrayValue(ty.of, elems...), nil

	case *structType:
		// Load fields seperately.
		fields := make([]value, len(ty.fields))
		for i, f := range ty.fields {
			v, err := br.load(f.ty, off+f.offset)
			if err != nil {
				return value{}, err
			}
			fields[i] = v
		}
		return structValue(ty, fields...), nil

	default:
		return value{}, todo("split load " + ty.String())
	}
}

func (br *memBranch) store(v value, off uint64, version uint64) (memTreeNode, error) {
	br = br.modify(version)
	err := br.doStore(v, off, version)
	if err != nil {
		return nil, err
	}
	return br, nil
}

func (br *memBranch) doStore(v value, off uint64, version uint64) error {
	ty := v.typ()
	size := ty.bytes()
	if size == 0 {
		return nil
	}
	if off+size > 8<<br.shift {
		return errors.New("store out of bounds")
	}

	if (off+size-1)>>br.shift == off>>br.shift {
		// Store directly into a child.
		idx := off >> br.shift
		c, err := br.sub[idx].store(v, off&((1<<br.shift)-1), version)
		if err != nil {
			return err
		}
		br.sub[idx] = c
		br.anything |= 1 << idx
		br.pending |= 1 << idx
		return nil
	}

	switch ty := ty.(type) {
	case nonAggTyp:
		// Slice the store into smaller integer stores.
		n := ty.bits()
		for i := uint64(0); i < n; {
			m := n - i
			if sectionBits := 8 * ((1 << br.shift) - (off & ((1 << br.shift) - 1))); sectionBits < m {
				m = sectionBits
			}
			err := br.doStore(slice(v, i, m), off, version)
			if err != nil {
				return err
			}
			i += m
			off += (m + 7) / 8
		}

	case arrType:
		// Store elements individually.
		elemSize := ty.of.bytes()
		for i := uint32(0); i < ty.n; i++ {
			err := br.doStore(extractValue(v, i), off+(uint64(i)*uint64(elemSize)), version)
			if err != nil {
				return err
			}
		}

	case *structType:
		// Store fields individually.
		fields := ty.fields
		var i uint32
		var j uint64
		for j < size {
			if i < uint32(len(fields)) && fields[i].offset == j {
				err := br.doStore(extractValue(v, i), off+j, version)
				if err != nil {
					return err
				}
				j += fields[i].ty.bytes()
				i++
			} else {
				padEnd := size
				if i < uint32(len(fields)) {
					padEnd = fields[i].offset
				}
				err := br.doFillUndef(off+j, padEnd-j, version)
				if err != nil {
					return err
				}
				j = padEnd
			}
		}

	default:
		return todo("split store of " + ty.String())
	}

	return nil
}

func (br *memBranch) fillUndef(off, size uint64, version uint64) (memTreeNode, error) {
	br = br.modify(version)
	err := br.doFillUndef(off, size, version)
	if err != nil {
		return nil, err
	}
	return br, nil
}

func (br *memBranch) doFillUndef(off, size uint64, version uint64) error {
	if off+size > 8<<br.shift {
		return errors.New("fillUndef out of bounds")
	}

	end := off + size
	for i := off; i < end; i = (i | ((1 << br.shift) - 1)) + 1 {
		j := end
		if j > (i|((1<<br.shift)-1))+1 {
			j = (i | ((1 << br.shift) - 1)) + 1
		}
		idx := i >> br.shift
		c, err := br.sub[idx].fillUndef(i&((1<<br.shift)-1), j-i, version)
		if err != nil {
			return err
		}
		br.anything |= 1 << idx
		br.pending |= 1 << idx
		br.sub[idx] = c
	}

	return nil
}

func (br *memBranch) memSet(off, size uint64, v byte, version uint64) (memTreeNode, error) {
	br = br.modify(version)
	err := br.doMemSet(off, size, v, version)
	if err != nil {
		return nil, err
	}
	return br, nil
}

func (br *memBranch) doMemSet(off, size uint64, v byte, version uint64) error {
	if off+size > 8<<br.shift {
		return errors.New("memSet out of bounds")
	}

	end := off + size
	for i := off; i < end; i = (i | ((1 << br.shift) - 1)) + 1 {
		// memSet the child.
		j := end
		if j > (i|((1<<br.shift)-1))+1 {
			j = (i | ((1 << br.shift) - 1)) + 1
		}
		idx := i >> br.shift
		c, err := br.sub[idx].memSet(i&((1<<br.shift)-1), j-i, v, version)
		if err != nil {
			return err
		}
		if c.hasPending() {
			br.pending |= 1 << idx
			br.anything |= 1 << idx
		}
		br.sub[idx] = c
	}

	return nil
}

func (br *memBranch) copyTo(dst memTreeNode, dstOff, srcOff, n uint64, version uint64) (memTreeNode, error) {
	if srcOff+n > 8<<br.shift {
		return nil, errors.New("copyTo out of bounds")
	}

	if dst, ok := dst.(*memBranch); ok && dst.shift > br.shift {
		if dstOff+n > 8<<dst.shift {
			return nil, errors.New("copyTo out of bounds")
		}

		dst = dst.modify(version)

		end := dstOff + n
		for i := dstOff; i < end; i = (i | ((1 << dst.shift) - 1)) + 1 {
			j := end
			if j > (i|((1<<dst.shift)-1))+1 {
				j = (i | ((1 << dst.shift) - 1)) + 1
			}
			idx := i >> dst.shift
			c, err := br.copyTo(dst.sub[idx], i&((1<<dst.shift)-1), srcOff+(i-dstOff), j-i, version)
			if err != nil {
				return nil, err
			}
			dst.pending |= 1 << idx
			dst.anything |= 1 << idx
			dst.sub[idx] = c
		}

		return dst, nil
	}

	end := srcOff + n
	for i := srcOff; i < end; i = (i | ((1 << br.shift) - 1)) + 1 {
		j := end
		if j > (i|((1<<br.shift)-1))+1 {
			j = (i | ((1 << br.shift) - 1)) + 1
		}
		idx := i >> br.shift
		var err error
		dst, err = br.sub[idx].copyTo(dst, dstOff+(i-srcOff), i&((1<<br.shift)-1), j-i, version)
		if err != nil {
			return nil, err
		}
	}

	return dst, nil
}

func (br *memBranch) copyFrom(src *memLeaf, dstOff, srcOff, n uint64, version uint64) (memTreeNode, error) {
	if dstOff+n > 8<<br.shift {
		return nil, errors.New("copyFrom out of bounds")
	}
	if n == 0 {
		return br, nil
	}

	br = br.modify(version)
	if dstOff>>br.shift == (dstOff+n-1)>>br.shift {
		idx := dstOff >> br.shift
		c, err := br.sub[idx].copyFrom(src, dstOff&((1<<br.shift)-1), srcOff, n, version)
		if err != nil {
			return nil, err
		}
		br.anything |= 1 << idx
		br.pending |= 1 << idx
		br.sub[idx] = c
	} else {
		idx := dstOff >> br.shift
		n1 := (1 << br.shift) - (dstOff & ((1 << br.shift) - 1))
		{
			c, err := br.sub[idx].copyFrom(src, dstOff&((1<<br.shift)-1), srcOff, n1, version)
			if err != nil {
				return nil, err
			}
			br.anything |= 1 << idx
			br.pending |= 1 << idx
			br.sub[idx] = c
		}
		idx++
		n2 := n - n1
		{
			c, err := br.sub[idx].copyFrom(src, 0, srcOff+n1, n2, version)
			if err != nil {
				return nil, err
			}
			br.anything |= 1 << idx
			br.pending |= 1 << idx
			br.sub[idx] = c
		}
	}

	return br, nil
}

func (br *memBranch) clobber(off, size uint64, version uint64) (memTreeNode, error) {
	br = br.modify(version)
	err := br.doClobber(off, size, version)
	if err != nil {
		return nil, err
	}
	return br, nil
}

func (br *memBranch) doClobber(off, size uint64, version uint64) error {
	if off+size > 8<<br.shift {
		return errors.New("clobber out of bounds")
	}

	end := off + size
	for i := off; i < end; i = (i | ((1 << br.shift) - 1)) + 1 {
		if br.anything&(1<<((i>>br.shift)%8)) == 0 {
			// There is nothing to clobber.
			continue
		}

		// Clobber the contents of the child.
		j := end
		if j > (i|((1<<br.shift)-1))+1 {
			j = (i | ((1 << br.shift) - 1)) + 1
		}
		idx := i >> br.shift
		c, err := br.sub[idx].clobber(i&((1<<br.shift)-1), j-i, version)
		if err != nil {
			return err
		}
		switch {
		case !c.hasAny():
			br.anything &^= 1 << idx
			fallthrough
		case !c.hasPending():
			br.pending &^= 1 << idx
		}
		br.sub[idx] = c
	}

	return nil
}

func (br *memBranch) markStored(off, size uint64, version uint64) (memTreeNode, error) {
	if off+size > 8<<br.shift {
		return nil, errors.New("clobber out of bounds")
	}
	br = br.modify(version)
	end := off + size
	for i := off; i < end; i = (i | ((1 << br.shift) - 1)) + 1 {
		if br.pending&(1<<((i>>br.shift)%8)) == 0 {
			continue
		}

		j := end
		if j > (i|((1<<br.shift)-1))+1 {
			j = (i | ((1 << br.shift) - 1)) + 1
		}
		idx := i >> br.shift
		c, err := br.sub[idx].markStored(i&((1<<br.shift)-1), j-i, version)
		if err != nil {
			return nil, err
		}
		if !c.hasPending() {
			br.pending &^= 1 << idx
		}
		br.sub[idx] = c
	}

	return br, nil
}

func (br *memBranch) blockInit(off, size uint64, version uint64) (memTreeNode, error) {
	if off+size > 8<<br.shift {
		return nil, errors.New("blockInit out of bounds")
	}

	br = br.modify(version)
	end := off + size
	for i := off; i < end; i = (i | ((1 << br.shift) - 1)) + 1 {
		j := end
		if j > (i|((1<<br.shift)-1))+1 {
			j = (i | ((1 << br.shift) - 1)) + 1
		}
		idx := i >> br.shift
		c, err := br.sub[idx].blockInit(i&((1<<br.shift)-1), j-i, version)
		if err != nil {
			return nil, err
		}
		br.sub[idx] = c
	}

	return br, nil
}

func (br *memBranch) canInit(off, size uint64) (bool, error) {
	if off+size > 8<<br.shift {
		return false, errors.New("canInit out of bounds")
	}

	end := off + size
	for i := off; i < end; i = (i | ((1 << br.shift) - 1)) + 1 {
		j := end
		if j > (i|((1<<br.shift)-1))+1 {
			j = (i | ((1 << br.shift) - 1)) + 1
		}
		ok, err := br.sub[i>>br.shift].canInit(i&((1<<br.shift)-1), j-i)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}

func (br *memBranch) hasAny() bool {
	return br.anything != 0
}

func (br *memBranch) hasPending() bool {
	return br.pending != 0
}

func (br *memBranch) modify(version uint64) *memBranch {
	if br.version < version {
		dup := *br
		dup.version = version
		return &dup
	}

	return br
}

func (br *memBranch) String() string {
	sub := make([]string, len(br.sub))
	for i, s := range br.sub {
		sub[i] = s.String()
	}
	return "br:\n\t" + strings.ReplaceAll(strings.Join(sub, "\n"), "\n", "\n\t")
}

type memBranchSpecial struct {
	v       value
	off     uint64
	pending bool
}

type memLeaf struct {
	specials []value

	version uint64

	pending uint64

	// ^(high | low): unknown
	// high &^ low: special
	// low &^ high: undef
	// high & low: raw bytes
	metaHigh, metaLow uint64

	specialStarts uint64

	noInit uint64

	data [64]byte
}

func (l *memLeaf) store(v value, off uint64, version uint64) (memTreeNode, error) {
	// Check size and index.
	size := v.typ().bytes()
	if off+size > 64 {
		return nil, errors.New("store out of bounds")
	}
	if size == 0 {
		return l, nil
	}

	l = l.modify(version)

	// Clobber the stored range.
	var mask uint64 = ((1 << size) - 1) << off
	if err := l.doClobber(mask); err != nil {
		return nil, err
	}

	// Store the value.
	if err := l.doStore(v, off, size); err != nil {
		return nil, err
	}

	if assert {
		l.check()
	}

	return l, nil
}

func (l *memLeaf) doStore(v value, idx, size uint64) error {
	switch t := v.typ().(type) {
	case nonAggTyp:
		return l.storeNonAgg(v, idx, size)

	case arrType:
		// Store array elements seperately.
		elemSize := t.of.bytes()
		n := t.n
		for i := uint32(0); i < n; i++ {
			err := l.doStore(extractValue(v, i), idx+(elemSize*uint64(i)), elemSize)
			if err != nil {
				return err
			}
		}
		return nil

	case *structType:
		// Store field values seperately.
		var padding uint64 = (1 << size) - 1
		for i, f := range t.fields {
			size := f.ty.bytes()
			err := l.doStore(extractValue(v, uint32(i)), idx+f.offset, size)
			if err != nil {
				return err
			}
			padding &^= ((1 << size) - 1) << f.offset
		}

		// Fill padding with undef.
		padding <<= idx
		l.metaLow |= padding
		l.pending |= padding

		if assert {
			l.check()
		}

		return nil

	default:
		return todo("store " + t.String())
	}
}

func (l *memLeaf) storeNonAgg(v value, idx, size uint64) error {
	if v.val == nil {
		panic("nil value")
	}
	switch val := v.val.(type) {
	case floatVal:
		// Store as an integer.
		return l.storeNonAgg(cast(i32, v), idx, size)

	case doubleVal:
		// Store as an integer.
		return l.storeNonAgg(cast(i64, v), idx, size)

	case smallInt:
		// Convert to little endian and store as raw bytes.
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], v.raw)
		copy(l.data[idx:], buf[:size])
		var mask uint64 = ((1 << size) - 1) << idx
		l.metaHigh |= mask
		l.metaLow |= mask
		l.pending |= mask
		return nil

	case undef:
		// Mark any whole bytes as undef.
		bits := v.typ().(nonAggTyp).bits()
		wholeBytes := bits / 8
		var mask uint64 = ((1 << wholeBytes) - 1) << idx
		l.metaLow |= mask
		l.pending |= mask
		if bits%8 == 0 {
			return nil
		}

		// Store remaining bits as a special value.
		// This is mainly for `i1 undef`.
		idx += wholeBytes
		size = 1
		v = undefValue(iType(bits % 8))

	case *offAddr:
		// Normalize to offPtr and store as a special.
		v.val = (*offPtr)(val)

	case *offPtr, uglyGEP, extractedValue, runtimeValue, bitSlice:
		// These values must be stored as specials.

	case castVal:
		src := value{val.val, v.raw}
		toBits := val.to.bits()
		fromBits := src.typ().(nonAggTyp).bits()
		if toBits >= fromBits {
			// Store the underlying value.
			// If it is not a multiple of byte size, it will be zero-extended later.
			fromBytes := (fromBits + 7) / 8
			err := l.storeNonAgg(src, idx, fromBytes)
			if err != nil {
				return err
			}

			// Zero-pad it.
			toBytes := (toBits + 7) / 8
			padding := l.data[idx:][fromBytes:toBytes]
			for i := range padding {
				padding[i] = 0
			}
			var mask uint64 = ((1 << toBytes) - (1 << fromBytes)) << idx
			l.metaHigh |= mask
			l.metaLow |= mask
			l.pending |= mask
			return nil
		}

	case *bitCat:
		parts := *val
		totalWidth := v.raw
		var width uint64
	split:
		for i := 0; i < len(parts); i++ {
			part := parts[i]
			partWidth := part.typ().(nonAggTyp).bits()
			if width != 0 {
				if 8-(width&7) < partWidth {
					switch part.val.(type) {
					case smallInt, undef:
						width = (width | 7) + 1
						break split
					}
				}
			}
			width += partWidth
			if width%8 == 0 {
				break
			}
		}
		if width < totalWidth {
			// Split this into more reasonable chunks.
			low, high := slice(v, 0, width), slice(v, width, totalWidth-width)
			lowSize, highSize := (width+7)/8, ((totalWidth-width)+7)/8
			err := l.storeNonAgg(low, idx, lowSize)
			if err != nil {
				return err
			}
			return l.storeNonAgg(high, idx+lowSize, highSize)
		}

	default:
		return todo("store " + v.String())
	}

	// Store as a special value.
	var mask uint64 = ((1 << size) - 1) << idx
	spIdx := bits.OnesCount64(l.specialStarts & ((1 << idx) - 1))
	l.specials = append(l.specials, value{})
	copy(l.specials[spIdx+1:], l.specials[spIdx:])
	l.specials[spIdx] = v
	l.metaHigh |= mask
	l.specialStarts |= 1 << idx
	l.pending |= mask

	if assert {
		l.check()
	}

	return nil
}

func (l *memLeaf) fillUndef(off, size uint64, version uint64) (memTreeNode, error) {
	if off+size > 64 {
		return nil, errors.New("fillUndef out of bounds")
	}

	l = l.modify(version)

	// Clobber the stored range.
	var mask uint64 = ((1 << size) - 1) << off
	if err := l.doClobber(mask); err != nil {
		return nil, err
	}

	// Fill with undef.
	l.metaLow |= mask

	if assert {
		l.check()
	}

	return l, nil
}

func (l *memLeaf) memSet(off, size uint64, v byte, version uint64) (memTreeNode, error) {
	if off+size > 64 {
		return nil, errors.New("memSet out of bounds")
	}

	l = l.modify(version)

	// Clobber the stored range.
	var mask uint64 = ((1 << size) - 1) << off
	if err := l.doClobber(mask); err != nil {
		return nil, err
	}

	// Set the bytes in the range.
	toFill := l.data[off:][:size]
	for i := range toFill {
		toFill[i] = v
	}
	l.metaHigh |= mask
	l.metaLow |= mask
	l.pending |= mask

	if assert {
		l.check()
	}

	return l, nil
}

func (l *memLeaf) copyTo(dst memTreeNode, dstOff, srcOff, n uint64, version uint64) (memTreeNode, error) {
	return dst.copyFrom(l, dstOff, srcOff, n, version)
}

func (l *memLeaf) copyFrom(src *memLeaf, dstOff, srcOff, n uint64, version uint64) (memTreeNode, error) {
	if dstOff+n > 64 || srcOff+n > 64 {
		return nil, errors.New("copy out of bounds")
	}

	// Check that the source data is available.
	var srcMask uint64 = ((1 << n) - 1) << srcOff
	if srcMask&^(src.metaHigh|src.metaLow) != 0 {
		return nil, errRuntime
	}

	// Clobber the destination range.
	l = l.modify(version)
	var dstMask uint64 = ((1 << n) - 1) << dstOff
	if err := l.doClobber(dstMask); err != nil {
		return nil, err
	}
	if l == src {
		panic("inconsistent state")
	}

	// Copy raw bytes and undefs.
	copy(l.data[dstOff:][:n], src.data[srcOff:][:n])
	rawCpyMask := ((src.metaLow & srcMask) >> srcOff) << dstOff
	l.metaHigh |= rawCpyMask
	l.pending |= rawCpyMask
	l.metaLow |= rawCpyMask

	if assert {
		l.check()
	}

	// Copy specials.
	for spMask := (src.metaHigh &^ src.metaLow) & srcMask; spMask != 0; {
		i := uint64(bits.TrailingZeros64(spMask))
		tmp := src.specialStarts & ((1 << (i + 1)) - 1)
		j := uint64(bits.OnesCount64(tmp)) - 1
		start := uint64(bits.Len64(tmp)) - 1
		end := uint64(bits.TrailingZeros64((src.specialStarts | ^spMask) & -(1 << (i + 1))))
		err := l.storeNonAgg(slice(src.specials[j], 8*(i-start), 8*(end-i)), (i-srcOff)+dstOff, end-i)
		if err != nil {
			return nil, err
		}
		spMask &= -(1 << end)
	}

	if assert {
		l.check()
	}

	return l, nil
}

func (l *memLeaf) clobber(off, size uint64, version uint64) (memTreeNode, error) {
	if off+size > 64 {
		return nil, errors.New("clobber out of bounds")
	}

	l = l.modify(version)

	// Clobber the range.
	var mask uint64 = ((1 << size) - 1) << off
	if err := l.doClobber(mask); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *memLeaf) doClobber(mask uint64) error {
	if spMask := mask & l.metaHigh &^ l.metaLow; spMask != 0 {
		// Clobber specials.
		oldSp := l.specials
		oldStarts := l.specialStarts
		splits := l.specialStarts | l.metaLow | ^l.metaHigh
		sp := make([]value, 64)
		i, j := 0, 0
		var newStarts uint64
		for m := oldStarts; m != 0; {
			// Select the next special from the input.
			start := bits.TrailingZeros64(m)
			m &^= 1 << start
			end := bits.TrailingZeros64(splits & -(1 << (start + 1)))
			v := oldSp[i]
			i++

			outMask := (((1 << end) - (1 << start)) &^ mask) >> start
			switch {
			case outMask == 0:
				// This is completely discarded.

			case outMask == (1<<(end-start))-1:
				// This is passed through as-is.
				newStarts |= 1 << start
				sp[j] = v
				j++

			default:
				// Slice it into pieces on the output.
				vBits := v.typ().(nonAggTyp).bits()
				for outMask != 0 {
					off := bits.TrailingZeros64(outMask)
					width := bits.TrailingZeros64(^(outMask >> off))
					outMask &^= ((1 << width) - 1) << off
					offBits, widthBits := 8*uint64(off), 8*uint64(width)
					if offBits+widthBits > vBits {
						widthBits = vBits - offBits
					}
					newStarts |= 1 << off
					sp[j] = slice(v, offBits, widthBits)
					j++
				}
			}
		}
		l.specialStarts = newStarts
		l.specials = append(oldSp[:0], sp[:j]...)
	}

	l.metaHigh &^= mask
	l.metaLow &^= mask
	l.pending &^= mask

	if assert {
		l.check()
	}

	return nil
}

func (l *memLeaf) markStored(off, size uint64, version uint64) (memTreeNode, error) {
	if off+size > 64 {
		return nil, errors.New("markStored out of bounds")
	}
	l = l.modify(version)
	l.pending &^= ((1 << size) - 1) << off

	if assert {
		l.check()
	}

	return l, nil
}

func (l *memLeaf) blockInit(off, size uint64, version uint64) (memTreeNode, error) {
	idx := off % 64
	if idx+size > 64 {
		return nil, errors.New("blockInit out of bounds")
	}
	l = l.modify(version)
	l.noInit |= ((1 << size) - 1) << idx
	return l, nil
}

func (l *memLeaf) load(to typ, off uint64) (value, error) {
	size := to.bytes()
	if off+size > 64 {
		return value{}, errors.New("store out of bounds")
	}

	return l.doLoad(to, off, size)
}

func (l *memLeaf) doLoad(to typ, idx, size uint64) (value, error) {
	switch to := to.(type) {
	case nonAggTyp:
		// Load the raw non-aggregate value.
		return l.loadNonAgg(to, idx, size)

	case arrType:
		// Load array elements seperatey.
		elemSize := to.of.bytes()
		elems := make([]value, to.n)
		for i := range elems {
			v, err := l.doLoad(to.of, idx+elemSize*uint64(i), elemSize)
			if err != nil {
				return value{}, err
			}
			elems[i] = v
		}
		return arrayValue(to.of, elems...), nil

	case *structType:
		// Load fields seperately.
		fields := make([]value, len(to.fields))
		for i, f := range to.fields {
			v, err := l.doLoad(f.ty, idx+f.offset, f.ty.bytes())
			if err != nil {
				return value{}, err
			}
			fields[i] = v
		}
		return structValue(to, fields...), nil

	default:
		return value{}, todo("load " + to.String())
	}
}

func (l *memLeaf) loadNonAgg(to nonAggTyp, idx, size uint64) (value, error) {
	if size > 8 {
		return value{}, todo("big load")
	}

	var mask uint64 = ((1 << size) - 1) << idx
	switch {
	case mask&l.metaHigh&l.metaLow == mask:
		// Load a raw value from bytes.
		var buf [8]byte
		copy(buf[:], l.data[idx:])
		return cast(to, smallIntValue(iType(to.bits()), binary.LittleEndian.Uint64(buf[:]))), nil

	case mask&l.metaHigh&^l.metaLow == mask && mask&l.specialStarts&^(1<<idx) == 0:
		// Load from a special.
		spStart := uint64(bits.Len64(l.specialStarts&^((-(1 << idx))<<1)) - 1)
		special := l.specials[bits.OnesCount64(l.specialStarts&((1<<spStart)-1))]
		shift := 8 * (idx - spStart)
		return cast(to, slice(special, shift, special.typ().(nonAggTyp).bits()-shift)), nil

	case mask&l.metaLow&^l.metaHigh == mask:
		// The result is undefined.
		return undefValue(to), nil

	case mask&(l.metaHigh|l.metaLow) != mask:
		// This requires a runtime load.
		return value{}, errRuntime

	default:
		// The result consists of multiple concatenated values.
		splits := (l.metaHigh ^ (l.metaHigh << 1)) |
			(l.metaLow ^ (l.metaLow << 1)) |
			l.specialStarts |
			(1 << idx) |
			(1 << (idx + size))
		vals := make([]value, bits.OnesCount64(splits&mask))[:0]
		for i := idx; i < idx+size; {
			end := uint64(bits.TrailingZeros64(splits & ((-(1 << i)) << 1)))
			v, err := l.loadNonAgg(iType(8*(end-i)), i, end-i)
			if err != nil {
				return value{}, err
			}
			vals = append(vals, v)
			i = end
		}
		return cast(to, cat(vals)), nil
	}
}

func (l *memLeaf) canInit(off, size uint64) (bool, error) {
	if off+size > 64 {
		return false, errors.New("canInit out of bounds")
	}

	return l.noInit&(((1<<size)-1)<<off) == 0, nil
}

func (l *memLeaf) hasAny() bool {
	return l.metaHigh|l.metaLow != 0
}

func (l *memLeaf) hasPending() bool {
	return l.pending != 0
}

func (l *memLeaf) modify(version uint64) *memLeaf {
	if l.version < version {
		dup := *l
		dup.version = version
		dup.specials = append([]value(nil), l.specials...)
		return &dup
	}

	return l
}

func (l *memLeaf) String() string {
	const tbl = "?USB"
	var meta [len(l.data)]byte
	for i := range meta {
		meta[i] = tbl[(((l.metaHigh>>i)&1)<<1)|
			((l.metaLow>>i)&1)]
		if (l.specialStarts>>i)&1 != 0 {
			if meta[i] != 'S' {
				panic("special start is not a special")
			}
			meta[i] = '!'
		}
	}
	specials := make([]string, len(l.specials))
	for i, v := range l.specials {
		specials[i] = v.String()
	}
	return fmt.Sprintf("leaf:\n\tversion %d\n\tmeta    %s\n\tnoinit  %s\n\tpending %s\n\tbytes %x\n\tspecials %v", l.version, string(meta[:]), binLEStr(l.noInit), binLEStr(l.pending), l.data, specials)
}

func (l *memLeaf) check() {
	for i := range l.data {
		if (l.specialStarts>>i)&1 != 0 {
			if (1<<i)&l.metaHigh&^l.metaLow == 0 {
				panic("special start is not a special")
			}
		}
	}
	if len(l.specials) != bits.OnesCount64(l.specialStarts) {
		panic("mismatched special count")
	}
	for i, v := range l.specials {
		if v == (value{}) {
			println(len(l.specials), i)
			panic("empty special value")
		}
	}
}

func binLEStr(mask uint64) string {
	var data [64]byte
	for i := range data {
		data[i] = '0' + byte((mask>>i)&1)
	}
	return string(data[:])
}
