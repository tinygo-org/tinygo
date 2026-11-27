package compiler

import (
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// This file implements volatile loads/stores in runtime/volatile.LoadT and
// runtime/volatile.StoreT as compiler builtins.

// createVolatileCall emits a call to runtime/volatile as volatile loads and
// stores at the call site, so that no call remains at any optimization level.
// It returns false for calls it leaves to the normal call path.
func (b *builder) createVolatileCall(instr *ssa.CallCommon, name string) (llvm.Value, bool, error) {
	pos := instr.Pos()
	switch {
	case strings.HasPrefix(name, "runtime/volatile.Load"):
		addr := b.getValue(instr.Args[0], pos)
		b.createNilCheck(instr.Args[0], addr, "deref")
		valType := b.getLLVMType(instr.Args[0].Type().Underlying().(*types.Pointer).Elem())
		val := b.CreateLoad(valType, addr, "")
		val.SetVolatile(true)
		return val, true, nil
	case strings.HasPrefix(name, "runtime/volatile.Store"):
		addr := b.getValue(instr.Args[0], pos)
		b.createNilCheck(instr.Args[0], addr, "deref")
		store := b.CreateStore(b.getValue(instr.Args[1], pos), addr)
		store.SetVolatile(true)
		return store, true, nil
	case strings.HasPrefix(name, "(*runtime/volatile.Register"):
		if strings.HasSuffix(name, ").Bit") || strings.Contains(name, "$") {
			// Bitband methods and bound or thunk wrappers use the normal call path.
			return llvm.Value{}, false, nil
		}
		val, err := b.createVolatileRegisterOp(name, instr.Args, pos)
		return val, true, err
	default:
		return llvm.Value{}, false, nil
	}
}

// createVolatileRegisterOp emits a volatile.RegisterN method as volatile loads
// and stores of its Reg field. args holds the receiver and the arguments.
func (b *builder) createVolatileRegisterOp(name string, args []ssa.Value, pos token.Pos) (llvm.Value, error) {
	method := name[strings.LastIndex(name, ".")+1:]
	recv := args[0]
	regType := recv.Type().Underlying().(*types.Pointer).Elem()
	fieldType := regType.Underlying().(*types.Struct).Field(0).Type()
	valType := b.getLLVMType(fieldType)
	// Reg is the only field, so the receiver is the register address.
	addr := b.getValue(recv, pos)
	switch recv.(type) {
	case *ssa.FieldAddr:
		// A FieldAddr result is never nil, its base was already checked.
	case *ssa.FreeVar:
		// A bound method receiver can be nil.
		b.createNilCheckPointer(addr, "deref")
	default:
		b.createNilCheck(recv, addr, "deref")
	}
	load := func() llvm.Value {
		val := b.CreateLoad(valType, addr, "")
		val.SetVolatile(true)
		return val
	}
	store := func(val llvm.Value) llvm.Value {
		st := b.CreateStore(val, addr)
		st.SetVolatile(true)
		return st
	}
	arg := func(i int) llvm.Value {
		return b.getValue(args[i], pos)
	}
	switch method {
	case "Get":
		return load(), nil
	case "Set":
		return store(arg(1)), nil
	case "SetBits":
		return store(b.CreateOr(load(), arg(1), "")), nil
	case "ClearBits":
		return store(b.CreateAnd(load(), b.CreateNot(arg(1), ""), "")), nil
	case "HasBits":
		bits := b.CreateAnd(load(), arg(1), "")
		return b.CreateICmp(llvm.IntNE, bits, llvm.ConstInt(valType, 0, false), ""), nil
	case "ReplaceBits":
		value, mask, shift := arg(1), arg(2), arg(3)
		old := load()
		shiftType := args[3].Type()
		shiftedMask, err := b.createBinOp(token.SHL, fieldType, shiftType, mask, shift, pos)
		if err != nil {
			return llvm.Value{}, err
		}
		shiftedValue, err := b.createBinOp(token.SHL, fieldType, shiftType, value, shift, pos)
		if err != nil {
			return llvm.Value{}, err
		}
		cleared := b.CreateAnd(old, b.CreateNot(shiftedMask, ""), "")
		return store(b.CreateOr(cleared, shiftedValue, "")), nil
	default:
		return llvm.Value{}, b.makeError(pos, "unsupported method in runtime/volatile: "+name)
	}
}

// createVolatileRegisterMethod defines a RegisterN method for method values, interfaces, defer and go.
// Depending on the optimization level, these calls may not be inlined.
func (b *builder) createVolatileRegisterMethod() {
	b.createFunctionStart(true)
	params := make([]ssa.Value, len(b.fn.Params))
	for i, param := range b.fn.Params {
		params[i] = param
	}
	val, err := b.createVolatileRegisterOp(b.fn.RelString(nil), params, getPos(b.fn))
	if err != nil {
		b.addError(b.fn.Pos(), err.Error())
		b.CreateUnreachable()
		return
	}
	if b.fn.Signature.Results().Len() == 0 {
		b.CreateRetVoid()
	} else {
		b.CreateRet(val)
	}
}

// createVolatileLoad defines runtime/volatile.LoadT for function values, defer and go.
// Depending on the optimization level, these calls may not be inlined.
func (b *builder) createVolatileLoad() {
	b.createFunctionStart(true)
	addr := b.getValue(b.fn.Params[0], getPos(b.fn))
	b.createNilCheck(b.fn.Params[0], addr, "deref")
	valType := b.getLLVMType(b.fn.Params[0].Type().(*types.Pointer).Elem())
	val := b.CreateLoad(valType, addr, "")
	val.SetVolatile(true)
	b.CreateRet(val)
}

// createVolatileStore defines runtime/volatile.StoreT for function values, defer and go.
// Depending on the optimization level, these calls may not be inlined.
func (b *builder) createVolatileStore() {
	b.createFunctionStart(true)
	addr := b.getValue(b.fn.Params[0], getPos(b.fn))
	val := b.getValue(b.fn.Params[1], getPos(b.fn))
	b.createNilCheck(b.fn.Params[0], addr, "deref")
	store := b.CreateStore(val, addr)
	store.SetVolatile(true)
	b.CreateRetVoid()
}
