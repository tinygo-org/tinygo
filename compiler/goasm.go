package compiler

// ...

// The stack in a Go function call looks like this:
//                      <- SP of caller
//   parameter values
//   return values
//   return address     <- SP of callee

import (
	"strings"

	"tinygo.org/x/go-llvm"
)

// arch returns the effective architecture (in the Go convention). So for
// example, if the triple is x86_64--linux, the returned architecture is amd64.
func (c *compilerContext) arch() string {
	arch := strings.Split(c.Triple, "-")[0]
	switch {
	case arch == "x86_64":
		return "amd64"
	case arch == "aarch64":
		return "arm64"
	default:
		return arch // unknown arch, return the first part as a guess
	}
}

// createGoAsmWrapper builds a wrapper function for a function written in Go
// assembly (and using the Go ABI0 ABI).
func (b *builder) createGoAsmWrapper(forwardName string) {
	entryBlock := llvm.AddBasicBlock(b.llvmFn, "entry")
	b.Builder.SetInsertPointAtEnd(entryBlock)
	if b.Debug {
		if b.fn.Syntax() != nil {
			b.difunc = b.attachDebugInfo(b.fn)
		}
		pos := b.program.Fset.Position(b.fn.Pos())
		b.SetCurrentDebugLocation(uint(pos.Line), 0, b.difunc, llvm.Metadata{})
	}

	b.loadFunctionParams(entryBlock)

	// Determine the stack layout that's used for the Go ABI.
	// The layout roughly follows this convention (from low to high address):
	//   - empty space (for the return pointer?)
	//   - parameters
	//   - return values
	// These structs are aligned: the return value will be aligned even if both
	// parameters and return values are of byte type for example.
	var paramFields []llvm.Type
	if b.arch() == "arm64" {
		paramFields = append(paramFields, b.uintptrType)
	}
	paramOffset := len(paramFields)
	for _, param := range b.fn.Params {
		llvmValue := b.getValue(param)
		llvmType := llvmValue.Type()
		paramFields = append(paramFields, llvmType)
	}
	allocaFields := []llvm.Type{
		b.ctx.StructType(paramFields, false),
	}
	returnType := b.llvmFn.Type().ElementType().ReturnType()
	if returnType != b.ctx.VoidType() {
		allocaFields = append(allocaFields, returnType)
	}
	allocaType := b.ctx.StructType(allocaFields, false)

	// Create the alloca.
	alloca := b.CreateAlloca(allocaType, "callframe")

	// Get the stack pointer. WARNING: this is not always the same as the alloca
	// above. We're guessing here that the alloca is at the top of the stack and
	// therefore we can use this space for the call sequence.
	allocaPtrType := llvm.PointerType(allocaType, 0)
	asmType := llvm.FunctionType(allocaPtrType, []llvm.Type{allocaPtrType}, false)
	var inlineAsm llvm.Value
	switch b.arch() {
	case "amd64":
		inlineAsm = llvm.InlineAsm(asmType, "movq %rsp, ${0}", "=r,r", false, false, 0, false)
	case "arm64":
		inlineAsm = llvm.InlineAsm(asmType, "mov ${0}, sp", "=r,r", false, false, 0, false)
	default:
		panic("foobar")
	}
	stackPointer := b.CreateCall(inlineAsm, []llvm.Value{alloca}, "")

	// Store parameters at the top of the stack.
	for i, param := range b.fn.Params {
		gep := b.CreateGEP(stackPointer, []llvm.Value{
			llvm.ConstInt(b.ctx.Int32Type(), 0, false),
			llvm.ConstInt(b.ctx.Int32Type(), 0, false),
			llvm.ConstInt(b.ctx.Int32Type(), uint64(paramOffset+i), false),
		}, "")
		b.CreateStore(b.getValue(param), gep)
	}

	// Call the Go assembly!
	fn := b.mod.NamedFunction(forwardName)
	if fn.IsNil() {
		fnType := llvm.FunctionType(b.ctx.VoidType(), nil, false)
		fn = llvm.AddFunction(b.mod, forwardName, fnType)
	}
	b.CreateCall(fn, nil, "")

	// Clobber necessary registers after the call.
	// This happens _during_ the call, not afterwards, but hopefully LLVM won't
	// insert any instructions in this gap.
	asmType = llvm.FunctionType(allocaPtrType, []llvm.Type{allocaPtrType}, false)
	switch b.arch() {
	case "amd64":
		// https://en.wikipedia.org/wiki/X86_calling_conventions#System_V_AMD64_ABI
		inlineAsm = llvm.InlineAsm(asmType, "movq %rsp, ${0}", "=r,r,~{rbx},~{r12},~{r13},~{r14},~{r15},~{memory}", true, true, 0, false)
	case "arm64":
		// https://developer.arm.com/documentation/ihi0055/d/
		// Save x19-x28, d8-d15.
		inlineAsm = llvm.InlineAsm(asmType, "mov ${0}, sp", "=r,r,~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{d8},~{d9},~{d10},~{d11},~{d12},~{d13},~{d14},~{d15},~{memory}", true, true, 0, false)
	}
	stackPointer = b.CreateCall(inlineAsm, []llvm.Value{alloca}, "")

	// Return the resulting value.
	if returnType != b.ctx.VoidType() {
		// There is a value to return. Load it and return it.
		resultGEP := b.CreateGEP(stackPointer, []llvm.Value{
			llvm.ConstInt(b.ctx.Int32Type(), 0, false),
			llvm.ConstInt(b.ctx.Int32Type(), 1, false),
		}, "result.gep")
		result := b.CreateLoad(resultGEP, "result")
		b.CreateRet(result)
	} else {
		// There is no value to return. Return void instead.
		b.CreateRetVoid()
	}

	// With all the unsafe stuff we do above, it seems better to mark this
	// function as noinline so that optimizations won't interfere too much with
	// it.
	noinline := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("noinline"), 0)
	b.llvmFn.AddFunctionAttr(noinline)
}

func (b *builder) createGoAsmExport(forwardName string) {
	// Create function that reads incoming parameters from the stack.
	llvmFnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{b.i8ptrType}, false)
	llvmFn := llvm.AddFunction(b.mod, b.info.linkName+"$goasmwrapper", llvmFnType)
	llvmFn.SetLinkage(llvm.InternalLinkage)
	noinline := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("noinline"), 0)
	llvmFn.AddFunctionAttr(noinline)

	bb := llvm.AddBasicBlock(llvmFn, "entry")
	b.SetInsertPointAtEnd(bb)
	if b.Debug {
		pos := b.program.Fset.Position(b.fn.Pos())
		difunc := b.attachDebugInfoRaw(b.fn, llvmFn, "$goasmwrapper", pos.Filename, pos.Line)
		b.SetCurrentDebugLocation(uint(pos.Line), 0, difunc, llvm.Metadata{})
	}

	var params []llvm.Value
	offset := uint64(0)
	switch b.arch() {
	case "amd64", "arm64":
		offset += 8 // return pointer is at the top of the stack
	}
	sp := llvmFn.Param(0)
	for _, param := range b.fn.Params {
		llvmType := b.getLLVMType(param.Type())
		gep := b.CreateGEP(sp, []llvm.Value{
			llvm.ConstInt(b.ctx.Int8Type(), offset, false),
		}, param.Name()+".gep")
		bitcast := b.CreateBitCast(gep, llvm.PointerType(llvmType, 0), param.Name()+".cast")
		value := b.CreateLoad(bitcast, param.Name())
		params = append(params, value)
		offset += b.targetData.TypeAllocSize(llvmType) // TODO: alignment
	}

	params = append(params, llvm.ConstNull(b.i8ptrType)) // context

	result := b.createCall(b.llvmFn, params, "result")
	gep := b.CreateGEP(sp, []llvm.Value{
		llvm.ConstInt(b.ctx.Int8Type(), offset, false),
	}, "result.gep")
	bitcast := b.CreateBitCast(gep, llvm.PointerType(result.Type(), 0), "result.cast")
	b.CreateStore(result, bitcast)

	b.CreateRetVoid()

	// this could be done easier with prologue data
	b.createGoAsmSPForward(forwardName, llvmFn)
}

func (b *builder) createGoAsmSPForward(forwardName string, forwardFunction llvm.Value) {
	// Create function that forwards the stack pointer.
	llvmFnType := llvm.FunctionType(b.ctx.VoidType(), nil, false)
	llvmFn := llvm.AddFunction(b.mod, forwardName, llvmFnType)
	bb := llvm.AddBasicBlock(llvmFn, "entry")
	b.Builder.SetInsertPointAtEnd(bb)
	if b.Debug {
		pos := b.program.Fset.Position(b.fn.Pos())
		b.SetCurrentDebugLocation(uint(pos.Line), 0, b.difunc, llvm.Metadata{})
	}
	asmType := llvm.FunctionType(b.i8ptrType, nil, false)
	var asmString string
	switch b.arch() {
	case "amd64":
		asmString = "mov %rsp, $0"
	case "arm64":
		asmString = "mov x0, sp"
	default:
		b.addError(b.fn.Pos(), "cannot wrap Go assembly: unknown architecture")
		b.CreateUnreachable()
		return
	}
	asm := llvm.InlineAsm(asmType, asmString, "=r", false, false, 0, false)
	sp := b.CreateCall(asm, nil, "sp")
	call := b.CreateCall(forwardFunction, []llvm.Value{sp}, "")
	call.SetTailCall(true)
	b.CreateRetVoid()

	// Make sure no prologue/epilogue is created.
	naked := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("naked"), 0)
	llvmFn.AddFunctionAttr(naked)
}
