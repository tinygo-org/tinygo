package transform

// This file implements several small optimizations of runtime and reflect
// calls.

import (
	"slices"
	"strings"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// OptimizeStringToBytes transforms runtime.stringToBytes(...) calls into const
// []byte slices whenever possible. This optimizes the following pattern:
//
//	w.Write([]byte("foo"))
//
// where Write does not store to the slice.
func OptimizeStringToBytes(mod llvm.Module) {
	stringToBytes := mod.NamedFunction("runtime.stringToBytes")
	if stringToBytes.IsNil() {
		// nothing to optimize
		return
	}

	for _, call := range getUses(stringToBytes) {
		strptr := call.Operand(0)
		strlen := call.Operand(1)

		// strptr is always constant because strings are always constant.

		var pointerUses []llvm.Value
		canConvertPointer := true
		for _, use := range getUses(call) {
			if use.IsAExtractValueInst().IsNil() {
				// Expected an extractvalue, but this is something else.
				canConvertPointer = false
				break
			}
			switch use.Type().TypeKind() {
			case llvm.IntegerTypeKind:
				// A length (len or cap). Propagate the length value.
				// This can always be done because the byte slice is always the
				// same length as the original string.
				use.ReplaceAllUsesWith(strlen)
				use.EraseFromParentAsInstruction()
			case llvm.PointerTypeKind:
				// The string pointer itself.
				if !isReadOnly(use) {
					// There is a store to the byte slice. This means that none
					// of the pointer uses can't be propagated.
					canConvertPointer = false
					break
				}
				// It may be that the pointer value can be propagated, if all of
				// the pointer uses are readonly.
				pointerUses = append(pointerUses, use)
			default:
				// should not happen
				panic("unknown return type of runtime.stringToBytes: " + use.Type().String())
			}
		}
		if canConvertPointer {
			// All pointer uses are readonly, so they can be converted.
			for _, use := range pointerUses {
				use.ReplaceAllUsesWith(strptr)
				use.EraseFromParentAsInstruction()
			}

			// Call to runtime.stringToBytes can be eliminated: both the input
			// and the output is constant.
			call.EraseFromParentAsInstruction()
		}
	}
}

// OptimizeStringEqual transforms runtime.stringEqual(...) calls into simple
// integer comparisons if at least one of the sides of the comparison is zero.
// This converts str == "" into len(str) == 0 and "" == "" into true.
func OptimizeStringEqual(mod llvm.Module) {
	stringEqual := mod.NamedFunction("runtime.stringEqual")
	if stringEqual.IsNil() {
		// nothing to optimize
		return
	}

	builder := mod.Context().NewBuilder()
	defer builder.Dispose()

	for _, call := range getUses(stringEqual) {
		str1len := call.Operand(1)
		str2len := call.Operand(3)

		zero := llvm.ConstInt(str1len.Type(), 0, false)
		if str1len == zero || str2len == zero {
			builder.SetInsertPointBefore(call)
			icmp := builder.CreateICmp(llvm.IntEQ, str1len, str2len, "")
			call.ReplaceAllUsesWith(icmp)
			call.EraseFromParentAsInstruction()
			continue
		}
	}
}

// OptimizeStringFromBytes removes the copy in runtime.stringFromBytes calls
// when the string is only read by calls in the same basic block. This
// optimizes the following pattern:
//
//	utf8.RuneCountInString(string(p))
func OptimizeStringFromBytes(mod llvm.Module) {
	stringFromBytes := mod.NamedFunction("runtime.stringFromBytes")
	if stringFromBytes.IsNil() {
		// nothing to optimize
		return
	}

	for _, call := range getUses(stringFromBytes) {
		var strLen, strPtr []llvm.Value
		ok := true
		for _, use := range getUses(call) {
			if use.IsAExtractValueInst().IsNil() {
				ok = false
				break
			}
			if use.Type().TypeKind() == llvm.PointerTypeKind {
				strPtr = append(strPtr, use)
			} else {
				strLen = append(strLen, use)
			}
		}
		for _, ptr := range strPtr {
			if !ok || !isTransientReadOnlyPointer(call, ptr) {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}

		srcPtr := call.Operand(0)
		srcLen := call.Operand(1)
		for _, ptr := range strPtr {
			ptr.ReplaceAllUsesWith(srcPtr)
			ptr.EraseFromParentAsInstruction()
		}
		for _, l := range strLen {
			l.ReplaceAllUsesWith(srcLen)
			l.EraseFromParentAsInstruction()
		}
		call.EraseFromParentAsInstruction()
	}
}

// isTransientReadOnlyPointer returns whether ptr can alias the source bytes
// when only calls read it, none keep it and nothing writes before the last read.
func isTransientReadOnlyPointer(conversion, ptr llvm.Value) bool {
	block := conversion.InstructionParent()
	uses := getUses(ptr)
	if len(uses) == 0 {
		return true
	}
	pending := make(map[llvm.Value]bool)
	source := originOf(conversion.Operand(0))
	for _, use := range uses {
		if isTrackPointer(use, ptr) {
			// The copy needs no GC root once it is removed.
			continue
		}
		// The string must be used only by direct calls, as in f(string(b)).
		if use.IsACallInst().IsNil() {
			return false
		}
		callee := use.CalledValue()
		if callee.IsAFunction().IsNil() {
			return false
		}
		// Only this block is walked, so a use elsewhere could run after a write.
		if use.InstructionParent() != block {
			return false
		}
		if !callWritesOnlyLocal(use, callee, source) {
			return false
		}
		// The callee must not keep the pointer, including by returning it.
		if !isNoCaptureArg(use, ptr) {
			return false
		}
		// Variadic positions have no parameter attributes, so they count as captures.
		for i := callee.ParamsCount(); i < use.OperandsCount(); i++ {
			if use.Operand(i) == ptr {
				return false
			}
		}
		pending[use] = true
	}

	// Walk from the conversion to the last use. Any instruction that might
	// write memory before then could modify the bytes the string refers to.
	for inst := llvm.NextInstruction(conversion); !inst.IsNil() && len(pending) > 0; inst = llvm.NextInstruction(inst) {
		if pending[inst] {
			delete(pending, inst)
			continue
		}
		switch inst.InstructionOpcode() {
		case llvm.Store:
			// Stores to allocas that cannot alias the source are compiler
			// temporaries.
			if source.mayAlias(originOf(inst.Operand(1))) {
				return false
			}
		// Calls that cannot write the source, such as readonly intrinsics or llvm.dbg.*,
		// could be accepted. The test corpus showed no benefit on native targets.
		case llvm.Call:
			name := inst.CalledValue().Name()
			if name == "runtime.trackPointer" {
				// Becomes a store to the GC stack slot in MakeGCStackSlots.
				break
			}
			// Lifetime markers make the source bytes undefined (LangRef, Lifetime
			// Intrinsics). LLVM 23 removed the size operand, so the pointer is last.
			if !strings.HasPrefix(name, "llvm.lifetime.") ||
				source.mayAlias(originOf(inst.Operand(inst.OperandsCount()-2))) {
				return false
			}
		case llvm.Alloca, llvm.ExtractValue, llvm.InsertValue, llvm.GetElementPtr, llvm.BitCast,
			llvm.ICmp, llvm.Select, llvm.ZExt, llvm.SExt, llvm.Trunc,
			llvm.Add, llvm.Sub, llvm.Mul, llvm.And, llvm.Or, llvm.Xor,
			llvm.Shl, llvm.LShr, llvm.AShr, llvm.Load:
		default:
			// Anything else may write, including memcpy, atomics, fences and other calls.
			return false
		}
	}
	return len(pending) == 0
}

// isReadOnlyFunction returns whether the function is known not to write any
// memory.
func isReadOnlyFunction(fn llvm.Value) bool {
	if !fn.GetEnumFunctionAttribute(llvm.AttributeKindID("readnone")).IsNil() ||
		!fn.GetEnumFunctionAttribute(llvm.AttributeKindID("readonly")).IsNil() {
		return true
	}
	// memory(...) uses two bits per location and 2 means mod. See ModRef.h in
	// LLVM.
	attr := fn.GetEnumFunctionAttribute(llvm.AttributeKindID("memory"))
	return !attr.IsNil() && attr.GetEnumValue()&0xaaaa == 0
}

// callWritesOnlyLocal returns whether the call writes only callee-local memory
// or allocas that do not alias source.
func callWritesOnlyLocal(call, callee llvm.Value, source pointerOrigin) bool {
	if !writesArgsOnly(callee) {
		return writesOnlyLocalMemory(callee, 0)
	}
	for i := 0; i < callee.ParamsCount(); i++ {
		arg := call.Operand(i)
		if arg.Type().TypeKind() != llvm.PointerTypeKind {
			continue
		}
		readonly := !callee.GetEnumAttributeAtIndex(i+1, llvm.AttributeKindID("readonly")).IsNil()
		local := !source.mayAlias(originOf(arg)) || !arg.IsAUndefValue().IsNil() || !arg.IsAConstantPointerNull().IsNil()
		if !readonly && !local {
			return false
		}
	}
	return true
}

func writesArgsOnly(fn llvm.Value) bool {
	return !fn.GetStringAttributeAtIndex(-1, llvmutil.WritesArgsOnlyAttrName).IsNil()
}

// writesOnlyLocalMemory returns whether fn writes only to its own allocas.
func writesOnlyLocalMemory(fn llvm.Value, depth int) bool {
	if isReadOnlyFunction(fn) {
		return true
	}
	if fn.IsDeclaration() || depth > 4 {
		return false
	}
	// At depth > 0 an argument is a caller alloca or a readonly parameter.
	isLocal := func(ptr llvm.Value) bool {
		if !ptr.IsAUndefValue().IsNil() || !ptr.IsAConstantPointerNull().IsNil() {
			return true
		}
		return isLocalAlloca(ptr) || (depth > 0 && isArgumentPointer(ptr))
	}
	for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
		for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
			if !inst.IsABranchInst().IsNil() {
				// Br was split in two opcodes in LLVM 23.
				continue
			}
			switch inst.InstructionOpcode() {
			case llvm.Store:
				if !isLocal(inst.Operand(1)) {
					return false
				}
			case llvm.Call:
				callee := inst.CalledValue()
				if callee.IsAFunction().IsNil() {
					return false
				}
				if strings.HasPrefix(callee.Name(), "llvm.lifetime.") || isReadOnlyFunction(callee) {
					continue
				}
				for i := 0; i < callee.ParamsCount(); i++ {
					arg := inst.Operand(i)
					if arg.Type().TypeKind() == llvm.PointerTypeKind && !isLocal(arg) &&
						callee.GetEnumAttributeAtIndex(i+1, llvm.AttributeKindID("readonly")).IsNil() {
						return false
					}
				}
				attr := callee.GetEnumFunctionAttribute(llvm.AttributeKindID("memory"))
				if (!attr.IsNil() && attr.GetEnumValue()&^0x3 == 0) || writesArgsOnly(callee) {
					// memory(argmem: ...) only writes through arguments, which were
					// checked above.
					continue
				}
				if !writesOnlyLocalMemory(callee, depth+1) {
					return false
				}
			case llvm.Ret, llvm.Switch, llvm.Unreachable, llvm.Load, llvm.Alloca,
				llvm.GetElementPtr, llvm.BitCast, llvm.PHI, llvm.Select, llvm.ICmp,
				llvm.ExtractValue, llvm.InsertValue, llvm.ZExt, llvm.SExt, llvm.Trunc,
				llvm.Add, llvm.Sub, llvm.Mul, llvm.And, llvm.Or, llvm.Xor,
				llvm.Shl, llvm.LShr, llvm.AShr, llvm.UDiv, llvm.URem:
			default:
				// Atomics, fences, invoke and anything unknown.
				return false
			}
		}
	}
	return true
}

// isLocalAlloca returns whether ptr is derived from an alloca.
func isLocalAlloca(ptr llvm.Value) bool {
	return derivedFrom(ptr, func(v llvm.Value) bool { return !v.IsAAllocaInst().IsNil() })
}

// isArgumentPointer returns whether ptr is derived from a function parameter.
func isArgumentPointer(ptr llvm.Value) bool {
	return derivedFrom(ptr, func(v llvm.Value) bool { return !v.IsAArgument().IsNil() })
}

func derivedFrom(ptr llvm.Value, match func(llvm.Value) bool) bool {
	for {
		switch {
		case match(ptr):
			return true
		case !ptr.IsAGetElementPtrInst().IsNil(), !ptr.IsABitCastInst().IsNil():
			ptr = ptr.Operand(0)
		default:
			return false
		}
	}
}

// isNoCaptureArg returns whether param is fully nocapture in call. This rejects
// captures(ret: ...), which allows a returned alias (LangRef, Parameter Attributes).
func isNoCaptureArg(call, param llvm.Value) bool {
	fn := call.CalledValue()
	kind := llvm.AttributeKindID(llvmutil.NoCaptureAttrName())
	for i := 0; i < fn.ParamsCount(); i++ {
		if call.Operand(i) == param && !llvmutil.IsNoCapture(fn.GetEnumAttributeAtIndex(i+1, kind)) {
			return false
		}
	}
	return true
}

// pointerOrigin describes what a pointer may point to. Allocas are tracked
// through GEPs, bitcasts, selects and phis. Anything else is unknown.
type pointerOrigin struct {
	allocas []llvm.Value
	// An unknown origin can only reach allocas whose address escaped.
	unknown bool
}

func originOf(ptr llvm.Value) pointerOrigin {
	var o pointerOrigin
	o.collect(ptr, map[llvm.Value]bool{})
	return o
}

func (o *pointerOrigin) collect(ptr llvm.Value, seen map[llvm.Value]bool) {
	if seen[ptr] {
		return
	}
	seen[ptr] = true
	switch {
	case !ptr.IsAAllocaInst().IsNil():
		o.allocas = append(o.allocas, ptr)
	case !ptr.IsAGetElementPtrInst().IsNil(), !ptr.IsABitCastInst().IsNil():
		o.collect(ptr.Operand(0), seen)
	case !ptr.IsASelectInst().IsNil():
		o.collect(ptr.Operand(1), seen)
		o.collect(ptr.Operand(2), seen)
	case !ptr.IsAPHINode().IsNil():
		for i := 0; i < ptr.IncomingCount(); i++ {
			o.collect(ptr.IncomingValue(i), seen)
		}
	default:
		o.unknown = true
	}
}

// mayAlias returns whether the memory o points to may overlap memory target
// points to. An unknown origin only reaches an alloca whose address escaped.
// If target is unknown the answer is always yes.
func (o pointerOrigin) mayAlias(target pointerOrigin) bool {
	if target.unknown {
		return true
	}
	for _, t := range target.allocas {
		if slices.Contains(o.allocas, t) {
			return true
		}
		if o.unknown && allocaEscapes(t, map[llvm.Value]bool{}) {
			return true
		}
	}
	return false
}

// allocaEscapes returns whether the address of the alloca may be stored or
// captured.
func allocaEscapes(ptr llvm.Value, seen map[llvm.Value]bool) bool {
	if seen[ptr] {
		return false
	}
	seen[ptr] = true
	for _, use := range getUses(ptr) {
		switch {
		case !use.IsALoadInst().IsNil():
		case !use.IsAStoreInst().IsNil():
			if use.Operand(0) == ptr {
				return true
			}
		case !use.IsAGetElementPtrInst().IsNil(), !use.IsABitCastInst().IsNil(),
			!use.IsASelectInst().IsNil(), !use.IsAPHINode().IsNil():
			if allocaEscapes(use, seen) {
				return true
			}
		case !use.IsACallInst().IsNil():
			callee := use.CalledValue()
			if callee.IsAFunction().IsNil() || !isNoCaptureArg(use, ptr) {
				return true
			}
			for i := callee.ParamsCount(); i < use.OperandsCount()-1; i++ {
				if use.Operand(i) == ptr {
					return true
				}
			}
		default:
			return true
		}
	}
	return false
}

// isTrackPointer returns whether use is a call to runtime.trackPointer that
// tracks ptr.
func isTrackPointer(use, ptr llvm.Value) bool {
	if use.IsACallInst().IsNil() || use.CalledValue().Name() != "runtime.trackPointer" {
		return false
	}
	return use.Operand(0) == ptr && use.Operand(1) != ptr
}
