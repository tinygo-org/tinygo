package compiler

// This file implements the 'go' keyword to start a new goroutine. See
// goroutine-lowering.go for more details.

import (
	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// createGoInstruction starts a new goroutine with the provided function pointer
// and parameters.
// In general, you should pass all regular parameters plus the context parameter.
// There is one exception: the task-based scheduler needs to have the function
// pointer passed in as a parameter too in addition to the context.
//
// Because a go statement doesn't return anything, return undef.
func (b *builder) createGoInstruction(funcPtr llvm.Value, params []llvm.Value) llvm.Value {
	paramBundle := b.emitPointerPack(params)
	var callee llvm.Value
	switch b.Scheduler() {
	case "none", "tasks":
		callee = b.createGoroutineStartWrapper(funcPtr)
	case "coroutines":
		callee = b.CreatePtrToInt(funcPtr, b.uintptrType, "")
	default:
		panic("unreachable")
	}
	b.createCall(b.mod.NamedFunction("internal/task.start"), []llvm.Value{callee, paramBundle, llvm.Undef(b.i8ptrType), llvm.ConstPointerNull(b.i8ptrType)}, "")
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
func (c *compilerContext) createGoroutineStartWrapper(fn llvm.Value) llvm.Value {
	var wrapper llvm.Value

	builder := c.ctx.NewBuilder()

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
		wrapper.SetLinkage(llvm.PrivateLinkage)
		wrapper.SetUnnamedAddr(true)
		entry := c.ctx.AddBasicBlock(wrapper, "entry")
		builder.SetInsertPointAtEnd(entry)

		// Create the list of params for the call.
		paramTypes := fn.Type().ElementType().ParamTypes()
		params := llvmutil.EmitPointerUnpack(builder, c.mod, wrapper.Param(0), paramTypes[:len(paramTypes)-1])
		params = append(params, llvm.Undef(c.i8ptrType))

		// Create the call.
		builder.CreateCall(fn, params, "")

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
		wrapper = llvm.AddFunction(c.mod, ".gowrapper", wrapperType)
		wrapper.SetLinkage(llvm.InternalLinkage)
		wrapper.SetUnnamedAddr(true)
		entry := c.ctx.AddBasicBlock(wrapper, "entry")
		builder.SetInsertPointAtEnd(entry)

		// Get the list of parameters, with the extra parameters at the end.
		paramTypes := fn.Type().ElementType().ParamTypes()
		paramTypes[len(paramTypes)-1] = fn.Type() // the last element is the function pointer
		params := llvmutil.EmitPointerUnpack(builder, c.mod, wrapper.Param(0), paramTypes)

		// Get the function pointer.
		fnPtr := params[len(params)-1]

		// Ignore the last param, which isn't used anymore.
		// TODO: avoid this extra "parent handle" parameter in most functions.
		params[len(params)-1] = llvm.Undef(c.i8ptrType)

		// Create the call.
		builder.CreateCall(fnPtr, params, "")
	}

	// Finish the function. Every basic block must end in a terminator, and
	// because goroutines never return a value we can simply return void.
	builder.CreateRetVoid()

	// Return a ptrtoint of the wrapper, not the function itself.
	return builder.CreatePtrToInt(wrapper, c.uintptrType, "")
}
