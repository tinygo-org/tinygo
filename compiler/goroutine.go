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
		// Mark this function as a 'go' invocation and break invalid
		// interprocedural optimizations. For example, heap-to-stack
		// transformations are not sound as goroutines can outlive their parent.
		calleeType := funcPtr.Type()
		calleeValue := c.builder.CreatePtrToInt(funcPtr, c.uintptrType, "")
		calleeValue = c.createRuntimeCall("makeGoroutine", []llvm.Value{calleeValue}, "")
		calleeValue = c.builder.CreateIntToPtr(calleeValue, calleeType, "")

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
	if fn.IsAFunction().IsNil() {
		panic("todo: goroutine start wrapper for func value")
	}

	// See whether this wrapper has already been created. If so, return it.
	name := fn.Name()
	wrapper := c.mod.NamedFunction(name + "$gowrapper")
	if !wrapper.IsNil() {
		return c.builder.CreateIntToPtr(wrapper, c.uintptrType, "")
	}

	// Save the current position in the IR builder.
	currentBlock := c.builder.GetInsertBlock()
	defer c.builder.SetInsertPointAtEnd(currentBlock)

	// Create the wrapper.
	wrapperType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.i8ptrType}, false)
	wrapper = llvm.AddFunction(c.mod, name+"$gowrapper", wrapperType)
	wrapper.SetLinkage(llvm.PrivateLinkage)
	wrapper.SetUnnamedAddr(true)
	entry := llvm.AddBasicBlock(wrapper, "entry")
	c.builder.SetInsertPointAtEnd(entry)
	paramTypes := fn.Type().ElementType().ParamTypes()
	params := c.emitPointerUnpack(wrapper.Param(0), paramTypes[:len(paramTypes)-2])
	params = append(params, llvm.Undef(c.i8ptrType), llvm.ConstPointerNull(c.i8ptrType))
	c.builder.CreateCall(fn, params, "")
	c.builder.CreateRetVoid()
	return c.builder.CreatePtrToInt(wrapper, c.uintptrType, "")
}
