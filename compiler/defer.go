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
	"fmt"
	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"github.com/tinygo-org/tinygo/ir"
	"go/types"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// deferInitFunc sets up this function for future deferred calls. It must be
// called from within the entry block when this function contains deferred
// calls.
func (b *builder) deferInitFunc() {
	// Some setup.
	b.deferFuncs = make(map[*ir.Function]int)
	b.deferInvokeFuncs = make(map[string]int)
	b.deferClosureFuncs = make(map[*ir.Function]int)
	b.deferExprFuncs = make(map[interface{}]deferExpr)
	b.deferBuiltinFuncs = make(map[interface{}]deferBuiltin)

	// Create defer list pointer.
	deferType := llvm.PointerType(b.getLLVMRuntimeType("_defer"), 0)
	b.deferPtr = b.CreateAlloca(deferType, "deferPtr")
	b.CreateStore(llvm.ConstPointerNull(deferType), b.deferPtr)
}

// isInLoop checks if there is a path from a basic block to itself.
func isInLoop(start *ssa.BasicBlock) bool {
	// Use a breadth-first search to scan backwards through the block graph.
	queue := []*ssa.BasicBlock{start}
	checked := map[*ssa.BasicBlock]struct{}{}

	for len(queue) > 0 {
		// pop a block off of the queue
		block := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		// Search through predecessors.
		// Searching backwards means that this is pretty fast when the block is close to the start of the function.
		// Defers are often placed near the start of the function.
		for _, pred := range block.Preds {
			if pred == start {
				// cycle found
				return true
			}

			if _, ok := checked[pred]; ok {
				// block already checked
				continue
			}

			// add to queue and checked map
			queue = append(queue, pred)
			checked[pred] = struct{}{}
		}
	}

	return false
}

// createDefer emits a single defer instruction, to be run when this function
// returns.
func (b *builder) createDefer(instr *ssa.Defer) {
	// The pointer to the previous defer struct, which we will replace to
	// make a linked list.
	next := b.CreateLoad(b.deferPtr, "defer.next")

	var values []llvm.Value
	valueTypes := []llvm.Type{b.uintptrType, next.Type()}
	if instr.Call.IsInvoke() {
		// Method call on an interface.

		// Get callback type number.
		methodName := instr.Call.Method.FullName()
		if _, ok := b.deferInvokeFuncs[methodName]; !ok {
			b.deferInvokeFuncs[methodName] = len(b.allDeferFuncs)
			b.allDeferFuncs = append(b.allDeferFuncs, &instr.Call)
		}
		callback := llvm.ConstInt(b.uintptrType, uint64(b.deferInvokeFuncs[methodName]), false)

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields, followed by the call parameters).
		itf := b.getValue(instr.Call.Value) // interface
		typecode := b.CreateExtractValue(itf, 0, "invoke.func.typecode")
		receiverValue := b.CreateExtractValue(itf, 1, "invoke.func.receiver")
		values = []llvm.Value{callback, next, typecode, receiverValue}
		valueTypes = append(valueTypes, b.uintptrType, b.i8ptrType)
		for _, arg := range instr.Call.Args {
			val := b.getValue(arg)
			values = append(values, val)
			valueTypes = append(valueTypes, val.Type())
		}

	} else if callee, ok := instr.Call.Value.(*ssa.Function); ok {
		// Regular function call.
		fn := b.ir.GetFunction(callee)

		if _, ok := b.deferFuncs[fn]; !ok {
			b.deferFuncs[fn] = len(b.allDeferFuncs)
			b.allDeferFuncs = append(b.allDeferFuncs, fn)
		}
		callback := llvm.ConstInt(b.uintptrType, uint64(b.deferFuncs[fn]), false)

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields).
		values = []llvm.Value{callback, next}
		for _, param := range instr.Call.Args {
			llvmParam := b.getValue(param)
			values = append(values, llvmParam)
			valueTypes = append(valueTypes, llvmParam.Type())
		}

	} else if makeClosure, ok := instr.Call.Value.(*ssa.MakeClosure); ok {
		// Immediately applied function literal with free variables.

		// Extract the context from the closure. We won't need the function
		// pointer.
		// TODO: ignore this closure entirely and put pointers to the free
		// variables directly in the defer struct, avoiding a memory allocation.
		closure := b.getValue(instr.Call.Value)
		context := b.CreateExtractValue(closure, 0, "")

		// Get the callback number.
		fn := b.ir.GetFunction(makeClosure.Fn.(*ssa.Function))
		if _, ok := b.deferClosureFuncs[fn]; !ok {
			b.deferClosureFuncs[fn] = len(b.allDeferFuncs)
			b.allDeferFuncs = append(b.allDeferFuncs, makeClosure)
		}
		callback := llvm.ConstInt(b.uintptrType, uint64(b.deferClosureFuncs[fn]), false)

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields, followed by all parameters including the
		// context pointer).
		values = []llvm.Value{callback, next}
		for _, param := range instr.Call.Args {
			llvmParam := b.getValue(param)
			values = append(values, llvmParam)
			valueTypes = append(valueTypes, llvmParam.Type())
		}
		values = append(values, context)
		valueTypes = append(valueTypes, context.Type())

	} else if  builtin, ok := instr.Call.Value.(*ssa.Builtin); ok {
		var funcName string
		switch builtin.Name() {
		case "close":
			funcName = "chanClose"
		default:
			b.addError(instr.Pos(), fmt.Sprint("TODO: Implement defer for ", builtin.Name()))
			return
		}

		if _, ok := b.deferBuiltinFuncs[instr.Call.Value]; !ok {
			b.deferBuiltinFuncs[instr.Call.Value] = deferBuiltin {
				funcName,
				len(b.allDeferFuncs),
			}
			b.allDeferFuncs = append(b.allDeferFuncs, instr.Call.Value)
		}
		callback := llvm.ConstInt(b.uintptrType, uint64(b.deferBuiltinFuncs[instr.Call.Value].callback), false)

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields).
		values = []llvm.Value{callback, next}
		for _, param := range instr.Call.Args {
			llvmParam := b.getValue(param)
			values = append(values, llvmParam)
			valueTypes = append(valueTypes, llvmParam.Type())
		}

	} else {
		var funcValue llvm.Value
		var sig *types.Signature

		switch expr := instr.Call.Value.(type) {
		case *ssa.Extract:
			value := b.getValue(expr.Tuple)
			funcValue = b.CreateExtractValue(value, expr.Index, "")
			sig = expr.Tuple.(*ssa.Call).Call.Value.(*ssa.Function).Signature.Results().At(expr.Index).Type().Underlying().(*types.Signature)
		case *ssa.Call:
			funcValue = b.getValue(expr)
			sig = expr.Call.Value.Type().Underlying().(*types.Signature).Results().At(0).Type().Underlying().(*types.Signature)
		case *ssa.UnOp:
			funcValue = b.getValue(expr)
			switch ty := expr.X.Type().(type) {
			case *types.Pointer:
				sig = ty.Elem().Underlying().(*types.Signature)
			default:
				sig = ty.Underlying().(*types.Signature).Results().At(0).Type().Underlying().(*types.Signature)
			}
		}

		if funcValue.IsNil() == false && sig != nil {
			//funcSig, context := b.decodeFuncValue(funcValue, sig)
			if _, ok := b.deferExprFuncs[instr.Call.Value]; !ok {
				b.deferExprFuncs[instr.Call.Value] = deferExpr{
					funcValueType: funcValue.Type(),
					signature: sig,
					callback: len(b.allDeferFuncs),
				}
				b.allDeferFuncs = append(b.allDeferFuncs, instr.Call.Value)
			}

			callback := llvm.ConstInt(b.uintptrType, uint64(b.deferExprFuncs[instr.Call.Value].callback), false)

			// Collect all values to be put in the struct (starting with
			// runtime._defer fields, followed by all parameters including the
			// context pointer).
			values = []llvm.Value{callback, next}
			for _, param := range instr.Call.Args {
				llvmParam := b.getValue(param)
				values = append(values, llvmParam)
				valueTypes = append(valueTypes, llvmParam.Type())
			}

			//Pass funcValue through defer frame
			values = append(values, funcValue)
			valueTypes = append(valueTypes, funcValue.Type())

		} else {
			b.addError(instr.Pos(), "todo: defer on uncommon function call type")
			return
		}
	}

	// Make a struct out of the collected values to put in the defer frame.
	deferFrameType := b.ctx.StructType(valueTypes, false)
	deferFrame := llvm.ConstNull(deferFrameType)
	for i, value := range values {
		deferFrame = b.CreateInsertValue(deferFrame, value, i, "")
	}

	// Put this struct in an allocation.
	var alloca llvm.Value
	if !isInLoop(instr.Block()) {
		// This can safely use a stack allocation.
		alloca = llvmutil.CreateEntryBlockAlloca(b.Builder, deferFrameType, "defer.alloca")
	} else {
		// This may be hit a variable number of times, so use a heap allocation.
		size := b.targetData.TypeAllocSize(deferFrameType)
		sizeValue := llvm.ConstInt(b.uintptrType, size, false)
		allocCall := b.createRuntimeCall("alloc", []llvm.Value{sizeValue}, "defer.alloc.call")
		alloca = b.CreateBitCast(allocCall, llvm.PointerType(deferFrameType, 0), "defer.alloc")
	}
	if b.NeedsStackObjects() {
		b.trackPointer(alloca)
	}
	b.CreateStore(deferFrame, alloca)

	// Push it on top of the linked list by replacing deferPtr.
	allocaCast := b.CreateBitCast(alloca, next.Type(), "defer.alloca.cast")
	b.CreateStore(allocaCast, b.deferPtr)
}

// createRunDefers emits code to run all deferred functions.
func (b *builder) createRunDefers() {
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
	loophead := b.ctx.AddBasicBlock(b.fn.LLVMFn, "rundefers.loophead")
	loop := b.ctx.AddBasicBlock(b.fn.LLVMFn, "rundefers.loop")
	unreachable := b.ctx.AddBasicBlock(b.fn.LLVMFn, "rundefers.default")
	end := b.ctx.AddBasicBlock(b.fn.LLVMFn, "rundefers.end")
	b.CreateBr(loophead)

	// Create loop head:
	//     for stack != nil {
	b.SetInsertPointAtEnd(loophead)
	deferData := b.CreateLoad(b.deferPtr, "")
	stackIsNil := b.CreateICmp(llvm.IntEQ, deferData, llvm.ConstPointerNull(deferData.Type()), "stackIsNil")
	b.CreateCondBr(stackIsNil, end, loop)

	// Create loop body:
	//     _stack := stack
	//     stack = stack.next
	//     switch stack.callback {
	b.SetInsertPointAtEnd(loop)
	nextStackGEP := b.CreateInBoundsGEP(deferData, []llvm.Value{
		llvm.ConstInt(b.ctx.Int32Type(), 0, false),
		llvm.ConstInt(b.ctx.Int32Type(), 1, false), // .next field
	}, "stack.next.gep")
	nextStack := b.CreateLoad(nextStackGEP, "stack.next")
	b.CreateStore(nextStack, b.deferPtr)
	gep := b.CreateInBoundsGEP(deferData, []llvm.Value{
		llvm.ConstInt(b.ctx.Int32Type(), 0, false),
		llvm.ConstInt(b.ctx.Int32Type(), 0, false), // .callback field
	}, "callback.gep")
	callback := b.CreateLoad(gep, "callback")
	sw := b.CreateSwitch(callback, unreachable, len(b.allDeferFuncs))

	for i, callback := range b.allDeferFuncs {
		// Create switch case, for example:
		//     case 0:
		//         // run first deferred call
		block := b.ctx.AddBasicBlock(b.fn.LLVMFn, "rundefers.callback")
		sw.AddCase(llvm.ConstInt(b.uintptrType, uint64(i), false), block)
		b.SetInsertPointAtEnd(block)
		switch callback := callback.(type) {
		case *ssa.CallCommon:
			// Call on an interface value.
			if !callback.IsInvoke() {
				panic("expected an invoke call, not a direct call")
			}

			// Get the real defer struct type and cast to it.
			valueTypes := []llvm.Type{b.uintptrType, llvm.PointerType(b.getLLVMRuntimeType("_defer"), 0), b.uintptrType, b.i8ptrType}
			for _, arg := range callback.Args {
				valueTypes = append(valueTypes, b.getLLVMType(arg.Type()))
			}
			deferFrameType := b.ctx.StructType(valueTypes, false)
			deferFramePtr := b.CreateBitCast(deferData, llvm.PointerType(deferFrameType, 0), "deferFrame")

			// Extract the params from the struct (including receiver).
			forwardParams := []llvm.Value{}
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := 2; i < len(valueTypes); i++ {
				gep := b.CreateInBoundsGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(i), false)}, "gep")
				forwardParam := b.CreateLoad(gep, "param")
				forwardParams = append(forwardParams, forwardParam)
			}

			// Isolate the typecode.
			typecode, forwardParams := forwardParams[0], forwardParams[1:]

			// Add the context parameter. An interface call cannot also be a
			// closure but we have to supply the parameter anyway for platforms
			// with a strict calling convention.
			forwardParams = append(forwardParams, llvm.Undef(b.i8ptrType))

			// Parent coroutine handle.
			forwardParams = append(forwardParams, llvm.Undef(b.i8ptrType))

			fnPtr := b.getInvokePtr(callback, typecode)
			b.createCall(fnPtr, forwardParams, "")

		case *ir.Function:
			// Direct call.

			// Get the real defer struct type and cast to it.
			valueTypes := []llvm.Type{b.uintptrType, llvm.PointerType(b.getLLVMRuntimeType("_defer"), 0)}
			for _, param := range callback.Params {
				valueTypes = append(valueTypes, b.getLLVMType(param.Type()))
			}
			deferFrameType := b.ctx.StructType(valueTypes, false)
			deferFramePtr := b.CreateBitCast(deferData, llvm.PointerType(deferFrameType, 0), "deferFrame")

			// Extract the params from the struct.
			forwardParams := []llvm.Value{}
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := range callback.Params {
				gep := b.CreateInBoundsGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(i+2), false)}, "gep")
				forwardParam := b.CreateLoad(gep, "param")
				forwardParams = append(forwardParams, forwardParam)
			}

			// Plain TinyGo functions add some extra parameters to implement async functionality and function recievers.
			// These parameters should not be supplied when calling into an external C/ASM function.
			if !callback.IsExported() {
				// Add the context parameter. We know it is ignored by the receiving
				// function, but we have to pass one anyway.
				forwardParams = append(forwardParams, llvm.Undef(b.i8ptrType))

				// Parent coroutine handle.
				forwardParams = append(forwardParams, llvm.Undef(b.i8ptrType))
			}

			// Call real function.
			b.createCall(callback.LLVMFn, forwardParams, "")

		case *ssa.MakeClosure:
			// Get the real defer struct type and cast to it.
			fn := b.ir.GetFunction(callback.Fn.(*ssa.Function))
			valueTypes := []llvm.Type{b.uintptrType, llvm.PointerType(b.getLLVMRuntimeType("_defer"), 0)}
			params := fn.Signature.Params()
			for i := 0; i < params.Len(); i++ {
				valueTypes = append(valueTypes, b.getLLVMType(params.At(i).Type()))
			}
			valueTypes = append(valueTypes, b.i8ptrType) // closure
			deferFrameType := b.ctx.StructType(valueTypes, false)
			deferFramePtr := b.CreateBitCast(deferData, llvm.PointerType(deferFrameType, 0), "deferFrame")

			// Extract the params from the struct.
			forwardParams := []llvm.Value{}
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := 2; i < len(valueTypes); i++ {
				gep := b.CreateInBoundsGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(i), false)}, "")
				forwardParam := b.CreateLoad(gep, "param")
				forwardParams = append(forwardParams, forwardParam)
			}

			// Parent coroutine handle.
			forwardParams = append(forwardParams, llvm.Undef(b.i8ptrType))

			// Call deferred function.
			b.createCall(fn.LLVMFn, forwardParams, "")
		case *ssa.Extract, *ssa.Call, *ssa.UnOp:
			expr := b.deferExprFuncs[callback]

			// Get the real defer struct type and cast to it.
			valueTypes := []llvm.Type{b.uintptrType, llvm.PointerType(b.getLLVMRuntimeType("_defer"), 0)}

			//Get signature from call results
			params := expr.signature.Params()
			for i := 0; i < params.Len(); i++ {
				valueTypes = append(valueTypes, b.getLLVMType(params.At(i).Type()))
			}

			valueTypes = append(valueTypes, expr.funcValueType)
			deferFrameType := b.ctx.StructType(valueTypes, false)
			deferFramePtr := b.CreateBitCast(deferData, llvm.PointerType(deferFrameType, 0), "deferFrame")

			// Extract the params from the struct.
			var forwardParams []llvm.Value
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			funcPtrIndex := len(valueTypes)-1
			for i := 2; i < funcPtrIndex; i++ {
				gep := b.CreateInBoundsGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(i), false)}, "")
				forwardParam := b.CreateLoad(gep, "param")
				forwardParams = append(forwardParams, forwardParam)
			}

			//Last one is funcValue
			gep := b.CreateInBoundsGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(funcPtrIndex), false)}, "")
			fun := b.CreateLoad(gep, "param.func")

			//Get funcValueWithSignature and context
			funcPtr, context := b.decodeFuncValue(fun, expr.signature)

			//Pass context
			forwardParams = append(forwardParams, context)

			// Parent coroutine handle.
			forwardParams = append(forwardParams, llvm.Undef(b.i8ptrType))

			// Call deferred function.
			b.createCall(funcPtr, forwardParams, "")
		case *ssa.Builtin:
			db := b.deferBuiltinFuncs[callback]
			fullName := "runtime." + db.funcName
			fn := b.mod.NamedFunction(fullName)

			//Get parameter types
			valueTypes := []llvm.Type{b.uintptrType, llvm.PointerType(b.getLLVMRuntimeType("_defer"), 0)}

			//Get signature from call results
			params := callback.Type().Underlying().(*types.Signature).Params()
			for i := 0; i < params.Len(); i++ {
				valueTypes = append(valueTypes, b.getLLVMType(params.At(i).Type()))
			}

			deferFrameType := b.ctx.StructType(valueTypes, false)
			deferFramePtr := b.CreateBitCast(deferData, llvm.PointerType(deferFrameType, 0), "deferFrame")

			// Extract the params from the struct.
			forwardParams := []llvm.Value{}
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := 0; i < params.Len(); i++ {
				gep := b.CreateInBoundsGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(i+2), false)}, "gep")
				forwardParam := b.CreateLoad(gep, "param")
				forwardParams = append(forwardParams, forwardParam)
			}

			// Add the context parameter. We know it is ignored by the receiving
			// function, but we have to pass one anyway.
			forwardParams = append(forwardParams, llvm.Undef(b.i8ptrType))

			// Parent coroutine handle.
			forwardParams = append(forwardParams, llvm.Undef(b.i8ptrType))

			// Call real function.
			b.createCall(fn, forwardParams, "")
		default:
			panic("unknown deferred function type")
		}

		// Branch back to the start of the loop.
		b.CreateBr(loophead)
	}

	// Create default unreachable block:
	//     default:
	//         unreachable
	//     }
	b.SetInsertPointAtEnd(unreachable)
	b.CreateUnreachable()

	// End of loop.
	b.SetInsertPointAtEnd(end)
}
