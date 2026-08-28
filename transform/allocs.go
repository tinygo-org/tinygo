package transform

// This file implements an escape analysis pass. It looks for calls to
// runtime.alloc and replaces these calls with a stack allocation if the
// allocated value does not escape. It uses the LLVM nocapture flag for
// interprocedural escape analysis.

import (
	"bufio"
	"fmt"
	"go/token"
	"os"
	"regexp"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// OptimizeAllocs tries to replace heap allocations with stack allocations
// whenever possible. It relies on the LLVM 'nocapture' flag for interprocedural
// escape analysis, and within a function looks whether an allocation can escape
// to the heap.
// If printAllocs is non-nil, it indicates the regexp of functions for which a
// heap allocation explanation should be printed (why the object can't be stack
// allocated).
func OptimizeAllocs(mod llvm.Module, printAllocs *regexp.Regexp, maxStackAlloc uint64, logger func(token.Position, string)) {
	// Find allocator functions.
	var allocators []llvm.Value
	for _, name := range []string{"runtime.alloc", "runtime.alloc_noheap"} {
		allocator := mod.NamedFunction(name)
		if !allocator.IsNil() {
			allocators = append(allocators, allocator)
		}
	}
	if len(allocators) == 0 {
		// nothing to optimize
		return
	}

	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	ctx := mod.Context()
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	// Determine the maximum alignment on this platform.
	complex128Type := ctx.StructType([]llvm.Type{ctx.DoubleType(), ctx.DoubleType()}, false)
	maxAlign := int64(targetData.ABITypeAlignment(complex128Type))

	// Find all allocator calls.
	var heapallocs []llvm.Value
	for _, allocator := range allocators {
		heapallocs = append(heapallocs, getUses(allocator)...)
	}

	for _, heapalloc := range heapallocs {
		logAllocs := printAllocs != nil && printAllocs.MatchString(heapalloc.InstructionParent().Parent().Name())
		if heapalloc.Operand(0).IsAConstantInt().IsNil() {
			// Do not allocate variable length arrays on the stack.
			if logAllocs {
				logger(getPosition(heapalloc), "size is not constant")
			}
			continue
		}

		size := heapalloc.Operand(0).ZExtValue()
		if size > maxStackAlloc {
			// The maximum size for a stack allocation.
			if logAllocs {
				logger(getPosition(heapalloc),
					fmt.Sprintf("object size %d exceeds maximum stack allocation size %d", size, maxStackAlloc))
			}
			continue
		}

		if size == 0 {
			// If the size is 0, the pointer is allowed to alias other
			// zero-sized pointers. Use the pointer to the global that would
			// also be returned by runtime.alloc.
			zeroSizedAlloc := mod.NamedGlobal("runtime.zeroSizedAlloc")
			if !zeroSizedAlloc.IsNil() {
				heapalloc.ReplaceAllUsesWith(zeroSizedAlloc)
				heapalloc.EraseFromParentAsInstruction()
			}
			continue
		}

		// In general the pattern is:
		//     %0 = call i8* @runtime.alloc(i32 %size, i8* null)
		//     %1 = bitcast i8* %0 to type*
		//     (use %1 only)
		// But the bitcast might sometimes be dropped when allocating an *i8.
		// The 'bitcast' variable below is thus usually a bitcast of the
		// heapalloc but not always.
		bitcast := heapalloc // instruction that creates the value
		if uses := getUses(heapalloc); len(uses) == 1 && !uses[0].IsABitCastInst().IsNil() {
			// getting only bitcast use
			bitcast = uses[0]
		}

		if at := valueEscapesAt(bitcast); !at.IsNil() {
			if logAllocs {
				atPos := getPosition(at)
				msg := "escapes at unknown line"
				if atPos.Line != 0 {
					msg = fmt.Sprintf("escapes at line %d", atPos.Line)
				}
				logger(getPosition(heapalloc), msg)
			}
			continue
		}
		// The pointer value does not escape.

		// Determine the appropriate alignment of the alloca.
		attr := heapalloc.GetCallSiteEnumAttribute(0, llvm.AttributeKindID("align"))
		alignment := int(maxAlign)
		if !attr.IsNil() {
			// 'align' return value attribute is set, so use it.
			// This is basically always the case, but to be sure we'll default
			// to maxAlign if it isn't.
			alignment = int(attr.GetEnumValue())
		}

		// Insert alloca in the entry block. Do it here so that mem2reg can
		// promote it to a SSA value.
		fn := bitcast.InstructionParent().Parent()
		builder.SetInsertPointBefore(fn.EntryBasicBlock().FirstInstruction())
		allocaType := llvm.ArrayType(mod.Context().Int8Type(), int(size))
		alloca := builder.CreateAlloca(allocaType, "stackalloc")
		alloca.SetAlignment(alignment)

		// Zero the allocation inside the block where the value was originally allocated.
		zero := llvm.ConstNull(alloca.AllocatedType())
		builder.SetInsertPointBefore(bitcast)
		store := builder.CreateStore(zero, alloca)
		store.SetAlignment(alignment)

		// Replace heap alloc bitcast with stack alloc bitcast.
		bitcast.ReplaceAllUsesWith(alloca)
		if heapalloc != bitcast {
			bitcast.EraseFromParentAsInstruction()
		}
		heapalloc.EraseFromParentAsInstruction()

		// Add lifetime markers so that LLVM can reuse the stack slot.
		addLifetimeMarkers(mod, builder, alloca, store, size)
	}
}

// addLifetimeMarkers adds llvm.lifetime.start and llvm.lifetime.end to a
// promoted allocation when all uses stay in the block of the allocation.
func addLifetimeMarkers(mod llvm.Module, builder llvm.Builder, alloca, firstUse llvm.Value, size uint64) {
	block := firstUse.InstructionParent()

	// Collect each instruction that uses the pointer or a derived pointer.
	// One call can use several of these pointers. Examine each argument.
	users := map[llvm.Value]struct{}{}
	visited := map[llvm.Value]struct{}{alloca: {}}
	worklist := []llvm.Value{alloca}
	push := func(derived llvm.Value) {
		if _, ok := visited[derived]; !ok {
			visited[derived] = struct{}{}
			worklist = append(worklist, derived)
		}
	}
	for len(worklist) != 0 {
		value := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		for _, use := range getUses(value) {
			if use.InstructionParent() != block {
				// A use in another block. Add no markers.
				return
			}
			users[use] = struct{}{}
			switch use.InstructionOpcode() {
			case llvm.GetElementPtr, llvm.BitCast:
				// The result is a new pointer to the allocation.
				push(use)
			case llvm.Load, llvm.ICmp:
				// The pointer does not leave through these.
			case llvm.Store:
				if use.Operand(0) == value {
					// A stored pointer escapes. Promotion prevents this, but
					// add no markers to be safe.
					return
				}
			case llvm.Call:
				// The callee cannot capture the pointer but can return it.
				// Follow a pointer result. Add no markers for an aggregate.
				if callMayReturnValue(use, value) {
					if use.Type().TypeKind() == llvm.PointerTypeKind {
						push(use)
					} else {
						return
					}
				}
			default:
				// An unknown use. Add no markers.
				return
			}
		}
	}

	// Find the last use in the block.
	last := block.LastInstruction()
	for !last.IsNil() {
		if _, ok := users[last]; ok {
			break
		}
		last = llvm.PrevInstruction(last)
	}
	if last.IsNil() {
		return // not reached, the zeroing store is always a use
	}
	after := llvm.NextInstruction(last)
	if after.IsNil() {
		// The last use is the terminator. The lifetime cannot end in
		// this block.
		return
	}

	sizeValue := llvm.ConstInt(mod.Context().Int64Type(), size, false)
	builder.SetInsertPointBefore(firstUse)
	llvmutil.EmitLifetimeStart(builder, mod, alloca, sizeValue)
	builder.SetInsertPointBefore(after)
	llvmutil.EmitLifetimeEnd(builder, mod, alloca, sizeValue)
}

// callMayReturnValue reports whether the call can return value, directly or
// inside an aggregate. It uses the same rules as callValueEscapesAt.
func callMayReturnValue(call, value llvm.Value) bool {
	called := call.CalledValue()
	if called.IsAFunction().IsNil() {
		// An unknown callee. Assume that it returns the pointer.
		return true
	}
	kindReturned := llvm.AttributeKindID("returned")
	for i := 0; i < called.ParamsCount(); i++ {
		if call.Operand(i) != value {
			continue
		}
		if !called.GetEnumAttributeAtIndex(i+1, kindReturned).IsNil() {
			return true
		}
		if called.IsDeclaration() {
			// A declaration can only return a nocapture argument when the
			// parameter has the returned attribute.
			continue
		}
		result := valueEscapesAtImpl(called.Param(i), true, nil)
		if !result.escapeAt.IsNil() || result.returned {
			return true
		}
	}
	return false
}

// FormatAllocReason renders the heap allocation in a human-readable format.
func FormatAllocReason(pos token.Position, reason string) string {
	return fmt.Sprintf("%s: object allocated on the heap: %s", pos.String(), reason)
}

// FormatAllocCover renders the heap allocation in the go coverage tool format.
func FormatAllocCover(pos token.Position) string {
	if pos.Filename == "" || pos.Line <= 0 {
		return "" // no position info; a blank line would corrupt the profile
	}
	// Highlight the whole line: column 1 to one past the last byte (the end
	// column is exclusive, so add 1 to the line length).
	endCol := max(lineLengthAt(pos.Filename, pos.Line), 1) + 1
	return fmt.Sprintf("%s:%d.1,%d.%d 1 0", pos.Filename, pos.Line, pos.Line, endCol)
}

// valueEscapesAt returns the instruction where the given value may escape and a
// nil llvm.Value if it definitely doesn't. The value must be an instruction.
func valueEscapesAt(value llvm.Value) llvm.Value {
	return valueEscapesAtImpl(value, false, nil).escapeAt
}

type escapeResult struct {
	escapeAt llvm.Value

	// returned is separate from escapeAt because values can flow to a return
	// through aggregate operations. LLVM can mark a scalar returned parameter,
	// but a slice data pointer returned inside {ptr, len, cap} is only visible
	// after walking insertvalue/ret uses in the callee.
	returned bool
}

func (r *escapeResult) merge(other escapeResult) bool {
	if !other.escapeAt.IsNil() {
		r.escapeAt = other.escapeAt
		return false
	}
	r.returned = r.returned || other.returned
	return true
}

func valueEscapesAtImpl(value llvm.Value, allowReturn bool, visiting map[llvm.Value]struct{}) escapeResult {
	if visiting == nil {
		visiting = make(map[llvm.Value]struct{})
	}
	if _, ok := visiting[value]; ok {
		// Recursive call graph while following returned parameters. Treat this
		// as escaping to keep the analysis conservative and bounded.
		return escapeResult{escapeAt: value}
	}
	visiting[value] = struct{}{}
	defer delete(visiting, value)

	var result escapeResult
	uses := getUses(value)
	for _, use := range uses {
		if use.IsAInstruction().IsNil() {
			panic("expected instruction use")
		}
		switch use.InstructionOpcode() {
		case llvm.GetElementPtr:
			if !result.merge(valueEscapesAtImpl(use, allowReturn, visiting)) {
				return result
			}
		case llvm.BitCast:
			// A bitcast escapes if the casted-to value escapes.
			if !result.merge(valueEscapesAtImpl(use, allowReturn, visiting)) {
				return result
			}
		case llvm.InsertValue:
			if !result.merge(valueEscapesAtImpl(use, allowReturn, visiting)) {
				return result
			}
		case llvm.ExtractValue:
			if typeHasPointers(use.Type()) {
				if !result.merge(valueEscapesAtImpl(use, allowReturn, visiting)) {
					return result
				}
			}
		case llvm.Load:
			// Load does not escape.
		case llvm.Store:
			// Store only escapes when the value is stored to, not when the
			// value is stored into another value.
			if use.Operand(0) == value {
				return escapeResult{escapeAt: use}
			}
		case llvm.Call:
			if !result.merge(callValueEscapesAt(use, value, allowReturn, visiting)) {
				return result
			}
		case llvm.ICmp:
			// Comparing pointers don't let the pointer escape.
			// This is often a compiler-inserted nil check.
		case llvm.Ret:
			if !allowReturn || use.Operand(0) != value {
				return escapeResult{escapeAt: use}
			}
			result.returned = true
		default:
			// Unknown instruction, might escape.
			return escapeResult{escapeAt: use}
		}
	}

	// Checked all uses, and none let the pointer value escape.
	return result
}

// callValueEscapesAt returns whether value escapes through this call. It also
// handles calls that return value unchanged, as long as the called function does
// not otherwise capture the parameter and the returned alias does not escape.
func callValueEscapesAt(call, value llvm.Value, allowReturn bool, visiting map[llvm.Value]struct{}) escapeResult {
	called := call.CalledValue()
	if called.IsAFunction().IsNil() {
		return escapeResult{escapeAt: call}
	}
	kindNoCapture := llvm.AttributeKindID(llvmutil.NoCaptureAttrName())
	kindReturned := llvm.AttributeKindID("returned")
	matched := false
	var result escapeResult
	for i := 0; i < called.ParamsCount(); i++ {
		if call.Operand(i) != value {
			continue
		}
		matched = true
		index := i + 1 // param attributes start at 1
		nocapture := llvmutil.IsNoCapture(called.GetEnumAttributeAtIndex(index, kindNoCapture))
		returnedParam := !called.GetEnumAttributeAtIndex(index, kindReturned).IsNil()
		if returnedParam {
			result.returned = true
		}
		if nocapture {
			continue
		}
		if called.IsDeclaration() {
			return escapeResult{escapeAt: call}
		}
		if !result.merge(valueEscapesAtImpl(called.Param(i), true, visiting)) {
			return result
		}
	}
	for i := called.ParamsCount(); i < call.OperandsCount(); i++ {
		if call.Operand(i) == value {
			return escapeResult{escapeAt: call}
		}
	}
	if !matched {
		return escapeResult{}
	}
	if result.returned {
		return valueEscapesAtImpl(call, allowReturn, visiting)
	}
	return escapeResult{}
}

func lineLengthAt(filename string, lineNumber int) int {
	f, err := os.Open(filename)
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	line := 1
	for scanner.Scan() {
		if line == lineNumber {
			return len(scanner.Text())
		}
		line++
	}
	return 0
}
