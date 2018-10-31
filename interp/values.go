package interp

// This file provides a litte bit of abstraction around LLVM values.

import (
	"strconv"

	"github.com/aykevl/go-llvm"
)

// A Value is a LLVM value with some extra methods attached for easier
// interpretation.
type Value interface {
	Value() llvm.Value            // returns a LLVM value
	Type() llvm.Type              // equal to Value().Type()
	IsConstant() bool             // returns true if this value is a constant value
	Load() llvm.Value             // dereference a pointer
	Store(llvm.Value)             // store to a pointer
	GetElementPtr([]uint32) Value // returns an interior pointer
	String() string               // string representation, for debugging
}

// A type that simply wraps a LLVM constant value.
type LocalValue struct {
	Eval       *Eval
	Underlying llvm.Value
}

// Value implements Value by returning the constant value itself.
func (v *LocalValue) Value() llvm.Value {
	return v.Underlying
}

func (v *LocalValue) Type() llvm.Type {
	return v.Underlying.Type()
}

func (v *LocalValue) IsConstant() bool {
	return v.Underlying.IsConstant()
}

// Load loads a constant value if this is a constant GEP, otherwise it panics.
func (v *LocalValue) Load() llvm.Value {
	switch v.Underlying.Opcode() {
	case llvm.GetElementPtr:
		indices := v.getConstGEPIndices()
		if indices[0] != 0 {
			panic("invalid GEP")
		}
		global := v.Eval.getValue(v.Underlying.Operand(0))
		agg := global.Load()
		return llvm.ConstExtractValue(agg, indices[1:])
	default:
		panic("interp: load from a constant")
	}
}

// Store stores to the underlying value if the value type is a constant GEP,
// otherwise it panics.
func (v *LocalValue) Store(value llvm.Value) {
	switch v.Underlying.Opcode() {
	case llvm.GetElementPtr:
		indices := v.getConstGEPIndices()
		if indices[0] != 0 {
			panic("invalid GEP")
		}
		global := &GlobalValue{v.Eval, v.Underlying.Operand(0)}
		agg := global.Load()
		agg = llvm.ConstInsertValue(agg, value, indices[1:])
		global.Store(agg)
		return
	default:
		panic("interp: store on a constant")
	}
}

// GetElementPtr returns a constant GEP when the underlying value is also a
// constant GEP. It panics when the underlying value is not a constant GEP:
// getting the pointer to a constant is not possible.
func (v *LocalValue) GetElementPtr(indices []uint32) Value {
	switch v.Underlying.Opcode() {
	case llvm.GetElementPtr, llvm.IntToPtr:
		int32Type := v.Underlying.Type().Context().Int32Type()
		llvmIndices := getLLVMIndices(int32Type, indices)
		return &LocalValue{v.Eval, llvm.ConstGEP(v.Underlying, llvmIndices)}
	default:
		panic("interp: GEP on a constant")
	}
}

func (v *LocalValue) String() string {
	isConstant := "false"
	if v.IsConstant() {
		isConstant = "true"
	}
	return "&LocalValue{Type: " + v.Type().String() + ", IsConstant: " + isConstant + "}"
}

// getConstGEPIndices returns indices of this constant GEP, if this is a GEP
// instruction. If it is not, the behavior is undefined.
func (v *LocalValue) getConstGEPIndices() []uint32 {
	indices := make([]uint32, v.Underlying.OperandsCount()-1)
	for i := range indices {
		operand := v.Underlying.Operand(i + 1)
		indices[i] = uint32(operand.ZExtValue())
	}
	return indices
}

// GlobalValue wraps a LLVM global variable.
type GlobalValue struct {
	Eval       *Eval
	Underlying llvm.Value
}

// Value returns the initializer for this global variable.
func (v *GlobalValue) Value() llvm.Value {
	return v.Underlying
}

// Type returns the type of this global variable, which is a pointer type. Use
// Type().ElementType() to get the actual global variable type.
func (v *GlobalValue) Type() llvm.Type {
	return v.Underlying.Type()
}

// IsConstant returns true if this global is not dirty, false otherwise.
func (v *GlobalValue) IsConstant() bool {
	if _, ok := v.Eval.dirtyGlobals[v.Underlying]; ok {
		return true
	}
	return false
}

// Load returns the initializer of the global variable.
func (v *GlobalValue) Load() llvm.Value {
	return v.Underlying.Initializer()
}

// Store sets the initializer of the global variable.
func (v *GlobalValue) Store(value llvm.Value) {
	if !value.IsConstant() {
		v.MarkDirty()
		v.Eval.builder.CreateStore(value, v.Underlying)
	} else {
		v.Underlying.SetInitializer(value)
	}
}

// GetElementPtr returns a constant GEP on this global, which can be used in
// load and store instructions.
func (v *GlobalValue) GetElementPtr(indices []uint32) Value {
	int32Type := v.Underlying.Type().Context().Int32Type()
	gep := llvm.ConstGEP(v.Underlying, getLLVMIndices(int32Type, indices))
	return &LocalValue{v.Eval, gep}
}

func (v *GlobalValue) String() string {
	return "&GlobalValue{" + v.Underlying.Name() + "}"
}

// MarkDirty marks this global as dirty, meaning that every load from and store
// to this global (from now on) must be performed at runtime.
func (v *GlobalValue) MarkDirty() {
	if !v.IsConstant() {
		return // already dirty
	}
	v.Eval.dirtyGlobals[v.Underlying] = struct{}{}
}

// An alloca represents a local alloca, which is a stack allocated variable.
// It is emulated by storing the constant of the alloca.
type AllocaValue struct {
	Eval       *Eval
	Underlying llvm.Value // the constant value itself if not dirty, otherwise the alloca instruction
	Dirty      bool       // this value must be evaluated at runtime
}

// Value turns this alloca into a runtime alloca instead of a compile-time
// constant (if not already converted), and returns the alloca itself.
func (v *AllocaValue) Value() llvm.Value {
	if !v.Dirty {
		// Mark this alloca a dirty, meaning it is run at runtime instead of
		// compile time.
		alloca := v.Eval.builder.CreateAlloca(v.Underlying.Type(), "")
		v.Eval.builder.CreateStore(v.Underlying, alloca)
		v.Dirty = true
		v.Underlying = alloca
	}
	return v.Underlying
}

// Type returns the type of this alloca, which is always a pointer.
func (v *AllocaValue) Type() llvm.Type {
	if v.Dirty {
		return v.Underlying.Type()
	} else {
		return llvm.PointerType(v.Underlying.Type(), 0)
	}
}

func (v *AllocaValue) IsConstant() bool {
	return !v.Dirty
}

// Load returns the value this alloca contains, which may be evaluated at
// runtime.
func (v *AllocaValue) Load() llvm.Value {
	if v.Dirty {
		ret := v.Eval.builder.CreateLoad(v.Underlying, "")
		if ret.IsNil() {
			panic("alloca is nil")
		}
		return ret
	} else {
		if v.Underlying.IsNil() {
			panic("alloca is nil")
		}
		return v.Underlying
	}
}

// Store updates the value of this alloca.
func (v *AllocaValue) Store(value llvm.Value) {
	if v.Underlying.Type() != value.Type() {
		panic("interp: trying to store to an alloca with a different type")
	}
	if v.Dirty || !value.IsConstant() {
		v.Eval.builder.CreateStore(value, v.Value())
	} else {
		v.Underlying = value
	}
}

// GetElementPtr returns a value (a *GetElementPtrValue) that keeps a reference
// to this alloca, so that Load() and Store() continue to work.
func (v *AllocaValue) GetElementPtr(indices []uint32) Value {
	return &GetElementPtrValue{v, indices}
}

func (v *AllocaValue) String() string {
	return "&AllocaValue{Type: " + v.Type().String() + "}"
}

// GetElementPtrValue wraps an alloca, keeping track of what the GEP points to
// so it can be used as a pointer value (with Load() and Store()).
type GetElementPtrValue struct {
	Alloca  *AllocaValue
	Indices []uint32
}

// Type returns the type of this GEP, which is always of type pointer.
func (v *GetElementPtrValue) Type() llvm.Type {
	if v.Alloca.Dirty {
		return v.Value().Type()
	} else {
		return llvm.PointerType(v.Load().Type(), 0)
	}
}

func (v *GetElementPtrValue) IsConstant() bool {
	return v.Alloca.IsConstant()
}

// Value creates the LLVM GEP instruction of this GetElementPtrValue wrapper and
// returns it.
func (v *GetElementPtrValue) Value() llvm.Value {
	if v.Alloca.Dirty {
		alloca := v.Alloca.Value()
		int32Type := v.Alloca.Type().Context().Int32Type()
		llvmIndices := getLLVMIndices(int32Type, v.Indices)
		return v.Alloca.Eval.builder.CreateGEP(alloca, llvmIndices, "")
	} else {
		panic("interp: todo: pointer to alloca gep")
	}
}

// Load deferences the pointer this GEP points to. For a constant GEP, it
// extracts the value from the underlying alloca.
func (v *GetElementPtrValue) Load() llvm.Value {
	if v.Alloca.Dirty {
		gep := v.Value()
		return v.Alloca.Eval.builder.CreateLoad(gep, "")
	} else {
		underlying := v.Alloca.Load()
		indices := v.Indices
		if indices[0] != 0 {
			panic("invalid GEP")
		}
		return llvm.ConstExtractValue(underlying, indices[1:])
	}
}

// Store stores to the pointer this GEP points to. For a constant GEP, it
// updates the underlying allloca.
func (v *GetElementPtrValue) Store(value llvm.Value) {
	if v.Alloca.Dirty || !value.IsConstant() {
		alloca := v.Alloca.Value()
		int32Type := v.Alloca.Type().Context().Int32Type()
		llvmIndices := getLLVMIndices(int32Type, v.Indices)
		gep := v.Alloca.Eval.builder.CreateGEP(alloca, llvmIndices, "")
		v.Alloca.Eval.builder.CreateStore(value, gep)
	} else {
		underlying := v.Alloca.Load()
		indices := v.Indices
		if indices[0] != 0 {
			panic("invalid GEP")
		}
		underlying = llvm.ConstInsertValue(underlying, value, indices[1:])
		v.Alloca.Store(underlying)
	}
}

func (v *GetElementPtrValue) GetElementPtr(indices []uint32) Value {
	if v.Alloca.Dirty {
		panic("interp: todo: gep on a dirty gep")
	} else {
		combined := append([]uint32{}, v.Indices...)
		combined[len(combined)-1] += indices[0]
		combined = append(combined, indices[1:]...)
		return &GetElementPtrValue{v.Alloca, combined}
	}
}

func (v *GetElementPtrValue) String() string {
	indices := ""
	for _, n := range v.Indices {
		if indices != "" {
			indices += ", "
		}
		indices += strconv.Itoa(int(n))
	}
	return "&GetElementPtrValue{Alloca: " + v.Alloca.String() + ", Indices: [" + indices + "]}"
}

// PointerCastValue represents a bitcast operation on a pointer.
type PointerCastValue struct {
	Eval       *Eval
	Underlying Value
	CastType   llvm.Type
}

// Value returns a constant bitcast value.
func (v *PointerCastValue) Value() llvm.Value {
	from := v.Underlying.Value()
	return llvm.ConstBitCast(from, v.CastType)
}

// Type returns the type this pointer has been cast to.
func (v *PointerCastValue) Type() llvm.Type {
	return v.CastType
}

func (v *PointerCastValue) IsConstant() bool {
	return v.Underlying.IsConstant()
}

// Load tries to load and bitcast the given value. If this value cannot be
// bitcasted, Load panics.
func (v *PointerCastValue) Load() llvm.Value {
	if v.Underlying.IsConstant() {
		typeFrom := v.Underlying.Type().ElementType()
		typeTo := v.CastType.ElementType()
		if isScalar(typeFrom) && isScalar(typeTo) && v.Eval.TargetData.TypeAllocSize(typeFrom) == v.Eval.TargetData.TypeAllocSize(typeTo) {
			return llvm.ConstBitCast(v.Underlying.Load(), v.CastType.ElementType())
		}
	}

	panic("interp: load from a pointer bitcast: " + v.String())
}

// Store panics: it is not (yet) possible to store directly to a bitcast.
func (v *PointerCastValue) Store(value llvm.Value) {
	panic("interp: store on a pointer bitcast")
}

// GetElementPtr panics: it is not (yet) possible to do a GEP operation on a
// bitcast.
func (v *PointerCastValue) GetElementPtr(indices []uint32) Value {
	panic("interp: GEP on a pointer bitcast")
}

func (v *PointerCastValue) String() string {
	return "&PointerCastValue{Value: " + v.Underlying.String() + ", CastType: " + v.CastType.String() + "}"
}

// MapValue implements a Go map which is created at compile time and stored as a
// global variable.
type MapValue struct {
	Eval       *Eval
	PkgName    string
	Underlying llvm.Value
	Keys       []Value
	Values     []Value
	KeySize    int
	ValueSize  int
	KeyType    llvm.Type
	ValueType  llvm.Type
}

func (v *MapValue) newBucket() llvm.Value {
	ctx := v.Eval.Mod.Context()
	i8ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	bucketType := ctx.StructType([]llvm.Type{
		llvm.ArrayType(ctx.Int8Type(), 8), // tophash
		i8ptrType,                         // next bucket
		llvm.ArrayType(v.KeyType, 8),      // key type
		llvm.ArrayType(v.ValueType, 8),    // value type
	}, false)
	bucketValue := getZeroValue(bucketType)
	bucket := llvm.AddGlobal(v.Eval.Mod, bucketType, v.PkgName+"$mapbucket")
	bucket.SetInitializer(bucketValue)
	bucket.SetLinkage(llvm.InternalLinkage)
	bucket.SetUnnamedAddr(true)
	return bucket
}

// Value returns a global variable which is a pointer to the actual hashmap.
func (v *MapValue) Value() llvm.Value {
	if !v.Underlying.IsNil() {
		return v.Underlying
	}

	ctx := v.Eval.Mod.Context()
	i8ptrType := llvm.PointerType(ctx.Int8Type(), 0)

	var firstBucketGlobal llvm.Value
	if len(v.Keys) == 0 {
		// there are no buckets
		firstBucketGlobal = llvm.ConstPointerNull(i8ptrType)
	} else {
		// create initial bucket
		firstBucketGlobal = v.newBucket()
	}

	// Insert each key/value pair in the hashmap.
	bucketGlobal := firstBucketGlobal
	for i, key := range v.Keys {
		var keyBuf []byte
		llvmKey := key.Value()
		llvmValue := v.Values[i].Value()
		if key.Type().TypeKind() == llvm.StructTypeKind && key.Type().StructName() == "runtime._string" {
			keyPtr := llvm.ConstExtractValue(llvmKey, []uint32{0})
			keyLen := llvm.ConstExtractValue(llvmKey, []uint32{1})
			keyPtrVal := v.Eval.getValue(keyPtr)
			keyBuf = getStringBytes(keyPtrVal, keyLen)
		} else if key.Type().TypeKind() == llvm.IntegerTypeKind {
			keyBuf = make([]byte, v.Eval.TargetData.TypeAllocSize(key.Type()))
			n := key.Value().ZExtValue()
			for i := range keyBuf {
				keyBuf[i] = byte(n)
				n >>= 8
			}
		} else {
			panic("interp: map key type not implemented: " + key.Type().String())
		}
		hash := v.hash(keyBuf)

		if i%8 == 0 && i != 0 {
			// Bucket is full, create a new one.
			newBucketGlobal := v.newBucket()
			zero := llvm.ConstInt(ctx.Int32Type(), 0, false)
			newBucketPtr := llvm.ConstInBoundsGEP(newBucketGlobal, []llvm.Value{zero})
			newBucketPtrCast := llvm.ConstBitCast(newBucketPtr, i8ptrType)
			// insert pointer into old bucket
			bucket := bucketGlobal.Initializer()
			bucket = llvm.ConstInsertValue(bucket, newBucketPtrCast, []uint32{1})
			bucketGlobal.SetInitializer(bucket)
			// switch to next bucket
			bucketGlobal = newBucketGlobal
		}

		tophashValue := llvm.ConstInt(ctx.Int8Type(), uint64(v.topHash(hash)), false)
		bucket := bucketGlobal.Initializer()
		bucket = llvm.ConstInsertValue(bucket, tophashValue, []uint32{0, uint32(i % 8)})
		bucket = llvm.ConstInsertValue(bucket, llvmKey, []uint32{2, uint32(i % 8)})
		bucket = llvm.ConstInsertValue(bucket, llvmValue, []uint32{3, uint32(i % 8)})
		bucketGlobal.SetInitializer(bucket)
	}

	// Create the hashmap itself.
	zero := llvm.ConstInt(ctx.Int32Type(), 0, false)
	bucketPtr := llvm.ConstInBoundsGEP(firstBucketGlobal, []llvm.Value{zero})
	hashmapType := v.Type()
	hashmap := llvm.ConstNamedStruct(hashmapType, []llvm.Value{
		llvm.ConstPointerNull(llvm.PointerType(hashmapType, 0)),                        // next
		llvm.ConstBitCast(bucketPtr, i8ptrType),                                        // buckets
		llvm.ConstInt(hashmapType.StructElementTypes()[2], uint64(len(v.Keys)), false), // count
		llvm.ConstInt(ctx.Int8Type(), uint64(v.KeySize), false),                        // keySize
		llvm.ConstInt(ctx.Int8Type(), uint64(v.ValueSize), false),                      // valueSize
		llvm.ConstInt(ctx.Int8Type(), 0, false),                                        // bucketBits
	})

	// Create a pointer to this hashmap.
	hashmapPtr := llvm.AddGlobal(v.Eval.Mod, hashmap.Type(), v.PkgName+"$map")
	hashmapPtr.SetInitializer(hashmap)
	hashmapPtr.SetLinkage(llvm.InternalLinkage)
	hashmapPtr.SetUnnamedAddr(true)
	v.Underlying = llvm.ConstInBoundsGEP(hashmapPtr, []llvm.Value{zero})
	return v.Underlying
}

// Type returns type runtime.hashmap, which is the actual hashmap type.
func (v *MapValue) Type() llvm.Type {
	return v.Eval.Mod.GetTypeByName("runtime.hashmap")
}

func (v *MapValue) IsConstant() bool {
	return true // TODO: dirty maps
}

// Load panics: maps are of reference type so cannot be dereferenced.
func (v *MapValue) Load() llvm.Value {
	panic("interp: load from a map")
}

// Store panics: maps are of reference type so cannot be stored to.
func (v *MapValue) Store(value llvm.Value) {
	panic("interp: store on a map")
}

// GetElementPtr panics: maps are of reference type so their (interior)
// addresses cannot be calculated.
func (v *MapValue) GetElementPtr(indices []uint32) Value {
	panic("interp: GEP on a map")
}

// PutString does a map assign operation, assuming that the map is of type
// map[string]T.
func (v *MapValue) PutString(keyBuf, keyLen, valPtr Value) {
	if !v.Underlying.IsNil() {
		panic("map already created")
	}

	var value llvm.Value
	switch valPtr := valPtr.(type) {
	case *PointerCastValue:
		value = valPtr.Underlying.Load()
		if v.ValueType.IsNil() {
			v.ValueType = value.Type()
			if int(v.Eval.TargetData.TypeAllocSize(v.ValueType)) != v.ValueSize {
				panic("interp: map store value type has the wrong size")
			}
		} else {
			if value.Type() != v.ValueType {
				panic("interp: map store value type is inconsistent")
			}
		}
	default:
		panic("interp: todo: handle map value pointer")
	}

	keyType := v.Eval.Mod.GetTypeByName("runtime._string")
	v.KeyType = keyType
	key := getZeroValue(keyType)
	key = llvm.ConstInsertValue(key, keyBuf.Value(), []uint32{0})
	key = llvm.ConstInsertValue(key, keyLen.Value(), []uint32{1})

	// TODO: avoid duplicate keys
	v.Keys = append(v.Keys, &LocalValue{v.Eval, key})
	v.Values = append(v.Values, &LocalValue{v.Eval, value})
}

// Get FNV-1a hash of this string.
//
// https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function#FNV-1a_hash
func (v *MapValue) hash(data []byte) uint32 {
	var result uint32 = 2166136261 // FNV offset basis
	for _, c := range data {
		result ^= uint32(c)
		result *= 16777619 // FNV prime
	}
	return result
}

// Get the topmost 8 bits of the hash, without using a special value (like 0).
func (v *MapValue) topHash(hash uint32) uint8 {
	tophash := uint8(hash >> 24)
	if tophash < 1 {
		// 0 means empty slot, so make it bigger.
		tophash += 1
	}
	return tophash
}

func (v *MapValue) String() string {
	return "&MapValue{KeySize: " + strconv.Itoa(v.KeySize) + ", ValueSize: " + strconv.Itoa(v.ValueSize) + "}"
}
