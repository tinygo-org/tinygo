package compiler

// This file emits the correct map intrinsics for map operations.

import (
	"fmt"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

const hashArrayUnrollLimit = 4

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

	// Resolve hash and equal functions for this key type. For string and
	// binary key types, reference the corresponding runtime functions
	// directly. For composite types, generate type-specific functions.
	var hashFn, equalFn llvm.Value
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		hashFn = b.getRuntimeFunctionValue("hashmapStringPtrHash", hashmapKeyHashSignature())
		equalFn = b.getRuntimeFunctionValue("hashmapStringEqual", hashmapKeyEqualSignature())
	} else if hashmapIsBinaryKey(keyType) {
		hashFn = b.getRuntimeFunctionValue("hash32", hashmapKeyHashSignature())
		equalFn = b.getRuntimeFunctionValue("memequal", hashmapKeyEqualSignature())
	} else {
		fn := b.getOrGenerateKeyHashFunc(keyType)
		hashFn = b.createFuncValue(fn, llvm.ConstNull(b.dataPtrType), hashmapKeyHashSignature())
		fn = b.getOrGenerateKeyEqualFunc(keyType)
		equalFn = b.createFuncValue(fn, llvm.ConstNull(b.dataPtrType), hashmapKeyEqualSignature())
	}

	hashmap := b.createRuntimeCall("hashmapMakeGeneric", []llvm.Value{
		llvmKeySize, llvmValueSize, sizeHint,
		hashFn, equalFn,
	}, "")
	return hashmap, nil
}

// getRuntimeFunctionValue returns a TinyGo function value (with nil context)
// for the named runtime function.
func (b *builder) getRuntimeFunctionValue(name string, sig *types.Signature) llvm.Value {
	member := b.program.ImportedPackage("runtime").Members[name]
	if member == nil {
		panic("unknown runtime function: " + name)
	}
	_, llvmFn := b.getFunction(member.(*ssa.Function))
	return b.createFuncValue(llvmFn, llvm.ConstNull(b.dataPtrType), sig)
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
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key, mapValueAlloca, mapValueSize}
		commaOkValue = b.createRuntimeCall("hashmapStringGet", params, "")
	} else {
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
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key, valueAlloca}
		b.createRuntimeInvoke("hashmapStringSet", params, "")
	} else {
		// Key stored at actual type.
		keyAlloca, keySize := b.createTemporaryAlloca(key.Type(), "hashmap.key")
		b.CreateStore(key, keyAlloca)
		fnName := "hashmapBinarySet"
		if !hashmapIsBinaryKey(keyType) {
			fnName = "hashmapGenericSet"
		}
		params := []llvm.Value{m, keyAlloca, valueAlloca}
		b.createRuntimeInvoke(fnName, params, "")
		b.emitLifetimeEnd(keyAlloca, keySize)
	}
	b.emitLifetimeEnd(valueAlloca, valueSize)
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
		fnName := "hashmapBinaryDelete"
		if !hashmapIsBinaryKey(keyType) {
			fnName = "hashmapGenericDelete"
		}
		params := []llvm.Value{m, keyAlloca}
		b.createRuntimeCall(fnName, params, "")
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
		return keyType.Info()&(types.IsBoolean|types.IsInteger) != 0 || keyType.Kind() == types.UnsafePointer
	case *types.Pointer:
		return true
	case *types.Array:
		return hashmapIsBinaryKey(keyType.Elem())
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
		arrayLen := keyType.Len()
		if hashmapIsBinaryKey(elemType) {
			// All elements are binary-comparable; hash the entire array as raw bytes.
			size := llvm.ConstInt(b.uintptrType, b.targetData.TypeAllocSize(llvmKeyType), false)
			return b.createRuntimeCall("hash32", []llvm.Value{keyPtr, size, seed}, "hash")
		}
		if arrayLen == 0 {
			return llvm.ConstInt(b.ctx.Int32Type(), 0, false)
		}
		if arrayLen <= hashArrayUnrollLimit {
			hash := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := 0; i < int(arrayLen); i++ {
				idx := llvm.ConstInt(b.uintptrType, uint64(i), false)
				elemPtr := b.CreateInBoundsGEP(llvmKeyType, keyPtr, []llvm.Value{zero, idx}, "")
				elemHash := b.generateKeyHash(elemType, llvmElemType, elemPtr, seed)
				hash = b.CreateXor(hash, elemHash, "")
			}
			return hash
		}
		initHash := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
		zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)

		loopEntry := b.GetInsertBlock()
		loopBody := b.ctx.AddBasicBlock(loopEntry.Parent(), "hash.array.body")
		loopDone := b.ctx.AddBasicBlock(loopEntry.Parent(), "hash.array.done")

		b.CreateBr(loopBody)
		b.SetInsertPointAtEnd(loopBody)

		phiI := b.CreatePHI(b.uintptrType, "i")
		phiHash := b.CreatePHI(b.ctx.Int32Type(), "hash.acc")

		elemPtr := b.CreateInBoundsGEP(llvmKeyType, keyPtr, []llvm.Value{zero, phiI}, "")
		elemHash := b.generateKeyHash(elemType, llvmElemType, elemPtr, seed)
		newHash := b.CreateXor(phiHash, elemHash, "")
		nextI := b.CreateAdd(phiI, llvm.ConstInt(b.uintptrType, 1, false), "")
		cond := b.CreateICmp(llvm.IntULT, nextI, llvm.ConstInt(b.uintptrType, uint64(arrayLen), false), "")
		b.CreateCondBr(cond, loopBody, loopDone)

		bodyEnd := b.GetInsertBlock()
		phiI.AddIncoming([]llvm.Value{llvm.ConstInt(b.uintptrType, 0, false), nextI},
			[]llvm.BasicBlock{loopEntry, bodyEnd})
		phiHash.AddIncoming([]llvm.Value{initHash, newHash},
			[]llvm.BasicBlock{loopEntry, bodyEnd})

		b.SetInsertPointAtEnd(loopDone)
		return newHash
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
		arrayLen := keyType.Len()
		if hashmapIsBinaryKey(elemType) {
			// All elements are binary-comparable; compare the entire array.
			size := llvm.ConstInt(b.uintptrType, b.targetData.TypeAllocSize(llvmKeyType), false)
			return b.createRuntimeCall("memequal", []llvm.Value{xPtr, yPtr, size}, "eq")
		}
		if arrayLen == 0 {
			return llvm.ConstInt(b.ctx.Int1Type(), 1, false)
		}
		if arrayLen <= hashArrayUnrollLimit {
			result := llvm.ConstInt(b.ctx.Int1Type(), 1, false)
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := 0; i < int(arrayLen); i++ {
				idx := llvm.ConstInt(b.uintptrType, uint64(i), false)
				xElemPtr := b.CreateInBoundsGEP(llvmKeyType, xPtr, []llvm.Value{zero, idx}, "")
				yElemPtr := b.CreateInBoundsGEP(llvmKeyType, yPtr, []llvm.Value{zero, idx}, "")
				elemEq := b.generateKeyEqual(elemType, llvmElemType, xElemPtr, yElemPtr, fn)
				result = b.CreateAnd(result, elemEq, "")
			}
			return result
		}
		zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)

		loopEntry := b.GetInsertBlock()
		loopBody := b.ctx.AddBasicBlock(loopEntry.Parent(), "eq.array.body")
		loopDone := b.ctx.AddBasicBlock(loopEntry.Parent(), "eq.array.done")

		b.CreateBr(loopBody)
		b.SetInsertPointAtEnd(loopBody)

		phiI := b.CreatePHI(b.uintptrType, "i")

		xElemPtr := b.CreateInBoundsGEP(llvmKeyType, xPtr, []llvm.Value{zero, phiI}, "")
		yElemPtr := b.CreateInBoundsGEP(llvmKeyType, yPtr, []llvm.Value{zero, phiI}, "")
		elemEq := b.generateKeyEqual(elemType, llvmElemType, xElemPtr, yElemPtr, fn)

		nextI := b.CreateAdd(phiI, llvm.ConstInt(b.uintptrType, 1, false), "")
		atEnd := b.CreateICmp(llvm.IntUGE, nextI, llvm.ConstInt(b.uintptrType, uint64(arrayLen), false), "")
		exitLoop := b.CreateOr(atEnd, b.CreateNot(elemEq, ""), "")
		b.CreateCondBr(exitLoop, loopDone, loopBody)

		bodyEnd := b.GetInsertBlock()
		phiI.AddIncoming([]llvm.Value{llvm.ConstInt(b.uintptrType, 0, false), nextI},
			[]llvm.BasicBlock{loopEntry, bodyEnd})

		b.SetInsertPointAtEnd(loopDone)
		return elemEq
	default:
		panic(fmt.Sprintf("unhandled key type for equal generation: %T", keyType))
	}
}
