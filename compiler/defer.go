package compiler

// This file implements the 'defer' keyword in Go.
// Defer statements are implemented by transforming the function in the
// following way:
//   * Creating an alloca in the entry block that contains a pointer (initially
//     null) to the linked list of defer frames.
//   * Every time a defer statement is executed, a new defer frame is created
//     using alloca with a pointer to the previous defer frame, and the head
//     pointer in the entry block is replaced with a pointer to this defer
//     frame.
//   * On return, runtime.rundefers is called which calls all deferred functions
//     from the head of the linked list until it has gone through all defer
//     frames.

import (
	"github.com/aykevl/go-llvm"
	"github.com/aykevl/tinygo/ir"
	"golang.org/x/tools/go/ssa"
)

// deferInitFunc sets up this function for future deferred calls. It must be
// called from within the entry block when this function contains deferred
// calls.
func (c *Compiler) deferInitFunc(frame *Frame) {
	// Some setup.
	frame.deferFuncs = make(map[*ir.Function]int)
	frame.deferInvokeFuncs = make(map[string]int)
	frame.deferClosureFuncs = make(map[*ir.Function]int)

	// Create defer list pointer.
	deferType := llvm.PointerType(c.mod.GetTypeByName("runtime._defer"), 0)
	frame.deferPtr = c.builder.CreateAlloca(deferType, "deferPtr")
	c.builder.CreateStore(llvm.ConstPointerNull(deferType), frame.deferPtr)
}

// emitDefer emits a single defer instruction, to be run when this function
// returns.
func (c *Compiler) emitDefer(frame *Frame, instr *ssa.Defer) error {
	// The pointer to the previous defer struct, which we will replace to
	// make a linked list.
	next := c.builder.CreateLoad(frame.deferPtr, "defer.next")

	var values []llvm.Value
	valueTypes := []llvm.Type{c.uintptrType, next.Type()}
	if instr.Call.IsInvoke() {
		// Method call on an interface.

		// Get callback type number.
		methodName := instr.Call.Method.FullName()
		if _, ok := frame.deferInvokeFuncs[methodName]; !ok {
			frame.deferInvokeFuncs[methodName] = len(frame.allDeferFuncs)
			frame.allDeferFuncs = append(frame.allDeferFuncs, &instr.Call)
		}
		callback := llvm.ConstInt(c.uintptrType, uint64(frame.deferInvokeFuncs[methodName]), false)

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields, followed by the call parameters).
		itf, err := c.parseExpr(frame, instr.Call.Value) // interface
		if err != nil {
			return err
		}
		receiverValue := c.builder.CreateExtractValue(itf, 1, "invoke.func.receiver")
		values = []llvm.Value{callback, next, receiverValue}
		valueTypes = append(valueTypes, c.i8ptrType)
		for _, arg := range instr.Call.Args {
			val, err := c.parseExpr(frame, arg)
			if err != nil {
				return err
			}
			values = append(values, val)
			valueTypes = append(valueTypes, val.Type())
		}

	} else if callee, ok := instr.Call.Value.(*ssa.Function); ok {
		// Regular function call.
		fn := c.ir.GetFunction(callee)

		if _, ok := frame.deferFuncs[fn]; !ok {
			frame.deferFuncs[fn] = len(frame.allDeferFuncs)
			frame.allDeferFuncs = append(frame.allDeferFuncs, fn)
		}
		callback := llvm.ConstInt(c.uintptrType, uint64(frame.deferFuncs[fn]), false)

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields).
		values = []llvm.Value{callback, next}
		for _, param := range instr.Call.Args {
			llvmParam, err := c.parseExpr(frame, param)
			if err != nil {
				return err
			}
			values = append(values, llvmParam)
			valueTypes = append(valueTypes, llvmParam.Type())
		}

	} else if makeClosure, ok := instr.Call.Value.(*ssa.MakeClosure); ok {
		// Immediately applied function literal with free variables.

		// Extract the context from the closure. We won't need the function
		// pointer.
		// TODO: ignore this closure entirely and put pointers to the free
		// variables directly in the defer struct, avoiding a memory allocation.
		closure, err := c.parseExpr(frame, instr.Call.Value)
		if err != nil {
			return err
		}
		context := c.builder.CreateExtractValue(closure, 0, "")

		// Get the callback number.
		fn := c.ir.GetFunction(makeClosure.Fn.(*ssa.Function))
		if _, ok := frame.deferClosureFuncs[fn]; !ok {
			frame.deferClosureFuncs[fn] = len(frame.allDeferFuncs)
			frame.allDeferFuncs = append(frame.allDeferFuncs, makeClosure)
		}
		callback := llvm.ConstInt(c.uintptrType, uint64(frame.deferClosureFuncs[fn]), false)

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields, followed by all parameters including the
		// context pointer).
		values = []llvm.Value{callback, next}
		for _, param := range instr.Call.Args {
			llvmParam, err := c.parseExpr(frame, param)
			if err != nil {
				return err
			}
			values = append(values, llvmParam)
			valueTypes = append(valueTypes, llvmParam.Type())
		}
		values = append(values, context)
		valueTypes = append(valueTypes, context.Type())

	} else {
		return c.makeError(instr.Pos(), "todo: defer on uncommon function call type")
	}

	// Make a struct out of the collected values to put in the defer frame.
	deferFrameType := c.ctx.StructType(valueTypes, false)
	deferFrame, err := c.getZeroValue(deferFrameType)
	if err != nil {
		return err
	}
	for i, value := range values {
		deferFrame = c.builder.CreateInsertValue(deferFrame, value, i, "")
	}

	// Put this struct in an alloca.
	alloca := c.builder.CreateAlloca(deferFrameType, "defer.alloca")
	c.builder.CreateStore(deferFrame, alloca)

	// Push it on top of the linked list by replacing deferPtr.
	allocaCast := c.builder.CreateBitCast(alloca, next.Type(), "defer.alloca.cast")
	c.builder.CreateStore(allocaCast, frame.deferPtr)
	return nil
}

// emitRunDefers emits code to run all deferred functions.
func (c *Compiler) emitRunDefers(frame *Frame) error {
	// Add a loop like the following:
	//     for stack != nil {
	//         _stack := stack
	//         stack = stack.next
	//         switch _stack.callback {
	//         case 0:
	//             // run first deferred call
	//         case 1:
	//             // run second deferred call
	//             // etc.
	//         default:
	//             unreachable
	//         }
	//     }

	// Create loop.
	loophead := llvm.AddBasicBlock(frame.fn.LLVMFn, "rundefers.loophead")
	loop := llvm.AddBasicBlock(frame.fn.LLVMFn, "rundefers.loop")
	unreachable := llvm.AddBasicBlock(frame.fn.LLVMFn, "rundefers.default")
	end := llvm.AddBasicBlock(frame.fn.LLVMFn, "rundefers.end")
	c.builder.CreateBr(loophead)

	// Create loop head:
	//     for stack != nil {
	c.builder.SetInsertPointAtEnd(loophead)
	deferData := c.builder.CreateLoad(frame.deferPtr, "")
	stackIsNil := c.builder.CreateICmp(llvm.IntEQ, deferData, llvm.ConstPointerNull(deferData.Type()), "stackIsNil")
	c.builder.CreateCondBr(stackIsNil, end, loop)

	// Create loop body:
	//     _stack := stack
	//     stack = stack.next
	//     switch stack.callback {
	c.builder.SetInsertPointAtEnd(loop)
	nextStackGEP := c.builder.CreateGEP(deferData, []llvm.Value{
		llvm.ConstInt(c.ctx.Int32Type(), 0, false),
		llvm.ConstInt(c.ctx.Int32Type(), 1, false), // .next field
	}, "stack.next.gep")
	nextStack := c.builder.CreateLoad(nextStackGEP, "stack.next")
	c.builder.CreateStore(nextStack, frame.deferPtr)
	gep := c.builder.CreateGEP(deferData, []llvm.Value{
		llvm.ConstInt(c.ctx.Int32Type(), 0, false),
		llvm.ConstInt(c.ctx.Int32Type(), 0, false), // .callback field
	}, "callback.gep")
	callback := c.builder.CreateLoad(gep, "callback")
	sw := c.builder.CreateSwitch(callback, unreachable, len(frame.allDeferFuncs))

	for i, callback := range frame.allDeferFuncs {
		// Create switch case, for example:
		//     case 0:
		//         // run first deferred call
		block := llvm.AddBasicBlock(frame.fn.LLVMFn, "rundefers.callback")
		sw.AddCase(llvm.ConstInt(c.uintptrType, uint64(i), false), block)
		c.builder.SetInsertPointAtEnd(block)
		switch callback := callback.(type) {
		case *ssa.CallCommon:
			// Call on an interface value.
			if !callback.IsInvoke() {
				panic("expected an invoke call, not a direct call")
			}

			// Get the real defer struct type and cast to it.
			valueTypes := []llvm.Type{c.uintptrType, llvm.PointerType(c.mod.GetTypeByName("runtime._defer"), 0), c.i8ptrType}
			for _, arg := range callback.Args {
				llvmType, err := c.getLLVMType(arg.Type())
				if err != nil {
					return err
				}
				valueTypes = append(valueTypes, llvmType)
			}
			deferFrameType := c.ctx.StructType(valueTypes, false)
			deferFramePtr := c.builder.CreateBitCast(deferData, llvm.PointerType(deferFrameType, 0), "deferFrame")

			// Extract the params from the struct (including receiver).
			forwardParams := []llvm.Value{}
			zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
			for i := 2; i < len(valueTypes); i++ {
				gep := c.builder.CreateGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false)}, "gep")
				forwardParam := c.builder.CreateLoad(gep, "param")
				forwardParams = append(forwardParams, forwardParam)
			}

			// Add the context parameter. An interface call cannot also be a
			// closure but we have to supply the parameter anyway for platforms
			// with a strict calling convention.
			forwardParams = append(forwardParams, llvm.Undef(c.i8ptrType))

			// Parent coroutine handle.
			forwardParams = append(forwardParams, llvm.Undef(c.i8ptrType))

			fnPtr, _, err := c.getInvokeCall(frame, callback)
			if err != nil {
				return err
			}
			c.createCall(fnPtr, forwardParams, "")

		case *ir.Function:
			// Direct call.

			// Get the real defer struct type and cast to it.
			valueTypes := []llvm.Type{c.uintptrType, llvm.PointerType(c.mod.GetTypeByName("runtime._defer"), 0)}
			for _, param := range callback.Params {
				llvmType, err := c.getLLVMType(param.Type())
				if err != nil {
					return err
				}
				valueTypes = append(valueTypes, llvmType)
			}
			deferFrameType := c.ctx.StructType(valueTypes, false)
			deferFramePtr := c.builder.CreateBitCast(deferData, llvm.PointerType(deferFrameType, 0), "deferFrame")

			// Extract the params from the struct.
			forwardParams := []llvm.Value{}
			zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
			for i := range callback.Params {
				gep := c.builder.CreateGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(c.ctx.Int32Type(), uint64(i+2), false)}, "gep")
				forwardParam := c.builder.CreateLoad(gep, "param")
				forwardParams = append(forwardParams, forwardParam)
			}

			// Add the context parameter. We know it is ignored by the receiving
			// function, but we have to pass one anyway.
			forwardParams = append(forwardParams, llvm.Undef(c.i8ptrType))

			// Parent coroutine handle.
			forwardParams = append(forwardParams, llvm.Undef(c.i8ptrType))

			// Call real function.
			c.createCall(callback.LLVMFn, forwardParams, "")

		case *ssa.MakeClosure:
			// Get the real defer struct type and cast to it.
			fn := c.ir.GetFunction(callback.Fn.(*ssa.Function))
			valueTypes := []llvm.Type{c.uintptrType, llvm.PointerType(c.mod.GetTypeByName("runtime._defer"), 0)}
			params := fn.Signature.Params()
			for i := 0; i < params.Len(); i++ {
				llvmType, err := c.getLLVMType(params.At(i).Type())
				if err != nil {
					return err
				}
				valueTypes = append(valueTypes, llvmType)
			}
			valueTypes = append(valueTypes, c.i8ptrType) // closure
			deferFrameType := c.ctx.StructType(valueTypes, false)
			deferFramePtr := c.builder.CreateBitCast(deferData, llvm.PointerType(deferFrameType, 0), "deferFrame")

			// Extract the params from the struct.
			forwardParams := []llvm.Value{}
			zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
			for i := 2; i < len(valueTypes); i++ {
				gep := c.builder.CreateGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false)}, "")
				forwardParam := c.builder.CreateLoad(gep, "param")
				forwardParams = append(forwardParams, forwardParam)
			}

			// Parent coroutine handle.
			forwardParams = append(forwardParams, llvm.Undef(c.i8ptrType))

			// Call deferred function.
			c.createCall(fn.LLVMFn, forwardParams, "")

		default:
			panic("unknown deferred function type")
		}

		// Branch back to the start of the loop.
		c.builder.CreateBr(loophead)
	}

	// Create default unreachable block:
	//     default:
	//         unreachable
	//     }
	c.builder.SetInsertPointAtEnd(unreachable)
	c.builder.CreateUnreachable()

	// End of loop.
	c.builder.SetInsertPointAtEnd(end)
	return nil
}
