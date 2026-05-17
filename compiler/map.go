package compiler

// This file emits the correct map intrinsics for map operations.

import (
	"fmt"
	"go/token"
	"go/types"

	"github.com/tinygo-org/tinygo/src/tinygo"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
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
	sizeHint := llvm.ConstInt(b.uintptrType, 8, false)
	if expr.Reserve != nil {
		sizeHint = b.getValue(expr.Reserve, getPos(expr))
		var err error
		sizeHint, err = b.createConvert(expr.Reserve.Type(), types.Typ[types.Uintptr], sizeHint, expr.Pos())
		if err != nil {
			return llvm.Value{}, err
		}
	}

	if hashmapCanGenerateHashEqual(keyType) && !hashmapIsBinaryKey(keyType) {
		// Composite keys: use compiler-generated hash/equal functions.
		// Binary and string keys use the more efficient dedicated paths
		// (hashmapMake with algorithm enum) which avoid function pointer
		// indirection.
		hashFn := b.getOrGenerateKeyHashFunc(keyType)
		equalFn := b.getOrGenerateKeyEqualFunc(keyType)
		hashFuncValue := b.createFuncValue(hashFn, llvm.ConstNull(b.dataPtrType), hashmapKeyHashSignature())
		equalFuncValue := b.createFuncValue(equalFn, llvm.ConstNull(b.dataPtrType), hashmapKeyEqualSignature())
		hashmap := b.createRuntimeCall("hashmapMakeGeneric", []llvm.Value{
			llvmKeySize, llvmValueSize, sizeHint,
			hashFuncValue, equalFuncValue,
		}, "")
		return hashmap, nil
	}

	var alg uint64
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		alg = uint64(tinygo.HashmapAlgorithmString)
	} else if hashmapIsBinaryKey(keyType) {
		alg = uint64(tinygo.HashmapAlgorithmBinary)
	} else {
		// Fallback for types not handled by hashmapCanGenerateHashEqual
		// (currently only unsafe.Pointer due to an interp issue).
		llvmKeyType = b.getLLVMRuntimeType("_interface")
		alg = uint64(tinygo.HashmapAlgorithmInterface)
	}
	algEnum := llvm.ConstInt(b.ctx.Int8Type(), alg, false)
	hashmap := b.createRuntimeCall("hashmapMake", []llvm.Value{llvmKeySize, llvmValueSize, sizeHint, algEnum}, "")
	return hashmap, nil
}

// createMapLookup returns the value in a map. It calls a runtime function
// depending on the map key type to load the map value and its comma-ok value.
func (b *builder) createMapLookup(keyType, valueType types.Type, m, key llvm.Value, commaOk bool, pos token.Pos) (llvm.Value, error) {
	llvmValueType := b.getLLVMType(valueType)

	// Allocate the memory for the resulting type. Do not zero this memory: it
	// will be zeroed by the hashmap get implementation if the key is not
	// present in the map.
	mapValueAlloca, mapValueAllocaSize := b.createTemporaryAlloca(llvmValueType, "hashmap.value")

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
	origKeyType := keyType
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key, mapValueAlloca, mapValueSize}
		commaOkValue = b.createRuntimeCall("hashmapStringGet", params, "")
	} else if hashmapIsBinaryKey(keyType) || hashmapCanGenerateHashEqual(keyType) {
		// Key stored at actual type: either binary-comparable or with
		// compiler-generated hash/equal.
		mapKeyAlloca, mapKeySize := b.createTemporaryAlloca(key.Type(), "hashmap.key")
		b.CreateStore(key, mapKeyAlloca)
		params := []llvm.Value{m, mapKeyAlloca, mapValueAlloca, mapValueSize}
		fnName := "hashmapBinaryGet"
		if !hashmapIsBinaryKey(keyType) {
			fnName = "hashmapGenericGet"
		}
		commaOkValue = b.createRuntimeCall(fnName, params, "")
		b.emitLifetimeEnd(mapKeyAlloca, mapKeySize)
	} else {
		// Not trivially comparable using memcmp. Make it an interface instead.
		itfKey := key
		if _, ok := keyType.(*types.Interface); !ok {
			// Not already an interface, so convert it to an interface now.
			itfKey = b.createMakeInterface(key, origKeyType, pos)
		}
		params := []llvm.Value{m, itfKey, mapValueAlloca, mapValueSize}
		commaOkValue = b.createRuntimeCall("hashmapInterfaceGet", params, "")
	}

	// Load the resulting value from the hashmap. The value is set to the zero
	// value if the key doesn't exist in the hashmap.
	mapValue := b.CreateLoad(llvmValueType, mapValueAlloca, "")
	b.emitLifetimeEnd(mapValueAlloca, mapValueAllocaSize)

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
	valueAlloca, valueSize := b.createTemporaryAlloca(value.Type(), "hashmap.value")
	b.CreateStore(value, valueAlloca)
	origKeyType := keyType
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key, valueAlloca}
		b.createRuntimeCall("hashmapStringSet", params, "")
	} else if hashmapIsBinaryKey(keyType) || hashmapCanGenerateHashEqual(keyType) {
		// Key stored at actual type.
		keyAlloca, keySize := b.createTemporaryAlloca(key.Type(), "hashmap.key")
		b.CreateStore(key, keyAlloca)
		fnName := "hashmapBinarySet"
		if !hashmapIsBinaryKey(keyType) {
			fnName = "hashmapGenericSet"
		}
		params := []llvm.Value{m, keyAlloca, valueAlloca}
		b.createRuntimeCall(fnName, params, "")
		b.emitLifetimeEnd(keyAlloca, keySize)
	} else {
		// Key is not trivially comparable, so compare it as an interface instead.
		itfKey := key
		if _, ok := keyType.(*types.Interface); !ok {
			// Not already an interface, so convert it to an interface first.
			itfKey = b.createMakeInterface(key, origKeyType, pos)
		}
		params := []llvm.Value{m, itfKey, valueAlloca}
		b.createRuntimeCall("hashmapInterfaceSet", params, "")
	}
	b.emitLifetimeEnd(valueAlloca, valueSize)
}

// createMapDelete deletes a key from a map by calling the appropriate runtime
// function. It is the implementation of the Go delete() builtin.
func (b *builder) createMapDelete(keyType types.Type, m, key llvm.Value, pos token.Pos) error {
	origKeyType := keyType
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key}
		b.createRuntimeCall("hashmapStringDelete", params, "")
		return nil
	} else if hashmapIsBinaryKey(keyType) || hashmapCanGenerateHashEqual(keyType) {
		// Key stored at actual type.
		keyAlloca, keySize := b.createTemporaryAlloca(key.Type(), "hashmap.key")
		b.CreateStore(key, keyAlloca)
		fnName := "hashmapBinaryDelete"
		if !hashmapIsBinaryKey(keyType) {
			fnName = "hashmapGenericDelete"
		}
		params := []llvm.Value{m, keyAlloca}
		b.createRuntimeCall(fnName, params, "")
		b.emitLifetimeEnd(keyAlloca, keySize)
		return nil
	} else {
		// Key is not trivially comparable, so compare it as an interface
		// instead.
		itfKey := key
		if _, ok := keyType.(*types.Interface); !ok {
			// Not already an interface, so convert it to an interface first.
			itfKey = b.createMakeInterface(key, origKeyType, pos)
		}
		params := []llvm.Value{m, itfKey}
		b.createRuntimeCall("hashmapInterfaceDelete", params, "")
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

	// Keys are stored as an interface value only for types not handled by
	// the binary or generic paths (currently only unsafe.Pointer).
	isKeyStoredAsInterface := false
	if t, ok := keyType.Underlying().(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
	} else if hashmapIsBinaryKey(keyType) || hashmapCanGenerateHashEqual(keyType) {
		// key stored at actual type
	} else {
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
	mapKeyAlloca, mapKeySize := b.createTemporaryAlloca(llvmStoredKeyType, "range.key")
	mapValueAlloca, mapValueSize := b.createTemporaryAlloca(llvmValueType, "range.value")
	ok := b.createRuntimeCall("hashmapNext", []llvm.Value{llvmRangeVal, it, mapKeyAlloca, mapValueAlloca}, "range.next")
	mapKey := b.CreateLoad(llvmStoredKeyType, mapKeyAlloca, "")
	mapValue := b.CreateLoad(llvmValueType, mapValueAlloca, "")

	if isKeyStoredAsInterface {
		// The key is stored as an interface but it isn't of interface type.
		// Extract the underlying value.
		mapKey = b.extractValueFromInterface(mapKey, llvmKeyType)
	}

	// End the lifetimes of the allocas, because we're done with them.
	b.emitLifetimeEnd(mapKeyAlloca, mapKeySize)
	b.emitLifetimeEnd(mapValueAlloca, mapValueSize)

	// Construct the *ssa.Next return value: {ok, mapKey, mapValue}
	tuple := llvm.Undef(b.ctx.StructType([]llvm.Type{b.ctx.Int1Type(), llvmKeyType, llvmValueType}, false))
	tuple = b.CreateInsertValue(tuple, ok, 0, "")
	tuple = b.CreateInsertValue(tuple, mapKey, 1, "")
	tuple = b.CreateInsertValue(tuple, mapValue, 2, "")

	return tuple
}

// Returns true if this key type does not contain strings, interfaces etc., so
// can be compared with runtime.memequal.  Note that padding bytes are undef
// and can alter two "equal" structs being equal when compared with memequal.
func hashmapIsBinaryKey(keyType types.Type) bool {
	switch keyType := keyType.Underlying().(type) {
	case *types.Basic:
		// TODO: unsafe.Pointer is also a binary key, but to support that we
		// need to fix an issue with interp first (see
		// https://github.com/tinygo-org/tinygo/pull/4898).
		return keyType.Info()&(types.IsBoolean|types.IsInteger) != 0
	case *types.Pointer:
		return true
	case *types.Array:
		return hashmapIsBinaryKey(keyType.Elem())
	default:
		return false
	}
}

// hashmapCanGenerateHashEqual returns true if the compiler can generate
// type-specific hash and equal functions for this key type. This covers all
// comparable types: integers, booleans, strings, floats, complex numbers,
// pointers, channels, interfaces, and composites (structs/arrays) of these.
func hashmapCanGenerateHashEqual(keyType types.Type) bool {
	switch keyType := keyType.Underlying().(type) {
	case *types.Basic:
		// Note: unsafe.Pointer is excluded (not IsBoolean/IsInteger/etc.)
		// due to a known interp issue (see hashmapIsBinaryKey).
		return keyType.Info()&(types.IsBoolean|types.IsInteger|types.IsString|types.IsFloat|types.IsComplex) != 0
	case *types.Pointer:
		return true
	case *types.Chan:
		return true
	case *types.Interface:
		return true
	case *types.Struct:
		for i := 0; i < keyType.NumFields(); i++ {
			fieldType := keyType.Field(i).Type().Underlying()
			if !hashmapCanGenerateHashEqual(fieldType) {
				return false
			}
		}
		return true
	case *types.Array:
		return hashmapCanGenerateHashEqual(keyType.Elem())
	default:
		return false
	}
}

// hashmapKeyHashSignature returns the Go type signature for hashmap key hash
// functions: func(key unsafe.Pointer, size, seed uintptr) uint32
func hashmapKeyHashSignature() *types.Signature {
	return types.NewSignatureType(nil, nil, nil,
		types.NewTuple(
			types.NewVar(token.NoPos, nil, "key", types.Typ[types.UnsafePointer]),
			types.NewVar(token.NoPos, nil, "size", types.Typ[types.Uintptr]),
			types.NewVar(token.NoPos, nil, "seed", types.Typ[types.Uintptr]),
		),
		types.NewTuple(
			types.NewVar(token.NoPos, nil, "", types.Typ[types.Uint32]),
		),
		false,
	)
}

// hashmapKeyEqualSignature returns the Go type signature for hashmap key equal
// functions: func(x, y unsafe.Pointer, n uintptr) bool
func hashmapKeyEqualSignature() *types.Signature {
	return types.NewSignatureType(nil, nil, nil,
		types.NewTuple(
			types.NewVar(token.NoPos, nil, "x", types.Typ[types.UnsafePointer]),
			types.NewVar(token.NoPos, nil, "y", types.Typ[types.UnsafePointer]),
			types.NewVar(token.NoPos, nil, "n", types.Typ[types.Uintptr]),
		),
		types.NewTuple(
			types.NewVar(token.NoPos, nil, "", types.Typ[types.Bool]),
		),
		false,
	)
}

// hashmapKeyFuncName returns a canonical name for a generated hash or equal
// function based on the key type's underlying structure. Named types are
// replaced with their underlying types so that structurally identical key
// types (e.g., struct{i1; str1} and struct{i2; str2} where both i1, i2 are
// int and str1, str2 are string) share the same generated function.
func hashmapKeyFuncName(prefix string, keyType types.Type) string {
	return prefix + "." + hashmapCanonicalTypeName(keyType)
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
		s := "struct{"
		for i := 0; i < t.NumFields(); i++ {
			if i > 0 {
				s += "; "
			}
			s += hashmapCanonicalTypeName(t.Field(i).Type())
		}
		return s + "}"
	case *types.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), hashmapCanonicalTypeName(t.Elem()))
	}
	return t.String()
}

// getOrGenerateKeyHashFunc returns an LLVM function that computes the hash
// of a key of the given type. The function is generated on first call and
// cached in the module.
func (b *builder) getOrGenerateKeyHashFunc(keyType types.Type) llvm.Value {
	name := hashmapKeyFuncName("hashmapKeyHash", keyType)
	if fn := b.mod.NamedFunction(name); !fn.IsNil() {
		return fn
	}

	// Create the LLVM function type:
	// (key ptr, size uintptr, seed uintptr, context ptr) -> i32
	fnType := llvm.FunctionType(b.ctx.Int32Type(), []llvm.Type{
		b.dataPtrType, b.uintptrType, b.uintptrType, b.dataPtrType,
	}, false)
	fn := llvm.AddFunction(b.mod, name, fnType)
	fn.SetLinkage(llvm.LinkOnceODRLinkage)
	fn.SetUnnamedAddr(true)
	b.addStandardAttributes(fn)

	// Generate the function body.
	savedBlock := b.GetInsertBlock()
	defer b.SetInsertPointAtEnd(savedBlock)

	entry := b.ctx.AddBasicBlock(fn, "entry")
	b.SetInsertPointAtEnd(entry)

	keyPtr := fn.Param(0)
	seed := fn.Param(2)
	llvmKeyType := b.getLLVMType(keyType)
	hash := b.generateKeyHash(keyType, llvmKeyType, keyPtr, seed)
	b.CreateRet(hash)

	return fn
}

// getOrGenerateKeyEqualFunc returns an LLVM function that compares two keys
// of the given type for equality. The function is generated on first call
// and cached in the module.
func (b *builder) getOrGenerateKeyEqualFunc(keyType types.Type) llvm.Value {
	name := hashmapKeyFuncName("hashmapKeyEqual", keyType)
	if fn := b.mod.NamedFunction(name); !fn.IsNil() {
		return fn
	}

	// Create the LLVM function type:
	// (x ptr, y ptr, n uintptr, context ptr) -> i1
	fnType := llvm.FunctionType(b.ctx.Int1Type(), []llvm.Type{
		b.dataPtrType, b.dataPtrType, b.uintptrType, b.dataPtrType,
	}, false)
	fn := llvm.AddFunction(b.mod, name, fnType)
	fn.SetLinkage(llvm.LinkOnceODRLinkage)
	fn.SetUnnamedAddr(true)
	b.addStandardAttributes(fn)

	// Generate the function body.
	savedBlock := b.GetInsertBlock()
	defer b.SetInsertPointAtEnd(savedBlock)

	entry := b.ctx.AddBasicBlock(fn, "entry")
	b.SetInsertPointAtEnd(entry)

	xPtr := fn.Param(0)
	yPtr := fn.Param(1)
	llvmKeyType := b.getLLVMType(keyType)
	result := b.generateKeyEqual(keyType, llvmKeyType, xPtr, yPtr, fn)
	b.CreateRet(result)

	return fn
}

// generateKeyHash generates IR that hashes a key value. Returns the i32 hash.
func (b *builder) generateKeyHash(keyType types.Type, llvmKeyType llvm.Type, keyPtr llvm.Value, seed llvm.Value) llvm.Value {
	switch keyType := keyType.Underlying().(type) {
	case *types.Basic:
		if keyType.Info()&types.IsString != 0 {
			// Hash the string contents. The size parameter is unused by
			// hashmapStringPtrHash (it dereferences the string header to
			// get the actual length), but we pass it for signature
			// consistency with other hash functions.
			size := llvm.ConstInt(b.uintptrType, b.targetData.TypeAllocSize(llvmKeyType), false)
			return b.createRuntimeCall("hashmapStringPtrHash", []llvm.Value{keyPtr, size, seed}, "hash")
		}
		if keyType.Info()&types.IsFloat != 0 {
			// Float hash: normalizes -0 to +0 before hashing.
			if keyType.Kind() == types.Float32 {
				return b.createRuntimeCall("hashmapFloat32Hash", []llvm.Value{keyPtr, seed}, "hash")
			}
			return b.createRuntimeCall("hashmapFloat64Hash", []llvm.Value{keyPtr, seed}, "hash")
		}
		if keyType.Info()&types.IsComplex != 0 {
			// Complex hash: hash real and imaginary parts as floats.
			if keyType.Kind() == types.Complex64 {
				realPtr := keyPtr
				imagPtr := b.CreateInBoundsGEP(b.ctx.Int8Type(), keyPtr, []llvm.Value{
					llvm.ConstInt(b.uintptrType, 4, false),
				}, "")
				realHash := b.createRuntimeCall("hashmapFloat32Hash", []llvm.Value{realPtr, seed}, "hash.real")
				imagHash := b.createRuntimeCall("hashmapFloat32Hash", []llvm.Value{imagPtr, seed}, "hash.imag")
				return b.CreateXor(realHash, imagHash, "")
			}
			realPtr := keyPtr
			imagPtr := b.CreateInBoundsGEP(b.ctx.Int8Type(), keyPtr, []llvm.Value{
				llvm.ConstInt(b.uintptrType, 8, false),
			}, "")
			realHash := b.createRuntimeCall("hashmapFloat64Hash", []llvm.Value{realPtr, seed}, "hash.real")
			imagHash := b.createRuntimeCall("hashmapFloat64Hash", []llvm.Value{imagPtr, seed}, "hash.imag")
			return b.CreateXor(realHash, imagHash, "")
		}
		// Integer/boolean: hash the raw bytes.
		size := llvm.ConstInt(b.uintptrType, b.targetData.TypeAllocSize(llvmKeyType), false)
		return b.createRuntimeCall("hash32", []llvm.Value{keyPtr, size, seed}, "hash")
	case *types.Pointer, *types.Chan:
		// Pointers and channels: hash as raw pointer-sized bytes.
		size := llvm.ConstInt(b.uintptrType, b.targetData.TypeAllocSize(llvmKeyType), false)
		return b.createRuntimeCall("hash32", []llvm.Value{keyPtr, size, seed}, "hash")
	case *types.Interface:
		// Interface: use runtime reflection-based hash.
		size := llvm.ConstInt(b.uintptrType, b.targetData.TypeAllocSize(llvmKeyType), false)
		return b.createRuntimeCall("hashmapInterfacePtrHash", []llvm.Value{keyPtr, size, seed}, "hash")
	case *types.Struct:
		hash := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
		zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
		for i := 0; i < keyType.NumFields(); i++ {
			if keyType.Field(i).Name() == "_" {
				continue // blank fields are ignored in Go equality
			}
			fieldType := keyType.Field(i).Type()
			llvmFieldType := b.getLLVMType(fieldType)
			if b.targetData.TypeAllocSize(llvmFieldType) == 0 {
				continue // skip zero-sized fields
			}
			idx := llvm.ConstInt(b.ctx.Int32Type(), uint64(i), false)
			fieldPtr := b.CreateInBoundsGEP(llvmKeyType, keyPtr, []llvm.Value{zero, idx}, "")
			fieldHash := b.generateKeyHash(fieldType, llvmFieldType, fieldPtr, seed)
			hash = b.CreateXor(hash, fieldHash, "")
		}
		return hash
	case *types.Array:
		elemType := keyType.Elem()
		llvmElemType := b.getLLVMType(elemType)
		hash := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
		zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
		for i := 0; i < int(keyType.Len()); i++ {
			idx := llvm.ConstInt(b.uintptrType, uint64(i), false)
			elemPtr := b.CreateInBoundsGEP(llvmKeyType, keyPtr, []llvm.Value{zero, idx}, "")
			elemHash := b.generateKeyHash(elemType, llvmElemType, elemPtr, seed)
			hash = b.CreateXor(hash, elemHash, "")
		}
		return hash
	default:
		panic(fmt.Sprintf("unhandled key type for hash generation: %T", keyType))
	}
}

// generateKeyEqual generates IR that compares two key values for equality.
// Returns an i1 result.
func (b *builder) generateKeyEqual(keyType types.Type, llvmKeyType llvm.Type, xPtr, yPtr llvm.Value, fn llvm.Value) llvm.Value {
	switch keyType := keyType.Underlying().(type) {
	case *types.Basic:
		if keyType.Info()&types.IsString != 0 {
			// Compare strings: load both string headers and compare.
			xStr := b.CreateLoad(llvmKeyType, xPtr, "x.str")
			yStr := b.CreateLoad(llvmKeyType, yPtr, "y.str")
			return b.createRuntimeCall("stringEqual", []llvm.Value{xStr, yStr}, "eq")
		}
		if keyType.Info()&types.IsFloat != 0 {
			// Float equality: fcmp oeq handles -0==+0 (true) and NaN==NaN (false).
			xVal := b.CreateLoad(llvmKeyType, xPtr, "x.float")
			yVal := b.CreateLoad(llvmKeyType, yPtr, "y.float")
			return b.CreateFCmp(llvm.FloatOEQ, xVal, yVal, "eq")
		}
		if keyType.Info()&types.IsComplex != 0 {
			// Complex equality: both real and imaginary parts must be equal.
			var floatType llvm.Type
			if keyType.Kind() == types.Complex64 {
				floatType = b.ctx.FloatType()
			} else {
				floatType = b.ctx.DoubleType()
			}
			floatSize := b.targetData.TypeAllocSize(floatType)
			imagOffset := llvm.ConstInt(b.uintptrType, floatSize, false)
			// Real parts
			xReal := b.CreateLoad(floatType, xPtr, "x.real")
			yReal := b.CreateLoad(floatType, yPtr, "y.real")
			realEq := b.CreateFCmp(llvm.FloatOEQ, xReal, yReal, "eq.real")
			// Imaginary parts
			xImagPtr := b.CreateInBoundsGEP(b.ctx.Int8Type(), xPtr, []llvm.Value{imagOffset}, "")
			yImagPtr := b.CreateInBoundsGEP(b.ctx.Int8Type(), yPtr, []llvm.Value{imagOffset}, "")
			xImag := b.CreateLoad(floatType, xImagPtr, "x.imag")
			yImag := b.CreateLoad(floatType, yImagPtr, "y.imag")
			imagEq := b.CreateFCmp(llvm.FloatOEQ, xImag, yImag, "eq.imag")
			return b.CreateAnd(realEq, imagEq, "")
		}
		// Integer/boolean: compare raw bytes.
		size := llvm.ConstInt(b.uintptrType, b.targetData.TypeAllocSize(llvmKeyType), false)
		return b.createRuntimeCall("memequal", []llvm.Value{xPtr, yPtr, size}, "eq")
	case *types.Pointer, *types.Chan:
		// Pointers and channels: compare as raw pointer-sized bytes.
		size := llvm.ConstInt(b.uintptrType, b.targetData.TypeAllocSize(llvmKeyType), false)
		return b.createRuntimeCall("memequal", []llvm.Value{xPtr, yPtr, size}, "eq")
	case *types.Interface:
		// Interface: use runtime interface equality.
		size := llvm.ConstInt(b.uintptrType, b.targetData.TypeAllocSize(llvmKeyType), false)
		return b.createRuntimeCall("hashmapInterfaceEqual", []llvm.Value{xPtr, yPtr, size}, "eq")
	case *types.Struct:
		result := llvm.ConstInt(b.ctx.Int1Type(), 1, false) // start with true
		zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
		for i := 0; i < keyType.NumFields(); i++ {
			if keyType.Field(i).Name() == "_" {
				continue // blank fields are ignored in Go equality
			}
			fieldType := keyType.Field(i).Type()
			llvmFieldType := b.getLLVMType(fieldType)
			if b.targetData.TypeAllocSize(llvmFieldType) == 0 {
				continue // skip zero-sized fields
			}
			idx := llvm.ConstInt(b.ctx.Int32Type(), uint64(i), false)
			xFieldPtr := b.CreateInBoundsGEP(llvmKeyType, xPtr, []llvm.Value{zero, idx}, "")
			yFieldPtr := b.CreateInBoundsGEP(llvmKeyType, yPtr, []llvm.Value{zero, idx}, "")
			fieldEq := b.generateKeyEqual(fieldType, llvmFieldType, xFieldPtr, yFieldPtr, fn)
			result = b.CreateAnd(result, fieldEq, "")
		}
		return result
	case *types.Array:
		elemType := keyType.Elem()
		llvmElemType := b.getLLVMType(elemType)
		result := llvm.ConstInt(b.ctx.Int1Type(), 1, false)
		zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
		for i := 0; i < int(keyType.Len()); i++ {
			idx := llvm.ConstInt(b.uintptrType, uint64(i), false)
			xElemPtr := b.CreateInBoundsGEP(llvmKeyType, xPtr, []llvm.Value{zero, idx}, "")
			yElemPtr := b.CreateInBoundsGEP(llvmKeyType, yPtr, []llvm.Value{zero, idx}, "")
			elemEq := b.generateKeyEqual(elemType, llvmElemType, xElemPtr, yElemPtr, fn)
			result = b.CreateAnd(result, elemEq, "")
		}
		return result
	default:
		panic(fmt.Sprintf("unhandled key type for equal generation: %T", keyType))
	}
}
