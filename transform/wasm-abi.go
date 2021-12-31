package transform

import (
	"errors"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"tinygo.org/x/go-llvm"
)

// ExternalInt64AsPtr converts i64 parameters in externally-visible functions to
// values passed by reference (*i64), to work around the lack of 64-bit integers
// in JavaScript (commonly used together with WebAssembly). Once that's
// resolved, this pass may be avoided. For more details:
// https://github.com/WebAssembly/design/issues/1172
//
// This pass can be enabled/disabled with the -wasm-abi flag, and is enabled by
// default as of december 2019.
func ExternalInt64AsPtr(mod llvm.Module, config *compileopts.Config) error {
	ctx := mod.Context()
	builder := ctx.NewBuilder()
	defer builder.Dispose()
	int64Type := ctx.Int64Type()
	int64PtrType := llvm.PointerType(int64Type, 0)

	// This builder is only used for creating new allocas in the entry block of
	// a function, avoiding many SetInsertPoint* calls.
	entryBlockBuilder := ctx.NewBuilder()
	defer entryBlockBuilder.Dispose()

	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if fn.Linkage() != llvm.ExternalLinkage {
			// Only change externally visible functions (exports and imports).
			continue
		}
		if strings.HasPrefix(fn.Name(), "llvm.") || strings.HasPrefix(fn.Name(), "runtime.") {
			// Do not try to modify the signature of internal LLVM functions and
			// assume that runtime functions are only temporarily exported for
			// transforms.
			continue
		}
		if !fn.GetStringAttributeAtIndex(-1, "tinygo-methods").IsNil() {
			// These are internal functions (interface method call, interface
			// type assert) that will be lowered by the interface lowering pass.
			// Don't transform them.
			continue
		}

		hasInt64 := false
		paramTypes := []llvm.Type{}

		// Check return type for 64-bit integer.
		fnType := fn.Type().ElementType()
		returnType := fnType.ReturnType()
		if returnType == int64Type {
			hasInt64 = true
			paramTypes = append(paramTypes, int64PtrType)
			returnType = ctx.VoidType()
		}

		// Check param types for 64-bit integers.
		for param := fn.FirstParam(); !param.IsNil(); param = llvm.NextParam(param) {
			if param.Type() == int64Type {
				hasInt64 = true
				paramTypes = append(paramTypes, int64PtrType)
			} else {
				paramTypes = append(paramTypes, param.Type())
			}
		}

		if !hasInt64 {
			// No i64 in the paramter list.
			continue
		}

		// Add $i64wrapper to the real function name as it is only used
		// internally.
		// Add a new function with the correct signature that is exported.
		name := fn.Name()
		fn.SetName(name + "$i64wrap")
		externalFnType := llvm.FunctionType(returnType, paramTypes, fnType.IsFunctionVarArg())
		externalFn := llvm.AddFunction(mod, name, externalFnType)
		AddStandardAttributes(fn, config)

		if fn.IsDeclaration() {
			// Just a declaration: the definition doesn't exist on the Go side
			// so it cannot be called from external code.
			// Update all users to call the external function.
			// The old $i64wrapper function could be removed, but it may as well
			// be left in place.
			for use := fn.FirstUse(); !use.IsNil(); use = use.NextUse() {
				call := use.User()
				entryBlockBuilder.SetInsertPointBefore(call.InstructionParent().Parent().EntryBasicBlock().FirstInstruction())
				builder.SetInsertPointBefore(call)
				callParams := []llvm.Value{}
				var retvalAlloca llvm.Value
				if fnType.ReturnType() == int64Type {
					retvalAlloca = entryBlockBuilder.CreateAlloca(int64Type, "i64asptr")
					callParams = append(callParams, retvalAlloca)
				}
				for i := 0; i < call.OperandsCount()-1; i++ {
					operand := call.Operand(i)
					if operand.Type() == int64Type {
						// Pass a stack-allocated pointer instead of the value
						// itself.
						alloca := entryBlockBuilder.CreateAlloca(int64Type, "i64asptr")
						builder.CreateStore(operand, alloca)
						callParams = append(callParams, alloca)
					} else {
						// Unchanged parameter.
						callParams = append(callParams, operand)
					}
				}
				var callName string
				if returnType.TypeKind() != llvm.VoidTypeKind {
					// Only use the name of the old call instruction if the new
					// call is not a void call.
					// A call instruction with an i64 return type may have had a
					// name, but it cannot have a name after this transform
					// because the return type will now be void.
					callName = call.Name()
				}
				if fnType.ReturnType() == int64Type {
					// Pass a stack-allocated pointer as the first parameter
					// where the return value should be stored, instead of using
					// the regular return value.
					builder.CreateCall(externalFn, callParams, callName)
					returnValue := builder.CreateLoad(retvalAlloca, "retval")
					call.ReplaceAllUsesWith(returnValue)
					call.EraseFromParentAsInstruction()
				} else {
					newCall := builder.CreateCall(externalFn, callParams, callName)
					call.ReplaceAllUsesWith(newCall)
					call.EraseFromParentAsInstruction()
				}
			}
		} else {
			// The function has a definition in Go. This means that it may still
			// be called both Go and from external code.
			// Keep existing calls with the existing convention in place (for
			// better performance), but export a new wrapper function with the
			// correct calling convention.
			fn.SetLinkage(llvm.InternalLinkage)
			fn.SetUnnamedAddr(true)
			entryBlock := ctx.AddBasicBlock(externalFn, "entry")
			builder.SetInsertPointAtEnd(entryBlock)
			var callParams []llvm.Value
			if fnType.ReturnType() == int64Type {
				return errors.New("not yet implemented: exported function returns i64 with -wasm-abi=js; " +
					"see https://tinygo.org/compiler-internals/calling-convention/")
			}
			for i, origParam := range fn.Params() {
				paramValue := externalFn.Param(i)
				if origParam.Type() == int64Type {
					paramValue = builder.CreateLoad(paramValue, "i64")
				}
				callParams = append(callParams, paramValue)
			}
			retval := builder.CreateCall(fn, callParams, "")
			if retval.Type().TypeKind() == llvm.VoidTypeKind {
				builder.CreateRetVoid()
			} else {
				builder.CreateRet(retval)
			}
		}
	}

	return nil
}
