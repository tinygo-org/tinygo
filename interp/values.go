package interp

// This file provides a litte bit of abstraction around LLVM values.

import (
	"strconv"

	"tinygo.org/x/go-llvm"
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
	if _, ok := v.Eval.dirtyGlobals[v.Underlying]; ok {
		return false
	}
	return v.Underlying.IsConstant()
}

// Load loads a constant value if this is a constant pointer.
func (v *LocalValue) Load() llvm.Value {
	if !v.Underlying.IsAGlobalVariable().IsNil() {
		return v.Underlying.Initializer()
	}
	switch v.Underlying.Opcode() {
	case llvm.GetElementPtr:
		indices := v.getConstGEPIndices()
		if indices[0] != 0 {
			panic("invalid GEP")
		}
		global := v.Eval.getValue(v.Underlying.Operand(0))
		agg := global.Load()
		return llvm.ConstExtractValue(agg, indices[1:])
	case llvm.BitCast:
		panic("interp: load from a bitcast")
	default:
		panic("interp: load from a constant")
	}
}

// Store stores to the underlying value if the value type is a pointer type,
// otherwise it panics.
func (v *LocalValue) Store(value llvm.Value) {
	if !v.Underlying.IsAGlobalVariable().IsNil() {
		if !value.IsConstant() {
			v.MarkDirty()
			v.Eval.builder.CreateStore(value, v.Underlying)
		} else {
			v.Underlying.SetInitializer(value)
		}
		return
	}
	switch v.Underlying.Opcode() {
	case llvm.GetElementPtr:
		indices := v.getConstGEPIndices()
		if indices[0] != 0 {
			panic("invalid GEP")
		}
		global := &LocalValue{v.Eval, v.Underlying.Operand(0)}
		agg := global.Load()
		agg = llvm.ConstInsertValue(agg, value, indices[1:])
		global.Store(agg)
		return
	default:
		panic("interp: store on a constant")
	}
}

// GetElementPtr returns a GEP when the underlying value is of pointer type.
func (v *LocalValue) GetElementPtr(indices []uint32) Value {
	if !v.Underlying.IsAGlobalVariable().IsNil() {
		int32Type := v.Underlying.Type().Context().Int32Type()
		gep := llvm.ConstGEP(v.Underlying, getLLVMIndices(int32Type, indices))
		return &LocalValue{v.Eval, gep}
	}
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

// MarkDirty marks this global as dirty, meaning that every load from and store
// to this global (from now on) must be performed at runtime.
func (v *LocalValue) MarkDirty() {
	if v.Underlying.IsAGlobalVariable().IsNil() {
		panic("trying to mark a non-global as dirty")
	}
	if !v.IsConstant() {
		return // already dirty
	}
	v.Eval.dirtyGlobals[v.Underlying] = struct{}{}
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
		} else if key.Type().TypeKind() == llvm.ArrayTypeKind &&
			key.Type().ElementType().TypeKind() == llvm.IntegerTypeKind &&
			key.Type().ElementType().IntTypeWidth() == 8 {
			keyBuf = make([]byte, v.Eval.TargetData.TypeAllocSize(key.Type()))
			for i := range keyBuf {
				keyBuf[i] = byte(llvm.ConstExtractValue(llvmKey, []uint32{uint32(i)}).ZExtValue())
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
func (v *MapValue) PutString(keyBuf, keyLen, valPtr *LocalValue) {
	if !v.Underlying.IsNil() {
		panic("map already created")
	}

	if valPtr.Underlying.Opcode() == llvm.BitCast {
		valPtr = &LocalValue{v.Eval, valPtr.Underlying.Operand(0)}
	}
	value := valPtr.Load()
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

	keyType := v.Eval.Mod.GetTypeByName("runtime._string")
	v.KeyType = keyType
	key := getZeroValue(keyType)
	key = llvm.ConstInsertValue(key, keyBuf.Value(), []uint32{0})
	key = llvm.ConstInsertValue(key, keyLen.Value(), []uint32{1})

	// TODO: avoid duplicate keys
	v.Keys = append(v.Keys, &LocalValue{v.Eval, key})
	v.Values = append(v.Values, &LocalValue{v.Eval, value})
}

// PutBinary does a map assign operation.
func (v *MapValue) PutBinary(keyPtr, valPtr *LocalValue) {
	if !v.Underlying.IsNil() {
		panic("map already created")
	}

	if valPtr.Underlying.Opcode() == llvm.BitCast {
		valPtr = &LocalValue{v.Eval, valPtr.Underlying.Operand(0)}
	}
	value := valPtr.Load()
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

	if keyPtr.Underlying.Opcode() == llvm.BitCast {
		keyPtr = &LocalValue{v.Eval, keyPtr.Underlying.Operand(0)}
	} else if keyPtr.Underlying.Opcode() == llvm.GetElementPtr {
		keyPtr = &LocalValue{v.Eval, keyPtr.Underlying.Operand(0)}
	}
	key := keyPtr.Load()
	if v.KeyType.IsNil() {
		v.KeyType = key.Type()
		if int(v.Eval.TargetData.TypeAllocSize(v.KeyType)) != v.KeySize {
			panic("interp: map store key type has the wrong size")
		}
	} else {
		if key.Type() != v.KeyType {
			panic("interp: map store key type is inconsistent")
		}
	}

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
