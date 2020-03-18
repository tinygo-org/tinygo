package compiler

import (
	"tinygo.org/x/go-llvm"
)

// For a description of the calling convention in prose, see:
// https://tinygo.org/compiler-internals/calling-convention/

// The maximum number of arguments that can be expanded from a single struct. If
// a struct contains more fields, it is passed as a struct without expanding.
const MaxFieldsPerParam = 3

// createCall creates a new call to runtime.<fnName> with the given arguments.
func (b *builder) createRuntimeCall(fnName string, args []llvm.Value, name string) llvm.Value {
	fullName := "runtime." + fnName
	fn := b.mod.NamedFunction(fullName)
	if fn.IsNil() {
		panic("trying to call non-existent function: " + fullName)
	}
	args = append(args, llvm.Undef(b.i8ptrType))            // unused context parameter
	args = append(args, llvm.ConstPointerNull(b.i8ptrType)) // coroutine handle
	return b.createCall(fn, args, name)
}

// createCall creates a call to the given function with the arguments possibly
// expanded.
func (b *builder) createCall(fn llvm.Value, args []llvm.Value, name string) llvm.Value {
	expanded := make([]llvm.Value, 0, len(args))
	for _, arg := range args {
		fragments := b.expandFormalParam(arg)
		expanded = append(expanded, fragments...)
	}
	return b.CreateCall(fn, expanded, name)
}

// Expand an argument type to a list that can be used in a function call
// parameter list.
func expandFormalParamType(t llvm.Type) []llvm.Type {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		fields := flattenAggregateType(t)
		if len(fields) <= MaxFieldsPerParam {
			return fields
		} else {
			// failed to lower
			return []llvm.Type{t}
		}
	default:
		// TODO: split small arrays
		return []llvm.Type{t}
	}
}

// expandFormalParamOffsets returns a list of offsets from the start of an
// object of type t after it would have been split up by expandFormalParam. This
// is useful for debug information, where it is necessary to know the offset
// from the start of the combined object.
func (b *builder) expandFormalParamOffsets(t llvm.Type) []uint64 {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		fields := b.flattenAggregateTypeOffsets(t)
		if len(fields) <= MaxFieldsPerParam {
			return fields
		} else {
			// failed to lower
			return []uint64{0}
		}
	default:
		// TODO: split small arrays
		return []uint64{0}
	}
}

// expandFormalParam splits a formal param value into pieces, so it can be
// passed directly as part of a function call. For example, it splits up small
// structs into individual fields. It is the equivalent of expandFormalParamType
// for parameter values.
func (b *builder) expandFormalParam(v llvm.Value) []llvm.Value {
	switch v.Type().TypeKind() {
	case llvm.StructTypeKind:
		fieldTypes := flattenAggregateType(v.Type())
		if len(fieldTypes) <= MaxFieldsPerParam {
			fields := b.flattenAggregate(v)
			if len(fields) != len(fieldTypes) {
				panic("type and value param lowering don't match")
			}
			return fields
		} else {
			// failed to lower
			return []llvm.Value{v}
		}
	default:
		// TODO: split small arrays
		return []llvm.Value{v}
	}
}

// Try to flatten a struct type to a list of types. Returns a 1-element slice
// with the passed in type if this is not possible.
func flattenAggregateType(t llvm.Type) []llvm.Type {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		fields := make([]llvm.Type, 0, t.StructElementTypesCount())
		for _, subfield := range t.StructElementTypes() {
			subfields := flattenAggregateType(subfield)
			fields = append(fields, subfields...)
		}
		return fields
	default:
		return []llvm.Type{t}
	}
}

// flattenAggregateTypeOffset returns the offsets from the start of an object of
// type t if this object were flattened like in flattenAggregate. Used together
// with flattenAggregate to know the start indices of each value in the
// non-flattened object.
//
// Note: this is an implementation detail, use expandFormalParamOffsets instead.
func (c *compilerContext) flattenAggregateTypeOffsets(t llvm.Type) []uint64 {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		fields := make([]uint64, 0, t.StructElementTypesCount())
		for fieldIndex, field := range t.StructElementTypes() {
			suboffsets := c.flattenAggregateTypeOffsets(field)
			offset := c.targetData.ElementOffset(t, fieldIndex)
			for i := range suboffsets {
				suboffsets[i] += offset
			}
			fields = append(fields, suboffsets...)
		}
		return fields
	default:
		return []uint64{0}
	}
}

// flattenAggregate breaks down a struct into its elementary values for argument
// passing. It is the value equivalent of flattenAggregateType
func (b *builder) flattenAggregate(v llvm.Value) []llvm.Value {
	switch v.Type().TypeKind() {
	case llvm.StructTypeKind:
		fields := make([]llvm.Value, 0, v.Type().StructElementTypesCount())
		for i := range v.Type().StructElementTypes() {
			subfield := b.CreateExtractValue(v, i, "")
			subfields := b.flattenAggregate(subfield)
			fields = append(fields, subfields...)
		}
		return fields
	default:
		return []llvm.Value{v}
	}
}

// collapseFormalParam combines an aggregate object back into the original
// value. This is used to join multiple LLVM parameters into a single Go value
// in the function entry block.
func (b *builder) collapseFormalParam(t llvm.Type, fields []llvm.Value) llvm.Value {
	param, remaining := b.collapseFormalParamInternal(t, fields)
	if len(remaining) != 0 {
		panic("failed to expand back all fields")
	}
	return param
}

// collapseFormalParamInternal is an implementation detail of
// collapseFormalParam: it works by recursing until there are no fields left.
func (b *builder) collapseFormalParamInternal(t llvm.Type, fields []llvm.Value) (llvm.Value, []llvm.Value) {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		if len(flattenAggregateType(t)) <= MaxFieldsPerParam {
			value := llvm.ConstNull(t)
			for i, subtyp := range t.StructElementTypes() {
				structField, remaining := b.collapseFormalParamInternal(subtyp, fields)
				fields = remaining
				value = b.CreateInsertValue(value, structField, i, "")
			}
			return value, fields
		} else {
			// this struct was not flattened
			return fields[0], fields[1:]
		}
	default:
		return fields[0], fields[1:]
	}
}
