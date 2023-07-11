package compiler

// This file implements the 'go' keyword to start a new goroutine. See
// goroutine-lowering.go for more details.

import (
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// createGo emits code to start a new goroutine.
func (b *builder) createGo(instr *ssa.Go) {
	// Get all function parameters to pass to the goroutine.
	var params []llvm.Value
	for _, param := range instr.Call.Args {
		params = append(params, b.getValue(param, getPos(instr)))
	}

	var prefix string
	var funcPtr llvm.Value
	var funcPtrType llvm.Type
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
		funcPtrType, funcPtr = b.getFunction(callee)
	} else if builtin, ok := instr.Call.Value.(*ssa.Builtin); ok {
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
	} else if instr.Call.IsInvoke() {
		// This is a method call on an interface value.
		itf := b.getValue(instr.Call.Value, getPos(instr))
		itfTypeCode := b.CreateExtractValue(itf, 0, "")
		itfValue := b.CreateExtractValue(itf, 1, "")
		funcPtr = b.getInvokeFunction(&instr.Call)
		funcPtrType = funcPtr.GlobalValueType()
		params = append([]llvm.Value{itfValue}, params...) // start with receiver
		params = append(params, itfTypeCode)               // end with typecode
	} else {
		// This is a function pointer.
		// At the moment, two extra params are passed to the newly started
		// goroutine:
		//   * The function context, for closures.
		//   * The function pointer (for tasks).
		var context llvm.Value
		funcPtrType, funcPtr, context = b.decodeFuncValue(b.getValue(instr.Call.Value, getPos(instr)), instr.Call.Value.Type().Underlying().(*types.Signature))
		params = append(params, context, funcPtr)
		hasContext = true
		prefix = b.fn.RelString(nil)
	}

	paramBundle := b.emitPointerPack(params)
	var stackSize llvm.Value
	callee := b.createGoroutineStartWrapper(funcPtrType, funcPtr, prefix, hasContext, instr.Pos())
	if b.AutomaticStackSize {
		// The stack size is not known until after linking. Call a dummy
		// function that will be replaced with a load from a special ELF
		// section that contains the stack size (and is modified after
		// linking).
		stackSizeFnType, stackSizeFn := b.getFunction(b.program.ImportedPackage("internal/task").Members["getGoroutineStackSize"].(*ssa.Function))
		stackSize = b.createCall(stackSizeFnType, stackSizeFn, []llvm.Value{callee, llvm.Undef(b.i8ptrType)}, "stacksize")
	} else {
		// The stack size is fixed at compile time. By emitting it here as a
		// constant, it can be optimized.
		if (b.Scheduler == "tasks" || b.Scheduler == "asyncify") && b.DefaultStackSize == 0 {
			b.addError(instr.Pos(), "default stack size for goroutines is not set")
		}
		stackSize = llvm.ConstInt(b.uintptrType, b.DefaultStackSize, false)
	}
	fnType, start := b.getFunction(b.program.ImportedPackage("internal/task").Members["start"].(*ssa.Function))
	b.createCall(fnType, start, []llvm.Value{callee, paramBundle, stackSize, llvm.Undef(b.i8ptrType)}, "")
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
func (c *compilerContext) createGoroutineStartWrapper(fnType llvm.Type, fn llvm.Value, prefix string, hasContext bool, pos token.Pos) llvm.Value {
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
		wrapper = c.mod.NamedFunction(name + "$gowrapper")
		if !wrapper.IsNil() {
			return llvm.ConstPtrToInt(wrapper, c.uintptrType)
		}

		// Create the wrapper.
		wrapperType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.i8ptrType}, false)
		wrapper = llvm.AddFunction(c.mod, name+"$gowrapper", wrapperType)
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

		// Create the list of params for the call.
		paramTypes := fnType.ParamTypes()
		if !hasContext {
			paramTypes = paramTypes[:len(paramTypes)-1] // strip context parameter
		}
		params := b.emitPointerUnpack(wrapper.Param(0), paramTypes)
		if !hasContext {
			params = append(params, llvm.Undef(c.i8ptrType)) // add dummy context parameter
		}

		// Create the call.
		b.CreateCall(fnType, fn, params, "")

		if c.Scheduler == "asyncify" {
			b.CreateCall(deadlockType, deadlock, []llvm.Value{
				llvm.Undef(c.i8ptrType),
			}, "")
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
		wrapperType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.i8ptrType}, false)
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
				llvm.Undef(c.i8ptrType),
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
	return b.CreatePtrToInt(wrapper, c.uintptrType, "")
}
