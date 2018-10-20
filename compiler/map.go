package compiler

// This file emits the correct map intrinsics for map operations.

import (
	"errors"
	"go/types"

	"github.com/aykevl/go-llvm"
)

func (c *Compiler) emitMapLookup(keyType, valueType types.Type, m, key llvm.Value) (llvm.Value, error) {
	switch keyType := keyType.Underlying().(type) {
	case *types.Basic:
		llvmValueType, err := c.getLLVMType(valueType)
		if err != nil {
			return llvm.Value{}, err
		}
		mapValueAlloca := c.builder.CreateAlloca(llvmValueType, "hashmap.value")
		mapValuePtr := c.builder.CreateBitCast(mapValueAlloca, c.i8ptrType, "hashmap.valueptr")
		if keyType.Info()&types.IsString != 0 {
			params := []llvm.Value{m, key, mapValuePtr}
			c.createRuntimeCall("hashmapStringGet", params, "")
			return c.builder.CreateLoad(mapValueAlloca, ""), nil
		} else if keyType.Info()&(types.IsBoolean|types.IsInteger) != 0 {
			keyAlloca := c.builder.CreateAlloca(key.Type(), "hashmap.key")
			c.builder.CreateStore(key, keyAlloca)
			keyPtr := c.builder.CreateBitCast(keyAlloca, c.i8ptrType, "hashmap.keyptr")
			params := []llvm.Value{m, keyPtr, mapValuePtr}
			c.createRuntimeCall("hashmapBinaryGet", params, "")
			return c.builder.CreateLoad(mapValueAlloca, ""), nil
		} else {
			return llvm.Value{}, errors.New("todo: map lookup key type: " + keyType.String())
		}
	default:
		return llvm.Value{}, errors.New("todo: map lookup key type: " + keyType.String())
	}
}

func (c *Compiler) emitMapUpdate(keyType types.Type, m, key, value llvm.Value) error {
	switch keyType := keyType.Underlying().(type) {
	case *types.Basic:
		valueAlloca := c.builder.CreateAlloca(value.Type(), "hashmap.value")
		c.builder.CreateStore(value, valueAlloca)
		valuePtr := c.builder.CreateBitCast(valueAlloca, c.i8ptrType, "hashmap.valueptr")
		if keyType.Info()&types.IsString != 0 {
			params := []llvm.Value{m, key, valuePtr}
			c.createRuntimeCall("hashmapStringSet", params, "")
			return nil
		} else if keyType.Info()&(types.IsBoolean|types.IsInteger) != 0 {
			keyAlloca := c.builder.CreateAlloca(key.Type(), "hashmap.key")
			c.builder.CreateStore(key, keyAlloca)
			keyPtr := c.builder.CreateBitCast(keyAlloca, c.i8ptrType, "hashmap.keyptr")
			params := []llvm.Value{m, keyPtr, valuePtr}
			c.createRuntimeCall("hashmapBinarySet", params, "")
			return nil
		} else {
			return errors.New("todo: map update key type: " + keyType.String())
		}
	default:
		return errors.New("todo: map update key type: " + keyType.String())
	}
}

func (c *Compiler) emitMapDelete(keyType types.Type, m, key llvm.Value) error {
	switch keyType := keyType.Underlying().(type) {
	case *types.Basic:
		if keyType.Info()&types.IsString != 0 {
			params := []llvm.Value{m, key}
			c.createRuntimeCall("hashmapStringDelete", params, "")
			return nil
		} else if keyType.Info()&(types.IsBoolean|types.IsInteger) != 0 {
			keyAlloca := c.builder.CreateAlloca(key.Type(), "hashmap.key")
			c.builder.CreateStore(key, keyAlloca)
			keyPtr := c.builder.CreateBitCast(keyAlloca, c.i8ptrType, "hashmap.keyptr")
			params := []llvm.Value{m, keyPtr}
			c.createRuntimeCall("hashmapBinaryDelete", params, "")
			return nil
		} else {
			return errors.New("todo: map lookup key type: " + keyType.String())
		}
	default:
		return errors.New("todo: map delete key type: " + keyType.String())
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
