package compiler

// This file implements the 'defer' keyword in Go. See src/runtime/defer.go for
// details.

import (
	"go/types"

	"github.com/aykevl/go-llvm"
	"golang.org/x/tools/go/ssa"
)

// A thunk for a defer that defers calling a function pointer with context.
type ContextDeferFunction struct {
	fn          llvm.Value
	deferStruct []llvm.Type
	signature   *types.Signature
}

// A thunk for a defer that defers calling an interface method.
type InvokeDeferFunction struct {
	method     *types.Func
	valueTypes []llvm.Type
}

// deferInitFunc sets up this function for future deferred calls.
func (c *Compiler) deferInitFunc(frame *Frame) {
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

	deferFuncType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{next.Type()}, false)

	var values []llvm.Value
	var valueTypes []llvm.Type
	if instr.Call.IsInvoke() {
		// Function call on an interface.
		fnPtr, args, err := c.getInvokeCall(frame, &instr.Call)
		if err != nil {
			return err
		}

		valueTypes = []llvm.Type{llvm.PointerType(deferFuncType, 0), next.Type(), fnPtr.Type()}
		for _, param := range args {
			valueTypes = append(valueTypes, param.Type())
		}

		// Create a thunk.
		deferName := instr.Call.Method.FullName() + "$defer"
		callback := c.mod.NamedFunction(deferName)
		if callback.IsNil() {
			// Not found, have to add it.
			callback = llvm.AddFunction(c.mod, deferName, deferFuncType)
			thunk := InvokeDeferFunction{
				method:     instr.Call.Method,
				valueTypes: valueTypes,
			}
			c.deferInvokeFuncs = append(c.deferInvokeFuncs, thunk)
		}

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields, followed by the function pointer to be
		// called).
		values = append([]llvm.Value{callback, next, fnPtr}, args...)

	} else if callee, ok := instr.Call.Value.(*ssa.Function); ok {
		// Regular function call.
		fn := c.ir.GetFunction(callee)

		// Try to find the wrapper $defer function.
		deferName := fn.LinkName() + "$defer"
		callback := c.mod.NamedFunction(deferName)
		if callback.IsNil() {
			// Not found, have to add it.
			callback = llvm.AddFunction(c.mod, deferName, deferFuncType)
			c.deferFuncs = append(c.deferFuncs, fn)
		}

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields).
		values = []llvm.Value{callback, next}
		valueTypes = []llvm.Type{callback.Type(), next.Type()}
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
		closure, err := c.parseExpr(frame, instr.Call.Value)
		if err != nil {
			return err
		}

		// Hopefully, LLVM will merge equivalent functions.
		deferName := frame.fn.LinkName() + "$fpdefer"
		callback := llvm.AddFunction(c.mod, deferName, deferFuncType)

		// Collect all values to be put in the struct (starting with
		// runtime._defer fields, followed by the closure).
		values = []llvm.Value{callback, next, closure}
		valueTypes = []llvm.Type{callback.Type(), next.Type(), closure.Type()}
		for _, param := range instr.Call.Args {
			llvmParam, err := c.parseExpr(frame, param)
			if err != nil {
				return err
			}
			values = append(values, llvmParam)
			valueTypes = append(valueTypes, llvmParam.Type())
		}

		thunk := ContextDeferFunction{
			callback,
			valueTypes,
			makeClosure.Fn.(*ssa.Function).Signature,
		}
		c.ctxDeferFuncs = append(c.ctxDeferFuncs, thunk)

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
	deferData := c.builder.CreateLoad(frame.deferPtr, "")
	c.createRuntimeCall("rundefers", []llvm.Value{deferData}, "")
	return nil
}

// finalizeDefers creates thunks for deferred functions.
func (c *Compiler) finalizeDefers() error {
	// Create deferred function wrappers.
	for _, fn := range c.deferFuncs {
		// This function gets a single parameter which is a pointer to a struct
		// (the defer frame).
		// This struct starts with the values of runtime._defer, but after that
		// follow the real function parameters.
		// The job of this wrapper is to extract these parameters and to call
		// the real function with them.
		llvmFn := c.mod.NamedFunction(fn.LinkName() + "$defer")
		llvmFn.SetLinkage(llvm.InternalLinkage)
		llvmFn.SetUnnamedAddr(true)
		entry := c.ctx.AddBasicBlock(llvmFn, "entry")
		c.builder.SetInsertPointAtEnd(entry)
		deferRawPtr := llvmFn.Param(0)

		// Get the real param type and cast to it.
		valueTypes := []llvm.Type{llvmFn.Type(), llvm.PointerType(c.mod.GetTypeByName("runtime._defer"), 0)}
		for _, param := range fn.Params {
			llvmType, err := c.getLLVMType(param.Type())
			if err != nil {
				return err
			}
			valueTypes = append(valueTypes, llvmType)
		}
		deferFrameType := c.ctx.StructType(valueTypes, false)
		deferFramePtr := c.builder.CreateBitCast(deferRawPtr, llvm.PointerType(deferFrameType, 0), "deferFrame")

		// Extract the params from the struct.
		forwardParams := []llvm.Value{}
		zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
		for i := range fn.Params {
			gep := c.builder.CreateGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(c.ctx.Int32Type(), uint64(i+2), false)}, "gep")
			forwardParam := c.builder.CreateLoad(gep, "param")
			forwardParams = append(forwardParams, forwardParam)
		}

		// Call real function (of which this is a wrapper).
		c.createCall(fn.LLVMFn, forwardParams, "")
		c.builder.CreateRetVoid()
	}

	// Create wrapper for deferred interface call.
	for _, thunk := range c.deferInvokeFuncs {
		// This function gets a single parameter which is a pointer to a struct
		// (the defer frame).
		// This struct starts with the values of runtime._defer, but after that
		// follow the real function parameters.
		// The job of this wrapper is to extract these parameters and to call
		// the real function with them.
		llvmFn := c.mod.NamedFunction(thunk.method.FullName() + "$defer")
		llvmFn.SetLinkage(llvm.InternalLinkage)
		llvmFn.SetUnnamedAddr(true)
		entry := c.ctx.AddBasicBlock(llvmFn, "entry")
		c.builder.SetInsertPointAtEnd(entry)
		deferRawPtr := llvmFn.Param(0)

		// Get the real param type and cast to it.
		deferFrameType := c.ctx.StructType(thunk.valueTypes, false)
		deferFramePtr := c.builder.CreateBitCast(deferRawPtr, llvm.PointerType(deferFrameType, 0), "deferFrame")

		// Extract the params from the struct.
		forwardParams := []llvm.Value{}
		zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
		for i := range thunk.valueTypes[3:] {
			gep := c.builder.CreateGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(c.ctx.Int32Type(), uint64(i+3), false)}, "gep")
			forwardParam := c.builder.CreateLoad(gep, "param")
			forwardParams = append(forwardParams, forwardParam)
		}

		// Call real function (of which this is a wrapper).
		fnGEP := c.builder.CreateGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(c.ctx.Int32Type(), 2, false)}, "fn.gep")
		fn := c.builder.CreateLoad(fnGEP, "fn")
		c.createCall(fn, forwardParams, "")
		c.builder.CreateRetVoid()
	}

	// Create wrapper for deferred function pointer call.
	for _, thunk := range c.ctxDeferFuncs {
		// This function gets a single parameter which is a pointer to a struct
		// (the defer frame).
		// This struct starts with the values of runtime._defer, but after that
		// follows the closure and then the real parameters.
		// The job of this wrapper is to extract this closure and these
		// parameters and to call the function pointer with them.
		llvmFn := thunk.fn
		llvmFn.SetLinkage(llvm.InternalLinkage)
		llvmFn.SetUnnamedAddr(true)
		entry := c.ctx.AddBasicBlock(llvmFn, "entry")
		// TODO: set the debug location - perhaps the location of the rundefers
		// call?
		c.builder.SetInsertPointAtEnd(entry)
		deferRawPtr := llvmFn.Param(0)

		// Get the real param type and cast to it.
		deferFrameType := c.ctx.StructType(thunk.deferStruct, false)
		deferFramePtr := c.builder.CreateBitCast(deferRawPtr, llvm.PointerType(deferFrameType, 0), "defer.frame")

		// Extract the params from the struct.
		forwardParams := []llvm.Value{}
		zero := llvm.ConstInt(c.ctx.Int32Type(), 0, false)
		for i := 3; i < len(thunk.deferStruct); i++ {
			gep := c.builder.CreateGEP(deferFramePtr, []llvm.Value{zero, llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false)}, "")
			forwardParam := c.builder.CreateLoad(gep, "param")
			forwardParams = append(forwardParams, forwardParam)
		}

		// Extract the closure from the struct.
		fpGEP := c.builder.CreateGEP(deferFramePtr, []llvm.Value{
			zero,
			llvm.ConstInt(c.ctx.Int32Type(), 2, false),
			llvm.ConstInt(c.ctx.Int32Type(), 1, false),
		}, "closure.fp.ptr")
		fp := c.builder.CreateLoad(fpGEP, "closure.fp")
		contextGEP := c.builder.CreateGEP(deferFramePtr, []llvm.Value{
			zero,
			llvm.ConstInt(c.ctx.Int32Type(), 2, false),
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
		}, "closure.context.ptr")
		context := c.builder.CreateLoad(contextGEP, "closure.context")
		forwardParams = append(forwardParams, context)

		// Cast the function pointer in the closure to the correct function
		// pointer type.
		closureType, err := c.getLLVMType(thunk.signature)
		if err != nil {
			return err
		}
		fpType := closureType.StructElementTypes()[1]
		fpCast := c.builder.CreateBitCast(fp, fpType, "closure.fp.cast")

		// Call real function (of which this is a wrapper).
		c.createCall(fpCast, forwardParams, "")
		c.builder.CreateRetVoid()
	}

	return nil
}
