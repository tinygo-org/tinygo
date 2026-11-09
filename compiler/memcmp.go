package compiler

import (
	"fmt"
	"go/types"
	"strconv"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// memHashWithPattern hashes a value in memory.
func (c *context) memHashWithPattern(b llvm.Builder, pattern memCmpPattern, ptr, seed llvm.Value, name string) llvm.Value {
	// Get or create the pattern hash function.
	fn := c.memHashFunc(pattern)

	// Call the function.
	size := pattern.size()
	call := b.CreateCall(fn.GlobalValueType(), fn, []llvm.Value{
		ptr,
		llvm.ConstInt(c.uintptrType, size, false),
		seed,
		c.dataPtrUndef,
	}, name)

	if size > 0 {
		// Mark the whole type memory as dereferenceable.
		// This permits insertion of speculative loads by LLVM.
		dereferenceableAttr := c.ctx.CreateEnumAttribute(llvm.AttributeKindID("dereferenceable"), size)
		call.AddCallSiteAttribute(1, dereferenceableAttr)
	}

	return call
}

// memHashFunc finds or creates a function to hash a value in memory.
func (c *context) memHashFunc(pattern memCmpPattern) llvm.Value {
	switch pattern := pattern.(type) {
	case memCmpBasic:
		_, fn := c.getRuntimeFunction(memHashPrimitiveFuncs[pattern&memCmpKindMask])
		return fn

	case memCmpArray:
		return c.memHashArrayFunc(pattern.element)

	case *memCmpStruct:
		return c.memHashStructFunc(pattern)

	default:
		panic("invalid node")
	}
}

var memHashPrimitiveFuncs = [...]string{
	memCmpKindNop:       "hashmapNopHash",
	memCmpKindByte:      "hash32",
	memCmpKindFloat32:   "hashmapFloat32Hash",
	memCmpKindFloat64:   "hashmapFloat64Hash",
	memCmpKindString:    "hashmapStringPtrHash",
	memCmpKindInterface: "hashmapInterfacePtrHash",
}

// memHashArrayFunc finds or creates a function to hash an array in memory.
// The provided pattern is that of the array element.
// The "size" (in bytes) argument is used to determine the array length.
func (c *context) memHashArrayFunc(pattern memCmpPattern) llvm.Value {
	if pattern == memCmpByte {
		// Use the specialized function to hash arrays.
		_, fn := c.getRuntimeFunction("hash32")
		return fn
	}

	// Check if the function already exists.
	funcName := string(pattern.appendName([]byte("_tinygo_array_hash_")))
	if fn := c.mod.NamedFunction(funcName); !fn.IsNil() {
		return fn
	}

	// Create the LLVM function.
	fn := c.defineMemHashFunc(funcName)
	keyPtr := fn.Param(0)
	size := fn.Param(1)
	seed := fn.Param(2)

	// Create the loop entry path.
	b := c.ctx.NewBuilder()
	defer b.Dispose()
	entryBlock := c.ctx.AddBasicBlock(fn, "entry")
	b.SetInsertPointAtEnd(entryBlock)
	loopBlock := c.ctx.AddBasicBlock(fn, "loop")
	b.CreateBr(loopBlock)
	b.SetInsertPointAtEnd(loopBlock)
	prevOffset := b.CreatePHI(c.uintptrType, "offset.prev")
	prevHash := b.CreatePHI(c.i32, "hash.prev")

	// Compute the new element offset.
	offset := b.CreateNSWSub(prevOffset, llvm.ConstInt(c.uintptrType, pattern.size(), false), "offset")

	// Compute the element pointer.
	element := b.CreateInBoundsGEP(c.i8, keyPtr, []llvm.Value{offset}, "elem")

	// Hash the element.
	elemHash := c.memHashWithPattern(b, pattern, element, seed, "hash.elem")

	// Combine the hash from the previous iteration.
	prevHashMul := b.CreateMul(prevHash, llvm.ConstInt(c.i32, 31, false), "hash.prev.mul")
	hash := b.CreateXor(prevHashMul, elemHash, "hash")

	// Exit the loop if there are no more elements.
	done := b.CreateICmp(llvm.IntEQ, offset, c.uintptrZero, "done")
	retBlock := c.ctx.AddBasicBlock(fn, "ret")
	b.CreateCondBr(done, retBlock, loopBlock)
	prevOffset.AddIncoming(
		[]llvm.Value{size, offset},
		[]llvm.BasicBlock{entryBlock, loopBlock},
	)
	prevHash.AddIncoming(
		[]llvm.Value{llvm.ConstNull(c.i32), hash},
		[]llvm.BasicBlock{entryBlock, loopBlock},
	)

	// Return the resulting hash.
	b.SetInsertPointAtEnd(retBlock)
	b.CreateRet(hash)

	return fn
}

// memHashStructFunc finds or creates a function to hash a struct in memory.
func (c *context) memHashStructFunc(pattern *memCmpStruct) llvm.Value {
	// Check if the function already exists.
	funcName := string(pattern.appendName([]byte("_tinygo_hash_")))
	if fn := c.mod.NamedFunction(funcName); !fn.IsNil() {
		return fn
	}

	// Create the LLVM function.
	fn := c.defineMemHashFunc(funcName)
	keyPtr := fn.Param(0)
	seed := fn.Param(2)

	// Hash the fields.
	b := c.ctx.NewBuilder()
	defer b.Dispose()
	b.SetInsertPointAtEnd(c.ctx.AddBasicBlock(fn, ""))
	multiplier := llvm.ConstInt(c.i32, 31, false)
	hash := llvm.ConstNull(c.i32)
	for _, field := range pattern.fields {
		// Compute the field pointer.
		offset := llvm.ConstInt(c.uintptrType, field.offset, false)
		fieldPtr := b.CreateInBoundsGEP(c.i8, keyPtr, []llvm.Value{offset}, "field.ptr")

		// Hash the field.
		fieldHash := c.memHashWithPattern(b, field.element, fieldPtr, seed, "field.hash")

		// Merge the field hash.
		hash = b.CreateXor(b.CreateMul(hash, multiplier, ""), fieldHash, "")
	}

	// Return the resulting hash.
	b.CreateRet(hash)

	return fn
}

func (c *context) defineMemHashFunc(name string) llvm.Value {
	// Create the function type.
	// (key ptr, size uintptr, seed uintptr, context ptr) -> i32
	fnType := llvm.FunctionType(c.i32, []llvm.Type{
		c.dataPtrType, c.uintptrType, c.uintptrType, c.dataPtrType,
	}, false)

	// Create the function.
	fn := llvm.AddFunction(c.mod, name, fnType)
	fn.SetLinkage(llvm.LinkOnceODRLinkage)
	fn.SetUnnamedAddr(true)
	c.addStandardAttributes(fn)

	// Set the argument names.
	fn.Param(0).SetName("key")
	fn.Param(1).SetName("size")
	fn.Param(2).SetName("seed")
	fn.Param(3).SetName("context")

	// The arguments are not captured.
	nocaptureAttr := c.ctx.CreateEnumAttribute(llvm.AttributeKindID(llvmutil.NoCaptureAttrName()), 0)
	fn.AddAttributeAtIndex(1, nocaptureAttr)

	// The arguments are not modified.
	readonlyAttr := c.ctx.CreateEnumAttribute(llvm.AttributeKindID("readonly"), 0)
	fn.AddAttributeAtIndex(1, readonlyAttr)

	return fn
}

// memCmpWithPattern creates a typed comparison of two values in memory.
func (c *context) memCmpWithPattern(b llvm.Builder, pattern memCmpPattern, x, y llvm.Value, align uint64, name string) llvm.Value {
	// Get or create the pattern comparison function.
	fn := c.memCmpFunc(pattern)

	// Call the function.
	size := pattern.size()
	call := b.CreateCall(fn.GlobalValueType(), fn, []llvm.Value{
		x,
		y,
		llvm.ConstInt(c.uintptrType, size, false),
		c.dataPtrUndef,
	}, name)

	if size > 0 {
		// Mark the whole type memory as dereferenceable.
		// This permits insertion of speculative loads by LLVM.
		dereferenceableAttr := c.ctx.CreateEnumAttribute(llvm.AttributeKindID("dereferenceable"), size)
		call.AddCallSiteAttribute(1, dereferenceableAttr)
		call.AddCallSiteAttribute(2, dereferenceableAttr)
	}

	if align > 1 {
		// Attach the type alignment to the arguments.
		alignAttr := c.ctx.CreateEnumAttribute(llvm.AttributeKindID("align"), align)
		call.AddCallSiteAttribute(1, alignAttr)
		call.AddCallSiteAttribute(2, alignAttr)
	}

	return call
}

// memCmpFunc finds or creates a function to compare values in memory.
func (c *context) memCmpFunc(pattern memCmpPattern) llvm.Value {
	switch pattern := pattern.(type) {
	case memCmpBasic:
		_, fn := c.getRuntimeFunction(memCmpPrimitiveFuncs[pattern&memCmpKindMask])
		return fn

	case memCmpArray:
		return c.memCmpArrayFunc(pattern.element)

	case *memCmpStruct:
		return c.memCmpStructFunc(pattern)

	default:
		panic("invalid node")
	}
}

var memCmpPrimitiveFuncs = [...]string{
	memCmpKindNop:       "memCompareNop",
	memCmpKindByte:      "memCompareBytes",
	memCmpKindFloat32:   "memCompareFloat32",
	memCmpKindFloat64:   "memCompareFloat64",
	memCmpKindString:    "memCompareString",
	memCmpKindInterface: "memCompareInterface",
}

// memCmpArrayFunc finds or creates a function to compare an array in memory.
// The provided pattern is that of the array element.
// The "size" (in bytes) argument is used to determine the array length.
func (c *context) memCmpArrayFunc(pattern memCmpPattern) llvm.Value {
	if pattern == memCmpByte {
		// Use the specialized function to compare byte arrays.
		_, fn := c.getRuntimeFunction("memCompareBytes")
		return fn
	}

	// Check if the function already exists.
	funcName := string(pattern.appendName([]byte("_tinygo_array_cmp_")))
	if fn := c.mod.NamedFunction(funcName); !fn.IsNil() {
		return fn
	}

	// Create the LLVM function.
	fn := c.defineMemCmpFunc(funcName)
	xPtr := fn.Param(0)
	yPtr := fn.Param(1)
	size := fn.Param(2)
	// TODO: apply a range attr to size once possible

	// Create the loop entry path.
	entryBlock := c.ctx.AddBasicBlock(fn, "entry")
	b := c.ctx.NewBuilder()
	defer b.Dispose()
	b.SetInsertPointAtEnd(entryBlock)
	loopBlock := c.ctx.AddBasicBlock(fn, "loop")
	b.CreateBr(loopBlock)
	b.SetInsertPointAtEnd(loopBlock)
	prevOffset := b.CreatePHI(c.uintptrType, "offset.prev")

	if pattern.group() == memCmpGroupInterface {
		// We must perform all comparisons.
		// We must use an and-chain instead of returning early.
		prevEq := b.CreatePHI(c.i1, "eq.prev")

		// Compute the new element offset.
		offset := []llvm.Value{b.CreateNSWSub(prevOffset, llvm.ConstInt(c.uintptrType, pattern.size(), false), "offset")}

		// Compute the element pointers.
		xElem := b.CreateInBoundsGEP(c.i8, xPtr, offset, "x.elem")
		yElem := b.CreateInBoundsGEP(c.i8, yPtr, offset, "y.elem")

		// Compare the elements.
		elemEq := c.memCmpWithPattern(b, pattern, xElem, yElem, 1, "eq.elem")

		// Merge the comparisons.
		eq := b.CreateAnd(prevEq, elemEq, "eq")

		// Exit the loop if there are no more elements.
		done := b.CreateICmp(llvm.IntEQ, offset[0], c.uintptrZero, "done")
		retBlock := c.ctx.AddBasicBlock(fn, "ret")
		b.CreateCondBr(done, retBlock, loopBlock)
		prevOffset.AddIncoming(
			[]llvm.Value{size, offset[0]},
			[]llvm.BasicBlock{entryBlock, loopBlock},
		)
		prevEq.AddIncoming(
			[]llvm.Value{c.i1True, eq},
			[]llvm.BasicBlock{entryBlock, loopBlock},
		)

		// Create the return.
		b.SetInsertPointAtEnd(retBlock)
		b.CreateRet(eq)
	} else {
		// Compute the new element offset.
		offset := []llvm.Value{b.CreateNSWSub(prevOffset, llvm.ConstInt(c.uintptrType, pattern.size(), false), "offset")}

		// Compute the element pointers.
		xElem := b.CreateInBoundsGEP(c.i8, xPtr, offset, "x.elem")
		yElem := b.CreateInBoundsGEP(c.i8, yPtr, offset, "y.elem")

		// Compare the elements.
		elemEq := c.memCmpWithPattern(b, pattern, xElem, yElem, 1, "eq.elem")

		// Exit the loop if the comparison failed.
		nextBlock := c.ctx.AddBasicBlock(fn, "next")
		retBlock := c.ctx.AddBasicBlock(fn, "ret")
		b.CreateCondBr(elemEq, nextBlock, retBlock)

		// Otherwise check for more elements.
		b.SetInsertPointAtEnd(nextBlock)
		done := b.CreateICmp(llvm.IntEQ, offset[0], c.uintptrZero, "done")
		b.CreateCondBr(done, retBlock, loopBlock)
		prevOffset.AddIncoming(
			[]llvm.Value{size, offset[0]},
			[]llvm.BasicBlock{entryBlock, nextBlock},
		)

		// Create the return.
		b.SetInsertPointAtEnd(retBlock)
		eq := b.CreatePHI(c.i1, "eq")
		eq.AddIncoming(
			[]llvm.Value{c.i1False, c.i1True},
			[]llvm.BasicBlock{loopBlock, nextBlock},
		)
		b.CreateRet(eq)
	}

	return fn
}

// memCmpStructFunc finds or creates a function to compare structs in memory.
func (c *context) memCmpStructFunc(pattern *memCmpStruct) llvm.Value {
	// Check if the function already exists.
	funcName := string(pattern.appendName([]byte("_tinygo_cmp_")))
	if fn := c.mod.NamedFunction(funcName); !fn.IsNil() {
		return fn
	}

	// Create the LLVM function.
	fn := c.defineMemCmpFunc(funcName)
	xPtr := fn.Param(0)
	yPtr := fn.Param(1)
	curBlock := c.ctx.AddBasicBlock(fn, "")
	b := c.ctx.NewBuilder()
	defer b.Dispose()
	b.SetInsertPointAtEnd(curBlock)

	// Group the fields into buckets.
	var buckets [4][]memCmpStructField
	for _, field := range pattern.fields {
		group := field.element.group()
		buckets[group] = append(buckets[group], field)
	}

	// Process fields that contain interfaces first.
	// These can never be skipped.
	var curEq llvm.Value
	for _, field := range buckets[memCmpGroupInterface] {
		// Compute the field pointers.
		offset := []llvm.Value{llvm.ConstInt(c.uintptrType, field.offset, false)}
		xField := b.CreateInBoundsGEP(c.i8, xPtr, offset, "x.field")
		yField := b.CreateInBoundsGEP(c.i8, yPtr, offset, "y.field")

		// Compare the field contents.
		prevEq := curEq
		curEq = c.memCmpWithPattern(b, field.element, xField, yField, 1, "field.eq")

		if !prevEq.IsNil() {
			// Merge the result with previous comparisons.
			curEq = b.CreateAnd(prevEq, curEq, "")
		}
	}

	// Process the rest of the fields in order of approximate comparison cost.
	// We can stop early with these fields.
	var earlyBlocks []llvm.BasicBlock
	var retBlock llvm.BasicBlock
	for _, bucket := range buckets[:memCmpGroupInterface] {
		for _, field := range bucket {
			if !curEq.IsNil() {
				// Return early if the previous comparison failed.
				if retBlock.IsNil() {
					retBlock = c.ctx.AddBasicBlock(fn, "return")
				}
				earlyBlocks = append(earlyBlocks, curBlock)
				nextBlock := c.ctx.InsertBasicBlock(retBlock, "")
				b.CreateCondBr(curEq, nextBlock, retBlock)
				b.SetInsertPointAtEnd(nextBlock)
				curBlock = nextBlock
			}

			// Compute the field pointers.
			offset := []llvm.Value{llvm.ConstInt(c.uintptrType, field.offset, false)}
			xField := b.CreateInBoundsGEP(c.i8, xPtr, offset, "x.field")
			yField := b.CreateInBoundsGEP(c.i8, yPtr, offset, "y.field")

			// Compare the field contents.
			curEq = c.memCmpWithPattern(b, field.element, xField, yField, 1, "field.eq")
		}
	}

	if retBlock.IsNil() {
		// The entire function is a single block.
		b.CreateRet(curEq)
		return fn
	}

	// Populate the return block.
	b.CreateBr(retBlock)
	b.SetInsertPointAtEnd(retBlock)
	phi := b.CreatePHI(c.i1, "eq")
	incomingValues := make([]llvm.Value, len(earlyBlocks)+1)
	for i := range earlyBlocks {
		incomingValues[i] = c.i1False
	}
	incomingValues[len(earlyBlocks)] = curEq
	phi.AddIncoming(incomingValues, append(earlyBlocks, curBlock))
	b.CreateRet(phi)

	return fn
}

func (c *context) defineMemCmpFunc(name string) llvm.Value {
	// Create the function type.
	// (x ptr, y ptr, size uintptr, context ptr) -> i1
	fnType := llvm.FunctionType(c.i1, []llvm.Type{
		c.dataPtrType, c.dataPtrType, c.uintptrType, c.dataPtrType,
	}, false)

	// Create the function.
	fn := llvm.AddFunction(c.mod, name, fnType)
	fn.SetLinkage(llvm.LinkOnceODRLinkage)
	fn.SetUnnamedAddr(true)
	c.addStandardAttributes(fn)

	// Set the argument names.
	fn.Param(0).SetName("x")
	fn.Param(1).SetName("y")
	fn.Param(2).SetName("size")
	fn.Param(3).SetName("context")

	// The arguments are not captured.
	nocaptureAttr := c.ctx.CreateEnumAttribute(llvm.AttributeKindID(llvmutil.NoCaptureAttrName()), 0)
	fn.AddAttributeAtIndex(1, nocaptureAttr)
	fn.AddAttributeAtIndex(2, nocaptureAttr)

	// The arguments are not modified.
	readonlyAttr := c.ctx.CreateEnumAttribute(llvm.AttributeKindID("readonly"), 0)
	fn.AddAttributeAtIndex(1, readonlyAttr)
	fn.AddAttributeAtIndex(2, readonlyAttr)

	return fn
}

// isByteComparable reports whether t can be compared with
// runtime.memCompareBytes.
func (c *context) isByteComparable(t types.Type) bool {
	// Get the memory comparison pattern for the type.
	pattern := c.memCmpPattern(t)

	if pattern == memCmpNop {
		// This is binary-comparable iff t is zero-sized.
		return c.targetData.TypeAllocSize(c.getLLVMType(t)) == 0
	}

	if arr, ok := pattern.(memCmpArray); ok {
		pattern = arr.element
	}

	return pattern == memCmpByte
}

// memCmpPattern creates a memory comparison pattern for the provided type.
// This may be used to create a hash or comparison function.
func (c *context) memCmpPattern(t types.Type) memCmpPattern {
	// Strip any name/alias from the type.
	t = t.Underlying()

	// Check for an existing pattern.
	if node := c.memCmpByType.At(t); node != nil {
		return node.(memCmpPattern)
	}

	var node memCmpPattern
	switch t := t.(type) {
	case *types.Basic:
		kind := t.Kind()
		switch kind {
		case types.Bool, types.Uint8, types.Int8:
			node = memCmpByte
		case types.Float32:
			node = memCmpFloat32
		case types.Float64:
			node = memCmpFloat64
		case types.Complex64:
			node = memCmpArray{
				element: memCmpFloat32,
				length:  2,
			}
		case types.Complex128:
			node = memCmpArray{
				element: memCmpFloat64,
				length:  2,
			}
		case types.String:
			node = memCmpKindString | memCmpBasic(c.targetData.TypeAllocSize(c.getLLVMRuntimeType("_string"))<<memCmpSizeShift)
		default:
			node = c.memCmpPatternByteArray(t)
		}

	case *types.Pointer, *types.Chan:
		node = c.memCmpPatternByteArray(t)

	case *types.Interface:
		node = memCmpKindInterface | memCmpBasic(c.targetData.TypeAllocSize(c.getLLVMRuntimeType("_interface"))<<memCmpSizeShift)

	case *types.Struct:
		node = c.memCmpPatternStruct(t)

	case *types.Array:
		node = memCmpPatternArray(c.memCmpPattern(t.Elem()), uint64(t.Len()))

	default:
		panic(fmt.Sprintf("unhandled type for memCmpPattern: %T", t))
	}

	// Save the pattern for future use.
	c.memCmpByType.Set(t, node)

	return node
}

// memCmpPattern is part of a tree representing the process to compare/hash a type in memory.
type memCmpPattern interface {
	// size of the type in bytes.
	size() uint64
	// appendName generates a unique identifier for this node and appends it to dst.
	appendName(dst []byte) []byte
	// fmt.Stringer returns the identifier generated by appendName.
	fmt.Stringer
	// group determines when a comparison should be performed relative to other nodes in a struct.
	group() memCmpGroup
}

// memCmpGroup determines when a comparison should be performed relative to other nodes in a struct.
type memCmpGroup uint8

const (
	// memCmpGroupBytes indicates that a memCmpPattern only contains byte comparisons.
	memCmpGroupBytes memCmpGroup = iota
	// memCmpGroupFloat indicates that a memCmpPattern only contains float/byte comparisons.
	// Floating-point comparisons are expensive without an FPU, so it is best to wait until all byte comparisons have passed.
	memCmpGroupFloat
	// memCmpGroupString indicates that a memCmpPattern contains string comparisons but no interfaces.
	// String comparisons are expensive, so it is best to wait until all other comparisons have passed.
	memCmpGroupString
	// memCmpGroupInterface indicates that a memCmpPattern contains interface comparisons.
	// It must always be performed, even if another comparison has failed.
	memCmpGroupInterface
)

func (c *context) memCmpPatternStruct(t *types.Struct) memCmpPattern {
	// Get the underlying type.
	llvmType := c.getLLVMType(t)

	// Convert and coalesce the fields.
	var fields []memCmpStructField
	for i := 0; i < t.NumFields(); i++ {
		// Get the field info.
		field := t.Field(i)
		if field.Name() == "_" {
			// Blank fields are ignored by comparisons.
			continue
		}

		// Get or create a pattern for the field type.
		element := c.memCmpPattern(field.Type())
		if element == memCmpNop {
			// Discard the nop.
			continue
		}

		// Compute the offset of the field.
		offset := c.targetData.ElementOffset(llvmType, i)

		if inner, ok := element.(*memCmpStruct); ok && len(inner.fields) == 1 {
			// The struct has a single field plus padding.
			// Inline the field.
			element = inner.fields[0].element
			offset += inner.fields[0].offset
		}

		if len(fields) > 0 {
			if prev := fields[len(fields)-1]; prev.offset+prev.element.size() == offset {
				// The new field is adjacent to the previous field.
				lhs := memCmpArray{
					element: prev.element,
					length:  1,
				}
				if arr, ok := lhs.element.(memCmpArray); ok {
					lhs = arr
				}
				rhs := memCmpArray{
					element: element,
					length:  1,
				}
				if arr, ok := rhs.element.(memCmpArray); ok {
					rhs = arr
				}
				if lhs.element == rhs.element {
					// Process struct{[a]T, [b]T} as struct{[a+b]T}.
					fields[len(fields)-1].element = memCmpPatternArray(lhs.element, lhs.length+rhs.length)
					continue
				}
			}
		}

		// Append the new field.
		fields = append(fields, memCmpStructField{
			element: element,
			offset:  offset,
		})
	}
	if len(fields) == 0 {
		// This is a struct{} or equivalent.
		return memCmpNop
	}

	// Compute the padded size of the struct.
	size := c.targetData.TypeAllocSize(llvmType)

	if len(fields) == 1 && fields[0].offset == 0 && fields[0].element.size() == size {
		// Process struct{T} as T.
		return fields[0].element
	}

	// Check if the struct pattern already exists.
	nameBytes := (&memCmpStruct{
		fields:     fields,
		paddedSize: size,
	}).appendName(nil)
	if dup := c.memCmpStructs[string(nameBytes)]; dup != nil {
		// There is already an identical struct pattern.
		return dup
	}

	// Register the struct pattern.
	pattern := &memCmpStruct{
		fields:     fields,
		paddedSize: size,
	}
	c.memCmpStructs[string(nameBytes)] = pattern

	return pattern
}

// memCmpStruct is used to hold a group of comparisons.
// NOTE: This may not correspond 1-to-1 with an actual struct in the source.
type memCmpStruct struct {
	// fields hold the individual component comparisons.
	// They are listed in ascending offset order and do not overlap.
	// There must always be at least one field.
	fields []memCmpStructField
	// paddedSize of the type in memory.
	paddedSize uint64
}

func (cmp *memCmpStruct) size() uint64 {
	return cmp.paddedSize
}

func (cmp *memCmpStruct) appendName(dst []byte) []byte {
	dst = strconv.AppendUint(dst, cmp.paddedSize, 10)
	dst = append(dst, '{')
	if fields := cmp.fields; len(fields) > 0 {
		dst = fields[0].appendName(dst)
		for _, f := range fields[1:] {
			dst = append(dst, ',')
			dst = f.appendName(dst)
		}
	}
	dst = append(dst, '}')
	return dst
}

func (cmp *memCmpStruct) String() string {
	return string(cmp.appendName(nil))
}

func (cmp *memCmpStruct) group() memCmpGroup {
	group := memCmpGroupBytes
	for _, f := range cmp.fields {
		group = max(group, f.element.group())
	}
	return group
}

type memCmpStructField struct {
	element memCmpPattern
	offset  uint64
}

func (cmp memCmpStructField) appendName(dst []byte) []byte {
	dst = strconv.AppendUint(dst, cmp.offset, 10)
	dst = append(dst, ':')
	dst = cmp.element.appendName(dst)
	return dst
}

func (cmp memCmpStructField) String() string {
	return string(cmp.appendName(nil))
}

// memCmpPatternByteArray creates a memCmpPattern for a binary-comparable type.
func (c *context) memCmpPatternByteArray(t types.Type) memCmpPattern {
	return memCmpPatternArray(memCmpByte, c.targetData.TypeAllocSize(c.getLLVMType(t)))
}

// memCmpPatternArray creates a normalized memCmpPattern for [len]elem.
func memCmpPatternArray(elem memCmpPattern, len uint64) memCmpPattern {
	switch len {
	case 0:
		// Process [0]T as struct{}.
		return memCmpNop
	case 1:
		// Process [1]T as T.
		return elem
	}

	if elem == memCmpNop {
		// Process [a]struct{} as struct{}.
		return memCmpNop
	}

	if arr, ok := elem.(memCmpArray); ok {
		// Process [a][b]T as [a*b]T.
		elem = arr.element
		len *= arr.length
	}

	return memCmpArray{
		element: elem,
		length:  len,
	}
}

// memCmpArray is used to compare arrays of simpler values.
type memCmpArray struct {
	// element pattern of the array.
	// This must not be memCmpNop or another array.
	element memCmpPattern

	// length of the array.
	// This must be at least 2.
	length uint64
}

func (cmp memCmpArray) size() uint64 {
	return cmp.element.size() * cmp.length
}

func (cmp memCmpArray) appendName(dst []byte) []byte {
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, cmp.length, 10)
	dst = append(dst, ']')
	dst = cmp.element.appendName(dst)
	return dst
}

func (cmp memCmpArray) String() string {
	return string(cmp.appendName(nil))
}

func (cmp memCmpArray) group() memCmpGroup {
	return cmp.element.group()
}

// memCmpBasic is a comparison of a builtin type.
type memCmpBasic uint8

const (
	memCmpNop     = memCmpKindNop | 0<<memCmpSizeShift
	memCmpByte    = memCmpKindByte | 1<<memCmpSizeShift
	memCmpFloat32 = memCmpKindFloat32 | 4<<memCmpSizeShift
	memCmpFloat64 = memCmpKindFloat64 | 8<<memCmpSizeShift
)

const (
	// memCmpKindNop is used for comparisons that will always pass.
	memCmpKindNop memCmpBasic = iota
	// memCmpKindByte is used to compare a pair of bytes.
	// This is wrapped by memCmpArray to represent other integer/pointer types.
	memCmpKindByte
	// memCmpKindFloat32 is used to compare a pair of float32 values.
	memCmpKindFloat32
	// memCmpKindFloat64 is used to compare a pair of float64 values.
	memCmpKindFloat64
	// memCmpKindString is used to compare a pair of strings.
	memCmpKindString
	// memCmpKindInterface is used to compare a pair of interfaces.
	memCmpKindInterface
)

const (
	memCmpKindMask  = 7
	memCmpSizeShift = 3
)

func (cmp memCmpBasic) size() uint64 {
	return uint64(cmp >> memCmpSizeShift)
}

func (cmp memCmpBasic) appendName(dst []byte) []byte {
	return append(dst, cmp.String()...)
}

func (cmp memCmpBasic) String() string {
	return memCmpBasicNames[cmp&memCmpKindMask]
}

var memCmpBasicNames = [...]string{
	memCmpKindNop:       "nop",
	memCmpKindByte:      "b",
	memCmpKindFloat32:   "f32",
	memCmpKindFloat64:   "f64",
	memCmpKindString:    "str",
	memCmpKindInterface: "itf",
}

func (cmp memCmpBasic) group() memCmpGroup {
	return memCmpBasicGroups[cmp&memCmpKindMask]
}

var memCmpBasicGroups = [...]memCmpGroup{
	memCmpNop:           memCmpGroupBytes,
	memCmpKindByte:      memCmpGroupBytes,
	memCmpKindFloat32:   memCmpGroupFloat,
	memCmpKindFloat64:   memCmpGroupFloat,
	memCmpKindString:    memCmpGroupString,
	memCmpKindInterface: memCmpGroupInterface,
}
