package compiler

// This file lowers channel operations (make/send/recv/close) to runtime calls
// or pseudo-operations that are lowered during goroutine lowering.

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// emitMakeChan returns a new channel value for the given channel type.
func (c *Compiler) emitMakeChan(expr *ssa.MakeChan) (llvm.Value, error) {
	chanType := c.getLLVMType(expr.Type())
	size := c.targetData.TypeAllocSize(chanType.ElementType())
	sizeValue := llvm.ConstInt(c.uintptrType, size, false)
	ptr := c.createRuntimeCall("alloc", []llvm.Value{sizeValue}, "chan.alloc")
	ptr = c.builder.CreateBitCast(ptr, chanType, "chan")
	// Set the elementSize field
	elementSizePtr := c.builder.CreateGEP(ptr, []llvm.Value{
		llvm.ConstInt(c.ctx.Int32Type(), 0, false),
		llvm.ConstInt(c.ctx.Int32Type(), 0, false),
	}, "")
	elementSize := c.targetData.TypeAllocSize(c.getLLVMType(expr.Type().(*types.Chan).Elem()))
	if elementSize > 0xffff {
		return ptr, c.makeError(expr.Pos(), fmt.Sprintf("element size is %d bytes, which is bigger than the maximum of %d bytes", elementSize, 0xffff))
	}
	elementSizeValue := llvm.ConstInt(c.ctx.Int16Type(), elementSize, false)
	c.builder.CreateStore(elementSizeValue, elementSizePtr)
	return ptr, nil
}

// emitChanSend emits a pseudo chan send operation. It is lowered to the actual
// channel send operation during goroutine lowering.
func (c *Compiler) emitChanSend(frame *Frame, instr *ssa.Send) {
	ch := c.getValue(frame, instr.Chan)
	chanValue := c.getValue(frame, instr.X)

	// store value-to-send
	valueType := c.getLLVMType(instr.X.Type())
	valueAlloca, valueAllocaCast, valueAllocaSize := c.createTemporaryAlloca(valueType, "chan.value")
	c.builder.CreateStore(chanValue, valueAlloca)

	// Do the send.
	coroutine := c.createRuntimeCall("getCoroutine", nil, "")
	c.createRuntimeCall("chanSend", []llvm.Value{coroutine, ch, valueAllocaCast}, "")

	// End the lifetime of the alloca.
	// This also works around a bug in CoroSplit, at least in LLVM 8:
	// https://bugs.llvm.org/show_bug.cgi?id=41742
	c.emitLifetimeEnd(valueAllocaCast, valueAllocaSize)
}

// emitChanRecv emits a pseudo chan receive operation. It is lowered to the
// actual channel receive operation during goroutine lowering.
func (c *Compiler) emitChanRecv(frame *Frame, unop *ssa.UnOp) llvm.Value {
	valueType := c.getLLVMType(unop.X.Type().(*types.Chan).Elem())
	ch := c.getValue(frame, unop.X)

	// Allocate memory to receive into.
	valueAlloca, valueAllocaCast, valueAllocaSize := c.createTemporaryAlloca(valueType, "chan.value")

	// Do the receive.
	coroutine := c.createRuntimeCall("getCoroutine", nil, "")
	c.createRuntimeCall("chanRecv", []llvm.Value{coroutine, ch, valueAllocaCast}, "")
	received := c.builder.CreateLoad(valueAlloca, "chan.received")
	c.emitLifetimeEnd(valueAllocaCast, valueAllocaSize)

	if unop.CommaOk {
		commaOk := c.createRuntimeCall("getTaskPromiseData", []llvm.Value{coroutine}, "chan.commaOk.wide")
		commaOk = c.builder.CreateTrunc(commaOk, c.ctx.Int1Type(), "chan.commaOk")
		tuple := llvm.Undef(c.ctx.StructType([]llvm.Type{valueType, c.ctx.Int1Type()}, false))
		tuple = c.builder.CreateInsertValue(tuple, received, 0, "")
		tuple = c.builder.CreateInsertValue(tuple, commaOk, 1, "")
		return tuple
	} else {
		return received
	}
}

// emitChanClose closes the given channel.
func (c *Compiler) emitChanClose(frame *Frame, param ssa.Value) {
	ch := c.getValue(frame, param)
	c.createRuntimeCall("chanClose", []llvm.Value{ch}, "")
}

// emitSelect emits all IR necessary for a select statements. That's a
// non-trivial amount of code because select is very complex to implement.
func (c *Compiler) emitSelect(frame *Frame, expr *ssa.Select) llvm.Value {
	if len(expr.States) == 0 {
		// Shortcuts for some simple selects.
		llvmType := c.getLLVMType(expr.Type())
		if expr.Blocking {
			// Blocks forever:
			//     select {}
			c.createRuntimeCall("deadlockStub", nil, "")
			return llvm.Undef(llvmType)
		} else {
			// No-op:
			//     select {
			//     default:
			//     }
			retval := llvm.Undef(llvmType)
			retval = c.builder.CreateInsertValue(retval, llvm.ConstInt(c.intType, 0xffffffffffffffff, true), 0, "")
			return retval // {-1, false}
		}
	}

	// This code create a (stack-allocated) slice containing all the select
	// cases and then calls runtime.chanSelect to perform the actual select
	// statement.
	// Simple selects (blocking and with just one case) are already transformed
	// into regular chan operations during SSA construction so we don't have to
	// optimize such small selects.

	// Go through all the cases. Create the selectStates slice and and
	// determine the receive buffer size and alignment.
	recvbufSize := uint64(0)
	recvbufAlign := 0
	hasReceives := false
	var selectStates []llvm.Value
	chanSelectStateType := c.getLLVMRuntimeType("chanSelectState")
	for _, state := range expr.States {
		ch := c.getValue(frame, state.Chan)
		selectState := c.getZeroValue(chanSelectStateType)
		selectState = c.builder.CreateInsertValue(selectState, ch, 0, "")
		switch state.Dir {
		case types.RecvOnly:
			// Make sure the receive buffer is big enough and has the correct alignment.
			llvmType := c.getLLVMType(state.Chan.Type().(*types.Chan).Elem())
			if size := c.targetData.TypeAllocSize(llvmType); size > recvbufSize {
				recvbufSize = size
			}
			if align := c.targetData.ABITypeAlignment(llvmType); align > recvbufAlign {
				recvbufAlign = align
			}
			hasReceives = true
		case types.SendOnly:
			// Store this value in an alloca and put a pointer to this alloca
			// in the send state.
			sendValue := c.getValue(frame, state.Send)
			alloca := c.createEntryBlockAlloca(sendValue.Type(), "select.send.value")
			c.builder.CreateStore(sendValue, alloca)
			ptr := c.builder.CreateBitCast(alloca, c.i8ptrType, "")
			selectState = c.builder.CreateInsertValue(selectState, ptr, 1, "")
		default:
			panic("unreachable")
		}
		selectStates = append(selectStates, selectState)
	}

	// Create a receive buffer, where the received value will be stored.
	recvbuf := llvm.Undef(c.i8ptrType)
	if hasReceives {
		allocaType := llvm.ArrayType(c.ctx.Int8Type(), int(recvbufSize))
		recvbufAlloca := c.builder.CreateAlloca(allocaType, "select.recvbuf.alloca")
		recvbufAlloca.SetAlignment(recvbufAlign)
		recvbuf = c.builder.CreateGEP(recvbufAlloca, []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
		}, "select.recvbuf")
	}

	// Create the states slice (allocated on the stack).
	statesAllocaType := llvm.ArrayType(chanSelectStateType, len(selectStates))
	statesAlloca := c.builder.CreateAlloca(statesAllocaType, "select.states.alloca")
	for i, state := range selectStates {
		// Set each slice element to the appropriate channel.
		gep := c.builder.CreateGEP(statesAlloca, []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false),
		}, "")
		c.builder.CreateStore(state, gep)
	}
	statesPtr := c.builder.CreateGEP(statesAlloca, []llvm.Value{
		llvm.ConstInt(c.ctx.Int32Type(), 0, false),
		llvm.ConstInt(c.ctx.Int32Type(), 0, false),
	}, "select.states")
	statesLen := llvm.ConstInt(c.uintptrType, uint64(len(selectStates)), false)

	// Convert the 'blocking' flag on this select into a LLVM value.
	blockingInt := uint64(0)
	if expr.Blocking {
		blockingInt = 1
	}
	blockingValue := llvm.ConstInt(c.ctx.Int1Type(), blockingInt, false)

	// Do the select in the runtime.
	results := c.createRuntimeCall("chanSelect", []llvm.Value{
		recvbuf,
		statesPtr, statesLen, statesLen, // []chanSelectState
		blockingValue,
	}, "")

	// The result value does not include all the possible received values,
	// because we can't load them in advance. Instead, the *ssa.Extract
	// instruction will treat a *ssa.Select specially and load it there inline.
	// Store the receive alloca in a sidetable until we hit this extract
	// instruction.
	if frame.selectRecvBuf == nil {
		frame.selectRecvBuf = make(map[*ssa.Select]llvm.Value)
	}
	frame.selectRecvBuf[expr] = recvbuf

	return results
}

// getChanSelectResult returns the special values from a *ssa.Extract expression
// when extracting a value from a select statement (*ssa.Select). Because
// *ssa.Select cannot load all values in advance, it does this later in the
// *ssa.Extract expression.
func (c *Compiler) getChanSelectResult(frame *Frame, expr *ssa.Extract) llvm.Value {
	if expr.Index == 0 {
		// index
		value := c.getValue(frame, expr.Tuple)
		index := c.builder.CreateExtractValue(value, expr.Index, "")
		if index.Type().IntTypeWidth() < c.intType.IntTypeWidth() {
			index = c.builder.CreateSExt(index, c.intType, "")
		}
		return index
	} else if expr.Index == 1 {
		// comma-ok
		value := c.getValue(frame, expr.Tuple)
		return c.builder.CreateExtractValue(value, expr.Index, "")
	} else {
		// Select statements are (index, ok, ...) where ... is a number of
		// received values, depending on how many receive statements there
		// are. They are all combined into one alloca (because only one
		// receive can proceed at a time) so we'll get that alloca, bitcast
		// it to the correct type, and dereference it.
		recvbuf := frame.selectRecvBuf[expr.Tuple.(*ssa.Select)]
		typ := llvm.PointerType(c.getLLVMType(expr.Type()), 0)
		ptr := c.builder.CreateBitCast(recvbuf, typ, "")
		return c.builder.CreateLoad(ptr, "")
	}
}
