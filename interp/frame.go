package interp

// This file implements the core interpretation routines, interpreting single
// functions.

import (
	"strings"

	"tinygo.org/x/go-llvm"
)

type frame struct {
	*evalPackage
	fn     llvm.Value
	locals map[llvm.Value]Value
}

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
				return nil, nil, fr.unsupportedInstructionError(inst)
			}

		// Memory operators
		case !inst.IsAAllocaInst().IsNil():
			allocType := inst.Type().ElementType()
			alloca := llvm.AddGlobal(fr.Mod, allocType, fr.packagePath+"$alloca")
			alloca.SetInitializer(llvm.ConstNull(allocType))
			alloca.SetLinkage(llvm.InternalLinkage)
			fr.locals[inst] = &LocalValue{
				Underlying: alloca,
				Eval:       fr.Eval,
			}
		case !inst.IsALoadInst().IsNil():
			operand := fr.getLocal(inst.Operand(0)).(*LocalValue)
			var value llvm.Value
			if !operand.IsConstant() || inst.IsVolatile() || (!operand.Underlying.IsAConstantExpr().IsNil() && operand.Underlying.Opcode() == llvm.BitCast) {
				value = fr.builder.CreateLoad(operand.Value(), inst.Name())
			} else {
				value = operand.Load()
			}
			if value.Type() != inst.Type() {
				return nil, nil, fr.errorAt(inst, "interp: load: type does not match")
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
					return nil, nil, fr.errorAt(inst, "todo: non-const gep")
				}
				indices[i] = uint32(operand.Value().ZExtValue())
			}
			result := value.GetElementPtr(indices)
			if result.Type() != inst.Type() {
				return nil, nil, fr.errorAt(inst, "interp: gep: type does not match")
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
			if _, ok := fr.getLocal(operand).(*MapValue); ok {
				// Special case for runtime.trackPointer calls.
				// Note: this might not be entirely sound in some rare cases
				// where the map is stored in a dirty global.
				uses := getUses(inst)
				if len(uses) == 1 {
					use := uses[0]
					if !use.IsACallInst().IsNil() && !use.CalledValue().IsAFunction().IsNil() && use.CalledValue().Name() == "runtime.trackPointer" {
						continue
					}
				}
				// It is not possible in Go to bitcast a map value to a pointer.
				return nil, nil, fr.errorAt(inst, "unimplemented: bitcast of map")
			}
			value := fr.getLocal(operand).(*LocalValue)
			fr.locals[inst] = &LocalValue{fr.Eval, fr.builder.CreateBitCast(value.Value(), inst.Type(), "")}

		// Other operators
		case !inst.IsAICmpInst().IsNil():
			lhs := fr.getLocal(inst.Operand(0)).(*LocalValue).Underlying
			rhs := fr.getLocal(inst.Operand(1)).(*LocalValue).Underlying
			predicate := inst.IntPredicate()
			if predicate == llvm.IntEQ {
				var lhsZero, rhsZero bool
				var ok1, ok2 bool
				if lhs.Type().TypeKind() == llvm.PointerTypeKind {
					// Unfortunately, the const propagation in the IR builder
					// doesn't handle pointer compares of inttoptr values. So we
					// implement it manually here.
					lhsZero, ok1 = isPointerNil(lhs)
					rhsZero, ok2 = isPointerNil(rhs)
				}
				if lhs.Type().TypeKind() == llvm.IntegerTypeKind {
					lhsZero, ok1 = isZero(lhs)
					rhsZero, ok2 = isZero(rhs)
				}
				if ok1 && ok2 {
					if lhsZero && rhsZero {
						// Both are zero, so this icmp is always evaluated to true.
						fr.locals[inst] = &LocalValue{fr.Eval, llvm.ConstInt(fr.Mod.Context().Int1Type(), 1, false)}
						continue
					}
					if lhsZero != rhsZero {
						// Only one of them is zero, so this comparison must return false.
						fr.locals[inst] = &LocalValue{fr.Eval, llvm.ConstInt(fr.Mod.Context().Int1Type(), 0, false)}
						continue
					}
				}
			}
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
						return nil, nil, fr.unsupportedInstructionError(inst)
					}
					elementCount = int(size / typeSize)
					allocType = llvm.ArrayType(allocType, elementCount)
				}
				alloc := llvm.AddGlobal(fr.Mod, allocType, fr.packagePath+"$alloc")
				alloc.SetInitializer(llvm.ConstNull(allocType))
				alloc.SetLinkage(llvm.InternalLinkage)
				result := &LocalValue{
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
					PkgName:   fr.packagePath,
					KeySize:   int(keySize),
					ValueSize: int(valueSize),
				}
			case callee.Name() == "runtime.hashmapStringSet":
				// set a string key in the map
				keyBuf := fr.getLocal(inst.Operand(1)).(*LocalValue)
				keyLen := fr.getLocal(inst.Operand(2)).(*LocalValue)
				valPtr := fr.getLocal(inst.Operand(3)).(*LocalValue)
				m, ok := fr.getLocal(inst.Operand(0)).(*MapValue)
				if !ok || !keyBuf.IsConstant() || !keyLen.IsConstant() || !valPtr.IsConstant() {
					// The mapassign operation could not be done at compile
					// time. Do it at runtime instead.
					m := fr.getLocal(inst.Operand(0)).Value()
					fr.markDirty(m)
					llvmParams := []llvm.Value{
						m,                                    // *runtime.hashmap
						fr.getLocal(inst.Operand(1)).Value(), // key.ptr
						fr.getLocal(inst.Operand(2)).Value(), // key.len
						fr.getLocal(inst.Operand(3)).Value(), // value (unsafe.Pointer)
						fr.getLocal(inst.Operand(4)).Value(), // context
						fr.getLocal(inst.Operand(5)).Value(), // parentHandle
					}
					fr.builder.CreateCall(callee, llvmParams, "")
					continue
				}
				// "key" is a Go string value, which in the TinyGo calling convention is split up
				// into separate pointer and length parameters.
				m.PutString(keyBuf, keyLen, valPtr)
			case callee.Name() == "runtime.hashmapBinarySet":
				// set a binary (int etc.) key in the map
				keyBuf := fr.getLocal(inst.Operand(1)).(*LocalValue)
				valPtr := fr.getLocal(inst.Operand(2)).(*LocalValue)
				m, ok := fr.getLocal(inst.Operand(0)).(*MapValue)
				if !ok || !keyBuf.IsConstant() || !valPtr.IsConstant() {
					// The mapassign operation could not be done at compile
					// time. Do it at runtime instead.
					m := fr.getLocal(inst.Operand(0)).Value()
					fr.markDirty(m)
					llvmParams := []llvm.Value{
						m,                                    // *runtime.hashmap
						fr.getLocal(inst.Operand(1)).Value(), // key
						fr.getLocal(inst.Operand(2)).Value(), // value
						fr.getLocal(inst.Operand(3)).Value(), // context
						fr.getLocal(inst.Operand(4)).Value(), // parentHandle
					}
					fr.builder.CreateCall(callee, llvmParams, "")
					continue
				}
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
				global := llvm.AddGlobal(fr.Mod, globalType, fr.packagePath+"$stringconcat")
				global.SetInitializer(globalValue)
				global.SetLinkage(llvm.InternalLinkage)
				global.SetGlobalConstant(true)
				global.SetUnnamedAddr(true)
				stringType := fr.Mod.GetTypeByName("runtime._string")
				retPtr := llvm.ConstGEP(global, getLLVMIndices(fr.Mod.Context().Int32Type(), []uint32{0, 0}))
				retLen := llvm.ConstInt(stringType.StructElementTypes()[1], uint64(len(result)), false)
				ret := llvm.ConstNull(stringType)
				ret = llvm.ConstInsertValue(ret, retPtr, []uint32{0})
				ret = llvm.ConstInsertValue(ret, retLen, []uint32{1})
				fr.locals[inst] = &LocalValue{fr.Eval, ret}
			case callee.Name() == "runtime.sliceCopy":
				elementSize := fr.getLocal(inst.Operand(4)).(*LocalValue).Value().ZExtValue()
				dstArray := fr.getLocal(inst.Operand(0)).(*LocalValue).stripPointerCasts()
				srcArray := fr.getLocal(inst.Operand(1)).(*LocalValue).stripPointerCasts()
				dstLen := fr.getLocal(inst.Operand(2)).(*LocalValue)
				srcLen := fr.getLocal(inst.Operand(3)).(*LocalValue)
				if elementSize != 1 && dstArray.Type().ElementType().TypeKind() == llvm.ArrayTypeKind && srcArray.Type().ElementType().TypeKind() == llvm.ArrayTypeKind {
					// Slice data pointers are created by adding a global array
					// and getting the address of the first element using a GEP.
					// However, before the compiler can pass it to
					// runtime.sliceCopy, it has to perform a bitcast to a *i8,
					// to make it a unsafe.Pointer. Now, when the IR builder
					// sees a bitcast of a GEP with zero indices, it will make
					// a bitcast of the original array instead of the GEP,
					// which breaks our assumptions.
					// Re-add this GEP, in the hope that it it is then of the correct type...
					dstArray = dstArray.GetElementPtr([]uint32{0, 0}).(*LocalValue)
					srcArray = srcArray.GetElementPtr([]uint32{0, 0}).(*LocalValue)
				}
				if fr.Eval.TargetData.TypeAllocSize(dstArray.Type().ElementType()) != elementSize {
					return nil, nil, fr.errorAt(inst, "interp: slice dst element size does not match pointer type")
				}
				if fr.Eval.TargetData.TypeAllocSize(srcArray.Type().ElementType()) != elementSize {
					return nil, nil, fr.errorAt(inst, "interp: slice src element size does not match pointer type")
				}
				if dstArray.Type() != srcArray.Type() {
					return nil, nil, fr.errorAt(inst, "interp: slice element types don't match")
				}
				length := dstLen.Value().SExtValue()
				if srcLength := srcLen.Value().SExtValue(); srcLength < length {
					length = srcLength
				}
				if length < 0 {
					return nil, nil, fr.errorAt(inst, "interp: trying to copy a slice with negative length?")
				}
				for i := int64(0); i < length; i++ {
					// *dst = *src
					dstArray.Store(srcArray.Load())
					// dst++
					dstArray = dstArray.GetElementPtr([]uint32{1}).(*LocalValue)
					// src++
					srcArray = srcArray.GetElementPtr([]uint32{1}).(*LocalValue)
				}
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
				global := llvm.AddGlobal(fr.Mod, globalType, fr.packagePath+"$bytes")
				global.SetInitializer(globalValue)
				global.SetLinkage(llvm.InternalLinkage)
				global.SetGlobalConstant(true)
				global.SetUnnamedAddr(true)
				sliceType := inst.Type()
				retPtr := llvm.ConstGEP(global, getLLVMIndices(fr.Mod.Context().Int32Type(), []uint32{0, 0}))
				retLen := llvm.ConstInt(sliceType.StructElementTypes()[1], uint64(len(result)), false)
				ret := llvm.ConstNull(sliceType)
				ret = llvm.ConstInsertValue(ret, retPtr, []uint32{0}) // ptr
				ret = llvm.ConstInsertValue(ret, retLen, []uint32{1}) // len
				ret = llvm.ConstInsertValue(ret, retLen, []uint32{2}) // cap
				fr.locals[inst] = &LocalValue{fr.Eval, ret}
			case callee.Name() == "runtime.interfaceImplements":
				typecode := fr.getLocal(inst.Operand(0)).(*LocalValue).Underlying
				interfaceMethodSet := fr.getLocal(inst.Operand(1)).(*LocalValue).Underlying
				if typecode.IsAConstantExpr().IsNil() || typecode.Opcode() != llvm.PtrToInt {
					return nil, nil, fr.errorAt(inst, "interp: expected typecode to be a ptrtoint")
				}
				typecode = typecode.Operand(0)
				if interfaceMethodSet.IsAConstantExpr().IsNil() || interfaceMethodSet.Opcode() != llvm.GetElementPtr {
					return nil, nil, fr.errorAt(inst, "interp: expected method set in runtime.interfaceImplements to be a constant gep")
				}
				interfaceMethodSet = interfaceMethodSet.Operand(0).Initializer()
				methodSet := llvm.ConstExtractValue(typecode.Initializer(), []uint32{1})
				if methodSet.IsAConstantExpr().IsNil() || methodSet.Opcode() != llvm.GetElementPtr {
					return nil, nil, fr.errorAt(inst, "interp: expected method set to be a constant gep")
				}
				methodSet = methodSet.Operand(0).Initializer()

				// Make a set of all the methods on the concrete type, for
				// easier checking in the next step.
				definedMethods := map[string]struct{}{}
				for i := 0; i < methodSet.Type().ArrayLength(); i++ {
					methodInfo := llvm.ConstExtractValue(methodSet, []uint32{uint32(i)})
					name := llvm.ConstExtractValue(methodInfo, []uint32{0}).Name()
					definedMethods[name] = struct{}{}
				}
				// Check whether all interface methods are also in the list
				// of defined methods calculated above.
				implements := uint64(1) // i1 true
				for i := 0; i < interfaceMethodSet.Type().ArrayLength(); i++ {
					name := llvm.ConstExtractValue(interfaceMethodSet, []uint32{uint32(i)}).Name()
					if _, ok := definedMethods[name]; !ok {
						// There is a method on the interface that is not
						// implemented by the type.
						implements = 0 // i1 false
						break
					}
				}
				fr.locals[inst] = &LocalValue{fr.Eval, llvm.ConstInt(fr.Mod.Context().Int1Type(), implements, false)}
			case callee.Name() == "runtime.nanotime":
				fr.locals[inst] = &LocalValue{fr.Eval, llvm.ConstInt(fr.Mod.Context().Int64Type(), 0, false)}
			case callee.Name() == "llvm.dbg.value":
				// do nothing
			case strings.HasPrefix(callee.Name(), "llvm.lifetime."):
				// do nothing
			case callee.Name() == "runtime.trackPointer":
				// do nothing
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
				scanResult, err := fr.hasSideEffects(callee)
				if err != nil {
					return nil, nil, err
				}
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
					ret, err = fr.function(callee, params, indent+"    ")
					if err != nil {
						return nil, nil, err
					}
				}
				if inst.Type().TypeKind() != llvm.VoidTypeKind {
					fr.locals[inst] = ret
				}
			default:
				// function pointers, etc.
				return nil, nil, fr.unsupportedInstructionError(inst)
			}
		case !inst.IsAExtractValueInst().IsNil():
			agg := fr.getLocal(inst.Operand(0)).(*LocalValue) // must be constant
			indices := inst.Indices()
			if agg.Underlying.IsConstant() {
				newValue := llvm.ConstExtractValue(agg.Underlying, indices)
				fr.locals[inst] = fr.getValue(newValue)
			} else {
				if len(indices) != 1 {
					return nil, nil, fr.errorAt(inst, "interp: cannot handle extractvalue with not exactly 1 index")
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
					return nil, nil, fr.errorAt(inst, "interp: cannot handle insertvalue with not exactly 1 index")
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
				return nil, nil, fr.errorAt(inst, "expected an i1 in a branch instruction")
			}
			thenBB := inst.Operand(1)
			elseBB := inst.Operand(2)
			if !cond.IsAInstruction().IsNil() {
				return nil, nil, fr.errorAt(inst, "interp: branch on a non-constant")
			}
			if !cond.IsAConstantExpr().IsNil() {
				// This may happen when the instruction builder could not
				// const-fold some instructions.
				return nil, nil, fr.errorAt(inst, "interp: branch on a non-const-propagated constant expression")
			}
			switch cond {
			case llvm.ConstInt(fr.Mod.Context().Int1Type(), 0, false): // false
				return nil, []llvm.Value{thenBB}, nil // then
			case llvm.ConstInt(fr.Mod.Context().Int1Type(), 1, false): // true
				return nil, []llvm.Value{elseBB}, nil // else
			default:
				return nil, nil, fr.errorAt(inst, "branch was not true or false")
			}
		case !inst.IsABranchInst().IsNil() && inst.OperandsCount() == 1:
			// unconditional branch (goto)
			return nil, []llvm.Value{inst.Operand(0)}, nil
		case !inst.IsAUnreachableInst().IsNil():
			// Unreachable was reached (e.g. after a call to panic()).
			// Report this as an error, as it is not supposed to happen.
			// This is a sentinel error value.
			return nil, nil, errUnreachable

		default:
			return nil, nil, fr.unsupportedInstructionError(inst)
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
		// This should not happen under normal circumstances.
		panic("cannot find value")
	}
}
