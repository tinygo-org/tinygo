package compiler

import (
	"go/types"

	"github.com/aykevl/llvm/bindings/go/llvm"
	"github.com/aykevl/tinygo/ir"
	"golang.org/x/tools/go/ssa"
)

// This file implements the calling convention used by Go.
// The calling convention is like the C calling convention (or, whatever LLVM
// makes of it) with the following modifications:
//   * Struct parameters are fully expanded to individual fields (recursively),
//     when the number of fields (combined) is 3 or less.
//     Examples:
//       {i8*, i32}         -> i8*, i32
//       {{i8*, i32}, i16}  -> i8*, i32, i16
//       {{i64}}            -> i64
//       {i8*, i32, i8, i8} -> {i8*, i32, i8, i8}
//     Note that all native Go data types that don't exist in LLVM (string,
//     slice, interface, fat function pointer) can be expanded this way, making
//     the work of LLVM optimizers easier.
//   * Closures have an extra paramter appended at the end of the argument list,
//     which is a pointer to a struct containing free variables.
//   * Blocking functions have a coroutine pointer prepended to the argument
//     list, see src/runtime/scheduler.go for details.

const MaxFieldsPerParam = 3

func (c *Compiler) createRuntimeCall(fnName string, args []llvm.Value, name string) llvm.Value {
	runtimePkg := c.ir.Program.ImportedPackage("runtime")
	member := runtimePkg.Members[fnName]
	if member == nil {
		panic("trying to call runtime." + fnName)
	}
	fn := c.ir.GetFunction(member.(*ssa.Function))
	return c.createCall(fn, args, name)
}

func (c *Compiler) createCall(fn *ir.Function, args []llvm.Value, name string) llvm.Value {
	return c.createIndirectCall(fn.Signature, fn.LLVMFn, args, name)
}

func (c *Compiler) createIndirectCall(sig *types.Signature, fn llvm.Value, args []llvm.Value, name string) llvm.Value {
	expanded := make([]llvm.Value, 0, len(args))
	for _, arg := range args {
		fragments := c.expandFormalParam(arg)
		expanded = append(expanded, fragments...)
	}
	return c.builder.CreateCall(fn, expanded, name)
}

func (c *Compiler) getLLVMParamTypes(t types.Type) ([]llvm.Type, error) {
	llvmType, err := c.getLLVMType(t)
	if err != nil {
		return nil, err
	}
	return c.expandFormalParamType(llvmType), nil
}

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

// Convert an argument to one that can be passed in a parameter.
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

func (c *Compiler) flattenAggregateType(t llvm.Type) []llvm.Type {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		fields := make([]llvm.Type, 0, len(t.Subtypes()))
		for _, subfield := range t.Subtypes() {
			subfields := c.flattenAggregateType(subfield)
			fields = append(fields, subfields...)
		}
		return fields
	default:
		return []llvm.Type{t}
	}
}

// Break down a struct into its elementary types for argument passing.
func (c *Compiler) flattenAggregate(v llvm.Value) []llvm.Value {
	switch v.Type().TypeKind() {
	case llvm.StructTypeKind:
		fields := make([]llvm.Value, 0, len(v.Type().Subtypes()))
		for i := range v.Type().Subtypes() {
			subfield := c.builder.CreateExtractValue(v, i, "")
			subfields := c.flattenAggregate(subfield)
			fields = append(fields, subfields...)
		}
		return fields
	default:
		return []llvm.Value{v}
	}
}

func (c *Compiler) collapseFormalParam(t llvm.Type, fields []llvm.Value) llvm.Value {
	param, remaining := c.collapseFormalParamInternal(t, fields)
	if len(remaining) != 0 {
		panic("failed to expand back all fields")
	}
	return param
}

// Returns (value, remainingFields).
func (c *Compiler) collapseFormalParamInternal(t llvm.Type, fields []llvm.Value) (llvm.Value, []llvm.Value) {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		if len(c.flattenAggregateType(t)) <= MaxFieldsPerParam {
			value, err := getZeroValue(t)
			if err != nil {
				panic("could not get zero value of struct: " + err.Error())
			}
			for i, subtyp := range t.Subtypes() {
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
