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
	"strings"

	"github.com/tinygo-org/tinygo/internal/llvmutil"
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
		logAllocs := printAllocs != nil
		report := func(reason string) {
			reportHeapAlloc(heapalloc, reason, printAllocs, logger)
		}
		if heapalloc.Operand(0).IsAConstantInt().IsNil() {
			// Do not allocate variable length arrays on the stack.
			if logAllocs {
				report("size is not constant")
			}
			continue
		}

		size := heapalloc.Operand(0).ZExtValue()
		if size > maxStackAlloc {
			// The maximum size for a stack allocation.
			if logAllocs {
				report(fmt.Sprintf("object size %d exceeds maximum stack allocation size %d", size, maxStackAlloc))
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

		if at := valueEscapesAt(heapalloc); !at.IsNil() {
			if logAllocs {
				atPos := getPosition(at)
				msg := "escapes at unknown line"
				if atPos.Line != 0 {
					msg = fmt.Sprintf("escapes at line %d", atPos.Line)
				}
				report(msg)
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
		fn := heapalloc.InstructionParent().Parent()
		builder.SetInsertPointBefore(fn.EntryBasicBlock().FirstInstruction())
		allocaType := llvm.ArrayType(mod.Context().Int8Type(), int(size))
		alloca := builder.CreateAlloca(allocaType, "stackalloc")
		alloca.SetAlignment(alignment)

		// Zero the allocation inside the block where the value was originally allocated.
		zero := llvm.ConstNull(alloca.AllocatedType())
		builder.SetInsertPointBefore(heapalloc)
		store := builder.CreateStore(zero, alloca)
		store.SetAlignment(alignment)

		// Replace heap alloc with stack alloc.
		heapalloc.ReplaceAllUsesWith(alloca)
		heapalloc.EraseFromParentAsInstruction()
	}
}

// reportHeapAlloc logs a heap allocation if its function matches the regexp.
// An allocation inside a runtime function is also logged at the calls to that
// function from other packages, so for example append is reported at its line.
func reportHeapAlloc(heapalloc llvm.Value, reason string, re *regexp.Regexp, logger func(token.Position, string)) {
	fn := heapalloc.InstructionParent().Parent()
	if re.MatchString(fn.Name()) {
		logger(getPosition(heapalloc), reason)
	}
	if !strings.HasPrefix(fn.Name(), "runtime.") {
		return
	}
	visited := map[llvm.Value]struct{}{fn: {}}
	work := []llvm.Value{fn}
	for len(work) > 0 {
		callee := work[len(work)-1]
		work = work[:len(work)-1]
		for call := range llvmutil.Uses(callee) {
			if call.IsACallInst().IsNil() || call.CalledValue() != callee {
				continue
			}
			caller := call.InstructionParent().Parent()
			if strings.HasPrefix(caller.Name(), "runtime.") {
				if _, ok := visited[caller]; !ok {
					visited[caller] = struct{}{}
					work = append(work, caller)
				}
			} else if re.MatchString(caller.Name()) {
				logger(getPosition(call), fmt.Sprintf("%s, allocated in %s", reason, fn.Name()))
			}
		}
	}
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
	analysis := escapeAnalysis{
		visiting: make(map[llvm.Value]struct{}),
		results:  make(map[escapeKey]escapeResult),
		originFn: value.InstructionParent().Parent(),
		// An allocation that runs more than once per call reuses its stack
		// slot, so older objects must not stay reachable through a PHI.
		originInCycle: value.InstructionParent().InCycle(),
	}
	return analysis.valueEscapesAt(value, false).escapeAt
}

type escapeKey struct {
	value       llvm.Value
	allowReturn bool
}

type escapeAnalysis struct {
	visiting map[llvm.Value]struct{}
	results  map[escapeKey]escapeResult

	originFn      llvm.Value
	originInCycle bool

	// cuts counts the PHIs that were reached again while they were still
	// being checked.
	cuts int
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

func (a *escapeAnalysis) valueEscapesAt(value llvm.Value, allowReturn bool) (result escapeResult) {
	if _, ok := a.visiting[value]; ok {
		if !value.IsAPHINode().IsNil() {
			// A value flowing around a loop back into the same PHI. The
			// first visit already checks every use of the PHI.
			a.cuts++
			return escapeResult{}
		}
		// Recursive call graph while following returned parameters. Treat this
		// as escaping to keep the analysis conservative and bounded.
		return escapeResult{escapeAt: value}
	}
	key := escapeKey{value, allowReturn}
	// Only results that do not escape are ever reused, since any escape ends
	// the analysis.
	if result, ok := a.results[key]; ok {
		return result
	}
	a.visiting[value] = struct{}{}
	cuts := a.cuts
	defer func() {
		delete(a.visiting, value)
		// A result that stopped at a PHI that was still being checked misses
		// what that PHI reaches, such as a return. It is only right for this
		// path, so don't reuse it.
		if a.cuts == cuts {
			a.results[key] = result
		}
	}()

	for use := range llvmutil.Uses(value) {
		if use.IsAInstruction().IsNil() {
			panic("expected instruction use")
		}
		switch use.InstructionOpcode() {
		case llvm.GetElementPtr:
			if !result.merge(a.valueEscapesAt(use, allowReturn)) {
				return result
			}
		case llvm.PHI:
			// A PHI in another function runs once per call of the allocating
			// function, so the object cannot be overwritten through it.
			if a.originInCycle && use.InstructionParent().Parent() == a.originFn &&
				use.InstructionParent().InCycle() {
				return escapeResult{escapeAt: use}
			}
			fallthrough
		case llvm.BitCast, llvm.Select:
			if !result.merge(a.valueEscapesAt(use, allowReturn)) {
				return result
			}
		case llvm.InsertValue:
			if !result.merge(a.valueEscapesAt(use, allowReturn)) {
				return result
			}
		case llvm.ExtractValue:
			if typeHasPointers(use.Type()) {
				if !result.merge(a.valueEscapesAt(use, allowReturn)) {
					return result
				}
			}
		case llvm.Load:
			// Load does not escape.
		case llvm.Store:
			// Store only escapes when the value is stored to, not when the
			// value is stored into another value.
			if use.Operand(0) == value {
				// Storing into a local stack slot is fine as long as the
				// pointers loaded from it don't escape either. An allocation
				// that runs in a loop reuses its stack slot, so the pointer
				// must not stay reachable in memory across iterations.
				if slot := use.Operand(1).UnderlyingObject(); !a.originInCycle && !slot.IsAAllocaInst().IsNil() &&
					slot.InstructionParent().Parent() == use.InstructionParent().Parent() {
					if !result.merge(a.stackSlotEscapesAt(slot, allowReturn)) {
						return result
					}
					break
				}
				return escapeResult{escapeAt: use}
			}
		case llvm.Call:
			if !result.merge(a.callValueEscapesAt(use, value, allowReturn)) {
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

// stackSlotEscapesAt reports whether a pointer stored in the given alloca can
// escape. The alloca itself must stay local and every pointer loaded from it
// must not escape.
func (a *escapeAnalysis) stackSlotEscapesAt(slot llvm.Value, allowReturn bool) escapeResult {
	if _, ok := a.visiting[slot]; ok {
		return escapeResult{escapeAt: slot}
	}
	a.visiting[slot] = struct{}{}
	defer delete(a.visiting, slot)

	var check func(ptr llvm.Value) escapeResult
	check = func(ptr llvm.Value) escapeResult {
		var res escapeResult
		for _, use := range getUses(ptr) {
			switch use.InstructionOpcode() {
			case llvm.GetElementPtr:
				if !res.merge(check(use)) {
					return res
				}
			case llvm.Store:
				if use.Operand(0) == ptr {
					// The address of the slot is stored somewhere.
					return escapeResult{escapeAt: use}
				}
			case llvm.Load:
				if typeHasPointers(use.Type()) {
					if !res.merge(a.valueEscapesAt(use, allowReturn)) {
						return res
					}
				}
			case llvm.Call:
				// Lifetime markers don't let the slot escape.
				callee := use.CalledValue()
				if callee.IsAFunction().IsNil() || !strings.HasPrefix(callee.Name(), "llvm.lifetime.") {
					return escapeResult{escapeAt: use}
				}
			default:
				return escapeResult{escapeAt: use}
			}
		}
		return res
	}
	return check(slot)
}

// callValueEscapesAt returns whether value escapes through this call. It also
// handles calls that return value unchanged, as long as the called function does
// not otherwise capture the parameter and the returned alias does not escape.
func (a *escapeAnalysis) callValueEscapesAt(call, value llvm.Value, allowReturn bool) escapeResult {
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
		nocapture := llvmutil.CapturesAtMostNullness(called.GetEnumAttributeAtIndex(index, kindNoCapture))
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
		if _, recursive := a.visiting[called.Param(i)]; recursive && !typeHasPointers(call.Type()) {
			// valueEscapesAt is already checking this parameter's other uses.
			continue
		}
		if !result.merge(a.valueEscapesAt(called.Param(i), true)) {
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
		return a.valueEscapesAt(call, allowReturn)
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
