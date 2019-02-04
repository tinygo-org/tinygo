package interp

// This file implements the core interpretation routines, interpreting single
// functions.

import (
	"errors"
	"strings"

	"tinygo.org/x/go-llvm"
)

type frame struct {
	*Eval
	fn      llvm.Value
	pkgName string
	locals  map[llvm.Value]Value
}

var ErrUnreachable = errors.New("interp: unreachable executed")

// evalBasicBlock evaluates a single basic block, returning the return value (if
// ending with a ret instruction), a list of outgoing basic blocks (if not
// ending with a ret instruction), or an error on failure.
// Most of it works at compile time. Some calls get translated into calls to be
// executed at runtime: calls to functions with side effects, external calls,
// and operations on the result of such instructions.
func (fr *frame) evalBasicBlock(bb, incoming llvm.BasicBlock, indent string) (retval Value, outgoing []llvm.Value, err error) {
	for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
		if fr.Debug {
			print(indent)
			inst.Dump()
			println()
		}
		switch {
		case !inst.IsABinaryOperator().IsNil():
			lhs := fr.getLocal(inst.Operand(0)).(*LocalValue).Underlying
			rhs := fr.getLocal(inst.Operand(1)).(*LocalValue).Underlying

			switch inst.InstructionOpcode() {
			// Standard binary operators
			case llvm.Add:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateAdd(lhs, rhs, "")}
			case llvm.FAdd:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateFAdd(lhs, rhs, "")}
			case llvm.Sub:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateSub(lhs, rhs, "")}
			case llvm.FSub:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateFSub(lhs, rhs, "")}
			case llvm.Mul:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateMul(lhs, rhs, "")}
			case llvm.FMul:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateFMul(lhs, rhs, "")}
			case llvm.UDiv:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateUDiv(lhs, rhs, "")}
			case llvm.SDiv:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateSDiv(lhs, rhs, "")}
			case llvm.FDiv:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateFDiv(lhs, rhs, "")}
			case llvm.URem:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateURem(lhs, rhs, "")}
			case llvm.SRem:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateSRem(lhs, rhs, "")}
			case llvm.FRem:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateFRem(lhs, rhs, "")}

			// Logical operators
			case llvm.Shl:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateShl(lhs, rhs, "")}
			case llvm.LShr:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateLShr(lhs, rhs, "")}
			case llvm.AShr:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateAShr(lhs, rhs, "")}
			case llvm.And:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateAnd(lhs, rhs, "")}
			case llvm.Or:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateOr(lhs, rhs, "")}
			case llvm.Xor:
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateXor(lhs, rhs, "")}

			default:
				return nil, nil, &Unsupported{inst}
			}

		// Memory operators
		case !inst.IsAAllocaInst().IsNil():
			fr.locals[inst] = &AllocaValue{
				Underlying: getZeroValue(inst.Type().ElementType()),
				Dirty:      false,
				Eval:       fr.Eval,
			}
		case !inst.IsALoadInst().IsNil():
			operand := fr.getLocal(inst.Operand(0))
			var value llvm.Value
			if !operand.IsConstant() || inst.IsVolatile() {
				value = fr.builder.CreateLoad(operand.Value(), inst.Name())
			} else {
				value = operand.Load()
			}
			if value.Type() != inst.Type() {
				panic("interp: load: type does not match")
			}
			fr.locals[inst] = fr.getValue(value)
		case !inst.IsAStoreInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			ptr := fr.getLocal(inst.Operand(1))
			if inst.IsVolatile() {
				fr.builder.CreateStore(value.Value(), ptr.Value())
			} else {
				ptr.Store(value.Value())
			}
		case !inst.IsAGetElementPtrInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			llvmIndices := make([]llvm.Value, inst.OperandsCount()-1)
			for i := range llvmIndices {
				llvmIndices[i] = inst.Operand(i + 1)
			}
			indices := make([]uint32, len(llvmIndices))
			for i, llvmIndex := range llvmIndices {
				operand := fr.getLocal(llvmIndex)
				if !operand.IsConstant() {
					// Not a constant operation.
					// This should be detected by the scanner, but isn't at the
					// moment.
					panic("todo: non-const gep")
				}
				indices[i] = uint32(operand.Value().ZExtValue())
			}
			result := value.GetElementPtr(indices)
			if result.Type() != inst.Type() {
				println(" expected:", inst.Type().String())
				println(" actual:  ", result.Type().String())
				panic("interp: gep: type does not match")
			}
			fr.locals[inst] = result

		// Cast operators
		case !inst.IsATruncInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateTrunc(value.(*LocalValue).Value(), inst.Type(), "")}
		case !inst.IsAZExtInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateZExt(value.(*LocalValue).Value(), inst.Type(), "")}
		case !inst.IsASExtInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateSExt(value.(*LocalValue).Value(), inst.Type(), "")}
		case !inst.IsAFPToUIInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateFPToUI(value.(*LocalValue).Value(), inst.Type(), "")}
		case !inst.IsAFPToSIInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateFPToSI(value.(*LocalValue).Value(), inst.Type(), "")}
		case !inst.IsAUIToFPInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateUIToFP(value.(*LocalValue).Value(), inst.Type(), "")}
		case !inst.IsASIToFPInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateSIToFP(value.(*LocalValue).Value(), inst.Type(), "")}
		case !inst.IsAFPTruncInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateFPTrunc(value.(*LocalValue).Value(), inst.Type(), "")}
		case !inst.IsAFPExtInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateFPExt(value.(*LocalValue).Value(), inst.Type(), "")}
		case !inst.IsAPtrToIntInst().IsNil():
			value := fr.getLocal(inst.Operand(0))
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreatePtrToInt(value.Value(), inst.Type(), "")}
		case !inst.IsABitCastInst().IsNil() && inst.Type().TypeKind() == llvm.PointerTypeKind:
			operand := inst.Operand(0)
			if !operand.IsACallInst().IsNil() {
				fn := operand.CalledValue()
				if !fn.IsAFunction().IsNil() && fn.Name() == "runtime.alloc" {
					continue // special case: bitcast of alloc
				}
			}
			value := fr.getLocal(operand)
			if bc, ok := value.(*PointerCastValue); ok {
				value = bc.Underlying // avoid double bitcasts
			}
			fr.locals[inst] = &PointerCastValue{Eval: fr.Eval, Underlying: value, CastType: inst.Type()}

		// Other operators
		case !inst.IsAICmpInst().IsNil():
			lhs := fr.getLocal(inst.Operand(0)).(*LocalValue).Underlying
			rhs := fr.getLocal(inst.Operand(1)).(*LocalValue).Underlying
			predicate := inst.IntPredicate()
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateICmp(predicate, lhs, rhs, "")}
		case !inst.IsAFCmpInst().IsNil():
			lhs := fr.getLocal(inst.Operand(0)).(*LocalValue).Underlying
			rhs := fr.getLocal(inst.Operand(1)).(*LocalValue).Underlying
			predicate := inst.FloatPredicate()
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateFCmp(predicate, lhs, rhs, "")}
		case !inst.IsAPHINode().IsNil():
			for i := 0; i < inst.IncomingCount(); i++ {
				if inst.IncomingBlock(i) == incoming {
					fr.locals[inst] = fr.getLocal(inst.IncomingValue(i))
				}
			}
		case !inst.IsACallInst().IsNil():
			callee := inst.CalledValue()
			switch {
			case callee.Name() == "runtime.alloc":
				// heap allocation
				users := getUses(inst)
				var resultInst = inst
				if len(users) == 1 && !users[0].IsABitCastInst().IsNil() {
					// happens when allocating something other than i8*
					resultInst = users[0]
				}
				size := fr.getLocal(inst.Operand(0)).(*LocalValue).Underlying.ZExtValue()
				allocType := resultInst.Type().ElementType()
				typeSize := fr.TargetData.TypeAllocSize(allocType)
				elementCount := 1
				if size != typeSize {
					// allocate an array
					if size%typeSize != 0 {
						return nil, nil, &Unsupported{inst}
					}
					elementCount = int(size / typeSize)
					allocType = llvm.ArrayType(allocType, elementCount)
				}
				alloc := llvm.AddGlobal(fr.Mod, allocType, fr.pkgName+"$alloc")
				alloc.SetInitializer(getZeroValue(allocType))
				alloc.SetLinkage(llvm.InternalLinkage)
				result := &GlobalValue{
					Underlying: alloc,
					Eval:       fr.Eval,
				}
				if elementCount == 1 {
					fr.locals[resultInst] = result
				} else {
					fr.locals[resultInst] = result.GetElementPtr([]uint32{0, 0})
				}
			case callee.Name() == "runtime.hashmapMake":
				// create a map
				keySize := inst.Operand(0).ZExtValue()
				valueSize := inst.Operand(1).ZExtValue()
				fr.locals[inst] = &MapValue{
					Eval:      fr.Eval,
					PkgName:   fr.pkgName,
					KeySize:   int(keySize),
					ValueSize: int(valueSize),
				}
			case callee.Name() == "runtime.hashmapStringSet":
				// set a string key in the map
				m := fr.getLocal(inst.Operand(0)).(*MapValue)
				// "key" is a Go string value, which in the TinyGo calling convention is split up
				// into separate pointer and length parameters.
				keyBuf := fr.getLocal(inst.Operand(1))
				keyLen := fr.getLocal(inst.Operand(2))
				valPtr := fr.getLocal(inst.Operand(3))
				m.PutString(keyBuf, keyLen, valPtr)
			case callee.Name() == "runtime.hashmapBinarySet":
				// set a binary (int etc.) key in the map
				m := fr.getLocal(inst.Operand(0)).(*MapValue)
				keyBuf := fr.getLocal(inst.Operand(1))
				valPtr := fr.getLocal(inst.Operand(2))
				m.PutBinary(keyBuf, valPtr)
			case callee.Name() == "runtime.stringConcat":
				// adding two strings together
				buf1Ptr := fr.getLocal(inst.Operand(0))
				buf1Len := fr.getLocal(inst.Operand(1))
				buf2Ptr := fr.getLocal(inst.Operand(2))
				buf2Len := fr.getLocal(inst.Operand(3))
				buf1 := getStringBytes(buf1Ptr, buf1Len.Value())
				buf2 := getStringBytes(buf2Ptr, buf2Len.Value())
				result := []byte(string(buf1) + string(buf2))
				vals := make([]llvm.Value, len(result))
				for i := range vals {
					vals[i] = llvm.ConstInt(fr.Mod.Context().Int8Type(), uint64(result[i]), false)
				}
				globalType := llvm.ArrayType(fr.Mod.Context().Int8Type(), len(result))
				globalValue := llvm.ConstArray(fr.Mod.Context().Int8Type(), vals)
				global := llvm.AddGlobal(fr.Mod, globalType, fr.pkgName+"$stringconcat")
				global.SetInitializer(globalValue)
				global.SetLinkage(llvm.InternalLinkage)
				global.SetGlobalConstant(true)
				global.SetUnnamedAddr(true)
				stringType := fr.Mod.GetTypeByName("runtime._string")
				retPtr := llvm.ConstGEP(global, getLLVMIndices(fr.Mod.Context().Int32Type(), []uint32{0, 0}))
				retLen := llvm.ConstInt(stringType.StructElementTypes()[1], uint64(len(result)), false)
				ret := getZeroValue(stringType)
				ret = llvm.ConstInsertValue(ret, retPtr, []uint32{0})
				ret = llvm.ConstInsertValue(ret, retLen, []uint32{1})
				fr.locals[inst] = &LocalValue{fr.Eval, ret}
			case callee.Name() == "runtime.stringToBytes":
				// convert a string to a []byte
				bufPtr := fr.getLocal(inst.Operand(0))
				bufLen := fr.getLocal(inst.Operand(1))
				result := getStringBytes(bufPtr, bufLen.Value())
				vals := make([]llvm.Value, len(result))
				for i := range vals {
					vals[i] = llvm.ConstInt(fr.Mod.Context().Int8Type(), uint64(result[i]), false)
				}
				globalType := llvm.ArrayType(fr.Mod.Context().Int8Type(), len(result))
				globalValue := llvm.ConstArray(fr.Mod.Context().Int8Type(), vals)
				global := llvm.AddGlobal(fr.Mod, globalType, fr.pkgName+"$bytes")
				global.SetInitializer(globalValue)
				global.SetLinkage(llvm.InternalLinkage)
				global.SetGlobalConstant(true)
				global.SetUnnamedAddr(true)
				sliceType := inst.Type()
				retPtr := llvm.ConstGEP(global, getLLVMIndices(fr.Mod.Context().Int32Type(), []uint32{0, 0}))
				retLen := llvm.ConstInt(sliceType.StructElementTypes()[1], uint64(len(result)), false)
				ret := getZeroValue(sliceType)
				ret = llvm.ConstInsertValue(ret, retPtr, []uint32{0}) // ptr
				ret = llvm.ConstInsertValue(ret, retLen, []uint32{1}) // len
				ret = llvm.ConstInsertValue(ret, retLen, []uint32{2}) // cap
				fr.locals[inst] = &LocalValue{fr.Eval, ret}
			case callee.Name() == "runtime.makeInterface":
				uintptrType := callee.Type().Context().IntType(fr.TargetData.PointerSize() * 8)
				fr.locals[inst] = &LocalValue{fr.Eval, llvm.ConstPtrToInt(inst.Operand(0), uintptrType)}
			case strings.HasPrefix(callee.Name(), "runtime.print") || callee.Name() == "runtime._panic":
				// This are all print instructions, which necessarily have side
				// effects but no results.
				// TODO: print an error when executing runtime._panic (with the
				// exact error message it would print at runtime).
				var params []llvm.Value
				for i := 0; i < inst.OperandsCount()-1; i++ {
					operand := fr.getLocal(inst.Operand(i)).Value()
					fr.markDirty(operand)
					params = append(params, operand)
				}
				// TODO: accurate debug info, including call chain
				fr.builder.CreateCall(callee, params, inst.Name())
			case !callee.IsAFunction().IsNil() && callee.IsDeclaration():
				// external functions
				var params []llvm.Value
				for i := 0; i < inst.OperandsCount()-1; i++ {
					operand := fr.getLocal(inst.Operand(i)).Value()
					fr.markDirty(operand)
					params = append(params, operand)
				}
				// TODO: accurate debug info, including call chain
				result := fr.builder.CreateCall(callee, params, inst.Name())
				if inst.Type().TypeKind() != llvm.VoidTypeKind {
					fr.markDirty(result)
					fr.locals[inst] = &LocalValue{fr.Eval, result}
				}
			case !callee.IsAFunction().IsNil():
				// regular function
				var params []Value
				dirtyParams := false
				for i := 0; i < inst.OperandsCount()-1; i++ {
					local := fr.getLocal(inst.Operand(i))
					if !local.IsConstant() {
						dirtyParams = true
					}
					params = append(params, local)
				}
				var ret Value
				scanResult := fr.Eval.hasSideEffects(callee)
				if scanResult.severity == sideEffectLimited || dirtyParams && scanResult.severity != sideEffectAll {
					// Side effect is bounded. This means the operation invokes
					// side effects (like calling an external function) but it
					// is known at compile time which side effects it invokes.
					// This means the function can be called at runtime and the
					// affected globals can be marked dirty at compile time.
					llvmParams := make([]llvm.Value, len(params))
					for i, param := range params {
						llvmParams[i] = param.Value()
					}
					result := fr.builder.CreateCall(callee, llvmParams, inst.Name())
					ret = &LocalValue{fr.Eval, result}
					// mark all mentioned globals as dirty
					for global := range scanResult.mentionsGlobals {
						fr.markDirty(global)
					}
				} else {
					// Side effect is one of:
					//   * None: no side effects, can be fully interpreted at
					//     compile time.
					//   * Unbounded: cannot call at runtime so we'll try to
					//     interpret anyway and hope for the best.
					ret, err = fr.function(callee, params, fr.pkgName, indent+"    ")
					if err != nil {
						return nil, nil, err
					}
				}
				if inst.Type().TypeKind() != llvm.VoidTypeKind {
					fr.locals[inst] = ret
				}
			default:
				// function pointers, etc.
				return nil, nil, &Unsupported{inst}
			}
		case !inst.IsAExtractValueInst().IsNil():
			agg := fr.getLocal(inst.Operand(0)).(*LocalValue) // must be constant
			indices := inst.Indices()
			if agg.Underlying.IsConstant() {
				newValue := llvm.ConstExtractValue(agg.Underlying, indices)
				fr.locals[inst] = fr.getValue(newValue)
			} else {
				if len(indices) != 1 {
					return nil, nil, errors.New("cannot handle extractvalue with not exactly 1 index")
				}
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateExtractValue(agg.Underlying, int(indices[0]), inst.Name())}
			}
		case !inst.IsAInsertValueInst().IsNil():
			agg := fr.getLocal(inst.Operand(0)).(*LocalValue) // must be constant
			val := fr.getLocal(inst.Operand(1))
			indices := inst.Indices()
			if agg.IsConstant() && val.IsConstant() {
				newValue := llvm.ConstInsertValue(agg.Underlying, val.Value(), indices)
				fr.locals[inst] = &LocalValue{fr.Eval, newValue}
			} else {
				if len(indices) != 1 {
					return nil, nil, errors.New("cannot handle insertvalue with not exactly 1 index")
				}
				fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateInsertValue(agg.Underlying, val.Value(), int(indices[0]), inst.Name())}
			}

		case !inst.IsAReturnInst().IsNil() && inst.OperandsCount() == 0:
			return nil, nil, nil // ret void
		case !inst.IsAReturnInst().IsNil() && inst.OperandsCount() == 1:
			return fr.getLocal(inst.Operand(0)), nil, nil
		case !inst.IsABranchInst().IsNil() && inst.OperandsCount() == 3:
			// conditional branch (if/then/else)
			cond := fr.getLocal(inst.Operand(0)).Value()
			if cond.Type() != fr.Mod.Context().Int1Type() {
				panic("expected an i1 in a branch instruction")
			}
			thenBB := inst.Operand(1)
			elseBB := inst.Operand(2)
			if !cond.IsConstant() {
				return nil, nil, errors.New("interp: branch on a non-constant")
			} else {
				switch cond.ZExtValue() {
				case 0: // false
					return nil, []llvm.Value{thenBB}, nil // then
				case 1: // true
					return nil, []llvm.Value{elseBB}, nil // else
				default:
					panic("branch was not true or false")
				}
			}
		case !inst.IsABranchInst().IsNil() && inst.OperandsCount() == 1:
			// unconditional branch (goto)
			return nil, []llvm.Value{inst.Operand(0)}, nil
		case !inst.IsAUnreachableInst().IsNil():
			// Unreachable was reached (e.g. after a call to panic()).
			// Report this as an error, as it is not supposed to happen.
			return nil, nil, ErrUnreachable

		default:
			return nil, nil, &Unsupported{inst}
		}
	}

	panic("interp: reached end of basic block without terminator")
}

// Get the Value for an operand, which is a constant value of some sort.
func (fr *frame) getLocal(v llvm.Value) Value {
	if ret, ok := fr.locals[v]; ok {
		return ret
	} else if value := fr.getValue(v); value != nil {
		return value
	} else {
		panic("cannot find value")
	}
}
