package compiler

// This file emits the correct map intrinsics for map operations.

import (
	"fmt"
	"go/token"
	"go/types"
	"math/big"

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
	layoutValue := b.createMapLayout(llvmKeyType, llvmValueType, expr.Pos())
	hashmap := b.createRuntimeCall("hashmapMake", []llvm.Value{llvmKeySize, llvmValueSize, sizeHint, layoutValue}, "")
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
func (b *builder) createMapUpdate(keyType, elemType types.Type, m, key, value llvm.Value, pos token.Pos) {
	valueAlloca, valuePtr, valueSize := b.createTemporaryAlloca(value.Type(), "hashmap.value")
	b.CreateStore(value, valueAlloca)
	keyType = keyType.Underlying()
	elemType = elemType.Underlying()
	layoutValue := b.createMapLayout(b.getLLVMType(keyType), b.getLLVMType(elemType), pos)

	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key, valuePtr, layoutValue}
		b.createRuntimeCall("hashmapStringSet", params, "")
	} else if hashmapIsBinaryKey(keyType) {
		// key can be compared with runtime.memequal
		keyAlloca, keyPtr, keySize := b.createTemporaryAlloca(key.Type(), "hashmap.key")
		b.CreateStore(key, keyAlloca)
		params := []llvm.Value{m, keyPtr, valuePtr, layoutValue}
		b.createRuntimeCall("hashmapBinarySet", params, "")
		b.emitLifetimeEnd(keyPtr, keySize)
	} else {
		// Key is not trivially comparable, so compare it as an interface instead.
		itfKey := key
		if _, ok := keyType.(*types.Interface); !ok {
			// Not already an interface, so convert it to an interface first.
			itfKey = b.createMakeInterface(key, keyType, pos)
		}
		params := []llvm.Value{m, itfKey, valuePtr, layoutValue}
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
	// Determine the type of the values to return from the *ssa.Next
	// instruction. It is returned as {bool, keyType, valueType}.
	keyType := rangeVal.Type().Underlying().(*types.Map).Key()
	valueType := rangeVal.Type().Underlying().(*types.Map).Elem()
	llvmKeyType := b.getLLVMType(keyType)
	llvmValueType := b.getLLVMType(valueType)

	// There is a special case in which keys are stored as an interface value
	// instead of the value they normally are. This happens for non-trivially
	// comparable types such as float32 or some structs.
	isKeyStoredAsInterface := false
	if t, ok := keyType.Underlying().(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
	} else if hashmapIsBinaryKey(keyType) {
		// key can be compared with runtime.memequal
	} else {
		// The key is stored as an interface value, and may or may not be an
		// interface type (for example, float32 keys are stored as an interface
		// value).
		if _, ok := keyType.Underlying().(*types.Interface); !ok {
			isKeyStoredAsInterface = true
		}
	}

	// Determine the type of the key as stored in the map.
	llvmStoredKeyType := llvmKeyType
	if isKeyStoredAsInterface {
		llvmStoredKeyType = b.getLLVMRuntimeType("_interface")
	}

	// Extract the key and value from the map.
	mapKeyAlloca, mapKeyPtr, mapKeySize := b.createTemporaryAlloca(llvmStoredKeyType, "range.key")
	mapValueAlloca, mapValuePtr, mapValueSize := b.createTemporaryAlloca(llvmValueType, "range.value")
	ok := b.createRuntimeCall("hashmapNext", []llvm.Value{llvmRangeVal, it, mapKeyPtr, mapValuePtr}, "range.next")
	mapKey := b.CreateLoad(mapKeyAlloca, "")
	mapValue := b.CreateLoad(mapValueAlloca, "")

	if isKeyStoredAsInterface {
		// The key is stored as an interface but it isn't of interface type.
		// Extract the underlying value.
		mapKey = b.extractValueFromInterface(mapKey, llvmKeyType)
	}

	// End the lifetimes of the allocas, because we're done with them.
	b.emitLifetimeEnd(mapKeyPtr, mapKeySize)
	b.emitLifetimeEnd(mapValuePtr, mapValueSize)

	// Construct the *ssa.Next return value: {ok, mapKey, mapValue}
	tuple := llvm.Undef(b.ctx.StructType([]llvm.Type{b.ctx.Int1Type(), llvmKeyType, llvmValueType}, false))
	tuple = b.CreateInsertValue(tuple, ok, 0, "")
	tuple = b.CreateInsertValue(tuple, mapKey, 1, "")
	tuple = b.CreateInsertValue(tuple, mapValue, 2, "")

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

// Returns the size in bytes and the bitmap for a hashmap's bucket allocation
// given the pointer size and sizes and bitmaps for the key and element.
func hashmapBucketLayout(
	pointerSize, pointerAlign uint64,
	tSizeBytes uint64, tbitmap *big.Int,
	eSizeBytes uint64, ebitmap *big.Int,
) (uint64, *big.Int) {
	t := Bits{
		uint(tSizeBytes / pointerAlign),
		tbitmap,
	}
	e := Bits{
		uint(eSizeBytes / pointerAlign),
		ebitmap,
	}
	t.AssertValidBitLen()
	e.AssertValidBitLen()

	// Currently, a hashmap bucket is some overhead followed by 8 keys followed by 8 values.
	// The overhead consists of 8 bytes followed by a pointer field.
	//
	// From src/runtime/hashmap.go: func hashmapInsertIntoNewBucket
	//		bucketBufSize := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*8
	//
	//		type hashmapBucket struct {
	//			tophash [8]uint8
	//			next    *hashmapBucket // next bucket (if there are more than 8 in a chain)
	//			// Followed by the actual keys, and then the actual values. These are
	//			// allocated but as they're of variable size they can't be shown here.
	//		}
	//
	//		So the overhead is 8 bytes plus the size of a uintptr (aka pointerSize).
	//		Will assume the alignment is taken care of by 8 bytes always preceding the 'next' pointer.

	tophash := Repeat(Zero(), 8/uint(pointerAlign))
	next := One()
	if pointerSize > pointerAlign {
		// Assuming the size of an alloc layout parameter represents the alignment were
		// a pointer can start, rather than the words that back-to-back pointers would
		// occupy, as is the case for the AVR where pointers can start on any byte but
		// are 2 bytes long: we want a bitmap of '1' followed by as many zeros as there
		// are surplus pointerSize bytes over the pointerAlign. For the AVR, this would
		// make next bitmap '10' and have a size of 2.
		next = Concat(One(), Repeat(Zero(), uint(pointerSize-pointerAlign)))
	} else if pointerSize < pointerAlign {
		println(pointerSize, pointerAlign)
		panic("we don't know how to create a layout yet when a pointer's size is smaller than it's alignment")
	}
	key := t
	if tSizeBytes < pointerAlign {
		key = Repeat(Zero(), uint((8*tSizeBytes)/pointerAlign))
	} else {
		key = By8(key)
	}
	elem := e
	if eSizeBytes < pointerAlign {
		elem = Repeat(Zero(), uint((8*eSizeBytes)/pointerAlign))
	} else {
		elem = By8(elem)
	}

	// Bucket represents the hashmap bucket layout.
	bucket := Concat(tophash, next, key, elem)

	// Each bit in a Bit represents a word in an allocation layout. A word is the same size as a
	// pointer. The overhead is thus the number of words in [8]uint8 and a pointer word.

	sizeBytes := uint64(bucket.Size()) * pointerAlign
	{
		// TODO A little extra sanity checking. Probably not worth keeping in
		// the long run as the actual allocation size will be tested to be a
		// multiple of the layout's advertised size anyway. While this is new,
		// it can't hurt.
		// Thu Mar 17 17:14:33 EDT 2022
		expectedSize := 8 + pointerAlign + 8*tSizeBytes + 8*eSizeBytes
		if expectedSize != sizeBytes {
			println(tSizeBytes, eSizeBytes, pointerAlign)
			panic(fmt.Sprintf("expectedSize != sizeBytes, %d, %d", expectedSize, sizeBytes))
		}
	}
	return sizeBytes, bucket.Int()
}
