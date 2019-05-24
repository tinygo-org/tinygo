package compiler

// This file lowers goroutine pseudo-functions into coroutines scheduled by a
// scheduler at runtime. It uses coroutine support in LLVM for this
// transformation: https://llvm.org/docs/Coroutines.html
//
// For example, take the following code:
//
//     func main() {
//         go foo()
//         time.Sleep(2 * time.Second)
//         println("some other operation")
//         i := bar()
//         println("done", *i)
//     }
//
//     func foo() {
//         for {
//             println("foo!")
//             time.Sleep(time.Second)
//         }
//     }
//
//     func bar() *int {
//         time.Sleep(time.Second)
//         println("blocking operation completed)
//         return new(int)
//     }
//
// It is transformed by the IR generator in compiler.go into the following
// pseudo-Go code:
//
//     func main() {
//         fn := runtime.makeGoroutine(foo)
//         fn()
//         time.Sleep(2 * time.Second)
//         println("some other operation")
//         i := bar() // imagine an 'await' keyword in front of this call
//         println("done", *i)
//     }
//
//     func foo() {
//         for {
//             println("foo!")
//             time.Sleep(time.Second)
//         }
//     }
//
//     func bar() *int {
//         time.Sleep(time.Second)
//         println("blocking operation completed)
//         return new(int)
//     }
//
// The pass in this file transforms this code even further, to the following
// async/await style pseudocode:
//
//     func main(parent) {
//         hdl := llvm.makeCoroutine()
//         foo(nil)                                // do not pass the parent coroutine: this is an independent goroutine
//         runtime.sleepTask(hdl, 2 * time.Second) // ask the scheduler to re-activate this coroutine at the right time
//         llvm.suspend(hdl)                       // suspend point
//         println("some other operation")
//         var i *int                              // allocate space on the stack for the return value
//         runtime.setTaskPromisePtr(hdl, &i)      // store return value alloca in our coroutine promise
//         bar(hdl)                                // await, pass a continuation (hdl) to bar
//         llvm.suspend(hdl)                       // suspend point, wait for the callee to re-activate
//         println("done", *i)
//         runtime.activateTask(parent)            // re-activate the parent (nop, there is no parent)
//     }
//
//     func foo(parent) {
//         hdl := llvm.makeCoroutine()
//         for {
//             println("foo!")
//             runtime.sleepTask(hdl, time.Second) // ask the scheduler to re-activate this coroutine at the right time
//             llvm.suspend(hdl)                   // suspend point
//         }
//     }
//
//     func bar(parent) {
//         hdl := llvm.makeCoroutine()
//         runtime.sleepTask(hdl, time.Second) // ask the scheduler to re-activate this coroutine at the right time
//         llvm.suspend(hdl)                   // suspend point
//         println("blocking operation completed)
//         runtime.activateTask(parent)        // re-activate the parent coroutine before returning
//     }
//
// The real LLVM code is more complicated, but this is the general idea.
//
// The LLVM coroutine passes will then process this file further transforming
// these three functions into coroutines. Most of the actual work is done by the
// scheduler, which runs in the background scheduling all coroutines.

import (
	"errors"
	"strings"

	"tinygo.org/x/go-llvm"
)

type asyncFunc struct {
	taskHandle       llvm.Value
	cleanupBlock     llvm.BasicBlock
	suspendBlock     llvm.BasicBlock
	unreachableBlock llvm.BasicBlock
}

// LowerGoroutines is a pass called during optimization that transforms the IR
// into one where all blocking functions are turned into goroutines and blocking
// calls into await calls.
func (c *Compiler) LowerGoroutines() error {
	needsScheduler, err := c.markAsyncFunctions()
	if err != nil {
		return err
	}

	uses := getUses(c.mod.NamedFunction("runtime.callMain"))
	if len(uses) != 1 || uses[0].IsACallInst().IsNil() {
		panic("expected exactly 1 call of runtime.callMain, check the entry point")
	}
	mainCall := uses[0]

	// Replace call of runtime.callMain() with a real call to main.main(),
	// optionally followed by a call to runtime.scheduler().
	c.builder.SetInsertPointBefore(mainCall)
	realMain := c.mod.NamedFunction(c.ir.MainPkg().Pkg.Path() + ".main")
	c.builder.CreateCall(realMain, []llvm.Value{llvm.Undef(c.i8ptrType), llvm.ConstPointerNull(c.i8ptrType)}, "")
	if needsScheduler {
		c.createRuntimeCall("scheduler", nil, "")
	}
	mainCall.EraseFromParentAsInstruction()

	if !needsScheduler {
		go_scheduler := c.mod.NamedFunction("go_scheduler")
		if !go_scheduler.IsNil() {
			// This is the WebAssembly backend.
			// There is no need to export the go_scheduler function, but it is
			// still exported. Make sure it is optimized away.
			go_scheduler.SetLinkage(llvm.InternalLinkage)
		}
	}

	// main.main was set to external linkage during IR construction. Set it to
	// internal linkage to enable interprocedural optimizations.
	realMain.SetLinkage(llvm.InternalLinkage)
	c.mod.NamedFunction("runtime.alloc").SetLinkage(llvm.InternalLinkage)
	c.mod.NamedFunction("runtime.free").SetLinkage(llvm.InternalLinkage)
	c.mod.NamedFunction("runtime.sleepTask").SetLinkage(llvm.InternalLinkage)
	c.mod.NamedFunction("runtime.setTaskPromisePtr").SetLinkage(llvm.InternalLinkage)
	c.mod.NamedFunction("runtime.getTaskPromisePtr").SetLinkage(llvm.InternalLinkage)
	c.mod.NamedFunction("runtime.scheduler").SetLinkage(llvm.InternalLinkage)

	return nil
}

// markAsyncFunctions does the bulk of the work of lowering goroutines. It
// determines whether a scheduler is needed, and if it is, it transforms
// blocking operations into goroutines and blocking calls into await calls.
//
// It does the following operations:
//    * Find all blocking functions.
//    * Determine whether a scheduler is necessary. If not, it skips the
//      following operations.
//    * Transform call instructions into await calls.
//    * Transform return instructions into final suspends.
//    * Set up the coroutine frames for async functions.
//    * Transform blocking calls into their async equivalents.
func (c *Compiler) markAsyncFunctions() (needsScheduler bool, err error) {
	var worklist []llvm.Value

	sleep := c.mod.NamedFunction("time.Sleep")
	if !sleep.IsNil() {
		worklist = append(worklist, sleep)
	}
	deadlockStub := c.mod.NamedFunction("runtime.deadlockStub")
	if !deadlockStub.IsNil() {
		worklist = append(worklist, deadlockStub)
	}
	chanSend := c.mod.NamedFunction("runtime.chanSend")
	if !chanSend.IsNil() {
		worklist = append(worklist, chanSend)
	}
	chanRecv := c.mod.NamedFunction("runtime.chanRecv")
	if !chanRecv.IsNil() {
		worklist = append(worklist, chanRecv)
	}

	if len(worklist) == 0 {
		// There are no blocking operations, so no need to transform anything.
		return false, c.lowerMakeGoroutineCalls()
	}

	// Find all async functions.
	// Keep reducing this worklist by marking a function as recursively async
	// from the worklist and pushing all its parents that are non-async.
	// This is somewhat similar to a worklist in a mark-sweep garbage collector:
	// the work items are then grey objects.
	asyncFuncs := make(map[llvm.Value]*asyncFunc)
	asyncList := make([]llvm.Value, 0, 4)
	for len(worklist) != 0 {
		// Pick the topmost.
		f := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		if _, ok := asyncFuncs[f]; ok {
			continue // already processed
		}
		// Add to set of async functions.
		asyncFuncs[f] = &asyncFunc{}
		asyncList = append(asyncList, f)

		// Add all callees to the worklist.
		for _, use := range getUses(f) {
			if use.IsConstant() && use.Opcode() == llvm.BitCast {
				bitcastUses := getUses(use)
				for _, call := range bitcastUses {
					if call.IsACallInst().IsNil() || call.CalledValue().Name() != "runtime.makeGoroutine" {
						return false, errors.New("async function " + f.Name() + " incorrectly used in bitcast, expected runtime.makeGoroutine")
					}
				}
				// This is a go statement. Do not mark the parent as async, as
				// starting a goroutine is not a blocking operation.
				continue
			}
			if use.IsACallInst().IsNil() {
				// Not a call instruction. Maybe a store to a global? In any
				// case, this requires support for async calls across function
				// pointers which is not yet supported.
				return false, errors.New("async function " + f.Name() + " used as function pointer")
			}
			parent := use.InstructionParent().Parent()
			for i := 0; i < use.OperandsCount()-1; i++ {
				if use.Operand(i) == f {
					return false, errors.New("async function " + f.Name() + " used as function pointer in " + parent.Name())
				}
			}
			worklist = append(worklist, parent)
		}
	}

	// Check whether a scheduler is needed.
	makeGoroutine := c.mod.NamedFunction("runtime.makeGoroutine")
	if c.GOOS == "js" && strings.HasPrefix(c.Triple, "wasm") {
		// JavaScript always needs a scheduler, as in general no blocking
		// operations are possible. Blocking operations block the browser UI,
		// which is very bad.
		needsScheduler = true
	} else {
		// Only use a scheduler when an async goroutine is started. When the
		// goroutine is not async (does not do any blocking operation), no
		// scheduler is necessary as it can be called directly.
		for _, use := range getUses(makeGoroutine) {
			// Input param must be const bitcast of function.
			bitcast := use.Operand(0)
			if !bitcast.IsConstant() || bitcast.Opcode() != llvm.BitCast {
				panic("expected const bitcast operand of runtime.makeGoroutine")
			}
			goroutine := bitcast.Operand(0)
			if _, ok := asyncFuncs[goroutine]; ok {
				needsScheduler = true
				break
			}
		}
	}

	if !needsScheduler {
		// No scheduler is needed. Do not transform all functions here.
		// However, make sure that all go calls (which are all non-async) are
		// transformed into regular calls.
		return false, c.lowerMakeGoroutineCalls()
	}

	// Create a few LLVM intrinsics for coroutine support.

	coroIdType := llvm.FunctionType(c.ctx.TokenType(), []llvm.Type{c.ctx.Int32Type(), c.i8ptrType, c.i8ptrType, c.i8ptrType}, false)
	coroIdFunc := llvm.AddFunction(c.mod, "llvm.coro.id", coroIdType)

	coroSizeType := llvm.FunctionType(c.ctx.Int32Type(), nil, false)
	coroSizeFunc := llvm.AddFunction(c.mod, "llvm.coro.size.i32", coroSizeType)

	coroBeginType := llvm.FunctionType(c.i8ptrType, []llvm.Type{c.ctx.TokenType(), c.i8ptrType}, false)
	coroBeginFunc := llvm.AddFunction(c.mod, "llvm.coro.begin", coroBeginType)

	coroSuspendType := llvm.FunctionType(c.ctx.Int8Type(), []llvm.Type{c.ctx.TokenType(), c.ctx.Int1Type()}, false)
	coroSuspendFunc := llvm.AddFunction(c.mod, "llvm.coro.suspend", coroSuspendType)

	coroEndType := llvm.FunctionType(c.ctx.Int1Type(), []llvm.Type{c.i8ptrType, c.ctx.Int1Type()}, false)
	coroEndFunc := llvm.AddFunction(c.mod, "llvm.coro.end", coroEndType)

	coroFreeType := llvm.FunctionType(c.i8ptrType, []llvm.Type{c.ctx.TokenType(), c.i8ptrType}, false)
	coroFreeFunc := llvm.AddFunction(c.mod, "llvm.coro.free", coroFreeType)

	// Transform all async functions into coroutines.
	for _, f := range asyncList {
		if f == sleep || f == deadlockStub || f == chanSend || f == chanRecv {
			continue
		}

		frame := asyncFuncs[f]
		frame.cleanupBlock = c.ctx.AddBasicBlock(f, "task.cleanup")
		frame.suspendBlock = c.ctx.AddBasicBlock(f, "task.suspend")
		frame.unreachableBlock = c.ctx.AddBasicBlock(f, "task.unreachable")

		// Scan for async calls and return instructions that need to have
		// suspend points inserted.
		var asyncCalls []llvm.Value
		var returns []llvm.Value
		for bb := f.EntryBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
			for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
				if !inst.IsACallInst().IsNil() {
					callee := inst.CalledValue()
					if _, ok := asyncFuncs[callee]; !ok || callee == sleep || callee == deadlockStub || callee == chanSend || callee == chanRecv {
						continue
					}
					asyncCalls = append(asyncCalls, inst)
				} else if !inst.IsAReturnInst().IsNil() {
					returns = append(returns, inst)
				}
			}
		}

		// Coroutine setup.
		c.builder.SetInsertPointBefore(f.EntryBasicBlock().FirstInstruction())
		taskState := c.builder.CreateAlloca(c.mod.GetTypeByName("runtime.taskState"), "task.state")
		stateI8 := c.builder.CreateBitCast(taskState, c.i8ptrType, "task.state.i8")
		id := c.builder.CreateCall(coroIdFunc, []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			stateI8,
			llvm.ConstNull(c.i8ptrType),
			llvm.ConstNull(c.i8ptrType),
		}, "task.token")
		size := c.builder.CreateCall(coroSizeFunc, nil, "task.size")
		if c.targetData.TypeAllocSize(size.Type()) > c.targetData.TypeAllocSize(c.uintptrType) {
			size = c.builder.CreateTrunc(size, c.uintptrType, "task.size.uintptr")
		} else if c.targetData.TypeAllocSize(size.Type()) < c.targetData.TypeAllocSize(c.uintptrType) {
			size = c.builder.CreateZExt(size, c.uintptrType, "task.size.uintptr")
		}
		data := c.createRuntimeCall("alloc", []llvm.Value{size}, "task.data")
		frame.taskHandle = c.builder.CreateCall(coroBeginFunc, []llvm.Value{id, data}, "task.handle")

		// Modify async calls so this function suspends right after the child
		// returns, because the child is probably not finished yet. Wait until
		// the child reactivates the parent.
		for _, inst := range asyncCalls {
			inst.SetOperand(inst.OperandsCount()-2, frame.taskHandle)

			// Split this basic block.
			await := c.splitBasicBlock(inst, llvm.NextBasicBlock(c.builder.GetInsertBlock()), "task.await")

			// Allocate space for the return value.
			var retvalAlloca llvm.Value
			if inst.Type().TypeKind() != llvm.VoidTypeKind {
				c.builder.SetInsertPointBefore(inst.InstructionParent().Parent().EntryBasicBlock().FirstInstruction())
				retvalAlloca = c.builder.CreateAlloca(inst.Type(), "coro.retvalAlloca")
				c.builder.SetInsertPointBefore(inst)
				data := c.builder.CreateBitCast(retvalAlloca, c.i8ptrType, "")
				c.createRuntimeCall("setTaskPromisePtr", []llvm.Value{frame.taskHandle, data}, "")
			}

			// Suspend.
			c.builder.SetInsertPointAtEnd(inst.InstructionParent())
			continuePoint := c.builder.CreateCall(coroSuspendFunc, []llvm.Value{
				llvm.ConstNull(c.ctx.TokenType()),
				llvm.ConstInt(c.ctx.Int1Type(), 0, false),
			}, "")
			sw := c.builder.CreateSwitch(continuePoint, frame.suspendBlock, 2)
			sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 0, false), await)
			sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 1, false), frame.cleanupBlock)

			if inst.Type().TypeKind() != llvm.VoidTypeKind {
				// Load the return value from the alloca. The callee has
				// written the return value to it.
				c.builder.SetInsertPointBefore(await.FirstInstruction())
				retval := c.builder.CreateLoad(retvalAlloca, "coro.retval")
				inst.ReplaceAllUsesWith(retval)
			}
		}

		// Replace return instructions with suspend points that should
		// reactivate the parent coroutine.
		for _, inst := range returns {
			// These properties were added by the functionattrs pass. Remove
			// them, because now we start using the parameter.
			// https://llvm.org/docs/Passes.html#functionattrs-deduce-function-attributes
			for _, kind := range []string{"nocapture", "readnone"} {
				kindID := llvm.AttributeKindID(kind)
				f.RemoveEnumAttributeAtIndex(f.ParamsCount(), kindID)
			}

			c.builder.SetInsertPointBefore(inst)

			var parentHandle llvm.Value
			if f.Linkage() == llvm.ExternalLinkage {
				// Exported function.
				// Note that getTaskPromisePtr will panic if it is called with
				// a nil pointer, so blocking exported functions that try to
				// return anything will not work.
				parentHandle = llvm.ConstPointerNull(c.i8ptrType)
			} else {
				parentHandle = f.LastParam()
				if parentHandle.IsNil() || parentHandle.Name() != "parentHandle" {
					// sanity check
					panic("trying to make exported function async")
				}
			}

			// Store return values.
			switch inst.OperandsCount() {
			case 0:
				// Nothing to return.
			case 1:
				// Return this value by writing to the pointer stored in the
				// parent handle. The parent coroutine has made an alloca that
				// we can write to to store our return value.
				returnValuePtr := c.createRuntimeCall("getTaskPromisePtr", []llvm.Value{parentHandle}, "coro.parentData")
				alloca := c.builder.CreateBitCast(returnValuePtr, llvm.PointerType(inst.Operand(0).Type(), 0), "coro.parentAlloca")
				c.builder.CreateStore(inst.Operand(0), alloca)
			default:
				panic("unreachable")
			}

			// Reactivate the parent coroutine. This adds it back to the run
			// queue, so it is started again by the scheduler when possible
			// (possibly right after the following suspend).
			c.createRuntimeCall("activateTask", []llvm.Value{parentHandle}, "")

			// Suspend this coroutine.
			// It would look like this is unnecessary, but if this
			// suspend point is left out, it leads to undefined
			// behavior somehow (with the unreachable instruction).
			continuePoint := c.builder.CreateCall(coroSuspendFunc, []llvm.Value{
				llvm.ConstNull(c.ctx.TokenType()),
				llvm.ConstInt(c.ctx.Int1Type(), 0, false),
			}, "ret")
			sw := c.builder.CreateSwitch(continuePoint, frame.suspendBlock, 2)
			sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 0, false), frame.unreachableBlock)
			sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 1, false), frame.cleanupBlock)
			inst.EraseFromParentAsInstruction()
		}

		// Coroutine cleanup. Free resources associated with this coroutine.
		c.builder.SetInsertPointAtEnd(frame.cleanupBlock)
		mem := c.builder.CreateCall(coroFreeFunc, []llvm.Value{id, frame.taskHandle}, "task.data.free")
		c.createRuntimeCall("free", []llvm.Value{mem}, "")
		c.builder.CreateBr(frame.suspendBlock)

		// Coroutine suspend. A call to llvm.coro.suspend() will branch here.
		c.builder.SetInsertPointAtEnd(frame.suspendBlock)
		c.builder.CreateCall(coroEndFunc, []llvm.Value{frame.taskHandle, llvm.ConstInt(c.ctx.Int1Type(), 0, false)}, "unused")
		returnType := f.Type().ElementType().ReturnType()
		if returnType.TypeKind() == llvm.VoidTypeKind {
			c.builder.CreateRetVoid()
		} else {
			c.builder.CreateRet(llvm.Undef(returnType))
		}

		// Coroutine exit. All final suspends (return instructions) will branch
		// here.
		c.builder.SetInsertPointAtEnd(frame.unreachableBlock)
		c.builder.CreateUnreachable()
	}

	// Replace calls to runtime.getCoroutineCall with the coroutine of this
	// frame.
	for _, getCoroutineCall := range getUses(c.mod.NamedFunction("runtime.getCoroutine")) {
		frame := asyncFuncs[getCoroutineCall.InstructionParent().Parent()]
		getCoroutineCall.ReplaceAllUsesWith(frame.taskHandle)
		getCoroutineCall.EraseFromParentAsInstruction()
	}

	// Transform calls to time.Sleep() into coroutine suspend points.
	for _, sleepCall := range getUses(sleep) {
		// sleepCall must be a call instruction.
		frame := asyncFuncs[sleepCall.InstructionParent().Parent()]
		duration := sleepCall.Operand(0)

		// Set task state to TASK_STATE_SLEEP and set the duration.
		c.builder.SetInsertPointBefore(sleepCall)
		c.createRuntimeCall("sleepTask", []llvm.Value{frame.taskHandle, duration}, "")

		// Yield to scheduler.
		continuePoint := c.builder.CreateCall(coroSuspendFunc, []llvm.Value{
			llvm.ConstNull(c.ctx.TokenType()),
			llvm.ConstInt(c.ctx.Int1Type(), 0, false),
		}, "")
		wakeup := c.splitBasicBlock(sleepCall, llvm.NextBasicBlock(c.builder.GetInsertBlock()), "task.wakeup")
		c.builder.SetInsertPointBefore(sleepCall)
		sw := c.builder.CreateSwitch(continuePoint, frame.suspendBlock, 2)
		sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 0, false), wakeup)
		sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 1, false), frame.cleanupBlock)
		sleepCall.EraseFromParentAsInstruction()
	}

	// Transform calls to runtime.deadlockStub into coroutine suspends (without
	// resume).
	for _, deadlockCall := range getUses(deadlockStub) {
		// deadlockCall must be a call instruction.
		frame := asyncFuncs[deadlockCall.InstructionParent().Parent()]

		// Exit coroutine.
		c.builder.SetInsertPointBefore(deadlockCall)
		continuePoint := c.builder.CreateCall(coroSuspendFunc, []llvm.Value{
			llvm.ConstNull(c.ctx.TokenType()),
			llvm.ConstInt(c.ctx.Int1Type(), 0, false),
		}, "")
		c.splitBasicBlock(deadlockCall, llvm.NextBasicBlock(c.builder.GetInsertBlock()), "task.wakeup.dead")
		c.builder.SetInsertPointBefore(deadlockCall)
		sw := c.builder.CreateSwitch(continuePoint, frame.suspendBlock, 2)
		sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 0, false), frame.unreachableBlock)
		sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 1, false), frame.cleanupBlock)
		deadlockCall.EraseFromParentAsInstruction()
	}

	// Transform calls to runtime.chanSend into channel send operations.
	for _, sendOp := range getUses(chanSend) {
		// sendOp must be a call instruction.
		frame := asyncFuncs[sendOp.InstructionParent().Parent()]

		// Yield to scheduler.
		c.builder.SetInsertPointBefore(llvm.NextInstruction(sendOp))
		continuePoint := c.builder.CreateCall(coroSuspendFunc, []llvm.Value{
			llvm.ConstNull(c.ctx.TokenType()),
			llvm.ConstInt(c.ctx.Int1Type(), 0, false),
		}, "")
		sw := c.builder.CreateSwitch(continuePoint, frame.suspendBlock, 2)
		wakeup := c.splitBasicBlock(sw, llvm.NextBasicBlock(c.builder.GetInsertBlock()), "task.sent")
		sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 0, false), wakeup)
		sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 1, false), frame.cleanupBlock)
	}

	// Transform calls to runtime.chanRecv into channel receive operations.
	for _, recvOp := range getUses(chanRecv) {
		// recvOp must be a call instruction.
		frame := asyncFuncs[recvOp.InstructionParent().Parent()]

		// Yield to scheduler.
		c.builder.SetInsertPointBefore(llvm.NextInstruction(recvOp))
		continuePoint := c.builder.CreateCall(coroSuspendFunc, []llvm.Value{
			llvm.ConstNull(c.ctx.TokenType()),
			llvm.ConstInt(c.ctx.Int1Type(), 0, false),
		}, "")
		sw := c.builder.CreateSwitch(continuePoint, frame.suspendBlock, 2)
		wakeup := c.splitBasicBlock(sw, llvm.NextBasicBlock(c.builder.GetInsertBlock()), "task.received")
		c.builder.SetInsertPointAtEnd(recvOp.InstructionParent())
		sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 0, false), wakeup)
		sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 1, false), frame.cleanupBlock)
	}

	return true, c.lowerMakeGoroutineCalls()
}

// Lower runtime.makeGoroutine calls to regular call instructions. This is done
// after the regular goroutine transformations. The started goroutines are
// either non-blocking (in which case they can be called directly) or blocking,
// in which case they will ask the scheduler themselves to be rescheduled.
func (c *Compiler) lowerMakeGoroutineCalls() error {
	// The following Go code:
	//   go startedGoroutine()
	//
	// Is translated to the following during IR construction, to preserve the
	// fact that this function should be called as a new goroutine.
	//   %0 = call i8* @runtime.makeGoroutine(i8* bitcast (void (i8*, i8*)* @main.startedGoroutine to i8*), i8* undef, i8* null)
	//   %1 = bitcast i8* %0 to void (i8*, i8*)*
	//   call void %1(i8* undef, i8* undef)
	//
	// This function rewrites it to a direct call:
	//   call void @main.startedGoroutine(i8* undef, i8* null)

	makeGoroutine := c.mod.NamedFunction("runtime.makeGoroutine")
	for _, goroutine := range getUses(makeGoroutine) {
		bitcastIn := goroutine.Operand(0)
		origFunc := bitcastIn.Operand(0)
		uses := getUses(goroutine)
		if len(uses) != 1 || uses[0].IsABitCastInst().IsNil() {
			return errors.New("expected exactly 1 bitcast use of runtime.makeGoroutine")
		}
		bitcastOut := uses[0]
		uses = getUses(bitcastOut)
		if len(uses) != 1 || uses[0].IsACallInst().IsNil() {
			return errors.New("expected exactly 1 call use of runtime.makeGoroutine bitcast")
		}
		realCall := uses[0]

		// Create call instruction.
		var params []llvm.Value
		for i := 0; i < realCall.OperandsCount()-1; i++ {
			params = append(params, realCall.Operand(i))
		}
		params[len(params)-1] = llvm.ConstPointerNull(c.i8ptrType) // parent coroutine handle (must be nil)
		c.builder.SetInsertPointBefore(realCall)
		c.builder.CreateCall(origFunc, params, "")
		realCall.EraseFromParentAsInstruction()
		bitcastOut.EraseFromParentAsInstruction()
		goroutine.EraseFromParentAsInstruction()
	}

	return nil
}
