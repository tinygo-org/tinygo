package compiler

// This file emits the correct map intrinsics for map operations.

import (
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// createMakeMap creates a new map object (runtime.hashmap) by allocating and
// initializing an appropriately sized object.
func (b *builder) createMakeMap(expr *ssa.MakeMap) (llvm.Value, error) {
	mapType := expr.Type().Underlying().(*types.Map)
	keyType := mapType.Key().Underlying()
	llvmValueType := b.getLLVMType(mapType.Elem().Underlying())
	var llvmKeyType llvm.Type
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// String keys.
		llvmKeyType = b.getLLVMType(keyType)
	} else if hashmapIsBinaryKey(keyType) {
		// Trivially comparable keys.
		llvmKeyType = b.getLLVMType(keyType)
	} else {
		// All other keys. Implemented as map[interface{}]valueType for ease of
		// implementation.
		llvmKeyType = b.getLLVMRuntimeType("_interface")
	}
	keySize := b.targetData.TypeAllocSize(llvmKeyType)
	valueSize := b.targetData.TypeAllocSize(llvmValueType)
	llvmKeySize := llvm.ConstInt(b.ctx.Int8Type(), keySize, false)
	llvmValueSize := llvm.ConstInt(b.ctx.Int8Type(), valueSize, false)
	sizeHint := llvm.ConstInt(b.uintptrType, 8, false)
	if expr.Reserve != nil {
		sizeHint = b.getValue(expr.Reserve)
		var err error
		sizeHint, err = b.createConvert(expr.Reserve.Type(), types.Typ[types.Uintptr], sizeHint, expr.Pos())
		if err != nil {
			return llvm.Value{}, err
		}
	}
	hashmap := b.createRuntimeCall("hashmapMake", []llvm.Value{llvmKeySize, llvmValueSize, sizeHint}, "")
	return hashmap, nil
}

// createMapLookup returns the value in a map. It calls a runtime function
// depending on the map key type to load the map value and its comma-ok value.
func (b *builder) createMapLookup(keyType, valueType types.Type, m, key llvm.Value, commaOk bool, pos token.Pos) (llvm.Value, error) {
	llvmValueType := b.getLLVMType(valueType)

	// Allocate the memory for the resulting type. Do not zero this memory: it
	// will be zeroed by the hashmap get implementation if the key is not
	// present in the map.
	mapValueAlloca, mapValuePtr, mapValueAllocaSize := b.createTemporaryAlloca(llvmValueType, "hashmap.value")

	// We need the map size (with type uintptr) to pass to the hashmap*Get
	// functions. This is necessary because those *Get functions are valid on
	// nil maps, and they'll need to zero the value pointer by that number of
	// bytes.
	mapValueSize := mapValueAllocaSize
	if mapValueSize.Type().IntTypeWidth() > b.uintptrType.IntTypeWidth() {
		mapValueSize = llvm.ConstTrunc(mapValueSize, b.uintptrType)
	}

	// Do the lookup. How it is done depends on the key type.
	var commaOkValue llvm.Value
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key, mapValuePtr, mapValueSize}
		commaOkValue = b.createRuntimeCall("hashmapStringGet", params, "")
	} else if hashmapIsBinaryKey(keyType) {
		// key can be compared with runtime.memequal
		// Store the key in an alloca, in the entry block to avoid dynamic stack
		// growth.
		mapKeyAlloca, mapKeyPtr, mapKeySize := b.createTemporaryAlloca(key.Type(), "hashmap.key")
		b.CreateStore(key, mapKeyAlloca)
		// Fetch the value from the hashmap.
		params := []llvm.Value{m, mapKeyPtr, mapValuePtr, mapValueSize}
		commaOkValue = b.createRuntimeCall("hashmapBinaryGet", params, "")
		b.emitLifetimeEnd(mapKeyPtr, mapKeySize)
	} else {
		// Not trivially comparable using memcmp. Make it an interface instead.
		itfKey := key
		if _, ok := keyType.(*types.Interface); !ok {
			// Not already an interface, so convert it to an interface now.
			itfKey = b.createMakeInterface(key, keyType, pos)
		}
		params := []llvm.Value{m, itfKey, mapValuePtr, mapValueSize}
		commaOkValue = b.createRuntimeCall("hashmapInterfaceGet", params, "")
	}

	// Load the resulting value from the hashmap. The value is set to the zero
	// value if the key doesn't exist in the hashmap.
	mapValue := b.CreateLoad(mapValueAlloca, "")
	b.emitLifetimeEnd(mapValuePtr, mapValueAllocaSize)

	if commaOk {
		tuple := llvm.Undef(b.ctx.StructType([]llvm.Type{llvmValueType, b.ctx.Int1Type()}, false))
		tuple = b.CreateInsertValue(tuple, mapValue, 0, "")
		tuple = b.CreateInsertValue(tuple, commaOkValue, 1, "")
		return tuple, nil
	} else {
		return mapValue, nil
	}
}

// createMapUpdate updates a map key to a given value, by creating an
// appropriate runtime call.
func (b *builder) createMapUpdate(keyType types.Type, m, key, value llvm.Value, pos token.Pos) {
	valueAlloca, valuePtr, valueSize := b.createTemporaryAlloca(value.Type(), "hashmap.value")
	b.CreateStore(value, valueAlloca)
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key, valuePtr}
		b.createRuntimeCall("hashmapStringSet", params, "")
	} else if hashmapIsBinaryKey(keyType) {
		// key can be compared with runtime.memequal
		keyAlloca, keyPtr, keySize := b.createTemporaryAlloca(key.Type(), "hashmap.key")
		b.CreateStore(key, keyAlloca)
		params := []llvm.Value{m, keyPtr, valuePtr}
		b.createRuntimeCall("hashmapBinarySet", params, "")
		b.emitLifetimeEnd(keyPtr, keySize)
	} else {
		// Key is not trivially comparable, so compare it as an interface instead.
		itfKey := key
		if _, ok := keyType.(*types.Interface); !ok {
			// Not already an interface, so convert it to an interface first.
			itfKey = b.createMakeInterface(key, keyType, pos)
		}
		params := []llvm.Value{m, itfKey, valuePtr}
		b.createRuntimeCall("hashmapInterfaceSet", params, "")
	}
	b.emitLifetimeEnd(valuePtr, valueSize)
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
	} else if hashmapIsBinaryKey(keyType) {
		keyAlloca, keyPtr, keySize := b.createTemporaryAlloca(key.Type(), "hashmap.key")
		b.CreateStore(key, keyAlloca)
		params := []llvm.Value{m, keyPtr}
		b.createRuntimeCall("hashmapBinaryDelete", params, "")
		b.emitLifetimeEnd(keyPtr, keySize)
		return nil
	} else {
		// Key is not trivially comparable, so compare it as an interface
		// instead.
		itfKey := key
		if _, ok := keyType.(*types.Interface); !ok {
			// Not already an interface, so convert it to an interface first.
			itfKey = b.createMakeInterface(key, keyType, pos)
		}
		params := []llvm.Value{m, itfKey}
		b.createRuntimeCall("hashmapInterfaceDelete", params, "")
		return nil
	}
}

// createMapIteratorNext lowers the *ssa.Next instruction for iterating over a
// map. It returns a tuple of {bool, key, value} with the result of the
// iteration.
func (b *builder) createMapIteratorNext(rangeVal ssa.Value, llvmRangeVal, it llvm.Value) llvm.Value {
	llvmKeyType := b.getLLVMType(rangeVal.Type().Underlying().(*types.Map).Key())
	llvmValueType := b.getLLVMType(rangeVal.Type().Underlying().(*types.Map).Elem())

	mapKeyAlloca, mapKeyPtr, mapKeySize := b.createTemporaryAlloca(llvmKeyType, "range.key")
	mapValueAlloca, mapValuePtr, mapValueSize := b.createTemporaryAlloca(llvmValueType, "range.value")
	ok := b.createRuntimeCall("hashmapNext", []llvm.Value{llvmRangeVal, it, mapKeyPtr, mapValuePtr}, "range.next")

	tuple := llvm.Undef(b.ctx.StructType([]llvm.Type{b.ctx.Int1Type(), llvmKeyType, llvmValueType}, false))
	tuple = b.CreateInsertValue(tuple, ok, 0, "")
	tuple = b.CreateInsertValue(tuple, b.CreateLoad(mapKeyAlloca, ""), 1, "")
	tuple = b.CreateInsertValue(tuple, b.CreateLoad(mapValueAlloca, ""), 2, "")
	b.emitLifetimeEnd(mapKeyPtr, mapKeySize)
	b.emitLifetimeEnd(mapValuePtr, mapValueSize)
	return tuple
}

// Returns true if this key type does not contain strings, interfaces etc., so
// can be compared with runtime.memequal.
func hashmapIsBinaryKey(keyType types.Type) bool {
	switch keyType := keyType.(type) {
	case *types.Basic:
		return keyType.Info()&(types.IsBoolean|types.IsInteger) != 0
	case *types.Pointer:
		return true
	case *types.Struct:
		for i := 0; i < keyType.NumFields(); i++ {
			fieldType := keyType.Field(i).Type().Underlying()
			if !hashmapIsBinaryKey(fieldType) {
				return false
			}
		}
		return true
	case *types.Array:
		return hashmapIsBinaryKey(keyType.Elem())
	case *types.Named:
		return hashmapIsBinaryKey(keyType.Underlying())
	default:
		return false
	}
}
