package transform

// This file lowers asynchronous functions and goroutine starts when using the coroutines scheduler.
// This is accomplished by inserting LLVM intrinsics which are used in order to save the states of functions.

import (
	"errors"
	"go/token"
	"strconv"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// LowerCoroutines turns async functions into coroutines.
// This must be run with the coroutines scheduler.
//
// Before this pass, goroutine starts are expressed as a call to an intrinsic called "internal/task.start".
// This intrinsic accepts the function pointer and a pointer to a struct containing the function's arguments.
//
// Before this pass, an intrinsic called "internal/task.Pause" is used to express suspensions of the current goroutine.
//
// This pass first accumulates a list of blocking functions.
// A function is considered "blocking" if it calls "internal/task.Pause" or any other blocking function.
//
// Blocking calls are implemented by turning blocking functions into a coroutine.
// The body of each blocking function is modified to start a new coroutine, and to return after the first suspend.
// After calling a blocking function, the caller coroutine suspends.
// The caller also provides a buffer to store the return value into.
// When a blocking function returns, the return value is written into this buffer and then the caller is queued to run.
//
// Goroutine starts which invoke non-blocking functions are implemented as direct calls.
// Goroutine starts are replaced with the creation of a new task data structure followed by a call to the start of the blocking function.
// The task structure is populated with a "noop" coroutine before invoking the blocking function.
// When the blocking function returns, it resumes this "noop" coroutine which does nothing.
// The goroutine starter is able to continue after the first suspend of the started goroutine.
//
// The transformation of a function to a coroutine is accomplished using LLVM's coroutines system (https://llvm.org/docs/Coroutines.html).
// The simplest implementation of a coroutine inserts a suspend point after every blocking call.
//
// Transforming blocking functions into coroutines and calls into suspend points is extremely expensive.
// In many cases, a blocking call is followed immediately by a function terminator (a return or an "unreachable" instruction).
// This is a blocking "tail call".
// In a non-returning tail call (call to a non-returning function, such as an infinite loop), the coroutine can exit without any extra work.
// In a returning tail call, the returned value must either be the return of the call or a value known before the call.
// If the return value of the caller is the return of the callee, the coroutine can exit without any extra work and the tailed call will instead return to the caller of the caller.
// If the return value is known in advance, this result can be stored into the parent's return buffer before the call so that a suspend is unnecessary.
// If the callee returns an unnecessary value, a return buffer can be allocated on the heap so that it will outlive the coroutine.
//
// In the implementation of time.Sleep, the current task is pushed onto a timer queue and then suspended.
// Since the only suspend point is a call to "internal/task.Pause" followed by a return, there is no need to transform this into a coroutine.
// This generalizes to all blocking functions in which all suspend points can be elided.
// This optimization saves a substantial amount of binary size.
func LowerCoroutines(mod llvm.Module, needStackSlots bool) error {
	ctx := mod.Context()

	builder := ctx.NewBuilder()
	defer builder.Dispose()

	target := llvm.NewTargetData(mod.DataLayout())
	defer target.Dispose()

	pass := &coroutineLoweringPass{
		mod:            mod,
		ctx:            ctx,
		builder:        builder,
		target:         target,
		needStackSlots: needStackSlots,
	}

	err := pass.load()
	if err != nil {
		return err
	}

	// Supply task operands to async calls.
	pass.supplyTaskOperands()

	// Analyze async returns.
	pass.returnAnalysisPass()

	// Categorize async calls.
	pass.categorizeCalls()

	// Lower async functions.
	pass.lowerFuncsPass()

	// Lower calls to internal/task.Current.
	pass.lowerCurrent()

	// Lower goroutine starts.
	pass.lowerStartsPass()

	// Fix annotations on async call params.
	pass.fixAnnotations()

	if needStackSlots {
		// Set up garbage collector tracking of tasks at start.
		err = pass.trackGoroutines()
		if err != nil {
			return err
		}
	}

	return nil
}

// CoroutinesError is an error returned when coroutine lowering failed, for
// example because an async function is exported.
type CoroutinesError struct {
	Msg       string
	Pos       token.Position
	Traceback []CoroutinesErrorLine
}

// CoroutinesErrorLine is a single line of a CoroutinesError traceback.
type CoroutinesErrorLine struct {
	Name     string         // function name
	Position token.Position // position in the function
}

// Error implements the error interface by returning a simple error message
// without the stack.
func (err CoroutinesError) Error() string {
	return err.Msg
}

type asyncCallInfo struct {
	fn   llvm.Value
	call llvm.Value
}

// asyncFunc is a metadata container for an asynchronous function.
type asyncFunc struct {
	// fn is the underlying function pointer.
	fn llvm.Value

	// rawTask is the parameter where the task pointer is passed in.
	rawTask llvm.Value

	// callers is a set of all functions which call this async function.
	callers map[llvm.Value]struct{}

	// returns is a list of returns in the function, along with metadata.
	returns []asyncReturn

	// calls is a list of all calls in the asyncFunc.
	// normalCalls is a list of all intermideate suspending calls in the asyncFunc.
	// tailCalls is a list of all tail calls in the asyncFunc.
	calls, normalCalls, tailCalls []llvm.Value
}

// asyncReturn is a metadata container for a return from an asynchronous function.
type asyncReturn struct {
	// block is the basic block terminated by the return.
	block llvm.BasicBlock

	// kind is the kind of the return.
	kind returnKind
}

// coroutineLoweringPass is a goroutine lowering pass which is used with the "coroutines" scheduler.
type coroutineLoweringPass struct {
	mod     llvm.Module
	ctx     llvm.Context
	builder llvm.Builder
	target  llvm.TargetData

	// asyncFuncs is a map of all asyncFuncs.
	// The map keys are function pointers.
	asyncFuncs map[llvm.Value]*asyncFunc

	asyncFuncsOrdered []*asyncFunc

	// calls is a slice of all of the async calls in the module.
	calls []llvm.Value

	i8ptr llvm.Type

	// memory management functions from the runtime
	alloc, free llvm.Value

	// coroutine intrinsics
	start, pause, current                                   llvm.Value
	setState, setRetPtr, getRetPtr, returnTo, returnCurrent llvm.Value
	createTask                                              llvm.Value

	// llvm.coro intrinsics
	coroId, coroSize, coroBegin, coroSuspend, coroEnd, coroFree, coroSave llvm.Value

	trackPointer   llvm.Value
	needStackSlots bool
}

// findAsyncFuncs finds all asynchronous functions.
// A function is considered asynchronous if it calls an asynchronous function or intrinsic.
func (c *coroutineLoweringPass) findAsyncFuncs() error {
	asyncs := map[llvm.Value]*asyncFunc{}
	asyncsOrdered := []llvm.Value{}
	calls := []llvm.Value{}
	callsAsyncFunction := map[llvm.Value]asyncCallInfo{}

	// Use a breadth-first search to find all async functions.
	worklist := []llvm.Value{c.pause}
	for len(worklist) > 0 {
		// Pop a function off the worklist.
		fn := worklist[0]
		worklist = worklist[1:]

		// Get task pointer argument.
		task := fn.LastParam()
		if fn != c.pause && (task.IsNil() || task.Name() != "parentHandle") {
			// Exported functions must not do async operations.
			err := CoroutinesError{
				Msg: "blocking operation in exported function: " + fn.Name(),
				Pos: getPosition(fn),
			}
			f := fn
			for !f.IsNil() && f != c.pause {
				data := callsAsyncFunction[f]
				err.Traceback = append(err.Traceback, CoroutinesErrorLine{f.Name(), getPosition(data.call)})
				f = data.fn
			}
			return err
		}

		// Search all uses of the function while collecting callers.
		callers := map[llvm.Value]struct{}{}
		for use := fn.FirstUse(); !use.IsNil(); use = use.NextUse() {
			user := use.User()
			if user.IsACallInst().IsNil() {
				// User is not a call instruction, so this is irrelevant.
				continue
			}
			if user.CalledValue() != fn {
				// Not the called value.
				continue
			}

			// Add to calls list.
			calls = append(calls, user)

			// Get the caller.
			caller := user.InstructionParent().Parent()

			// Add as caller.
			callers[caller] = struct{}{}

			if _, ok := asyncs[caller]; ok {
				// Already marked caller as async.
				continue
			}

			// Mark the caller as async.
			// Use nil as a temporary value. It will be replaced later.
			asyncs[caller] = nil
			asyncsOrdered = append(asyncsOrdered, caller)

			// Track which calls caused this function to be marked async (for
			// better diagnostics).
			callsAsyncFunction[caller] = asyncCallInfo{
				fn:   fn,
				call: user,
			}

			// Put the caller on the worklist.
			worklist = append(worklist, caller)
		}

		asyncs[fn] = &asyncFunc{
			fn:      fn,
			rawTask: task,
			callers: callers,
		}
	}

	// Flip the order of the async functions so that the top ones are lowered first.
	for i := 0; i < len(asyncsOrdered)/2; i++ {
		asyncsOrdered[i], asyncsOrdered[len(asyncsOrdered)-(i+1)] = asyncsOrdered[len(asyncsOrdered)-(i+1)], asyncsOrdered[i]
	}

	// Map the elements of asyncsOrdered to *asyncFunc.
	asyncFuncsOrdered := make([]*asyncFunc, len(asyncsOrdered))
	for i, v := range asyncsOrdered {
		asyncFuncsOrdered[i] = asyncs[v]
	}

	c.asyncFuncs = asyncs
	c.asyncFuncsOrdered = asyncFuncsOrdered
	c.calls = calls

	return nil
}

func (c *coroutineLoweringPass) load() error {
	// Find memory management functions from the runtime.
	c.alloc = c.mod.NamedFunction("runtime.alloc")
	if c.alloc.IsNil() {
		return ErrMissingIntrinsic{"runtime.alloc"}
	}
	c.free = c.mod.NamedFunction("runtime.free")
	if c.free.IsNil() {
		return ErrMissingIntrinsic{"runtime.free"}
	}

	// Find intrinsics.
	c.pause = c.mod.NamedFunction("internal/task.Pause")
	if c.pause.IsNil() {
		return ErrMissingIntrinsic{"internal/task.Pause"}
	}
	c.start = c.mod.NamedFunction("internal/task.start")
	if c.start.IsNil() {
		return ErrMissingIntrinsic{"internal/task.start"}
	}
	c.current = c.mod.NamedFunction("internal/task.Current")
	if c.current.IsNil() {
		return ErrMissingIntrinsic{"internal/task.Current"}
	}
	c.setState = c.mod.NamedFunction("(*internal/task.Task).setState")
	if c.setState.IsNil() {
		return ErrMissingIntrinsic{"(*internal/task.Task).setState"}
	}
	c.setRetPtr = c.mod.NamedFunction("(*internal/task.Task).setReturnPtr")
	if c.setRetPtr.IsNil() {
		return ErrMissingIntrinsic{"(*internal/task.Task).setReturnPtr"}
	}
	c.getRetPtr = c.mod.NamedFunction("(*internal/task.Task).getReturnPtr")
	if c.getRetPtr.IsNil() {
		return ErrMissingIntrinsic{"(*internal/task.Task).getReturnPtr"}
	}
	c.returnTo = c.mod.NamedFunction("(*internal/task.Task).returnTo")
	if c.returnTo.IsNil() {
		return ErrMissingIntrinsic{"(*internal/task.Task).returnTo"}
	}
	c.returnCurrent = c.mod.NamedFunction("(*internal/task.Task).returnCurrent")
	if c.returnCurrent.IsNil() {
		return ErrMissingIntrinsic{"(*internal/task.Task).returnCurrent"}
	}
	c.createTask = c.mod.NamedFunction("internal/task.createTask")
	if c.createTask.IsNil() {
		return ErrMissingIntrinsic{"internal/task.createTask"}
	}

	if c.needStackSlots {
		c.trackPointer = c.mod.NamedFunction("runtime.trackPointer")
		if c.trackPointer.IsNil() {
			return ErrMissingIntrinsic{"runtime.trackPointer"}
		}
	}

	// Find async functions.
	err := c.findAsyncFuncs()
	if err != nil {
		return err
	}

	// Get i8* type.
	c.i8ptr = llvm.PointerType(c.ctx.Int8Type(), 0)

	// Build LLVM coroutine intrinsic.
	coroIdType := llvm.FunctionType(c.ctx.TokenType(), []llvm.Type{c.ctx.Int32Type(), c.i8ptr, c.i8ptr, c.i8ptr}, false)
	c.coroId = llvm.AddFunction(c.mod, "llvm.coro.id", coroIdType)

	sizeT := c.alloc.Param(0).Type()
	coroSizeType := llvm.FunctionType(sizeT, nil, false)
	c.coroSize = llvm.AddFunction(c.mod, "llvm.coro.size.i"+strconv.Itoa(sizeT.IntTypeWidth()), coroSizeType)

	coroBeginType := llvm.FunctionType(c.i8ptr, []llvm.Type{c.ctx.TokenType(), c.i8ptr}, false)
	c.coroBegin = llvm.AddFunction(c.mod, "llvm.coro.begin", coroBeginType)

	coroSuspendType := llvm.FunctionType(c.ctx.Int8Type(), []llvm.Type{c.ctx.TokenType(), c.ctx.Int1Type()}, false)
	c.coroSuspend = llvm.AddFunction(c.mod, "llvm.coro.suspend", coroSuspendType)

	coroEndType := llvm.FunctionType(c.ctx.Int1Type(), []llvm.Type{c.i8ptr, c.ctx.Int1Type()}, false)
	c.coroEnd = llvm.AddFunction(c.mod, "llvm.coro.end", coroEndType)

	coroFreeType := llvm.FunctionType(c.i8ptr, []llvm.Type{c.ctx.TokenType(), c.i8ptr}, false)
	c.coroFree = llvm.AddFunction(c.mod, "llvm.coro.free", coroFreeType)

	coroSaveType := llvm.FunctionType(c.ctx.TokenType(), []llvm.Type{c.i8ptr}, false)
	c.coroSave = llvm.AddFunction(c.mod, "llvm.coro.save", coroSaveType)

	return nil
}

func (c *coroutineLoweringPass) track(ptr llvm.Value) {
	if c.needStackSlots {
		if ptr.Type() != c.i8ptr {
			ptr = c.builder.CreateBitCast(ptr, c.i8ptr, "track.bitcast")
		}
		c.builder.CreateCall(c.trackPointer, []llvm.Value{ptr, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")
	}
}

// lowerStartSync lowers a goroutine start of a synchronous function to a synchronous call.
func (c *coroutineLoweringPass) lowerStartSync(start llvm.Value) {
	c.builder.SetInsertPointBefore(start)

	// Get function to call.
	fn := start.Operand(0).Operand(0)

	// Create the list of params for the call.
	paramTypes := fn.Type().ElementType().ParamTypes()
	params := llvmutil.EmitPointerUnpack(c.builder, c.mod, start.Operand(1), paramTypes[:len(paramTypes)-1])
	params = append(params, llvm.Undef(c.i8ptr))

	// Generate call to function.
	c.builder.CreateCall(fn, params, "")

	// Remove start call.
	start.EraseFromParentAsInstruction()
}

// supplyTaskOperands fills in the task operands of async calls.
func (c *coroutineLoweringPass) supplyTaskOperands() {
	var curCalls []llvm.Value
	for use := c.current.FirstUse(); !use.IsNil(); use = use.NextUse() {
		curCalls = append(curCalls, use.User())
	}
	for _, call := range append(curCalls, c.calls...) {
		c.builder.SetInsertPointBefore(call)
		task := c.asyncFuncs[call.InstructionParent().Parent()].rawTask
		call.SetOperand(call.OperandsCount()-2, task)
	}
}

// returnKind is a classification of a type of function terminator.
type returnKind uint8

const (
	// returnNormal is a terminator that returns a value normally from a function.
	returnNormal returnKind = iota

	// returnVoid is a terminator that exits normally without returning a value.
	returnVoid

	// returnVoidTail is a terminator which is a tail call to a void-returning function in a void-returning function.
	returnVoidTail

	// returnTail is a terinator which is a tail call to a value-returning function where the value is returned by the callee.
	returnTail

	// returnDeadTail is a terminator which is a call to a non-returning asynchronous function.
	returnDeadTail

	// returnAlternateTail is a terminator which is a tail call to a value-returning function where a previously acquired value is returned by the callee.
	returnAlternateTail

	// returnDitchedTail is a terminator which is a tail call to a value-returning function, where the callee returns void.
	returnDitchedTail

	// returnDelayedValue is a terminator in which a void-returning tail call is followed by a return of a previous value.
	returnDelayedValue
)

// isAsyncCall returns whether the specified call is async.
func (c *coroutineLoweringPass) isAsyncCall(call llvm.Value) bool {
	_, ok := c.asyncFuncs[call.CalledValue()]
	return ok
}

// analyzeFuncReturns analyzes and classifies the returns of a function.
func (c *coroutineLoweringPass) analyzeFuncReturns(fn *asyncFunc) {
	returns := []asyncReturn{}
	if fn.fn == c.pause {
		// Skip pause.
		fn.returns = returns
		return
	}

	for _, bb := range fn.fn.BasicBlocks() {
		last := bb.LastInstruction()
		switch last.InstructionOpcode() {
		case llvm.Ret:
			// Check if it is a void return.
			isVoid := fn.fn.Type().ElementType().ReturnType().TypeKind() == llvm.VoidTypeKind

			// Analyze previous instruction.
			prev := llvm.PrevInstruction(last)
			switch {
			case prev.IsNil():
				fallthrough
			case prev.IsACallInst().IsNil():
				fallthrough
			case !c.isAsyncCall(prev):
				// This is not any form of asynchronous tail call.
				if isVoid {
					returns = append(returns, asyncReturn{
						block: bb,
						kind:  returnVoid,
					})
				} else {
					returns = append(returns, asyncReturn{
						block: bb,
						kind:  returnNormal,
					})
				}
			case isVoid:
				if prev.CalledValue().Type().ElementType().ReturnType().TypeKind() == llvm.VoidTypeKind {
					// This is a tail call to a void-returning function from a function with a void return.
					returns = append(returns, asyncReturn{
						block: bb,
						kind:  returnVoidTail,
					})
				} else {
					// This is a tail call to a value-returning function from a function with a void return.
					// The returned value will be ditched.
					returns = append(returns, asyncReturn{
						block: bb,
						kind:  returnDitchedTail,
					})
				}
			case last.Operand(0) == prev:
				// This is a regular tail call. The return of the callee is returned to the parent.
				returns = append(returns, asyncReturn{
					block: bb,
					kind:  returnTail,
				})
			case prev.CalledValue().Type().ElementType().ReturnType().TypeKind() == llvm.VoidTypeKind:
				// This is a tail call that returns a previous value after waiting on a void function.
				returns = append(returns, asyncReturn{
					block: bb,
					kind:  returnDelayedValue,
				})
			default:
				// This is a tail call that returns a value that is available before the function call.
				returns = append(returns, asyncReturn{
					block: bb,
					kind:  returnAlternateTail,
				})
			}
		case llvm.Unreachable:
			prev := llvm.PrevInstruction(last)

			if prev.IsNil() || prev.IsACallInst().IsNil() || !c.isAsyncCall(prev) {
				// This unreachable instruction does not behave as an asynchronous return.
				continue
			}

			// This is an asyncnhronous tail call to function that does not return.
			returns = append(returns, asyncReturn{
				block: bb,
				kind:  returnDeadTail,
			})
		}
	}

	fn.returns = returns
}

// returnAnalysisPass runs an analysis pass which classifies the returns of all async functions.
func (c *coroutineLoweringPass) returnAnalysisPass() {
	for _, async := range c.asyncFuncsOrdered {
		c.analyzeFuncReturns(async)
	}
}

// categorizeCalls categorizes all asynchronous calls into regular vs. async and matches them to their callers.
func (c *coroutineLoweringPass) categorizeCalls() {
	// Sort calls into their respective callers.
	for _, call := range c.calls {
		caller := c.asyncFuncs[call.InstructionParent().Parent()]
		caller.calls = append(caller.calls, call)
	}

	// Seperate regular and tail calls.
	for _, async := range c.asyncFuncsOrdered {
		// Search returns for tail calls.
		tails := map[llvm.Value]struct{}{}
		for _, ret := range async.returns {
			switch ret.kind {
			case returnVoidTail, returnTail, returnDeadTail, returnAlternateTail, returnDitchedTail, returnDelayedValue:
				// This is a tail return. The previous instruction is a tail call.
				tails[llvm.PrevInstruction(ret.block.LastInstruction())] = struct{}{}
			}
		}

		// Seperate tail calls and regular calls.
		normalCalls, tailCalls := []llvm.Value{}, []llvm.Value{}
		for _, call := range async.calls {
			if _, ok := tails[call]; ok {
				// This is a tail call.
				tailCalls = append(tailCalls, call)
			} else {
				// This is a regular call.
				normalCalls = append(normalCalls, call)
			}
		}

		async.normalCalls = normalCalls
		async.tailCalls = tailCalls
	}
}

// lowerFuncsPass lowers all functions, turning them into coroutines if necessary.
func (c *coroutineLoweringPass) lowerFuncsPass() {
	for _, fn := range c.asyncFuncs {
		if fn.fn == c.pause {
			// Skip. It is an intrinsic.
			continue
		}

		if len(fn.normalCalls) == 0 && fn.fn.FirstBasicBlock().FirstInstruction().IsAAllocaInst().IsNil() {
			// No suspend points or stack allocations. Lower without turning it into a coroutine.
			c.lowerFuncFast(fn)
		} else {
			// There are suspend points or stack allocations, so it is necessary to turn this into a coroutine.
			c.lowerFuncCoro(fn)
		}
	}
}

func (async *asyncFunc) hasValueStoreReturn() bool {
	for _, ret := range async.returns {
		switch ret.kind {
		case returnNormal, returnAlternateTail, returnDelayedValue:
			return true
		}
	}

	return false
}

// heapAlloc creates a heap allocation large enough to hold the supplied type.
// The allocation is returned as a raw i8* pointer.
// This allocation is not automatically tracked by the garbage collector, and should thus be stored into a tracked memory object immediately.
func (c *coroutineLoweringPass) heapAlloc(t llvm.Type, name string) llvm.Value {
	sizeT := c.alloc.FirstParam().Type()
	size := llvm.ConstInt(sizeT, c.target.TypeAllocSize(t), false)
	return c.builder.CreateCall(c.alloc, []llvm.Value{size, llvm.ConstNull(c.i8ptr), llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, name)
}

// lowerFuncFast lowers an async function that has no suspend points.
func (c *coroutineLoweringPass) lowerFuncFast(fn *asyncFunc) {
	// Get return type.
	retType := fn.fn.Type().ElementType().ReturnType()

	// Get task value.
	c.insertPointAfterAllocas(fn.fn)
	task := c.builder.CreateCall(c.current, []llvm.Value{llvm.Undef(c.i8ptr), fn.rawTask}, "task")

	// Get return pointer if applicable.
	var rawRetPtr, retPtr llvm.Value
	if fn.hasValueStoreReturn() {
		rawRetPtr = c.builder.CreateCall(c.getRetPtr, []llvm.Value{task, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "ret.ptr")
		retType = fn.fn.Type().ElementType().ReturnType()
		retPtr = c.builder.CreateBitCast(rawRetPtr, llvm.PointerType(retType, 0), "ret.ptr.bitcast")
	}

	// Lower returns.
	for _, ret := range fn.returns {
		// Get terminator.
		terminator := ret.block.LastInstruction()

		// Get tail call if applicable.
		var call llvm.Value
		switch ret.kind {
		case returnVoidTail, returnTail, returnDeadTail, returnAlternateTail, returnDitchedTail, returnDelayedValue:
			call = llvm.PrevInstruction(terminator)
		}

		switch ret.kind {
		case returnNormal:
			c.builder.SetInsertPointBefore(terminator)

			// Store value into return pointer.
			c.builder.CreateStore(terminator.Operand(0), retPtr)

			// Resume caller.
			c.builder.CreateCall(c.returnCurrent, []llvm.Value{task, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")

			// Erase return argument.
			terminator.SetOperand(0, llvm.Undef(retType))
		case returnVoid:
			c.builder.SetInsertPointBefore(terminator)

			// Resume caller.
			c.builder.CreateCall(c.returnCurrent, []llvm.Value{task, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")
		case returnVoidTail:
			// Nothing to do. There is already a tail call followed by a void return.
		case returnTail:
			// Erase return argument.
			terminator.SetOperand(0, llvm.Undef(retType))
		case returnDeadTail:
			// Replace unreachable with immediate return, without resuming the caller.
			c.builder.SetInsertPointBefore(terminator)
			if retType.TypeKind() == llvm.VoidTypeKind {
				c.builder.CreateRetVoid()
			} else {
				c.builder.CreateRet(llvm.Undef(retType))
			}
			terminator.EraseFromParentAsInstruction()
		case returnAlternateTail:
			c.builder.SetInsertPointBefore(call)

			// Store return value.
			c.builder.CreateStore(terminator.Operand(0), retPtr)

			// Heap-allocate a return buffer for the discarded return.
			alternateBuf := c.heapAlloc(call.Type(), "ret.alternate")
			c.builder.CreateCall(c.setRetPtr, []llvm.Value{task, alternateBuf, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")

			// Erase return argument.
			terminator.SetOperand(0, llvm.Undef(retType))
		case returnDitchedTail:
			c.builder.SetInsertPointBefore(call)

			// Heap-allocate a return buffer for the discarded return.
			ditchBuf := c.heapAlloc(call.Type(), "ret.ditch")
			c.builder.CreateCall(c.setRetPtr, []llvm.Value{task, ditchBuf, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")
		case returnDelayedValue:
			c.builder.SetInsertPointBefore(call)

			// Store value into return pointer.
			c.builder.CreateStore(terminator.Operand(0), retPtr)

			// Erase return argument.
			terminator.SetOperand(0, llvm.Undef(retType))
		}

		// Delete call if it is a pause, because it has already been lowered.
		if !call.IsNil() && call.CalledValue() == c.pause {
			call.EraseFromParentAsInstruction()
		}
	}
}

// insertPointAfterAllocas sets the insert point of the builder to be immediately after the last alloca in the entry block.
func (c *coroutineLoweringPass) insertPointAfterAllocas(fn llvm.Value) {
	inst := fn.EntryBasicBlock().FirstInstruction()
	for !inst.IsAAllocaInst().IsNil() {
		inst = llvm.NextInstruction(inst)
	}
	c.builder.SetInsertPointBefore(inst)
}

// lowerCallReturn lowers the return value of an async call by creating a return buffer and loading the returned value from it.
func (c *coroutineLoweringPass) lowerCallReturn(caller *asyncFunc, call llvm.Value) {
	// Get return type.
	retType := call.Type()
	if retType.TypeKind() == llvm.VoidTypeKind {
		// Void return. Nothing to do.
		return
	}

	// Create alloca for return buffer.
	alloca := llvmutil.CreateInstructionAlloca(c.builder, c.mod, retType, call, "call.return")

	// Store new return buffer into task before call.
	c.builder.SetInsertPointBefore(call)
	task := c.builder.CreateCall(c.current, []llvm.Value{llvm.Undef(c.i8ptr), caller.rawTask}, "call.task")
	retPtr := c.builder.CreateBitCast(alloca, c.i8ptr, "call.return.bitcast")
	c.builder.CreateCall(c.setRetPtr, []llvm.Value{task, retPtr, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")

	// Load return value after call.
	c.builder.SetInsertPointBefore(llvm.NextInstruction(call))
	ret := c.builder.CreateLoad(alloca, "call.return.load")

	// Replace call value with loaded return.
	call.ReplaceAllUsesWith(ret)
}

// lowerFuncCoro transforms an async function into a coroutine by lowering async operations to `llvm.coro` intrinsics.
// See https://llvm.org/docs/Coroutines.html for more information on these intrinsics.
func (c *coroutineLoweringPass) lowerFuncCoro(fn *asyncFunc) {
	// Ensure that any alloca instructions in the entry block are at the start.
	// Otherwise, block splitting would result in unintended behavior.
	{
		// Skip alloca instructions at the start of the block.
		inst := fn.fn.FirstBasicBlock().FirstInstruction()
		for !inst.IsAAllocaInst().IsNil() {
			inst = llvm.NextInstruction(inst)
		}

		// Find any other alloca instructions and move them after the other allocas.
		c.builder.SetInsertPointBefore(inst)
		for !inst.IsNil() {
			next := llvm.NextInstruction(inst)
			if !inst.IsAAllocaInst().IsNil() {
				inst.RemoveFromParentAsInstruction()
				c.builder.Insert(inst)
			}
			inst = next
		}
	}

	returnType := fn.fn.Type().ElementType().ReturnType()

	// Prepare coroutine state.
	c.insertPointAfterAllocas(fn.fn)
	// %coro.id = call token @llvm.coro.id(i32 0, i8* null, i8* null, i8* null)
	coroId := c.builder.CreateCall(c.coroId, []llvm.Value{
		llvm.ConstInt(c.ctx.Int32Type(), 0, false),
		llvm.ConstNull(c.i8ptr),
		llvm.ConstNull(c.i8ptr),
		llvm.ConstNull(c.i8ptr),
	}, "coro.id")
	// %coro.size = call i32 @llvm.coro.size.i32()
	coroSize := c.builder.CreateCall(c.coroSize, []llvm.Value{}, "coro.size")
	// %coro.alloc = call i8* runtime.alloc(i32 %coro.size)
	coroAlloc := c.builder.CreateCall(c.alloc, []llvm.Value{coroSize, llvm.ConstNull(c.i8ptr), llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "coro.alloc")
	// %coro.state = call noalias i8* @llvm.coro.begin(token %coro.id, i8* %coro.alloc)
	coroState := c.builder.CreateCall(c.coroBegin, []llvm.Value{coroId, coroAlloc}, "coro.state")
	c.track(coroState)
	// Store state into task.
	task := c.builder.CreateCall(c.current, []llvm.Value{llvm.Undef(c.i8ptr), fn.rawTask}, "task")
	parentState := c.builder.CreateCall(c.setState, []llvm.Value{task, coroState, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "task.state.parent")
	// Get return pointer if needed.
	var retPtrRaw, retPtr llvm.Value
	if returnType.TypeKind() != llvm.VoidTypeKind {
		retPtrRaw = c.builder.CreateCall(c.getRetPtr, []llvm.Value{task, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "task.retPtr")
		retPtr = c.builder.CreateBitCast(retPtrRaw, llvm.PointerType(fn.fn.Type().ElementType().ReturnType(), 0), "task.retPtr.bitcast")
	}

	// Build suspend block.
	// This is executed when the coroutine is about to suspend.
	suspend := c.ctx.AddBasicBlock(fn.fn, "suspend")
	c.builder.SetInsertPointAtEnd(suspend)
	// %unused = call i1 @llvm.coro.end(i8* %coro.state, i1 false)
	c.builder.CreateCall(c.coroEnd, []llvm.Value{coroState, llvm.ConstInt(c.ctx.Int1Type(), 0, false)}, "unused")
	// Insert return.
	if returnType.TypeKind() == llvm.VoidTypeKind {
		c.builder.CreateRetVoid()
	} else {
		c.builder.CreateRet(llvm.Undef(returnType))
	}

	// Build cleanup block.
	// This is executed before the function returns in order to clean up resources.
	cleanup := c.ctx.AddBasicBlock(fn.fn, "cleanup")
	c.builder.SetInsertPointAtEnd(cleanup)
	// %coro.memFree = call i8* @llvm.coro.free(token %coro.id, i8* %coro.state)
	coroMemFree := c.builder.CreateCall(c.coroFree, []llvm.Value{coroId, coroState}, "coro.memFree")
	// call i8* runtime.free(i8* %coro.memFree)
	c.builder.CreateCall(c.free, []llvm.Value{coroMemFree, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")
	// Branch to suspend block.
	c.builder.CreateBr(suspend)

	// Restore old state before tail calls.
	for _, call := range fn.tailCalls {
		if !llvm.NextInstruction(call).IsAUnreachableInst().IsNil() {
			// Callee never returns, so the state restore is ineffectual.
			continue
		}

		c.builder.SetInsertPointBefore(call)
		c.builder.CreateCall(c.setState, []llvm.Value{task, parentState, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "coro.state.restore")
	}

	// Lower returns.
	var postTail llvm.BasicBlock
	for _, ret := range fn.returns {
		// Get terminator instruction.
		terminator := ret.block.LastInstruction()

		// Get tail call if applicable.
		var call llvm.Value
		switch ret.kind {
		case returnVoidTail, returnTail, returnDeadTail, returnAlternateTail, returnDitchedTail, returnDelayedValue:
			call = llvm.PrevInstruction(terminator)
		}

		switch ret.kind {
		case returnNormal:
			c.builder.SetInsertPointBefore(terminator)

			// Store value into return pointer.
			c.builder.CreateStore(terminator.Operand(0), retPtr)

			// Resume caller.
			c.builder.CreateCall(c.returnTo, []llvm.Value{task, parentState, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")
		case returnVoid:
			c.builder.SetInsertPointBefore(terminator)

			// Resume caller.
			c.builder.CreateCall(c.returnTo, []llvm.Value{task, parentState, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")
		case returnVoidTail, returnDeadTail:
			// Nothing to do.
		case returnTail:
			c.builder.SetInsertPointBefore(call)

			// Restore the return pointer so that the caller can store into it.
			c.builder.CreateCall(c.setRetPtr, []llvm.Value{task, retPtrRaw, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")
		case returnAlternateTail:
			c.builder.SetInsertPointBefore(call)

			// Store return value.
			c.builder.CreateStore(terminator.Operand(0), retPtr)

			// Heap-allocate a return buffer for the discarded return.
			alternateBuf := c.heapAlloc(call.Type(), "ret.alternate")
			c.builder.CreateCall(c.setRetPtr, []llvm.Value{task, alternateBuf, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")
		case returnDitchedTail:
			c.builder.SetInsertPointBefore(call)

			// Heap-allocate a return buffer for the discarded return.
			ditchBuf := c.heapAlloc(call.Type(), "ret.ditch")
			c.builder.CreateCall(c.setRetPtr, []llvm.Value{task, ditchBuf, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")
		case returnDelayedValue:
			c.builder.SetInsertPointBefore(call)

			// Store return value.
			c.builder.CreateStore(terminator.Operand(0), retPtr)
		}

		// Delete call if it is a pause, because it has already been lowered.
		if !call.IsNil() && call.CalledValue() == c.pause {
			call.EraseFromParentAsInstruction()
		}

		// Replace terminator with a branch to the exit.
		var exit llvm.BasicBlock
		if ret.kind == returnNormal || ret.kind == returnVoid || fn.fn.FirstBasicBlock().FirstInstruction().IsAAllocaInst().IsNil() {
			// Exit through the cleanup path.
			exit = cleanup
		} else {
			if postTail.IsNil() {
				// Create a path with a suspend that never reawakens.
				postTail = c.ctx.AddBasicBlock(fn.fn, "post.tail")
				c.builder.SetInsertPointAtEnd(postTail)
				// %coro.save = call token @llvm.coro.save(i8* %coro.state)
				save := c.builder.CreateCall(c.coroSave, []llvm.Value{coroState}, "coro.save")
				// %call.suspend = llvm.coro.suspend(token %coro.save, i1 false)
				// switch i8 %call.suspend, label %suspend [i8 0, label %wakeup
				//                                          i8 1, label %cleanup]
				suspendValue := c.builder.CreateCall(c.coroSuspend, []llvm.Value{save, llvm.ConstInt(c.ctx.Int1Type(), 0, false)}, "call.suspend")
				sw := c.builder.CreateSwitch(suspendValue, suspend, 2)
				unreachableBlock := c.ctx.AddBasicBlock(fn.fn, "unreachable")
				sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 0, false), unreachableBlock)
				sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 1, false), cleanup)
				c.builder.SetInsertPointAtEnd(unreachableBlock)
				c.builder.CreateUnreachable()
			}

			// Exit through a permanent suspend.
			exit = postTail
		}

		terminator.EraseFromParentAsInstruction()
		c.builder.SetInsertPointAtEnd(ret.block)
		c.builder.CreateBr(exit)
	}

	// Lower regular calls.
	for _, call := range fn.normalCalls {
		// Lower return value of call.
		c.lowerCallReturn(fn, call)

		// Get originating basic block.
		bb := call.InstructionParent()

		// Split block.
		wakeup := llvmutil.SplitBasicBlock(c.builder, call, llvm.NextBasicBlock(bb), "wakeup")

		// Insert suspension and switch.
		c.builder.SetInsertPointAtEnd(bb)
		// %coro.save = call token @llvm.coro.save(i8* %coro.state)
		save := c.builder.CreateCall(c.coroSave, []llvm.Value{coroState}, "coro.save")
		// %call.suspend = llvm.coro.suspend(token %coro.save, i1 false)
		// switch i8 %call.suspend, label %suspend [i8 0, label %wakeup
		//                                          i8 1, label %cleanup]
		suspendValue := c.builder.CreateCall(c.coroSuspend, []llvm.Value{save, llvm.ConstInt(c.ctx.Int1Type(), 0, false)}, "call.suspend")
		sw := c.builder.CreateSwitch(suspendValue, suspend, 2)
		sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 0, false), wakeup)
		sw.AddCase(llvm.ConstInt(c.ctx.Int8Type(), 1, false), cleanup)

		// Delete call if it is a pause, because it has already been lowered.
		if call.CalledValue() == c.pause {
			call.EraseFromParentAsInstruction()
		}

		c.builder.SetInsertPointBefore(wakeup.FirstInstruction())
		c.track(coroState)
	}
}

// lowerCurrent lowers calls to internal/task.Current to bitcasts.
func (c *coroutineLoweringPass) lowerCurrent() error {
	taskType := c.current.Type().ElementType().ReturnType()
	deleteQueue := []llvm.Value{}
	for use := c.current.FirstUse(); !use.IsNil(); use = use.NextUse() {
		// Get user.
		user := use.User()

		if user.IsACallInst().IsNil() || user.CalledValue() != c.current {
			return errorAt(user, "unexpected non-call use of task.Current")
		}

		// Replace with bitcast.
		c.builder.SetInsertPointBefore(user)
		raw := user.Operand(1)
		if !raw.IsAUndefValue().IsNil() || raw.IsNull() {
			return errors.New("undefined task")
		}
		task := c.builder.CreateBitCast(raw, taskType, "task.current")
		user.ReplaceAllUsesWith(task)
		deleteQueue = append(deleteQueue, user)
	}

	// Delete calls.
	for _, inst := range deleteQueue {
		inst.EraseFromParentAsInstruction()
	}

	return nil
}

// lowerStart lowers a goroutine start into a task creation and call or a synchronous call.
func (c *coroutineLoweringPass) lowerStart(start llvm.Value) {
	c.builder.SetInsertPointBefore(start)

	// Get function to call.
	fn := start.Operand(0).Operand(0)

	if _, ok := c.asyncFuncs[fn]; !ok {
		// Turn into synchronous call.
		c.lowerStartSync(start)
		return
	}

	// Create the list of params for the call.
	paramTypes := fn.Type().ElementType().ParamTypes()
	params := llvmutil.EmitPointerUnpack(c.builder, c.mod, start.Operand(1), paramTypes[:len(paramTypes)-1])

	// Create task.
	task := c.builder.CreateCall(c.createTask, []llvm.Value{llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "start.task")
	rawTask := c.builder.CreateBitCast(task, c.i8ptr, "start.task.bitcast")
	params = append(params, rawTask)

	// Generate a return buffer if necessary.
	returnType := fn.Type().ElementType().ReturnType()
	if returnType.TypeKind() == llvm.VoidTypeKind {
		// No return buffer necessary for a void return.
	} else {
		// Check for any undead returns.
		var undead bool
		for _, ret := range c.asyncFuncs[fn].returns {
			if ret.kind != returnDeadTail {
				// This return results in a value being eventually stored.
				undead = true
				break
			}
		}
		if undead {
			// The function stores a value into a return buffer, so we need to create one.
			retBuf := c.heapAlloc(returnType, "ret.ditch")
			c.builder.CreateCall(c.setRetPtr, []llvm.Value{task, retBuf, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")
		}
	}

	// Generate call to function.
	c.builder.CreateCall(fn, params, "")

	// Erase start call.
	start.EraseFromParentAsInstruction()
}

// lowerStartsPass lowers all goroutine starts.
func (c *coroutineLoweringPass) lowerStartsPass() {
	starts := []llvm.Value{}
	for use := c.start.FirstUse(); !use.IsNil(); use = use.NextUse() {
		starts = append(starts, use.User())
	}
	for _, start := range starts {
		c.lowerStart(start)
	}
}

func (c *coroutineLoweringPass) fixAnnotations() {
	for f := range c.asyncFuncs {
		// These properties were added by the functionattrs pass. Remove
		// them, because now we start using the parameter.
		// https://llvm.org/docs/Passes.html#functionattrs-deduce-function-attributes
		for _, kind := range []string{"nocapture", "readnone"} {
			kindID := llvm.AttributeKindID(kind)
			n := f.ParamsCount()
			for i := 0; i <= n; i++ {
				f.RemoveEnumAttributeAtIndex(i, kindID)
			}
		}
	}
}

// trackGoroutines adds runtime.trackPointer calls to track goroutine starts and data.
func (c *coroutineLoweringPass) trackGoroutines() error {
	trackPointer := c.mod.NamedFunction("runtime.trackPointer")
	if trackPointer.IsNil() {
		return ErrMissingIntrinsic{"runtime.trackPointer"}
	}

	trackFunctions := []llvm.Value{c.createTask, c.setState, c.getRetPtr}
	for _, fn := range trackFunctions {
		for use := fn.FirstUse(); !use.IsNil(); use = use.NextUse() {
			call := use.User()

			c.builder.SetInsertPointBefore(llvm.NextInstruction(call))
			ptr := call
			if ptr.Type() != c.i8ptr {
				ptr = c.builder.CreateBitCast(call, c.i8ptr, "")
			}
			c.builder.CreateCall(trackPointer, []llvm.Value{ptr, llvm.Undef(c.i8ptr), llvm.Undef(c.i8ptr)}, "")
		}
	}

	return nil
}
