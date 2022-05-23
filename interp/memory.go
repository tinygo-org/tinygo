package interp

// This file implements memory as used by interp in a reversible way.
// Each new function call creates a new layer which is merged in the parent on
// successful return and is thrown away when the function couldn't complete (in
// which case the function call is done at runtime).
// Memory is not typed, except that there is a difference between pointer and
// non-pointer data. A pointer always points to an object. This implies:
//   * Nil pointers are zero, and are not considered a pointer.
//   * Pointers for memory-mapped I/O point to numeric pointer values, and are
//     thus not considered pointers but regular values. Dereferencing them cannot be
//     done in interp and results in a revert.
//
// Right now the memory is assumed to be little endian. This will need an update
// for big endian arcitectures, if TinyGo ever adds support for one.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"tinygo.org/x/go-llvm"
)

// An object is a memory buffer that may be an already existing global or a
// global created with runtime.alloc or the alloca instruction. If llvmGlobal is
// set, that's the global for this object, otherwise it needs to be created (if
// it is still reachable when the package initializer returns). The
// llvmLayoutType is not necessarily a complete type: it may need to be
// repeated (for example, for a slice value).
//
// Objects are copied in a memory view when they are stored to, to provide the
// ability to roll back interpreting a function.
type object struct {
	llvmGlobal     llvm.Value
	llvmType       llvm.Type // must match llvmGlobal.Type() if both are set, may be unset if llvmGlobal is set
	llvmLayoutType llvm.Type // LLVM type based on runtime.alloc layout parameter, if available
	globalName     string    // name, if not yet created (not guaranteed to be the final name)
	buffer         value     // buffer with value as given by interp, nil if external
	size           uint32    // must match buffer.len(), if available
	constant       bool      // true if this is a constant global
	marked         uint8     // 0 means unmarked, 1 means external read, 2 means external write
}

// clone() returns a cloned version of this object, for when an object needs to
// be written to for example.
func (obj object) clone() object {
	if obj.buffer != nil {
		obj.buffer = obj.buffer.clone()
	}
	return obj
}

// A memoryView is bound to a function activation. Loads are done from this view
// or a parent view (up to the *runner if it isn't included in a view). Stores
// copy the object to the current view.
//
// For details, see the README in the package.
type memoryView struct {
	r       *runner
	parent  *memoryView
	objects map[uint32]object

	// These instructions were added to runtime.initAll while interpreting a
	// function. They are stored here in a list so they can be removed if the
	// execution of the function needs to be rolled back.
	instructions []llvm.Value
}

// extend integrates the changes done by the sub-memoryView into this memory
// view. This happens when a function is successfully interpreted and returns to
// the parent, in which case all changed objects should be included in this
// memory view.
func (mv *memoryView) extend(sub memoryView) {
	if mv.objects == nil && len(sub.objects) != 0 {
		mv.objects = make(map[uint32]object)
	}
	for key, value := range sub.objects {
		mv.objects[key] = value
	}
	mv.instructions = append(mv.instructions, sub.instructions...)
}

// revert undoes changes done in this memory view: it removes all instructions
// created in this memoryView. Do not reuse this memoryView.
func (mv *memoryView) revert() {
	// Erase instructions in reverse order.
	for i := len(mv.instructions) - 1; i >= 0; i-- {
		llvmInst := mv.instructions[i]
		if llvmInst.IsAInstruction().IsNil() {
			// The IR builder will try to create constant versions of
			// instructions whenever possible. If it does this, it's not an
			// instruction and thus shouldn't be removed.
			continue
		}
		llvmInst.EraseFromParentAsInstruction()
	}
}

// markExternalLoad marks the given LLVM value as having an external read. That
// means that the interpreter can still read from it, but cannot write to it as
// that would mean the external read (done at runtime) reads from a state that
// would not exist had the whole initialization been done at runtime.
func (mv *memoryView) markExternalLoad(llvmValue llvm.Value) error {
	return mv.markExternal(llvmValue, 1)
}

// markExternalStore marks the given LLVM value as having an external write.
// This means that the interpreter can no longer read from it or write to it, as
// that would happen in a different order than if all initialization were
// happening at runtime.
func (mv *memoryView) markExternalStore(llvmValue llvm.Value) error {
	return mv.markExternal(llvmValue, 2)
}

// markExternal is a helper for markExternalLoad and markExternalStore, and
// should not be called directly.
func (mv *memoryView) markExternal(llvmValue llvm.Value, mark uint8) error {
	if llvmValue.IsUndef() || llvmValue.IsNull() {
		// Null and undef definitely don't contain (valid) pointers.
		return nil
	}
	if !llvmValue.IsAInstruction().IsNil() || !llvmValue.IsAArgument().IsNil() {
		// These are considered external by default, there is nothing to mark.
		return nil
	}

	if !llvmValue.IsAGlobalValue().IsNil() {
		objectIndex := mv.r.getValue(llvmValue).(pointerValue).index()
		obj := mv.get(objectIndex)
		if obj.marked < mark {
			obj = obj.clone()
			obj.marked = mark
			if mv.objects == nil {
				mv.objects = make(map[uint32]object)
			}
			mv.objects[objectIndex] = obj
			if !llvmValue.IsAGlobalVariable().IsNil() {
				initializer := llvmValue.Initializer()
				if !initializer.IsNil() {
					// Using mark '2' (which means read/write access) because
					// even from an object that is only read from, the resulting
					// loaded pointer can be written to.
					err := mv.markExternal(initializer, 2)
					if err != nil {
						return err
					}
				}
			} else {
				// This is a function. Go through all instructions and mark all
				// objects in there.
				for bb := llvmValue.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
					for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
						opcode := inst.InstructionOpcode()
						if opcode == llvm.Call {
							calledValue := inst.CalledValue()
							if !calledValue.IsAFunction().IsNil() {
								functionName := calledValue.Name()
								if functionName == "llvm.dbg.value" || strings.HasPrefix(functionName, "llvm.lifetime.") {
									continue
								}
							}
						}
						if opcode == llvm.Br || opcode == llvm.Switch {
							// These don't affect memory. Skipped here because
							// they also have a label as operand.
							continue
						}
						numOperands := inst.OperandsCount()
						for i := 0; i < numOperands; i++ {
							// Using mark '2' (which means read/write access)
							// because this might be a store instruction.
							err := mv.markExternal(inst.Operand(i), 2)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	} else if !llvmValue.IsAConstantExpr().IsNil() {
		switch llvmValue.Opcode() {
		case llvm.IntToPtr, llvm.PtrToInt, llvm.BitCast, llvm.GetElementPtr:
			err := mv.markExternal(llvmValue.Operand(0), mark)
			if err != nil {
				return err
			}
		case llvm.Add, llvm.Sub, llvm.Mul, llvm.UDiv, llvm.SDiv, llvm.URem, llvm.SRem, llvm.Shl, llvm.LShr, llvm.AShr, llvm.And, llvm.Or, llvm.Xor:
			// Integer binary operators. Mark both operands.
			err := mv.markExternal(llvmValue.Operand(0), mark)
			if err != nil {
				return err
			}
			err = mv.markExternal(llvmValue.Operand(1), mark)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("interp: unknown constant expression '%s'", instructionNameMap[llvmValue.Opcode()])
		}
	} else if !llvmValue.IsAInlineAsm().IsNil() {
		// Inline assembly can modify globals but only exported globals. Let's
		// hope the author knows what they're doing.
	} else {
		llvmType := llvmValue.Type()
		switch llvmType.TypeKind() {
		case llvm.IntegerTypeKind, llvm.FloatTypeKind, llvm.DoubleTypeKind:
			// Nothing to do here. Integers and floats aren't pointers so don't
			// need any marking.
		case llvm.StructTypeKind:
			numElements := llvmType.StructElementTypesCount()
			for i := 0; i < numElements; i++ {
				element := llvm.ConstExtractValue(llvmValue, []uint32{uint32(i)})
				err := mv.markExternal(element, mark)
				if err != nil {
					return err
				}
			}
		case llvm.ArrayTypeKind:
			numElements := llvmType.ArrayLength()
			for i := 0; i < numElements; i++ {
				element := llvm.ConstExtractValue(llvmValue, []uint32{uint32(i)})
				err := mv.markExternal(element, mark)
				if err != nil {
					return err
				}
			}
		default:
			return errors.New("interp: unknown type kind in markExternalValue")
		}
	}
	return nil
}

// hasExternalLoadOrStore returns true if this object has an external load or
// store. If this has happened, it is not possible for the interpreter to load
// from the object or store to it without affecting the behavior of the program.
func (mv *memoryView) hasExternalLoadOrStore(v pointerValue) bool {
	obj := mv.get(v.index())
	return obj.marked >= 1
}

// hasExternalStore returns true if this object has an external store. If this
// is true, stores to this object are no longer allowed by the interpreter.
// It returns false if it only has an external load, in which case it is still
// possible for the interpreter to read from the object.
func (mv *memoryView) hasExternalStore(v pointerValue) bool {
	obj := mv.get(v.index())
	return obj.marked >= 2 && !obj.constant
}

// get returns an object that can only be read from, as it may return an object
// of a parent view.
func (mv *memoryView) get(index uint32) object {
	if obj, ok := mv.objects[index]; ok {
		return obj
	}
	if mv.parent != nil {
		return mv.parent.get(index)
	}
	return mv.r.objects[index]
}

// getWritable returns an object that can be written to.
func (mv *memoryView) getWritable(index uint32) object {
	if obj, ok := mv.objects[index]; ok {
		// Object is already in the current memory view, so can be modified.
		return obj
	}
	// Object is not currently in this view. Get it, and clone it for use.
	obj := mv.get(index).clone()
	mv.r.objects[index] = obj
	return obj
}

// Replace the object (indicated with index) with the given object. This put is
// only done at the current memory view, so that if this memory view is reverted
// the object is not changed.
func (mv *memoryView) put(index uint32, obj object) {
	if mv.objects == nil {
		mv.objects = make(map[uint32]object)
	}
	if checks && mv.get(index).buffer == nil {
		panic("writing to external object")
	}
	if checks && mv.get(index).buffer.len(mv.r) != obj.buffer.len(mv.r) {
		panic("put() with a differently-sized object")
	}
	if checks && obj.constant {
		panic("interp: store to a constant")
	}
	mv.objects[index] = obj
}

// Load the value behind the given pointer. Returns nil if the pointer points to
// an external global.
func (mv *memoryView) load(p pointerValue, size uint32) value {
	if checks && mv.hasExternalStore(p) {
		panic("interp: load from object with external store")
	}
	obj := mv.get(p.index())
	if obj.buffer == nil {
		// External global, return nil.
		return nil
	}
	if p.offset() == 0 && size == obj.size {
		return obj.buffer.clone()
	}
	if checks && p.offset()+size > obj.size {
		panic("interp: load out of bounds")
	}
	v := obj.buffer.asRawValue(mv.r)
	loadedValue := rawValue{
		buf: v.buf[p.offset() : p.offset()+size],
	}
	return loadedValue
}

// Store to the value behind the given pointer. This overwrites the value in the
// memory view, so that the changed value is discarded when the memory view is
// reverted. Returns true on success, false if the object to store to is
// external.
func (mv *memoryView) store(v value, p pointerValue) bool {
	if checks && mv.hasExternalLoadOrStore(p) {
		panic("interp: store to object with external load/store")
	}
	obj := mv.get(p.index())
	if obj.buffer == nil {
		// External global, return false (for a failure).
		return false
	}
	if checks && p.offset()+v.len(mv.r) > obj.size {
		panic("interp: store out of bounds")
	}
	if p.offset() == 0 && v.len(mv.r) == obj.buffer.len(mv.r) {
		obj.buffer = v
	} else {
		obj = obj.clone()
		buffer := obj.buffer.asRawValue(mv.r)
		obj.buffer = buffer
		v := v.asRawValue(mv.r)
		for i := uint32(0); i < v.len(mv.r); i++ {
			buffer.buf[p.offset()+i] = v.buf[i]
		}
	}
	mv.put(p.index(), obj)
	return true // success
}

// value is some sort of value, comparable to a LLVM constant. It can be
// implemented in various ways for efficiency, but the fallback value (that all
// implementations can be converted to except for localValue) is rawValue.
type value interface {
	// len returns the length in bytes.
	len(r *runner) uint32
	clone() value
	asPointer(*runner) (pointerValue, error)
	asRawValue(*runner) rawValue
	Uint() uint64
	Int() int64
	toLLVMValue(llvm.Type, *memoryView) (llvm.Value, error)
	String() string
}

// literalValue contains simple integer values that don't need to be stored in a
// buffer.
type literalValue struct {
	value interface{}
}

func (v literalValue) len(r *runner) uint32 {
	switch v.value.(type) {
	case uint64:
		return 8
	case uint32:
		return 4
	case uint16:
		return 2
	case uint8:
		return 1
	default:
		panic("unknown value type")
	}
}

func (v literalValue) String() string {
	return strconv.FormatInt(v.Int(), 10)
}

func (v literalValue) clone() value {
	return v
}

func (v literalValue) asPointer(r *runner) (pointerValue, error) {
	return pointerValue{}, errIntegerAsPointer
}

func (v literalValue) asRawValue(r *runner) rawValue {
	var buf []byte
	switch value := v.value.(type) {
	case uint64:
		buf = make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, value)
	case uint32:
		buf = make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(value))
	case uint16:
		buf = make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(value))
	case uint8:
		buf = []byte{uint8(value)}
	default:
		panic("unknown value type")
	}
	raw := newRawValue(uint32(len(buf)))
	for i, b := range buf {
		raw.buf[i] = uint64(b)
	}
	return raw
}

func (v literalValue) Uint() uint64 {
	switch value := v.value.(type) {
	case uint64:
		return value
	case uint32:
		return uint64(value)
	case uint16:
		return uint64(value)
	case uint8:
		return uint64(value)
	default:
		panic("inpterp: unknown literal type")
	}
}

func (v literalValue) Int() int64 {
	switch value := v.value.(type) {
	case uint64:
		return int64(value)
	case uint32:
		return int64(int32(value))
	case uint16:
		return int64(int16(value))
	case uint8:
		return int64(int8(value))
	default:
		panic("inpterp: unknown literal type")
	}
}

func (v literalValue) toLLVMValue(llvmType llvm.Type, mem *memoryView) (llvm.Value, error) {
	switch llvmType.TypeKind() {
	case llvm.IntegerTypeKind:
		switch value := v.value.(type) {
		case uint64:
			return llvm.ConstInt(llvmType, value, false), nil
		case uint32:
			return llvm.ConstInt(llvmType, uint64(value), false), nil
		case uint16:
			return llvm.ConstInt(llvmType, uint64(value), false), nil
		case uint8:
			return llvm.ConstInt(llvmType, uint64(value), false), nil
		default:
			return llvm.Value{}, errors.New("interp: unknown literal type")
		}
	case llvm.DoubleTypeKind:
		return llvm.ConstFloat(llvmType, math.Float64frombits(v.value.(uint64))), nil
	case llvm.FloatTypeKind:
		return llvm.ConstFloat(llvmType, float64(math.Float32frombits(v.value.(uint32)))), nil
	default:
		return v.asRawValue(mem.r).toLLVMValue(llvmType, mem)
	}
}

// pointerValue contains a single pointer, with an offset into the underlying
// object.
type pointerValue struct {
	pointer uint64 // low 32 bits are offset, high 32 bits are index
}

func newPointerValue(r *runner, index, offset int) pointerValue {
	return pointerValue{
		pointer: uint64(index)<<32 | uint64(offset),
	}
}

func (v pointerValue) index() uint32 {
	return uint32(v.pointer >> 32)
}

func (v pointerValue) offset() uint32 {
	return uint32(v.pointer)
}

// addOffset essentially does a GEP operation (pointer arithmetic): it adds the
// offset to the pointer. It also checks that the offset doesn't overflow the
// maximum offset size (which is 4GB).
func (v pointerValue) addOffset(offset uint32) pointerValue {
	result := pointerValue{v.pointer + uint64(offset)}
	if checks && v.index() != result.index() {
		panic("interp: offset out of range")
	}
	return result
}

func (v pointerValue) len(r *runner) uint32 {
	return r.pointerSize
}

func (v pointerValue) String() string {
	name := strconv.Itoa(int(v.index()))
	if v.offset() == 0 {
		return "<" + name + ">"
	}
	return "<" + name + "+" + strconv.Itoa(int(v.offset())) + ">"
}

func (v pointerValue) clone() value {
	return v
}

func (v pointerValue) asPointer(r *runner) (pointerValue, error) {
	return v, nil
}

func (v pointerValue) asRawValue(r *runner) rawValue {
	rv := newRawValue(r.pointerSize)
	for i := range rv.buf {
		rv.buf[i] = v.pointer
	}
	return rv
}

func (v pointerValue) Uint() uint64 {
	panic("cannot convert pointer to integer")
}

func (v pointerValue) Int() int64 {
	panic("cannot convert pointer to integer")
}

func (v pointerValue) equal(rhs pointerValue) bool {
	return v.pointer == rhs.pointer
}

func (v pointerValue) llvmValue(mem *memoryView) llvm.Value {
	return mem.get(v.index()).llvmGlobal
}

// toLLVMValue returns the LLVM value for this pointer, which may be a GEP or
// bitcast. The llvm.Type parameter is optional, if omitted the pointer type may
// be different than expected.
func (v pointerValue) toLLVMValue(llvmType llvm.Type, mem *memoryView) (llvm.Value, error) {
	// If a particular LLVM type is requested, cast to it.
	if !llvmType.IsNil() && llvmType.TypeKind() != llvm.PointerTypeKind {
		// The LLVM value has (or should have) the same bytes once compiled, but
		// does not have the right LLVM type. This can happen for example when
		// storing to a struct with a single pointer field: this pointer may
		// then become the value even though the pointer should be wrapped in a
		// struct.
		// This can be worked around by simply converting to a raw value,
		// rawValue knows how to create such structs.
		return v.asRawValue(mem.r).toLLVMValue(llvmType, mem)
	}

	// Obtain the llvmValue, creating it if it doesn't exist yet.
	llvmValue := v.llvmValue(mem)
	if llvmValue.IsNil() {
		// The global does not yet exist. Probably this is the result of a
		// runtime.alloc.
		// First allocate a new global for this object.
		obj := mem.get(v.index())
		if obj.llvmType.IsNil() && obj.llvmLayoutType.IsNil() {
			// Create an initializer without knowing the global type.
			// This is probably the result of a runtime.alloc call.
			initializer, err := obj.buffer.asRawValue(mem.r).rawLLVMValue(mem)
			if err != nil {
				return llvm.Value{}, err
			}
			globalType := initializer.Type()
			llvmValue = llvm.AddGlobal(mem.r.mod, globalType, obj.globalName)
			llvmValue.SetInitializer(initializer)
			llvmValue.SetAlignment(mem.r.maxAlign)
			obj.llvmGlobal = llvmValue
			mem.put(v.index(), obj)
		} else {
			// The global type is known, or at least its structure.
			var globalType llvm.Type
			if !obj.llvmType.IsNil() {
				// The exact type is known.
				globalType = obj.llvmType.ElementType()
			} else { // !obj.llvmLayoutType.IsNil()
				// The exact type isn't known, but the object layout is known.
				globalType = obj.llvmLayoutType
				// The layout may not span the full size of the global because
				// of repetition. One example would be make([]string, 5) which
				// would be 10 words in size but the layout would only be two
				// words (for the string type).
				typeSize := mem.r.targetData.TypeAllocSize(globalType)
				if typeSize != uint64(obj.size) {
					globalType = llvm.ArrayType(globalType, int(uint64(obj.size)/typeSize))
				}
			}
			if checks && mem.r.targetData.TypeAllocSize(globalType) != uint64(obj.size) {
				panic("size of the globalType isn't the same as the object size")
			}
			llvmValue = llvm.AddGlobal(mem.r.mod, globalType, obj.globalName)
			obj.llvmGlobal = llvmValue
			mem.put(v.index(), obj)

			// Set the initializer for the global. Do this after creation to avoid
			// infinite recursion between creating the global and creating the
			// contents of the global (if the global contains itself).
			initializer, err := obj.buffer.toLLVMValue(globalType, mem)
			if err != nil {
				return llvm.Value{}, err
			}
			if checks && initializer.Type() != globalType {
				return llvm.Value{}, errors.New("interp: allocated value does not match allocated type")
			}
			llvmValue.SetInitializer(initializer)
			if obj.llvmType.IsNil() {
				// The exact type isn't known (only the layout), so use the
				// alignment that would normally be expected from runtime.alloc.
				llvmValue.SetAlignment(mem.r.maxAlign)
			}
		}

		// It should be included in r.globals because otherwise markExternal
		// would consider it a new global (and would fail to mark this global as
		// having an externa load/store).
		mem.r.globals[llvmValue] = int(v.index())
		llvmValue.SetLinkage(llvm.InternalLinkage)
	}

	if v.offset() != 0 {
		// If there is an offset, make sure to use a GEP to index into the
		// pointer.
		// Cast to an i8* first (if needed) for easy indexing.
		if llvmValue.Type() != mem.r.i8ptrType {
			llvmValue = llvm.ConstBitCast(llvmValue, mem.r.i8ptrType)
		}
		llvmValue = llvm.ConstInBoundsGEP(llvmValue, []llvm.Value{
			llvm.ConstInt(llvmValue.Type().Context().Int32Type(), uint64(v.offset()), false),
		})
	}

	// If a particular LLVM pointer type is requested, cast to it.
	if !llvmType.IsNil() && llvmType != llvmValue.Type() {
		llvmValue = llvm.ConstBitCast(llvmValue, llvmType)
	}

	return llvmValue, nil
}

// rawValue is a raw memory buffer that can store either pointers or regular
// data. This is the fallback data for everything that isn't clearly a
// literalValue or pointerValue.
type rawValue struct {
	// An integer in buf contains either pointers or bytes.
	// If it is a byte, it is smaller than 256.
	// If it is a pointer, the index is contained in the upper 32 bits and the
	// offset is contained in the lower 32 bits.
	buf []uint64
}

func newRawValue(size uint32) rawValue {
	return rawValue{make([]uint64, size)}
}

func (v rawValue) len(r *runner) uint32 {
	return uint32(len(v.buf))
}

func (v rawValue) String() string {
	if len(v.buf) == 2 || len(v.buf) == 4 || len(v.buf) == 8 {
		// Format as a pointer if the entire buf is this pointer.
		if v.buf[0] > 255 {
			isPointer := true
			for _, p := range v.buf {
				if p != v.buf[0] {
					isPointer = false
					break
				}
			}
			if isPointer {
				return pointerValue{v.buf[0]}.String()
			}
		}
		// Format as number if none of the buf is a pointer.
		if !v.hasPointer() {
			return strconv.FormatInt(v.Int(), 10)
		}
	}
	return "<[…" + strconv.Itoa(len(v.buf)) + "]>"
}

func (v rawValue) clone() value {
	newValue := v
	newValue.buf = make([]uint64, len(v.buf))
	copy(newValue.buf, v.buf)
	return newValue
}

func (v rawValue) asPointer(r *runner) (pointerValue, error) {
	if v.buf[0] <= 255 {
		// Probably a null pointer or memory-mapped I/O.
		return pointerValue{}, errIntegerAsPointer
	}
	return pointerValue{v.buf[0]}, nil
}

func (v rawValue) asRawValue(r *runner) rawValue {
	return v
}

func (v rawValue) bytes() []byte {
	buf := make([]byte, len(v.buf))
	for i, p := range v.buf {
		if p > 255 {
			panic("cannot convert pointer value to byte")
		}
		buf[i] = byte(p)
	}
	return buf
}

func (v rawValue) Uint() uint64 {
	buf := v.bytes()

	switch len(v.buf) {
	case 1:
		return uint64(buf[0])
	case 2:
		return uint64(binary.LittleEndian.Uint16(buf))
	case 4:
		return uint64(binary.LittleEndian.Uint32(buf))
	case 8:
		return binary.LittleEndian.Uint64(buf)
	default:
		panic("unknown integer size")
	}
}

func (v rawValue) Int() int64 {
	switch len(v.buf) {
	case 1:
		return int64(int8(v.Uint()))
	case 2:
		return int64(int16(v.Uint()))
	case 4:
		return int64(int32(v.Uint()))
	case 8:
		return int64(int64(v.Uint()))
	default:
		panic("unknown integer size")
	}
}

// equal returns true if (and only if) the value matches rhs.
func (v rawValue) equal(rhs rawValue) bool {
	if len(v.buf) != len(rhs.buf) {
		panic("comparing values of different size")
	}
	for i, p := range v.buf {
		if rhs.buf[i] != p {
			return false
		}
	}
	return true
}

// rawLLVMValue returns a llvm.Value for this rawValue, making up a type as it
// goes. The resulting value does not have a specified type, but it will be the
// same size and have the same bytes if it was created with a provided LLVM type
// (through toLLVMValue).
func (v rawValue) rawLLVMValue(mem *memoryView) (llvm.Value, error) {
	var structFields []llvm.Value
	ctx := mem.r.mod.Context()
	int8Type := ctx.Int8Type()

	var bytesBuf []llvm.Value
	// addBytes can be called after adding to bytesBuf to flush remaining bytes
	// to a new array in structFields.
	addBytes := func() {
		if len(bytesBuf) == 0 {
			return
		}
		if len(bytesBuf) == 1 {
			structFields = append(structFields, bytesBuf[0])
		} else {
			structFields = append(structFields, llvm.ConstArray(int8Type, bytesBuf))
		}
		bytesBuf = nil
	}

	// Create structFields, converting the rawValue to a LLVM value.
	for i := uint32(0); i < uint32(len(v.buf)); {
		if v.buf[i] > 255 {
			addBytes()
			field, err := pointerValue{v.buf[i]}.toLLVMValue(llvm.Type{}, mem)
			if err != nil {
				return llvm.Value{}, err
			}
			elementType := field.Type().ElementType()
			if elementType.TypeKind() == llvm.StructTypeKind {
				// There are some special pointer types that should be used as a
				// ptrtoint, so that they can be used in certain optimizations.
				name := elementType.StructName()
				if name == "runtime.typecodeID" || name == "runtime.funcValueWithSignature" {
					uintptrType := ctx.IntType(int(mem.r.pointerSize) * 8)
					field = llvm.ConstPtrToInt(field, uintptrType)
				}
			}
			structFields = append(structFields, field)
			i += mem.r.pointerSize
			continue
		}
		val := llvm.ConstInt(int8Type, uint64(v.buf[i]), false)
		bytesBuf = append(bytesBuf, val)
		i++
	}
	addBytes()

	// Return the created data.
	if len(structFields) == 1 {
		return structFields[0], nil
	}
	return ctx.ConstStruct(structFields, false), nil
}

func (v rawValue) toLLVMValue(llvmType llvm.Type, mem *memoryView) (llvm.Value, error) {
	isZero := true
	for _, p := range v.buf {
		if p != 0 {
			isZero = false
			break
		}
	}
	if isZero {
		return llvm.ConstNull(llvmType), nil
	}
	switch llvmType.TypeKind() {
	case llvm.IntegerTypeKind:
		if v.buf[0] > 255 {
			ptr, err := v.asPointer(mem.r)
			if err != nil {
				panic(err)
			}
			if checks && mem.r.targetData.TypeAllocSize(llvmType) != mem.r.targetData.TypeAllocSize(mem.r.i8ptrType) {
				// Probably trying to serialize a pointer to a byte array,
				// perhaps as a result of rawLLVMValue() in a previous interp
				// run.
				return llvm.Value{}, errInvalidPtrToIntSize
			}
			v, err := ptr.toLLVMValue(llvm.Type{}, mem)
			if err != nil {
				return llvm.Value{}, err
			}
			return llvm.ConstPtrToInt(v, llvmType), nil
		}
		var n uint64
		switch llvmType.IntTypeWidth() {
		case 64:
			n = rawValue{v.buf[:8]}.Uint()
		case 32:
			n = rawValue{v.buf[:4]}.Uint()
		case 16:
			n = rawValue{v.buf[:2]}.Uint()
		case 8:
			n = uint64(v.buf[0])
		case 1:
			n = uint64(v.buf[0])
			if n != 0 && n != 1 {
				panic("bool must be 0 or 1")
			}
		default:
			panic("unknown integer size")
		}
		return llvm.ConstInt(llvmType, n, false), nil
	case llvm.StructTypeKind:
		fieldTypes := llvmType.StructElementTypes()
		fields := make([]llvm.Value, len(fieldTypes))
		for i, fieldType := range fieldTypes {
			offset := mem.r.targetData.ElementOffset(llvmType, i)
			field := rawValue{
				buf: v.buf[offset:],
			}
			var err error
			fields[i], err = field.toLLVMValue(fieldType, mem)
			if err != nil {
				return llvm.Value{}, err
			}
		}
		if llvmType.StructName() != "" {
			return llvm.ConstNamedStruct(llvmType, fields), nil
		}
		return llvmType.Context().ConstStruct(fields, false), nil
	case llvm.ArrayTypeKind:
		numElements := llvmType.ArrayLength()
		childType := llvmType.ElementType()
		childTypeSize := mem.r.targetData.TypeAllocSize(childType)
		fields := make([]llvm.Value, numElements)
		for i := range fields {
			offset := i * int(childTypeSize)
			field := rawValue{
				buf: v.buf[offset:],
			}
			var err error
			fields[i], err = field.toLLVMValue(childType, mem)
			if err != nil {
				return llvm.Value{}, err
			}
			if checks && fields[i].Type() != childType {
				panic("child type doesn't match")
			}
		}
		return llvm.ConstArray(childType, fields), nil
	case llvm.PointerTypeKind:
		if v.buf[0] > 255 {
			// This is a regular pointer.
			llvmValue, err := pointerValue{v.buf[0]}.toLLVMValue(llvm.Type{}, mem)
			if err != nil {
				return llvm.Value{}, err
			}
			if llvmValue.Type() != llvmType {
				if llvmValue.Type().PointerAddressSpace() != llvmType.PointerAddressSpace() {
					// Special case for AVR function pointers.
					// Because go-llvm doesn't have addrspacecast at the moment,
					// do it indirectly with a ptrtoint/inttoptr pair.
					llvmValue = llvm.ConstIntToPtr(llvm.ConstPtrToInt(llvmValue, mem.r.uintptrType), llvmType)
				} else {
					llvmValue = llvm.ConstBitCast(llvmValue, llvmType)
				}
			}
			return llvmValue, nil
		}
		// This is either a null pointer or a raw pointer for memory-mapped I/O
		// (such as 0xe000ed00).
		ptr := rawValue{v.buf[:mem.r.pointerSize]}.Uint()
		if ptr == 0 {
			// Null pointer.
			return llvm.ConstNull(llvmType), nil
		}
		var ptrValue llvm.Value // the underlying int
		switch mem.r.pointerSize {
		case 8:
			ptrValue = llvm.ConstInt(llvmType.Context().Int64Type(), ptr, false)
		case 4:
			ptrValue = llvm.ConstInt(llvmType.Context().Int32Type(), ptr, false)
		case 2:
			ptrValue = llvm.ConstInt(llvmType.Context().Int16Type(), ptr, false)
		default:
			return llvm.Value{}, errors.New("interp: unknown pointer size")
		}
		return llvm.ConstIntToPtr(ptrValue, llvmType), nil
	case llvm.DoubleTypeKind:
		b := rawValue{v.buf[:8]}.Uint()
		f := math.Float64frombits(b)
		return llvm.ConstFloat(llvmType, f), nil
	case llvm.FloatTypeKind:
		b := uint32(rawValue{v.buf[:4]}.Uint())
		f := math.Float32frombits(b)
		return llvm.ConstFloat(llvmType, float64(f)), nil
	default:
		return llvm.Value{}, errors.New("interp: todo: raw value to LLVM value: " + llvmType.String())
	}
}

func (v *rawValue) set(llvmValue llvm.Value, r *runner) {
	if llvmValue.IsNull() {
		// A zero value is common so check that first.
		return
	}
	if !llvmValue.IsAGlobalValue().IsNil() {
		ptrSize := r.pointerSize
		ptr, err := r.getValue(llvmValue).asPointer(r)
		if err != nil {
			panic(err)
		}
		for i := uint32(0); i < ptrSize; i++ {
			v.buf[i] = ptr.pointer
		}
	} else if !llvmValue.IsAConstantExpr().IsNil() {
		switch llvmValue.Opcode() {
		case llvm.IntToPtr, llvm.PtrToInt, llvm.BitCast:
			// All these instructions effectively just reinterprets the bits
			// (like a bitcast) while no bits change and keeping the same
			// length, so just read its contents.
			v.set(llvmValue.Operand(0), r)
		case llvm.GetElementPtr:
			ptr := llvmValue.Operand(0)
			index := llvmValue.Operand(1)
			numOperands := llvmValue.OperandsCount()
			elementType := ptr.Type().ElementType()
			totalOffset := r.targetData.TypeAllocSize(elementType) * index.ZExtValue()
			for i := 2; i < numOperands; i++ {
				indexValue := llvmValue.Operand(i)
				if checks && indexValue.IsAConstantInt().IsNil() {
					panic("expected const gep index to be a constant integer")
				}
				index := indexValue.ZExtValue()
				switch elementType.TypeKind() {
				case llvm.StructTypeKind:
					// Indexing into a struct field.
					offsetInBytes := r.targetData.ElementOffset(elementType, int(index))
					totalOffset += offsetInBytes
					elementType = elementType.StructElementTypes()[index]
				default:
					// Indexing into an array.
					elementType = elementType.ElementType()
					elementSize := r.targetData.TypeAllocSize(elementType)
					totalOffset += index * elementSize
				}
			}
			ptrSize := r.pointerSize
			ptrValue, err := r.getValue(ptr).asPointer(r)
			if err != nil {
				panic(err)
			}
			ptrValue.pointer += totalOffset
			for i := uint32(0); i < ptrSize; i++ {
				v.buf[i] = ptrValue.pointer
			}
		default:
			llvmValue.Dump()
			println()
			panic("unknown constant expr")
		}
	} else if llvmValue.IsUndef() {
		// Let undef be zero, by lack of an explicit 'undef' marker.
	} else {
		if checks && llvmValue.IsAConstant().IsNil() {
			panic("expected a constant")
		}
		llvmType := llvmValue.Type()
		switch llvmType.TypeKind() {
		case llvm.IntegerTypeKind:
			n := llvmValue.ZExtValue()
			switch llvmValue.Type().IntTypeWidth() {
			case 64:
				var buf [8]byte
				binary.LittleEndian.PutUint64(buf[:], n)
				for i, b := range buf {
					v.buf[i] = uint64(b)
				}
			case 32:
				var buf [4]byte
				binary.LittleEndian.PutUint32(buf[:], uint32(n))
				for i, b := range buf {
					v.buf[i] = uint64(b)
				}
			case 16:
				var buf [2]byte
				binary.LittleEndian.PutUint16(buf[:], uint16(n))
				for i, b := range buf {
					v.buf[i] = uint64(b)
				}
			case 8, 1:
				v.buf[0] = n
			default:
				panic("unknown integer size")
			}
		case llvm.StructTypeKind:
			numElements := llvmType.StructElementTypesCount()
			for i := 0; i < numElements; i++ {
				offset := r.targetData.ElementOffset(llvmType, i)
				field := rawValue{
					buf: v.buf[offset:],
				}
				field.set(llvm.ConstExtractValue(llvmValue, []uint32{uint32(i)}), r)
			}
		case llvm.ArrayTypeKind:
			numElements := llvmType.ArrayLength()
			childType := llvmType.ElementType()
			childTypeSize := r.targetData.TypeAllocSize(childType)
			for i := 0; i < numElements; i++ {
				offset := i * int(childTypeSize)
				field := rawValue{
					buf: v.buf[offset:],
				}
				field.set(llvm.ConstExtractValue(llvmValue, []uint32{uint32(i)}), r)
			}
		case llvm.DoubleTypeKind:
			f, _ := llvmValue.DoubleValue()
			var buf [8]byte
			binary.LittleEndian.PutUint64(buf[:], math.Float64bits(f))
			for i, b := range buf {
				v.buf[i] = uint64(b)
			}
		case llvm.FloatTypeKind:
			f, _ := llvmValue.DoubleValue()
			var buf [4]byte
			binary.LittleEndian.PutUint32(buf[:], math.Float32bits(float32(f)))
			for i, b := range buf {
				v.buf[i] = uint64(b)
			}
		default:
			llvmValue.Dump()
			println()
			panic("unknown constant")
		}
	}
}

// hasPointer returns true if this raw value contains a pointer somewhere in the
// buffer.
func (v rawValue) hasPointer() bool {
	for _, p := range v.buf {
		if p > 255 {
			return true
		}
	}
	return false
}

// localValue is a special implementation of the value interface. It is a
// placeholder for other values in instruction operands, and is replaced with
// one of the others before executing.
type localValue struct {
	value llvm.Value
}

func (v localValue) len(r *runner) uint32 {
	panic("interp: localValue.len")
}

func (v localValue) String() string {
	return "<!>"
}

func (v localValue) clone() value {
	panic("interp: localValue.clone()")
}

func (v localValue) asPointer(r *runner) (pointerValue, error) {
	return pointerValue{}, errors.New("interp: localValue.asPointer called")
}

func (v localValue) asRawValue(r *runner) rawValue {
	panic("interp: localValue.asRawValue")
}

func (v localValue) Uint() uint64 {
	panic("interp: localValue.Uint")
}

func (v localValue) Int() int64 {
	panic("interp: localValue.Int")
}

func (v localValue) toLLVMValue(llvmType llvm.Type, mem *memoryView) (llvm.Value, error) {
	return v.value, nil
}

func (r *runner) getValue(llvmValue llvm.Value) value {
	if checks && llvmValue.IsNil() {
		panic("nil llvmValue")
	}
	if !llvmValue.IsAGlobalValue().IsNil() {
		index, ok := r.globals[llvmValue]
		if !ok {
			obj := object{
				llvmGlobal: llvmValue,
			}
			index = len(r.objects)
			r.globals[llvmValue] = index
			r.objects = append(r.objects, obj)
			if !llvmValue.IsAGlobalVariable().IsNil() {
				obj.size = uint32(r.targetData.TypeAllocSize(llvmValue.Type().ElementType()))
				if initializer := llvmValue.Initializer(); !initializer.IsNil() {
					obj.buffer = r.getValue(initializer)
					obj.constant = llvmValue.IsGlobalConstant()
				}
			} else if !llvmValue.IsAFunction().IsNil() {
				// OK
			} else {
				panic("interp: unknown global value")
			}
			// Update the object after it has been created. This avoids an
			// infinite recursion when using getValue on a global that contains
			// a reference to itself.
			r.objects[index] = obj
		}
		return newPointerValue(r, index, 0)
	} else if !llvmValue.IsAConstant().IsNil() {
		if !llvmValue.IsAConstantInt().IsNil() {
			n := llvmValue.ZExtValue()
			switch llvmValue.Type().IntTypeWidth() {
			case 64:
				return literalValue{n}
			case 32:
				return literalValue{uint32(n)}
			case 16:
				return literalValue{uint16(n)}
			case 8, 1:
				return literalValue{uint8(n)}
			default:
				panic("unknown integer size")
			}
		}
		size := r.targetData.TypeAllocSize(llvmValue.Type())
		v := newRawValue(uint32(size))
		v.set(llvmValue, r)
		return v
	} else if !llvmValue.IsAInstruction().IsNil() || !llvmValue.IsAArgument().IsNil() {
		return localValue{llvmValue}
	} else if !llvmValue.IsAInlineAsm().IsNil() {
		return localValue{llvmValue}
	} else {
		llvmValue.Dump()
		println()
		panic("unknown value")
	}
}

// readObjectLayout reads the object layout as it is stored by the compiler. It
// returns the size in the number of words and the bitmap.
func (r *runner) readObjectLayout(layoutValue value) (uint64, *big.Int) {
	pointerSize := layoutValue.len(r)
	if checks && uint64(pointerSize) != r.targetData.TypeAllocSize(r.i8ptrType) {
		panic("inconsistent pointer size")
	}

	// The object layout can be stored in a global variable, directly as an
	// integer value, or can be nil.
	ptr, err := layoutValue.asPointer(r)
	if err == errIntegerAsPointer {
		// It's an integer, which means it's a small object or unknown.
		layout := layoutValue.Uint()
		if layout == 0 {
			// Nil pointer, which means the layout is unknown.
			return 0, nil
		}
		if layout%2 != 1 {
			// Sanity check: the least significant bit must be set. This is how
			// the runtime can separate pointers from integers.
			panic("unexpected layout")
		}

		// Determine format of bitfields in the integer.
		pointerBits := uint64(pointerSize * 8)
		var sizeFieldBits uint64
		switch pointerBits {
		case 16:
			sizeFieldBits = 4
		case 32:
			sizeFieldBits = 5
		case 64:
			sizeFieldBits = 6
		default:
			panic("unknown pointer size")
		}

		// Extract fields.
		objectSizeWords := (layout >> 1) & (1<<sizeFieldBits - 1)
		bitmap := new(big.Int).SetUint64(layout >> (1 + sizeFieldBits))
		return objectSizeWords, bitmap
	}

	// Read the object size in words and the bitmap from the global.
	buf := r.objects[ptr.index()].buffer.(rawValue)
	objectSizeWords := rawValue{buf: buf.buf[:r.pointerSize]}.Uint()
	rawByteValues := buf.buf[r.pointerSize:]
	rawBytes := make([]byte, len(rawByteValues))
	for i, v := range rawByteValues {
		if uint64(byte(v)) != v {
			panic("found pointer in data array?") // sanity check
		}
		rawBytes[i] = byte(v)
	}
	bitmap := new(big.Int).SetBytes(rawBytes)
	return objectSizeWords, bitmap
}

// getLLVMTypeFromLayout returns the 'layout type', which is an approximation of
// the real type. Pointers are in the correct location but the actual object may
// have some additional repetition, for example in the buffer of a slice.
func (r *runner) getLLVMTypeFromLayout(layoutValue value) llvm.Type {
	objectSizeWords, bitmap := r.readObjectLayout(layoutValue)
	if bitmap == nil {
		// No information available.
		return llvm.Type{}
	}

	if bitmap.BitLen() == 0 {
		// There are no pointers in this object, so treat this as a raw byte
		// buffer. This is important because objects without pointers may have
		// lower alignment.
		return r.mod.Context().Int8Type()
	}

	// Create the LLVM type.
	pointerSize := layoutValue.len(r)
	pointerAlignment := r.targetData.PrefTypeAlignment(r.i8ptrType)
	var fields []llvm.Type
	for i := 0; i < int(objectSizeWords); {
		if bitmap.Bit(i) != 0 {
			// Pointer field.
			fields = append(fields, r.i8ptrType)
			i += int(pointerSize / uint32(pointerAlignment))
		} else {
			// Byte/word field.
			fields = append(fields, r.mod.Context().IntType(pointerAlignment*8))
			i += 1
		}
	}
	var llvmLayoutType llvm.Type
	if len(fields) == 1 {
		llvmLayoutType = fields[0]
	} else {
		llvmLayoutType = r.mod.Context().StructType(fields, false)
	}

	objectSizeBytes := objectSizeWords * uint64(pointerAlignment)
	if checks && r.targetData.TypeAllocSize(llvmLayoutType) != objectSizeBytes {
		panic("unexpected size") // sanity check
	}
	return llvmLayoutType
}
