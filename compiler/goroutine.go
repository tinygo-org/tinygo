package compiler

// This file implements the 'go' keyword to start a new goroutine. See
// goroutine-lowering.go for more details.

import "tinygo.org/x/go-llvm"

// emitStartGoroutine starts a new goroutine with the provided function pointer
// and parameters.
//
// Because a go statement doesn't return anything, return undef.
func (c *Compiler) emitStartGoroutine(funcPtr llvm.Value, params []llvm.Value) llvm.Value {
	switch c.selectScheduler() {
	case "tasks":
		paramBundle := c.emitPointerPack(params)
		paramBundle = c.builder.CreatePtrToInt(paramBundle, c.uintptrType, "")

		calleeValue := c.createGoroutineStartWrapper(funcPtr)
		c.createRuntimeCall("startGoroutine", []llvm.Value{calleeValue, paramBundle}, "")
	case "coroutines":
		// We roundtrip through runtime.makeGoroutine as a signal (to find these
		// calls) and to break any optimizations LLVM will try to do: they are
		// invalid if we called this as a regular function to be updated later.
		calleeValue := c.builder.CreatePtrToInt(funcPtr, c.uintptrType, "")
		calleeValue = c.createRuntimeCall("makeGoroutine", []llvm.Value{calleeValue}, "")
		calleeValue = c.builder.CreateIntToPtr(calleeValue, funcPtr.Type(), "")
		c.createCall(calleeValue, params, "")
	default:
		panic("unreachable")
	}
	return llvm.Undef(funcPtr.Type().ElementType().ReturnType())
}

// createGoroutineStartWrapper creates a wrapper for the task-based
// implementation of goroutines. For example, to call a function like this:
//
//     func add(x, y int) int { ... }
//
// It creates a wrapper like this:
//
//     func add$gowrapper(ptr *unsafe.Pointer) {
//         args := (*struct{
//             x, y int
//         })(ptr)
//         add(args.x, args.y)
//     }
//
// This is useful because the task-based goroutine start implementation only
// allows a single (pointer) argument to the newly started goroutine. Also, it
// ignores the return value because newly started goroutines do not have a
// return value.
func (c *Compiler) createGoroutineStartWrapper(fn llvm.Value) llvm.Value {
	var wrapper llvm.Value

	if !fn.IsAFunction().IsNil() {
		// See whether this wrapper has already been created. If so, return it.
		name := fn.Name()
		wrapper = c.mod.NamedFunction(name + "$gowrapper")
		if !wrapper.IsNil() {
			return c.builder.CreatePtrToInt(wrapper, c.uintptrType, "")
		}

		// Save the current position in the IR builder.
		currentBlock := c.builder.GetInsertBlock()
		defer c.builder.SetInsertPointAtEnd(currentBlock)

		// Create the wrapper.
		wrapperType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.i8ptrType}, false)
		wrapper = llvm.AddFunction(c.mod, name+"$gowrapper", wrapperType)
		if wrapper.IsNil() {
			panic("wrapper ($gowrapper) is nil")
		}
		wrapper.SetLinkage(llvm.PrivateLinkage)
		wrapper.SetUnnamedAddr(true)
		entry := c.ctx.AddBasicBlock(wrapper, "entry")
		c.builder.SetInsertPointAtEnd(entry)

		// Create the list of params for the call.
		paramTypes := fn.Type().ElementType().ParamTypes()
		params := c.emitPointerUnpack(wrapper.Param(0), paramTypes[:len(paramTypes)-2])
		params = append(params, llvm.Undef(c.i8ptrType), llvm.ConstPointerNull(c.i8ptrType))

		// Create the call.
		c.builder.CreateCall(fn, params, "")

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

		// Save the current position in the IR builder.
		currentBlock := c.builder.GetInsertBlock()
		defer c.builder.SetInsertPointAtEnd(currentBlock)

		// Create the wrapper.
		wrapperType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.i8ptrType}, false)
		wrapper = llvm.AddFunction(c.mod, ".gowrapper", wrapperType)
		wrapper.SetLinkage(llvm.InternalLinkage)
		if wrapper.IsNil() {
			panic("wrapper (.gowrapper) is nil")
		}
		wrapper.SetUnnamedAddr(true)
		entry := c.ctx.AddBasicBlock(wrapper, "entry")
		c.builder.SetInsertPointAtEnd(entry)

		// Get the list of parameters, with the extra parameters at the end.
		paramTypes := fn.Type().ElementType().ParamTypes()
		paramTypes[len(paramTypes)-1] = fn.Type() // the last element is the function pointer
		params := c.emitPointerUnpack(wrapper.Param(0), paramTypes)

		// Get the function pointer.
		fnPtr := params[len(params)-1]

		// Ignore the last param, which isn't used anymore.
		// TODO: avoid this extra "parent handle" parameter in most functions.
		params[len(params)-1] = llvm.Undef(c.i8ptrType)

		// Create the call.
		c.builder.CreateCall(fnPtr, params, "")
	}

	// Finish the function. Every basic block must end in a terminator, and
	// because goroutines never return a value we can simply return void.
	c.builder.CreateRetVoid()

	// Return a ptrtoint of the wrapper, not the function itself.
	return c.builder.CreatePtrToInt(wrapper, c.uintptrType, "")
}
