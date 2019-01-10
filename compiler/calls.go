package compiler

import (
	"github.com/aykevl/go-llvm"
	"golang.org/x/tools/go/ssa"
)

// For a description of the calling convention in prose, see:
// https://tinygo.org/compiler-internals/calling-convention/

// The maximum number of arguments that can be expanded from a single struct. If
// a struct contains more fields, it is passed as a struct without expanding.
const MaxFieldsPerParam = 3

// Shortcut: create a call to runtime.<fnName> with the given arguments.
func (c *Compiler) createRuntimeCall(fnName string, args []llvm.Value, name string) llvm.Value {
	runtimePkg := c.ir.Program.ImportedPackage("runtime")
	member := runtimePkg.Members[fnName]
	if member == nil {
		panic("trying to call runtime." + fnName)
	}
	fn := c.ir.GetFunction(member.(*ssa.Function))
	if !fn.IsExported() {
		args = append(args, llvm.Undef(c.i8ptrType))            // unused context parameter
		args = append(args, llvm.ConstPointerNull(c.i8ptrType)) // coroutine handle
	}
	return c.createCall(fn.LLVMFn, args, name)
}

// Create a call to the given function with the arguments possibly expanded.
func (c *Compiler) createCall(fn llvm.Value, args []llvm.Value, name string) llvm.Value {
	expanded := make([]llvm.Value, 0, len(args))
	for _, arg := range args {
		fragments := c.expandFormalParam(arg)
		expanded = append(expanded, fragments...)
	}
	return c.builder.CreateCall(fn, expanded, name)
}

// Expand an argument type to a list that can be used in a function call
// paramter list.
func (c *Compiler) expandFormalParamType(t llvm.Type) []llvm.Type {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		fields := c.flattenAggregateType(t)
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

// Equivalent of expandFormalParamType for parameter values.
func (c *Compiler) expandFormalParam(v llvm.Value) []llvm.Value {
	switch v.Type().TypeKind() {
	case llvm.StructTypeKind:
		fieldTypes := c.flattenAggregateType(v.Type())
		if len(fieldTypes) <= MaxFieldsPerParam {
			fields := c.flattenAggregate(v)
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
func (c *Compiler) flattenAggregateType(t llvm.Type) []llvm.Type {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		fields := make([]llvm.Type, 0, t.StructElementTypesCount())
		for _, subfield := range t.StructElementTypes() {
			subfields := c.flattenAggregateType(subfield)
			fields = append(fields, subfields...)
		}
		return fields
	default:
		return []llvm.Type{t}
	}
}

// Break down a struct into its elementary types for argument passing. The value
// equivalent of flattenAggregateType
func (c *Compiler) flattenAggregate(v llvm.Value) []llvm.Value {
	switch v.Type().TypeKind() {
	case llvm.StructTypeKind:
		fields := make([]llvm.Value, 0, v.Type().StructElementTypesCount())
		for i := range v.Type().StructElementTypes() {
			subfield := c.builder.CreateExtractValue(v, i, "")
			subfields := c.flattenAggregate(subfield)
			fields = append(fields, subfields...)
		}
		return fields
	default:
		return []llvm.Value{v}
	}
}

// Collapse a list of fields into its original value.
func (c *Compiler) collapseFormalParam(t llvm.Type, fields []llvm.Value) llvm.Value {
	param, remaining := c.collapseFormalParamInternal(t, fields)
	if len(remaining) != 0 {
		panic("failed to expand back all fields")
	}
	return param
}

// Returns (value, remainingFields). Used by collapseFormalParam.
func (c *Compiler) collapseFormalParamInternal(t llvm.Type, fields []llvm.Value) (llvm.Value, []llvm.Value) {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		if len(c.flattenAggregateType(t)) <= MaxFieldsPerParam {
			value, err := c.getZeroValue(t)
			if err != nil {
				panic("could not get zero value of struct: " + err.Error())
			}
			for i, subtyp := range t.StructElementTypes() {
				structField, remaining := c.collapseFormalParamInternal(subtyp, fields)
				fields = remaining
				value = c.builder.CreateInsertValue(value, structField, i, "")
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
