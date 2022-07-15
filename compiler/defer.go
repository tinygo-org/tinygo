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
	"strconv"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// supportsRecover returns whether the compiler supports the recover() builtin
// for the current architecture.
func (b *builder) supportsRecover() bool {
	switch b.archFamily() {
	case "wasm32":
		// Probably needs to be implemented using the exception handling
		// proposal of WebAssembly:
		// https://github.com/WebAssembly/exception-handling
		return false
	case "riscv64", "xtensa":
		// TODO: add support for these architectures
		return false
	default:
		return true
	}
}

// hasDeferFrame returns whether the current function needs to catch panics and
// run defers.
func (b *builder) hasDeferFrame() bool {
	if b.fn.Recover == nil {
		return false
	}
	return b.supportsRecover()
}

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

	if b.hasDeferFrame() {
		// Set up the defer frame with the current stack pointer.
		// This assumes that the stack pointer doesn't move outside of the
		// function prologue/epilogue (an invariant maintained by TinyGo but
		// possibly broken by the C alloca function).
		// The frame pointer is _not_ saved, because it is marked as clobbered
		// in the setjmp-like inline assembly.
		deferFrameType := b.getLLVMRuntimeType("deferFrame")
		b.deferFrame = b.CreateAlloca(deferFrameType, "deferframe.buf")
		stackPointer := b.readStackPointer()
		b.createRuntimeCall("setupDeferFrame", []llvm.Value{b.deferFrame, stackPointer}, "")

		// Create the landing pad block, which is where control transfers after
		// a panic.
		b.landingpad = b.ctx.AddBasicBlock(b.llvmFn, "lpad")
	}
}

// createLandingPad fills in the landing pad block. This block runs the deferred
// functions and returns (by jumping to the recover block). If the function is
// still panicking after the defers are run, the panic will be re-raised in
// destroyDeferFrame.
func (b *builder) createLandingPad() {
	b.SetInsertPointAtEnd(b.landingpad)

	// Add debug info, if needed.
	// The location used is the closing bracket of the function.
	if b.Debug {
		pos := b.program.Fset.Position(b.fn.Syntax().End())
		b.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), b.difunc, llvm.Metadata{})
	}

	b.createRunDefers()

	// Continue at the 'recover' block, which returns to the parent in an
	// appropriate way.
	b.CreateBr(b.blockEntries[b.fn.Recover])
}

// createInvokeCheckpoint saves the function state at the given point, to
// continue at the landing pad if a panic happened. This is implemented using a
// setjmp-like construct.
func (b *builder) createInvokeCheckpoint() {
	// Construct inline assembly equivalents of setjmp.
	// The assembly works as follows:
	//   * All registers (both callee-saved and caller saved) are clobbered
	//     after the inline assembly returns.
	//   * The assembly stores the address just past the end of the assembly
	//     into the jump buffer.
	//   * The return value (eax, rax, r0, etc) is set to zero in the inline
	//     assembly but set to an unspecified non-zero value when jumping using
	//     a longjmp.
	var asmString, constraints string
	resultType := b.uintptrType
	switch b.archFamily() {
	case "i386":
		asmString = `
xorl %eax, %eax
movl $$1f, 4(%ebx)
1:`
		constraints = "={eax},{ebx},~{ebx},~{ecx},~{edx},~{esi},~{edi},~{ebp},~{xmm0},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory}"
		// This doesn't include the floating point stack because TinyGo uses
		// newer floating point instructions.
	case "x86_64":
		asmString = `
leaq 1f(%rip), %rax
movq %rax, 8(%rbx)
xorq %rax, %rax
1:`
		constraints = "={rax},{rbx},~{rbx},~{rcx},~{rdx},~{rsi},~{rdi},~{rbp},~{r8},~{r9},~{r10},~{r11},~{r12},~{r13},~{r14},~{r15},~{xmm0},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{xmm8},~{xmm9},~{xmm10},~{xmm11},~{xmm12},~{xmm13},~{xmm14},~{xmm15},~{xmm16},~{xmm17},~{xmm18},~{xmm19},~{xmm20},~{xmm21},~{xmm22},~{xmm23},~{xmm24},~{xmm25},~{xmm26},~{xmm27},~{xmm28},~{xmm29},~{xmm30},~{xmm31},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory}"
		// This list doesn't include AVX/AVX512 registers because TinyGo
		// doesn't currently enable support for AVX instructions.
	case "arm":
		// Note: the following assembly takes into account that the PC is
		// always 4 bytes ahead on ARM. The PC that is stored always points
		// to the instruction just after the assembly fragment so that
		// tinygo_longjmp lands at the correct instruction.
		if b.isThumb() {
			// Instructions are 2 bytes in size.
			asmString = `
movs r0, #0
mov r2, pc
str r2, [r1, #4]`
		} else {
			// Instructions are 4 bytes in size.
			asmString = `
str pc, [r1, #4]
movs r0, #0`
		}
		constraints = "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"
	case "aarch64":
		asmString = `
adr x2, 1f
str x2, [x1, #8]
mov x0, #0
1:
`
		constraints = "={x0},{x1},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{q16},~{q17},~{q18},~{q19},~{q20},~{q21},~{q22},~{q23},~{q24},~{q25},~{q26},~{q27},~{q28},~{q29},~{q30},~{nzcv},~{ffr},~{vg},~{memory}"
		if b.GOOS != "darwin" {
			// These registers cause the following warning when compiling for
			// MacOS:
			//     warning: inline asm clobber list contains reserved registers:
			//     X18, FP
			//     Reserved registers on the clobber list may not be preserved
			//     across the asm statement, and clobbering them may lead to
			//     undefined behaviour.
			constraints += ",~{x18},~{fp}"
		}
		// TODO: SVE registers, which we don't use in TinyGo at the moment.
	case "avr":
		// Note: the Y register (R28:R29) is a fixed register and therefore
		// needs to be saved manually. TODO: do this only once per function with
		// a defer frame, not for every call.
		resultType = b.ctx.Int8Type()
		asmString = `
ldi r24, pm_lo8(1f)
ldi r25, pm_hi8(1f)
std z+2, r24
std z+3, r25
std z+4, r28
std z+5, r29
ldi r24, 0
1:`
		constraints = "={r24},z,~{r0},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{r13},~{r14},~{r15},~{r16},~{r17},~{r18},~{r19},~{r20},~{r21},~{r22},~{r23},~{r25},~{r26},~{r27}"
	case "riscv32":
		asmString = `
la a2, 1f
sw a2, 4(a1)
li a0, 0
1:`
		constraints = "={a0},{a1},~{a1},~{a2},~{a3},~{a4},~{a5},~{a6},~{a7},~{s0},~{s1},~{s2},~{s3},~{s4},~{s5},~{s6},~{s7},~{s8},~{s9},~{s10},~{s11},~{t0},~{t1},~{t2},~{t3},~{t4},~{t5},~{t6},~{ra},~{f0},~{f1},~{f2},~{f3},~{f4},~{f5},~{f6},~{f7},~{f8},~{f9},~{f10},~{f11},~{f12},~{f13},~{f14},~{f15},~{f16},~{f17},~{f18},~{f19},~{f20},~{f21},~{f22},~{f23},~{f24},~{f25},~{f26},~{f27},~{f28},~{f29},~{f30},~{f31},~{memory}"
	default:
		// This case should have been handled by b.supportsRecover().
		b.addError(b.fn.Pos(), "unknown architecture for defer: "+b.archFamily())
	}
	asmType := llvm.FunctionType(resultType, []llvm.Type{b.deferFrame.Type()}, false)
	asm := llvm.InlineAsm(asmType, asmString, constraints, false, false, 0, false)
	result := b.CreateCall(asm, []llvm.Value{b.deferFrame}, "setjmp")
	result.AddCallSiteAttribute(-1, b.ctx.CreateEnumAttribute(llvm.AttributeKindID("returns_twice"), 0))
	isZero := b.CreateICmp(llvm.IntEQ, result, llvm.ConstInt(resultType, 0, false), "setjmp.result")
	continueBB := b.insertBasicBlock("")
	b.CreateCondBr(isZero, continueBB, b.landingpad)
	b.SetInsertPointAtEnd(continueBB)
	b.blockExits[b.currentBlock] = continueBB
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

	// Make a struct out of the collected values to put in the deferred call
	// struct.
	deferredCallType := b.ctx.StructType(valueTypes, false)
	deferredCall := llvm.ConstNull(deferredCallType)
	for i, value := range values {
		deferredCall = b.CreateInsertValue(deferredCall, value, i, "")
	}

	// Put this struct in an allocation.
	var alloca llvm.Value
	if !isInLoop(instr.Block()) {
		// This can safely use a stack allocation.
		alloca = llvmutil.CreateEntryBlockAlloca(b.Builder, deferredCallType, "defer.alloca")
	} else {
		// This may be hit a variable number of times, so use a heap allocation.
		size := b.targetData.TypeAllocSize(deferredCallType)
		sizeValue := llvm.ConstInt(b.uintptrType, size, false)
		nilPtr := llvm.ConstNull(b.i8ptrType)
		allocCall := b.createRuntimeCall("alloc", []llvm.Value{sizeValue, nilPtr}, "defer.alloc.call")
		alloca = b.CreateBitCast(allocCall, llvm.PointerType(deferredCallType, 0), "defer.alloc")
	}
	if b.NeedsStackObjects {
		b.trackPointer(alloca)
	}
	b.CreateStore(deferredCall, alloca)

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

	// Create loop, in the order: loophead, loop, callback0, callback1, ..., unreachable, end.
	end := b.insertBasicBlock("rundefers.end")
	unreachable := b.ctx.InsertBasicBlock(end, "rundefers.default")
	loop := b.ctx.InsertBasicBlock(unreachable, "rundefers.loop")
	loophead := b.ctx.InsertBasicBlock(loop, "rundefers.loophead")
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
		block := b.insertBasicBlock("rundefers.callback" + strconv.Itoa(i))
		sw.AddCase(llvm.ConstInt(b.uintptrType, uint64(i), false), block)
		b.SetInsertPointAtEnd(block)
		switch callback := callback.(type) {
		case *ssa.CallCommon:
			// Call on an value or interface value.

			// Get the real defer struct type and cast to it.
			valueTypes := []llvm.Type{b.uintptrType, llvm.PointerType(b.getLLVMRuntimeType("_defer"), 0)}

			if !callback.IsInvoke() {
				//Expect funcValue to be passed through the deferred call.
				valueTypes = append(valueTypes, b.getFuncType(callback.Signature()))
			} else {
				//Expect typecode
				valueTypes = append(valueTypes, b.uintptrType, b.i8ptrType)
			}

			for _, arg := range callback.Args {
				valueTypes = append(valueTypes, b.getLLVMType(arg.Type()))
			}

			deferredCallType := b.ctx.StructType(valueTypes, false)
			deferredCallPtr := b.CreateBitCast(deferData, llvm.PointerType(deferredCallType, 0), "defercall")

			// Extract the params from the struct (including receiver).
			forwardParams := []llvm.Value{}
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := 2; i < len(valueTypes); i++ {
				gep := b.CreateInBoundsGEP(deferredCallPtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(i), false)}, "gep")
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
			deferredCallType := b.ctx.StructType(valueTypes, false)
			deferredCallPtr := b.CreateBitCast(deferData, llvm.PointerType(deferredCallType, 0), "defercall")

			// Extract the params from the struct.
			forwardParams := []llvm.Value{}
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := range getParams(callback.Signature) {
				gep := b.CreateInBoundsGEP(deferredCallPtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(i+2), false)}, "gep")
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

			// Call real function.
			b.createInvoke(b.getFunction(callback), forwardParams, "")

		case *ssa.MakeClosure:
			// Get the real defer struct type and cast to it.
			fn := callback.Fn.(*ssa.Function)
			valueTypes := []llvm.Type{b.uintptrType, llvm.PointerType(b.getLLVMRuntimeType("_defer"), 0)}
			params := fn.Signature.Params()
			for i := 0; i < params.Len(); i++ {
				valueTypes = append(valueTypes, b.getLLVMType(params.At(i).Type()))
			}
			valueTypes = append(valueTypes, b.i8ptrType) // closure
			deferredCallType := b.ctx.StructType(valueTypes, false)
			deferredCallPtr := b.CreateBitCast(deferData, llvm.PointerType(deferredCallType, 0), "defercall")

			// Extract the params from the struct.
			forwardParams := []llvm.Value{}
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := 2; i < len(valueTypes); i++ {
				gep := b.CreateInBoundsGEP(deferredCallPtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(i), false)}, "")
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

			deferredCallType := b.ctx.StructType(valueTypes, false)
			deferredCallPtr := b.CreateBitCast(deferData, llvm.PointerType(deferredCallType, 0), "defercall")

			// Extract the params from the struct.
			var argValues []llvm.Value
			zero := llvm.ConstInt(b.ctx.Int32Type(), 0, false)
			for i := 0; i < params.Len(); i++ {
				gep := b.CreateInBoundsGEP(deferredCallPtr, []llvm.Value{zero, llvm.ConstInt(b.ctx.Int32Type(), uint64(i+2), false)}, "gep")
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
