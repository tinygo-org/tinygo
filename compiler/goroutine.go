package compiler

// This file implements the 'go' keyword to start a new goroutine. See
// goroutine-lowering.go for more details.

import (
	"go/token"
	"go/types"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// createGo emits code to start a new goroutine.
func (b *builder) createGo(instr *ssa.Go) {
	if builtin, ok := instr.Call.Value.(*ssa.Builtin); ok {
		// We cheat. None of the builtins do any long or blocking operation, so
		// we might as well run these builtins right away without the program
		// noticing the difference.
		// Possible exceptions:
		//   - copy: this is a possibly long operation, but not a blocking
		//     operation. Semantically it makes no difference to run it right
		//     away (not in a goroutine). However, in practice it makes no sense
		//     to run copy in a goroutine as there is no way to (safely) know
		//     when it is finished.
		//   - panic: the error message would appear in the parent goroutine.
		//     But because `go panic("err")` would halt the program anyway
		//     (there is no recover), panicking right away would give the same
		//     behavior as creating a goroutine, switching the scheduler to that
		//     goroutine, and panicking there. So this optimization seems
		//     correct.
		//   - recover: because it runs in a new goroutine, it is never a
		//     deferred function. Thus this is a no-op.
		if builtin.Name() == "recover" {
			// This is a no-op, even in a deferred function:
			//   go recover()
			return
		}
		var argTypes []types.Type
		var argValues []llvm.Value
		for _, arg := range instr.Call.Args {
			argTypes = append(argTypes, arg.Type())
			argValues = append(argValues, b.getValue(arg, getPos(instr)))
		}
		b.createBuiltin(argTypes, argValues, builtin.Name(), instr.Pos())
		return
	}

	// Get all function parameters to pass to the goroutine.
	var params []llvm.Value
	for _, param := range instr.Call.Args {
		params = append(params, b.expandFormalParam(b.getValue(param, getPos(instr)))...)
	}

	var prefix string
	var funcPtr llvm.Value
	var funcType llvm.Type
	hasContext := false
	if callee := instr.Call.StaticCallee(); callee != nil {
		// Static callee is known. This makes it easier to start a new
		// goroutine.
		var context llvm.Value
		switch value := instr.Call.Value.(type) {
		case *ssa.Function:
			// Goroutine call is regular function call. No context is necessary.
		case *ssa.MakeClosure:
			// A goroutine call on a func value, but the callee is trivial to find. For
			// example: immediately applied functions.
			funcValue := b.getValue(value, getPos(instr))
			context = b.extractFuncContext(funcValue)
		default:
			panic("StaticCallee returned an unexpected value")
		}
		if !context.IsNil() {
			params = append(params, context) // context parameter
			hasContext = true
		}
		funcType, funcPtr = b.getFunction(callee)
	} else if instr.Call.IsInvoke() {
		// This is a method call on an interface value.
		itf := b.getValue(instr.Call.Value, getPos(instr))
		itfTypeCode := b.CreateExtractValue(itf, 0, "")
		itfValue := b.CreateExtractValue(itf, 1, "")
		funcPtr = b.getInvokeFunction(&instr.Call)
		funcType = funcPtr.GlobalValueType()
		params = append([]llvm.Value{itfValue}, params...) // start with receiver
		params = append(params, itfTypeCode)               // end with typecode
	} else {
		// This is a function pointer.
		// At the moment, two extra params are passed to the newly started
		// goroutine:
		//   * The function context, for closures.
		//   * The function pointer (for tasks).
		var context llvm.Value
		funcPtr, context = b.decodeFuncValue(b.getValue(instr.Call.Value, getPos(instr)))
		funcType = b.getLLVMFunctionType(instr.Call.Value.Type().Underlying().(*types.Signature))
		params = append(params, context, funcPtr)
		hasContext = true
		prefix = b.fn.RelString(nil)
	}

	paramBundle := b.emitPointerPack(params)
	var stackSize llvm.Value
	callee := b.createGoroutineStartWrapper(funcType, funcPtr, prefix, hasContext, false, instr.Pos())
	if b.AutomaticStackSize {
		// The stack size is not known until after linking. Call a dummy
		// function that will be replaced with a load from a special ELF
		// section that contains the stack size (and is modified after
		// linking).
		stackSizeFnType, stackSizeFn := b.getFunction(b.program.ImportedPackage("internal/task").Members["getGoroutineStackSize"].(*ssa.Function))
		stackSize = b.createCall(stackSizeFnType, stackSizeFn, []llvm.Value{callee, llvm.Undef(b.dataPtrType)}, "stacksize")
	} else {
		// The stack size is fixed at compile time. By emitting it here as a
		// constant, it can be optimized.
		if (b.Scheduler == "tasks" || b.Scheduler == "asyncify") && b.DefaultStackSize == 0 {
			b.addError(instr.Pos(), "default stack size for goroutines is not set")
		}
		stackSize = llvm.ConstInt(b.uintptrType, b.DefaultStackSize, false)
	}
	fnType, start := b.getFunction(b.program.ImportedPackage("internal/task").Members["start"].(*ssa.Function))
	b.createCall(fnType, start, []llvm.Value{callee, paramBundle, stackSize, llvm.Undef(b.dataPtrType)}, "")
}

// Create an exported wrapper function for functions with the //go:wasmexport
// pragma. This wrapper function is quite complex when the scheduler is enabled:
// it needs to start a new goroutine each time the exported function is called.
func (b *builder) createWasmExport() {
	pos := b.info.wasmExportPos
	if b.info.exported {
		// //export really shouldn't be used anymore when //go:wasmexport is
		// available, because //go:wasmexport is much better defined.
		b.addError(pos, "cannot use //export and //go:wasmexport at the same time")
		return
	}

	const suffix = "#wasmexport"

	// Declare the exported function.
	paramTypes := b.llvmFnType.ParamTypes()
	exportedFnType := llvm.FunctionType(b.llvmFnType.ReturnType(), paramTypes[:len(paramTypes)-1], false)
	exportedFn := llvm.AddFunction(b.mod, b.fn.RelString(nil)+suffix, exportedFnType)
	b.addStandardAttributes(exportedFn)
	llvmutil.AppendToGlobal(b.mod, "llvm.used", exportedFn)
	exportedFn.AddFunctionAttr(b.ctx.CreateStringAttribute("wasm-export-name", b.info.wasmExport))

	// Create a builder for this wrapper function.
	builder := newBuilder(b.compilerContext, b.ctx.NewBuilder(), b.fn)
	defer builder.Dispose()

	// Define this function as a separate function in DWARF
	if b.Debug {
		if b.fn.Syntax() != nil {
			// Create debug info file if needed.
			pos := b.program.Fset.Position(pos)
			builder.difunc = builder.attachDebugInfoRaw(b.fn, exportedFn, suffix, pos.Filename, pos.Line)
		}
		builder.setDebugLocation(pos)
	}

	// Create a single basic block inside of it.
	bb := llvm.AddBasicBlock(exportedFn, "entry")
	builder.SetInsertPointAtEnd(bb)

	// Insert an assertion to make sure this //go:wasmexport function is not
	// called at a time when it is not allowed (for example, before the runtime
	// is initialized).
	builder.createRuntimeCall("wasmExportCheckRun", nil, "")

	if b.Scheduler == "none" {
		// When the scheduler has been disabled, this is really trivial: just
		// call the function.
		params := exportedFn.Params()
		params = append(params, llvm.ConstNull(b.dataPtrType)) // context parameter
		retval := builder.CreateCall(b.llvmFnType, b.llvmFn, params, "")
		if b.fn.Signature.Results() == nil {
			builder.CreateRetVoid()
		} else {
			builder.CreateRet(retval)
		}

	} else {
		// The scheduler is enabled, so we need to start a new goroutine, wait
		// for it to complete, and read the result value.

		// Build a function that looks like this:
		//
		//   func foo#wasmexport(param0, param1, ..., paramN) {
		//       var state *stateStruct
		//
		//       // 'done' must be explicitly initialized ('state' is not zeroed)
		//       state.done = false
		//
		//       // store the parameters in the state object
		//       state.param0 = param0
		//       state.param1 = param1
		//       ...
		//       state.paramN = paramN
		//
		//       // create a goroutine and push it to the runqueue
		//       task.start(uintptr(gowrapper), &state)
		//
		//       // run the scheduler
		//       runtime.wasmExportRun(&state.done)
		//
		//       // if there is a return value, load it and return
		//       return state.result
		//   }

		hasReturn := b.fn.Signature.Results() != nil

		// Build the state struct type.
		// It stores the function parameters, the 'done' flag, and reserves
		// space for a return value if needed.
		stateFields := exportedFnType.ParamTypes()
		numParams := len(stateFields)
		stateFields = append(stateFields, b.ctx.Int1Type()) // 'done' field
		if hasReturn {
			stateFields = append(stateFields, b.llvmFnType.ReturnType())
		}
		stateStruct := b.ctx.StructType(stateFields, false)

		// Allocate the state struct on the stack.
		statePtr := builder.CreateAlloca(stateStruct, "status")

		// Initialize the 'done' field.
		doneGEP := builder.CreateInBoundsGEP(stateStruct, statePtr, []llvm.Value{
			llvm.ConstInt(b.ctx.Int32Type(), 0, false),
			llvm.ConstInt(b.ctx.Int32Type(), uint64(numParams), false),
		}, "done.gep")
		builder.CreateStore(llvm.ConstNull(b.ctx.Int1Type()), doneGEP)

		// Store all parameters in the state object.
		for i, param := range exportedFn.Params() {
			gep := builder.CreateInBoundsGEP(stateStruct, statePtr, []llvm.Value{
				llvm.ConstInt(b.ctx.Int32Type(), 0, false),
				llvm.ConstInt(b.ctx.Int32Type(), uint64(i), false),
			}, "")
			builder.CreateStore(param, gep)
		}

		// Create a new goroutine and add it to the runqueue.
		wrapper := b.createGoroutineStartWrapper(b.llvmFnType, b.llvmFn, "", false, true, pos)
		stackSize := llvm.ConstInt(b.uintptrType, b.DefaultStackSize, false)
		taskStartFnType, taskStartFn := builder.getFunction(b.program.ImportedPackage("internal/task").Members["start"].(*ssa.Function))
		builder.createCall(taskStartFnType, taskStartFn, []llvm.Value{wrapper, statePtr, stackSize, llvm.Undef(b.dataPtrType)}, "")

		// Run the scheduler.
		builder.createRuntimeCall("wasmExportRun", []llvm.Value{doneGEP}, "")

		// Read the return value (if any) and return to the caller of the
		// //go:wasmexport function.
		if hasReturn {
			gep := builder.CreateInBoundsGEP(stateStruct, statePtr, []llvm.Value{
				llvm.ConstInt(b.ctx.Int32Type(), 0, false),
				llvm.ConstInt(b.ctx.Int32Type(), uint64(numParams)+1, false),
			}, "")
			retval := builder.CreateLoad(b.llvmFnType.ReturnType(), gep, "retval")
			builder.CreateRet(retval)
		} else {
			builder.CreateRetVoid()
		}
	}
}

// createGoroutineStartWrapper creates a wrapper for the task-based
// implementation of goroutines. For example, to call a function like this:
//
//	func add(x, y int) int { ... }
//
// It creates a wrapper like this:
//
//	func add$gowrapper(ptr *unsafe.Pointer) {
//	    args := (*struct{
//	        x, y int
//	    })(ptr)
//	    add(args.x, args.y)
//	}
//
// This is useful because the task-based goroutine start implementation only
// allows a single (pointer) argument to the newly started goroutine. Also, it
// ignores the return value because newly started goroutines do not have a
// return value.
//
// The hasContext parameter indicates whether the context parameter (the second
// to last parameter of the function) is used for this wrapper. If hasContext is
// false, the parameter bundle is assumed to have no context parameter and undef
// is passed instead.
func (c *compilerContext) createGoroutineStartWrapper(fnType llvm.Type, fn llvm.Value, prefix string, hasContext, isWasmExport bool, pos token.Pos) llvm.Value {
	var wrapper llvm.Value

	b := &builder{
		compilerContext: c,
		Builder:         c.ctx.NewBuilder(),
	}
	defer b.Dispose()

	var deadlock llvm.Value
	var deadlockType llvm.Type
	if c.Scheduler == "asyncify" {
		deadlockType, deadlock = c.getFunction(c.program.ImportedPackage("runtime").Members["deadlock"].(*ssa.Function))
	}

	if !fn.IsAFunction().IsNil() {
		// See whether this wrapper has already been created. If so, return it.
		name := fn.Name()
		wrapperName := name + "$gowrapper"
		if isWasmExport {
			wrapperName += "-wasmexport"
		}
		wrapper = c.mod.NamedFunction(wrapperName)
		if !wrapper.IsNil() {
			return llvm.ConstPtrToInt(wrapper, c.uintptrType)
		}

		// Create the wrapper.
		wrapperType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.dataPtrType}, false)
		wrapper = llvm.AddFunction(c.mod, wrapperName, wrapperType)
		c.addStandardAttributes(wrapper)
		wrapper.SetLinkage(llvm.LinkOnceODRLinkage)
		wrapper.SetUnnamedAddr(true)
		wrapper.AddAttributeAtIndex(-1, c.ctx.CreateStringAttribute("tinygo-gowrapper", name))
		entry := c.ctx.AddBasicBlock(wrapper, "entry")
		b.SetInsertPointAtEnd(entry)

		if c.Debug {
			pos := c.program.Fset.Position(pos)
			diFuncType := c.dibuilder.CreateSubroutineType(llvm.DISubroutineType{
				File:       c.getDIFile(pos.Filename),
				Parameters: nil, // do not show parameters in debugger
				Flags:      0,   // ?
			})
			difunc := c.dibuilder.CreateFunction(c.getDIFile(pos.Filename), llvm.DIFunction{
				Name:         "<goroutine wrapper>",
				File:         c.getDIFile(pos.Filename),
				Line:         pos.Line,
				Type:         diFuncType,
				LocalToUnit:  true,
				IsDefinition: true,
				ScopeLine:    0,
				Flags:        llvm.FlagPrototyped,
				Optimized:    true,
			})
			wrapper.SetSubprogram(difunc)
			b.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), difunc, llvm.Metadata{})
		}

		if !isWasmExport {
			// Regular 'go' instruction.

			// Create the list of params for the call.
			paramTypes := fnType.ParamTypes()
			if !hasContext {
				paramTypes = paramTypes[:len(paramTypes)-1] // strip context parameter
			}

			params := b.emitPointerUnpack(wrapper.Param(0), paramTypes)
			if !hasContext {
				params = append(params, llvm.Undef(c.dataPtrType)) // add dummy context parameter
			}

			// Create the call.
			b.CreateCall(fnType, fn, params, "")

			if c.Scheduler == "asyncify" {
				b.CreateCall(deadlockType, deadlock, []llvm.Value{
					llvm.Undef(c.dataPtrType),
				}, "")
			}
		} else {
			// Goroutine started from a //go:wasmexport pragma.
			// The function looks like this:
			//
			//   func foo$gowrapper-wasmexport(state *stateStruct) {
			//       // load values
			//       param0 := state.params[0]
			//       param1 := state.params[1]
			//
			//       // call wrapped functions
			//       result := foo(param0, param1, ...)
			//
			//       // store result value (if there is any)
			//       state.result = result
			//
			//       // finish exported function
			//       state.done = true
			//       runtime.wasmExportExit()
			//   }
			//
			// The state object here looks like:
			//
			//   struct state {
			//       param0
			//       param1
			//       param* // etc
			//       done bool
			//       result returnType
			//   }

			returnType := fnType.ReturnType()
			hasReturn := returnType != b.ctx.VoidType()
			statePtr := wrapper.Param(0)

			// Create the state struct (it must match the type in createWasmExport).
			stateFields := fnType.ParamTypes()
			numParams := len(stateFields) - 1
			stateFields = stateFields[:numParams:numParams]     // strip 'context' parameter
			stateFields = append(stateFields, c.ctx.Int1Type()) // 'done' bool
			if hasReturn {
				stateFields = append(stateFields, returnType)
			}
			stateStruct := b.ctx.StructType(stateFields, false)

			// Extract parameters from the state object, and call the function
			// that's being wrapped.
			var callParams []llvm.Value
			for i := 0; i < numParams; i++ {
				gep := b.CreateInBoundsGEP(stateStruct, statePtr, []llvm.Value{
					llvm.ConstInt(b.ctx.Int32Type(), 0, false),
					llvm.ConstInt(b.ctx.Int32Type(), uint64(i), false),
				}, "")
				param := b.CreateLoad(stateFields[i], gep, "")
				callParams = append(callParams, param)
			}
			callParams = append(callParams, llvm.ConstNull(c.dataPtrType)) // add 'context' parameter
			result := b.CreateCall(fnType, fn, callParams, "")

			// Store the return value back into the shared state.
			// Unlike regular goroutines, these special //go:wasmexport
			// goroutines can return a value.
			if hasReturn {
				gep := b.CreateInBoundsGEP(stateStruct, statePtr, []llvm.Value{
					llvm.ConstInt(c.ctx.Int32Type(), 0, false),
					llvm.ConstInt(c.ctx.Int32Type(), uint64(numParams)+1, false),
				}, "result.ptr")
				b.CreateStore(result, gep)
			}

			// Mark this function as having finished executing.
			// This is important so the runtime knows the exported function
			// didn't block.
			doneGEP := b.CreateInBoundsGEP(stateStruct, statePtr, []llvm.Value{
				llvm.ConstInt(c.ctx.Int32Type(), 0, false),
				llvm.ConstInt(c.ctx.Int32Type(), uint64(numParams), false),
			}, "done.gep")
			b.CreateStore(llvm.ConstInt(b.ctx.Int1Type(), 1, false), doneGEP)

			// Call back into the runtime. This will exit the goroutine, switch
			// back to the scheduler, which will in turn return from the
			// //go:wasmexport function.
			b.createRuntimeCall("wasmExportExit", nil, "")
		}

	} else {
		// For a function pointer like this:
		//
		//     var funcPtr func(x, y int) int
		//
		// A wrapper like the following is created:
		//
		//     func .gowrapper(ptr *unsafe.Pointer) {
		//         args := (*struct{
		//             x, y int
		//             fn   func(x, y int) int
		//         })(ptr)
		//         args.fn(x, y)
		//     }
		//
		// With a bit of luck, identical wrapper functions like these can be
		// merged into one.

		// Create the wrapper.
		wrapperType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.dataPtrType}, false)
		wrapper = llvm.AddFunction(c.mod, prefix+".gowrapper", wrapperType)
		c.addStandardAttributes(wrapper)
		wrapper.SetLinkage(llvm.LinkOnceODRLinkage)
		wrapper.SetUnnamedAddr(true)
		wrapper.AddAttributeAtIndex(-1, c.ctx.CreateStringAttribute("tinygo-gowrapper", ""))
		entry := c.ctx.AddBasicBlock(wrapper, "entry")
		b.SetInsertPointAtEnd(entry)

		if c.Debug {
			pos := c.program.Fset.Position(pos)
			diFuncType := c.dibuilder.CreateSubroutineType(llvm.DISubroutineType{
				File:       c.getDIFile(pos.Filename),
				Parameters: nil, // do not show parameters in debugger
				Flags:      0,   // ?
			})
			difunc := c.dibuilder.CreateFunction(c.getDIFile(pos.Filename), llvm.DIFunction{
				Name:         "<goroutine wrapper>",
				File:         c.getDIFile(pos.Filename),
				Line:         pos.Line,
				Type:         diFuncType,
				LocalToUnit:  true,
				IsDefinition: true,
				ScopeLine:    0,
				Flags:        llvm.FlagPrototyped,
				Optimized:    true,
			})
			wrapper.SetSubprogram(difunc)
			b.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), difunc, llvm.Metadata{})
		}

		// Get the list of parameters, with the extra parameters at the end.
		paramTypes := fnType.ParamTypes()
		paramTypes = append(paramTypes, fn.Type()) // the last element is the function pointer
		params := b.emitPointerUnpack(wrapper.Param(0), paramTypes)

		// Get the function pointer.
		fnPtr := params[len(params)-1]
		params = params[:len(params)-1]

		// Create the call.
		b.CreateCall(fnType, fnPtr, params, "")

		if c.Scheduler == "asyncify" {
			b.CreateCall(deadlockType, deadlock, []llvm.Value{
				llvm.Undef(c.dataPtrType),
			}, "")
		}
	}

	if c.Scheduler == "asyncify" {
		// The goroutine was terminated via deadlock.
		b.CreateUnreachable()
	} else {
		// Finish the function. Every basic block must end in a terminator, and
		// because goroutines never return a value we can simply return void.
		b.CreateRetVoid()
	}

	// Return a ptrtoint of the wrapper, not the function itself.
	return llvm.ConstPtrToInt(wrapper, c.uintptrType)
}
