package compiler

// This file emits the correct map intrinsics for map operations.

import (
	"fmt"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

const (
	hashmapBucketSlots  = 8
	hashmapMaxKeySize   = 128
	hashmapMaxValueSize = 128
)

// createMakeMap creates a new map object (runtime.hashmap) by allocating and
// initializing an appropriately sized object.
func (b *builder) createMakeMap(expr *ssa.MakeMap) (llvm.Value, error) {
	mapType := expr.Type().Underlying().(*types.Map)
	keyType := mapType.Key().Underlying()
	llvmValueType := b.getLLVMType(mapType.Elem().Underlying())
	llvmKeyType := b.getLLVMType(keyType)

	keySize := b.targetData.TypeAllocSize(llvmKeyType)
	valueSize := b.targetData.TypeAllocSize(llvmValueType)
	llvmKeySize := llvm.ConstInt(b.uintptrType, keySize, false)
	llvmValueSize := llvm.ConstInt(b.uintptrType, valueSize, false)
	mapLayout := b.getHashmapTypeInfo(mapType, expr.Pos())

	sizeHint := llvm.ConstInt(b.uintptrType, 8, false)
	if expr.Reserve != nil {
		sizeHint = b.getValue(expr.Reserve, getPos(expr))
		var err error
		sizeHint, err = b.createConvert(expr.Reserve.Type(), types.Typ[types.Uintptr], sizeHint, expr.Pos())
		if err != nil {
			return llvm.Value{}, err
		}
	}

	// Resolve hash and equal functions for this key type.
	pattern := b.memCmpPattern(keyType)
	hashFn := b.memHashFunc(pattern)
	equalFn := b.memCmpFunc(pattern)
	hashFn = b.ctx.ConstStruct([]llvm.Value{b.dataPtrNull, hashFn}, false)
	equalFn = b.ctx.ConstStruct([]llvm.Value{b.dataPtrNull, equalFn}, false)

	hashmap := b.createRuntimeCall("hashmapMakeGeneric", []llvm.Value{
		llvmKeySize, llvmValueSize, sizeHint,
		mapLayout,
		hashFn, equalFn,
	}, "")
	return hashmap, nil
}

func (c *context) getHashmapTypeInfo(mapType *types.Map, pos token.Pos) llvm.Value {
	llvmKeyType := c.getLLVMType(mapType.Key().Underlying())
	llvmValueType := c.getLLVMType(mapType.Elem().Underlying())
	keySize := c.targetData.TypeAllocSize(llvmKeyType)
	valueSize := c.targetData.TypeAllocSize(llvmValueType)
	keyLayout := c.createObjectLayout(llvmKeyType, pos)
	valueLayout := c.createObjectLayout(llvmValueType, pos)

	llvmKeySlotType := llvmKeyType
	if keySize > hashmapMaxKeySize {
		llvmKeySlotType = c.dataPtrType
	}
	llvmValueSlotType := llvmValueType
	if valueSize > hashmapMaxValueSize {
		llvmValueSlotType = c.dataPtrType
	}

	// Keep this in sync with runtime.hashmapBucket and
	// runtime.hashmapBucketHeaderSize.
	pointerSize := uint64(c.ptrBits) / 8
	headerSize := (uint64(8) + pointerSize + 7) &^ 7
	headerPadding := headerSize - uint64(8) - pointerSize
	bucketFields := []llvm.Type{
		llvm.ArrayType(c.i8, hashmapBucketSlots),
		c.dataPtrType,
	}
	if headerPadding != 0 {
		bucketFields = append(bucketFields, llvm.ArrayType(c.i8, int(headerPadding)))
	}
	bucketFields = append(bucketFields,
		llvm.ArrayType(llvmKeySlotType, hashmapBucketSlots),
		llvm.ArrayType(llvmValueSlotType, hashmapBucketSlots),
	)
	bucketType := c.ctx.StructType(bucketFields, true)
	bucketSize := headerSize +
		c.targetData.TypeAllocSize(llvmKeySlotType)*hashmapBucketSlots +
		c.targetData.TypeAllocSize(llvmValueSlotType)*hashmapBucketSlots
	if c.targetData.TypeAllocSize(bucketType) != bucketSize {
		panic("compiler hashmap bucket layout does not match runtime")
	}
	bucketLayout := c.createObjectLayout(bucketType, pos)
	mapLayoutName := "runtime.hashmapType:" +
		hashmapCanonicalTypeName(mapType.Key()) + ":" +
		hashmapCanonicalTypeName(mapType.Elem())
	mapLayout := c.mod.NamedGlobal(mapLayoutName)
	if mapLayout.IsNil() {
		initializer := c.ctx.ConstStruct([]llvm.Value{
			keyLayout,
			valueLayout,
			bucketLayout,
		}, false)
		mapLayout = llvm.AddGlobal(c.mod, initializer.Type(), mapLayoutName)
		mapLayout.SetInitializer(initializer)
		mapLayout.SetGlobalConstant(true)
		mapLayout.SetUnnamedAddr(true)
		mapLayout.SetLinkage(llvm.LinkOnceODRLinkage)
	}
	return mapLayout
}

// createMapLookup returns the value in a map. It calls a runtime function
// depending on the map key type to load the map value and its comma-ok value.
func (b *builder) createMapLookup(keyType, valueType types.Type, m llvm.Value, key ssa.Value, commaOk bool, pos token.Pos) (llvm.Value, error) {
	llvmValueType := b.getLLVMType(valueType)

	// Allocate the memory for the resulting type. Do not zero this memory: it
	// will be zeroed by the hashmap get implementation if the key is not
	// present in the map.
	result := b.createRuntimeValueResult(llvmValueType, commaOk, false, "hashmap")
	mapValueAlloca := result.valuePtr

	// We need the map size (with type uintptr) to pass to the hashmap*Get
	// functions. This is necessary because those *Get functions are valid on
	// nil maps, and they'll need to zero the value pointer by that number of
	// bytes.
	mapValueSize := result.valueSize
	if mapValueSize.Type().IntTypeWidth() > b.ptrBits {
		mapValueSize = llvm.ConstTrunc(mapValueSize, b.uintptrType)
	}

	// Do the lookup. How it is done depends on the key type.
	var commaOkValue llvm.Value
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, b.getValue(key, getPos(key)), mapValueAlloca, mapValueSize}
		commaOkValue = b.createRuntimeCall("hashmapStringGet", params, "")
	} else {
		// Key stored at actual type: either binary-comparable or with
		// compiler-generated hash/equal.
		mapKey := b.getValueStorage(key, "hashmap.key")
		params := []llvm.Value{m, mapKey.ptr, mapValueAlloca, mapValueSize}
		if b.isByteComparable(keyType) {
			commaOkValue = b.createRuntimeCall("hashmapBinaryGet", params, "")
		} else {
			commaOkValue = b.createRuntimeInvoke("hashmapGenericGet", params, "")
		}
		b.endValueStorage(mapKey)
	}

	// The value is set to the zero value if the key doesn't exist.
	return result.finish(b, commaOkValue, ""), nil
}

// createMapUpdate updates a map key to a given value, by creating an
// appropriate runtime call.
func (b *builder) createMapUpdate(keyType types.Type, m llvm.Value, key, value ssa.Value, pos token.Pos) {
	storedValue := b.getValueStorage(value, "hashmap.value")
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, b.getValue(key, getPos(key)), storedValue.ptr}
		b.createRuntimeInvoke("hashmapStringSet", params, "")
	} else {
		// Key stored at actual type.
		keyStorage := b.getValueStorage(key, "hashmap.key")
		fnName := "hashmapBinarySet"
		if !b.isByteComparable(keyType) {
			fnName = "hashmapGenericSet"
		}
		params := []llvm.Value{m, keyStorage.ptr, storedValue.ptr}
		b.createRuntimeInvoke(fnName, params, "")
		b.endValueStorage(keyStorage)
	}
	b.endValueStorage(storedValue)
}

// createMapDelete deletes a key from a map by calling the appropriate runtime
// function. It is the implementation of the Go delete() builtin.
func (b *builder) createMapDelete(keyType types.Type, m, key llvm.Value, pos token.Pos) error {
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key}
		b.createRuntimeCall("hashmapStringDelete", params, "")
		return nil
	} else {
		// Key stored at actual type.
		keyAlloca, keySize := b.createTemporaryAlloca(key.Type(), "hashmap.key")
		b.CreateStore(key, keyAlloca)
		params := []llvm.Value{m, keyAlloca}
		if b.isByteComparable(keyType) {
			b.createRuntimeCall("hashmapBinaryDelete", params, "")
		} else {
			b.createRuntimeInvoke("hashmapGenericDelete", params, "")
		}
		b.emitLifetimeEnd(keyAlloca, keySize)
		return nil
	}
}

// Clear the given map.
func (b *builder) createMapClear(m llvm.Value) {
	b.createRuntimeCall("hashmapClear", []llvm.Value{m}, "")
}

// createMapIteratorNext lowers the *ssa.Next instruction for iterating over a
// map. It returns a tuple of {bool, key, value} with the result of the
// iteration.
func (b *builder) createMapIteratorNext(rangeVal ssa.Value, llvmRangeVal, it llvm.Value) llvm.Value {
	// Determine the type of the values to return from the *ssa.Next
	// instruction. It is returned as {bool, keyType, valueType}.
	keyType := rangeVal.Type().Underlying().(*types.Map).Key()
	valueType := rangeVal.Type().Underlying().(*types.Map).Elem()
	llvmKeyType := b.getLLVMType(keyType)
	llvmValueType := b.getLLVMType(valueType)

	// All key types are now stored at their declared type (no interface wrapping).

	// Extract the key and value from the map.
	mapKeyAlloca, mapKeySize := b.createTemporaryAlloca(llvmKeyType, "range.key")
	mapValueAlloca, mapValueSize := b.createTemporaryAlloca(llvmValueType, "range.value")
	ok := b.createRuntimeCall("hashmapNext", []llvm.Value{llvmRangeVal, it, mapKeyAlloca, mapValueAlloca}, "range.next")
	mapKey := b.CreateLoad(llvmKeyType, mapKeyAlloca, "")
	mapValue := b.CreateLoad(llvmValueType, mapValueAlloca, "")

	// End the lifetimes of the allocas, because we're done with them.
	b.emitLifetimeEnd(mapKeyAlloca, mapKeySize)
	b.emitLifetimeEnd(mapValueAlloca, mapValueSize)

	// Construct the *ssa.Next return value: {ok, mapKey, mapValue}
	tuple := llvm.Undef(b.ctx.StructType([]llvm.Type{b.i1, llvmKeyType, llvmValueType}, false))
	tuple = b.CreateInsertValue(tuple, ok, 0, "")
	tuple = b.CreateInsertValue(tuple, mapKey, 1, "")
	tuple = b.CreateInsertValue(tuple, mapValue, 2, "")

	return tuple
}

// hashmapCanonicalTypeName returns a string representation of the hash/equal
// operations needed for a type, stripping named types where the operation does
// not depend on the name. Pointer and channel names do not include the element
// type because their hash/equal operations only use the pointer word.
func hashmapCanonicalTypeName(t types.Type) string {
	switch t := t.Underlying().(type) {
	case *types.Basic:
		return t.Name()
	case *types.Pointer:
		return "*"
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			return "chan"
		case types.SendOnly:
			return "chan<-"
		case types.RecvOnly:
			return "<-chan"
		}
	case *types.Interface:
		if t.NumMethods() == 0 {
			return "interface{}"
		}
		return t.String()
	case *types.Struct:
		var s strings.Builder
		s.WriteString("struct{")
		for i := 0; i < t.NumFields(); i++ {
			if i > 0 {
				s.WriteString("; ")
			}
			s.WriteString(hashmapCanonicalTypeName(t.Field(i).Type()))
		}
		return s.String() + "}"
	case *types.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), hashmapCanonicalTypeName(t.Elem()))
	}
	return t.String()
}
