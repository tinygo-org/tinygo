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
	"go/types"
	"strings"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// deferInitFunc sets up this function for future deferred calls. It must be
// called from within the entry block when this function contains deferred
// calls.
func (b *builder) deferInitFunc() {
	// Some setup.
	b.deferFuncs = make(map[*ssa.Function]int)
	b.deferInvokeFuncs = make(map[string]int)
	b.deferClosureFuncs = make(map[*ssa.Function]int)
	b.deferExprFuncs = make(map[ssa.Value]int)
	b.deferBuiltinFuncs = make(map[ssa.Value]deferBuiltin)

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
		if _, ok := b.deferFuncs[callee]; !ok {
			b.deferFuncs[callee] = len(b.allDeferFuncs)
			b.allDeferFuncs = append(b.allDeferFuncs, callee)
		}
		callback := llvm.ConstInt(b.uintptrType, uint64(b.deferFuncs[callee]), false)

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
		fn := makeClosure.Fn.(*ssa.Function)
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

	} else if builtin, ok := instr.Call.Value.(*ssa.Builtin); ok {
		var argTypes []types.Type
		var argValues []llvm.Value
		for _, arg := range instr.Call.Args {
			argTypes = append(argTypes, arg.Type())
			argValues = append(argValues, b.getValue(arg))
		}

		if _, ok := b.deferBuiltinFuncs[instr.Call.Value]; !ok {
			b.deferBuiltinFuncs[instr.Call.Value] = deferBuiltin{
				callName: builtin.Name(),
				pos:      builtin.Pos(),
				argTypes: argTypes,
				callback: len(b.allDeferFuncs),
			}
			b.allDeferFuncs = append(b.allDeferFuncs, instr.Call.Value)
		}
		callback := llvm.ConstInt(b.uintptrType, uint64(b.deferBuiltinFuncs[instr.Call.Value].callback), false)

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields).
		values = []llvm.Value{callback, next}
		for _, param := range argValues {
			values = append(values, param)
			valueTypes = append(valueTypes, param.Type())
		}

	} else {
		funcValue := b.getValue(instr.Call.Value)

		if _, ok := b.deferExprFuncs[instr.Call.Value]; !ok {
			b.deferExprFuncs[instr.Call.Value] = len(b.allDeferFuncs)
			b.allDeferFuncs = append(b.allDeferFuncs, &instr.Call)
		}

		callback := llvm.ConstInt(b.uintptrType, uint64(b.deferExprFuncs[instr.Call.Value]), false)

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields, followed by all parameters including the
		// context pointer).
		values = []llvm.Value{callback, next, funcValue}
		valueTypes = append(valueTypes, funcValue.Type())
		for _, param := range instr.Call.Args {
			llvmParam := b.getValue(param)
			values = append(values, llvmParam)
			valueTypes = append(valueTypes, llvmParam.Type())
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
		nilPtr := llvm.ConstNull(b.i8ptrType)
		allocCall := b.createRuntimeCall("alloc", []llvm.Value{sizeValue, nilPtr}, "defer.alloc.call")
		alloca = b.CreateBitCast(allocCall, llvm.PointerType(deferFrameType, 0), "defer.alloc")
	}
	if b.NeedsStackObjects {
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
	loophead := b.ctx.AddBasicBlock(b.llvmFn, "rundefers.loophead")
	loop := b.ctx.AddBasicBlock(b.llvmFn, "rundefers.loop")
	unreachable := b.ctx.AddBasicBlock(b.llvmFn, "rundefers.default")
	end := b.ctx.AddBasicBlock(b.llvmFn, "rundefers.end")
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
		block := b.ctx.AddBasicBlock(b.llvmFn, "rundefers.callback")
		sw.AddCase(llvm.ConstInt(b.uintptrType, uint64(i), false), block)
		b.SetInsertPointAtEnd(block)
		switch callback := callback.(type) {
		case *ssa.CallCommon:
			// Call on an value or interface value.

			// Get the real defer struct type and cast to it.
			valueTypes := []llvm.Type{b.uintptrType, llvm.PointerType(b.getLLVMRuntimeType("_defer"), 0)}

			if !callback.IsInvoke() {
				//Expect funcValue to be passed through the defer frame.
				valueTypes = append(valueTypes, b.getFuncType(callback.Signature()))
			} else {
				//Expect typecode
				valueTypes = append(valueTypes, b.uintptrType, b.i8ptrType)
			}

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

			var fnPtr llvm.Value

			if !callback.IsInvoke() {
				// Isolate the func value.
				funcValue := forwardParams[0]
				forwardParams = forwardParams[1:]

				//Get function pointer and context
				fp, context := b.decodeFuncValue(funcValue, callback.Signature())
				fnPtr = fp

				//Pass context
				forwardParams = append(forwardParams, context)
			} else {
				// Move typecode from the start to the end of the list of
				// parameters.
				forwardParams = append(forwardParams[1:], forwardParams[0])
				fnPtr = b.getInvokeFunction(callback)

				// Add the context parameter. An interface call cannot also be a
				// closure but we have to supply the parameter anyway for platforms
				// with a strict calling convention.
				forwardParams = append(forwardParams, llvm.Undef(b.i8ptrType))
			}

			b.createCall(fnPtr, forwardParams, "")

		case *ssa.Function:
			// Direct call.

			// Get the real defer struct type and cast to it.
			valueTypes := []llvm.Type{b.uintptrType, llvm.PointerType(b.getLLVMRuntimeType("_defer"), 0)}
			for _, param := range getParams(callback.Signature) {
				valueTypes = append(valueTypes, b.getLLVMType(param.Type()))
			}
			deferFrameType := b.ctx.StructType(valueTypes, false)
			deferFramePtr := b.CreateBitCast(deferData, llvm.PointerType(deferFrameType, 0), "deferFrame")

			// Extract the params from the struct.
			forwardParams := []llvm.Value{}
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := range getParams(callback.Signature) {
				gep := b.CreateInBoundsGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(i+2), false)}, "gep")
				forwardParam := b.CreateLoad(gep, "param")
				forwardParams = append(forwardParams, forwardParam)
			}

			// Plain TinyGo functions add some extra parameters to implement async functionality and function recievers.
			// These parameters should not be supplied when calling into an external C/ASM function.
			if !b.getFunctionInfo(callback).exported {
				// Add the context parameter. We know it is ignored by the receiving
				// function, but we have to pass one anyway.
				forwardParams = append(forwardParams, llvm.Undef(b.i8ptrType))
			}

			createdAtomicOp := false
			fqName := callback.RelString(nil)
			// If this is an atomic operation, don't generate a function call
			// instead just inline an atomic operation
			if strings.HasPrefix(fqName, "sync/atomic.") {
				_, createdAtomicOp = b.createAtomicOp(callback, forwardParams)
			}

			// this was not an atomic operation, just do a normal function call
			if !createdAtomicOp {
				b.createCall(b.getFunction(callback), forwardParams, "")
			}
		case *ssa.MakeClosure:
			// Get the real defer struct type and cast to it.
			fn := callback.Fn.(*ssa.Function)
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

			// Call deferred function.
			b.createCall(b.getFunction(fn), forwardParams, "")
		case *ssa.Builtin:
			db := b.deferBuiltinFuncs[callback]

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
			var argValues []llvm.Value
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := 0; i < params.Len(); i++ {
				gep := b.CreateInBoundsGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(i+2), false)}, "gep")
				forwardParam := b.CreateLoad(gep, "param")
				argValues = append(argValues, forwardParam)
			}

			_, err := b.createBuiltin(db.argTypes, argValues, db.callName, db.pos)
			if err != nil {
				b.diagnostics = append(b.diagnostics, err)
			}
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
