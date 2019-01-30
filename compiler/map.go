package compiler

// This file emits the correct map intrinsics for map operations.

import (
	"go/token"
	"go/types"

	"tinygo.org/x/go-llvm"
)

func (c *Compiler) emitMapLookup(keyType, valueType types.Type, m, key llvm.Value, commaOk bool, pos token.Pos) (llvm.Value, error) {
	llvmValueType, err := c.getLLVMType(valueType)
	if err != nil {
		return llvm.Value{}, err
	}
	mapValueAlloca := c.builder.CreateAlloca(llvmValueType, "hashmap.value")
	mapValuePtr := c.builder.CreateBitCast(mapValueAlloca, c.i8ptrType, "hashmap.valueptr")
	var commaOkValue llvm.Value
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key, mapValuePtr}
		commaOkValue = c.createRuntimeCall("hashmapStringGet", params, "")
	} else if hashmapIsBinaryKey(keyType) {
		// key can be compared with runtime.memequal
		keyAlloca := c.builder.CreateAlloca(key.Type(), "hashmap.key")
		c.builder.CreateStore(key, keyAlloca)
		keyPtr := c.builder.CreateBitCast(keyAlloca, c.i8ptrType, "hashmap.keyptr")
		params := []llvm.Value{m, keyPtr, mapValuePtr}
		commaOkValue = c.createRuntimeCall("hashmapBinaryGet", params, "")
	} else {
		return llvm.Value{}, c.makeError(pos, "only strings, bools, ints or structs of bools/ints are supported as map keys, but got: "+keyType.String())
	}
	mapValue := c.builder.CreateLoad(mapValueAlloca, "")
	if commaOk {
		tuple := llvm.Undef(c.ctx.StructType([]llvm.Type{llvmValueType, c.ctx.Int1Type()}, false))
		tuple = c.builder.CreateInsertValue(tuple, mapValue, 0, "")
		tuple = c.builder.CreateInsertValue(tuple, commaOkValue, 1, "")
		return tuple, nil
	} else {
		return mapValue, nil
	}
}

func (c *Compiler) emitMapUpdate(keyType types.Type, m, key, value llvm.Value, pos token.Pos) error {
	valueAlloca := c.builder.CreateAlloca(value.Type(), "hashmap.value")
	c.builder.CreateStore(value, valueAlloca)
	valuePtr := c.builder.CreateBitCast(valueAlloca, c.i8ptrType, "hashmap.valueptr")
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key, valuePtr}
		c.createRuntimeCall("hashmapStringSet", params, "")
		return nil
	} else if hashmapIsBinaryKey(keyType) {
		// key can be compared with runtime.memequal
		keyAlloca := c.builder.CreateAlloca(key.Type(), "hashmap.key")
		c.builder.CreateStore(key, keyAlloca)
		keyPtr := c.builder.CreateBitCast(keyAlloca, c.i8ptrType, "hashmap.keyptr")
		params := []llvm.Value{m, keyPtr, valuePtr}
		c.createRuntimeCall("hashmapBinarySet", params, "")
		return nil
	} else {
		return c.makeError(pos, "todo: map update key type: "+keyType.String())
	}
}

func (c *Compiler) emitMapDelete(keyType types.Type, m, key llvm.Value, pos token.Pos) error {
	keyType = keyType.Underlying()
	if t, ok := keyType.(*types.Basic); ok && t.Info()&types.IsString != 0 {
		// key is a string
		params := []llvm.Value{m, key}
		c.createRuntimeCall("hashmapStringDelete", params, "")
		return nil
	} else if hashmapIsBinaryKey(keyType) {
		keyAlloca := c.builder.CreateAlloca(key.Type(), "hashmap.key")
		c.builder.CreateStore(key, keyAlloca)
		keyPtr := c.builder.CreateBitCast(keyAlloca, c.i8ptrType, "hashmap.keyptr")
		params := []llvm.Value{m, keyPtr}
		c.createRuntimeCall("hashmapBinaryDelete", params, "")
		return nil
	} else {
		return c.makeError(pos, "todo: map delete key type: "+keyType.String())
	}
}

// Get FNV-1a hash of this string.
//
// https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function#FNV-1a_hash
func hashmapHash(data []byte) uint32 {
	var result uint32 = 2166136261 // FNV offset basis
	for _, c := range data {
		result ^= uint32(c)
		result *= 16777619 // FNV prime
	}
	return result
}

// Get the topmost 8 bits of the hash, without using a special value (like 0).
func hashmapTopHash(hash uint32) uint8 {
	tophash := uint8(hash >> 24)
	if tophash < 1 {
		// 0 means empty slot, so make it bigger.
		tophash += 1
	}
	return tophash
}

// Returns true if this key type does not contain strings, interfaces etc., so
// can be compared with runtime.memequal.
func hashmapIsBinaryKey(keyType types.Type) bool {
	switch keyType := keyType.(type) {
	case *types.Basic:
		return keyType.Info()&(types.IsBoolean|types.IsInteger) != 0
	case *types.Struct:
		for i := 0; i < keyType.NumFields(); i++ {
			fieldType := keyType.Field(i).Type().Underlying()
			if !hashmapIsBinaryKey(fieldType) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
