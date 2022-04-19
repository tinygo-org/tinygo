package transform

// This file implements an escape analysis pass. It looks for calls to
// runtime.alloc and replaces these calls with a stack allocation if the
// allocated value does not escape. It uses the LLVM nocapture flag for
// interprocedural escape analysis.

import (
	"fmt"
	"go/token"
	"regexp"

	"tinygo.org/x/go-llvm"
)

// maxStackAlloc is the maximum size of an object that will be allocated on the
// stack. Bigger objects have increased risk of stack overflows and thus will
// always be heap allocated.
//
// TODO: tune this, this is just a random value.
// This value is also used in the compiler when translating ssa.Alloc nodes.
const maxStackAlloc = 256

// OptimizeAllocs tries to replace heap allocations with stack allocations
// whenever possible. It relies on the LLVM 'nocapture' flag for interprocedural
// escape analysis, and within a function looks whether an allocation can escape
// to the heap.
// If printAllocs is non-nil, it indicates the regexp of functions for which a
// heap allocation explanation should be printed (why the object can't be stack
// allocated).
func OptimizeAllocs(mod llvm.Module, printAllocs *regexp.Regexp, logger func(token.Position, string)) {
	allocator := mod.NamedFunction("runtime.alloc")
	if allocator.IsNil() {
		// nothing to optimize
		return
	}

	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	i8ptrType := llvm.PointerType(mod.Context().Int8Type(), 0)
	builder := mod.Context().NewBuilder()
	defer builder.Dispose()

	for _, heapalloc := range getUses(allocator) {
		logAllocs := printAllocs != nil && printAllocs.MatchString(heapalloc.InstructionParent().Parent().Name())
		if heapalloc.Operand(0).IsAConstantInt().IsNil() {
			// Do not allocate variable length arrays on the stack.
			if logAllocs {
				logAlloc(logger, heapalloc, "size is not constant")
			}
			continue
		}

		size := heapalloc.Operand(0).ZExtValue()
		if size > maxStackAlloc {
			// The maximum size for a stack allocation.
			if logAllocs {
				logAlloc(logger, heapalloc, fmt.Sprintf("object size %d exceeds maximum stack allocation size %d", size, maxStackAlloc))
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
				logAlloc(logger, heapalloc, msg)
			}
			continue
		}
		// The pointer value does not escape.

		// Determine the appropriate alignment of the alloca. The size of the
		// allocation gives us a hint what the alignment should be.
		var alignment int
		if size%2 != 0 {
			alignment = 1
		} else if size%4 != 0 {
			alignment = 2
		} else if size%8 != 0 {
			alignment = 4
		} else {
			alignment = 8
		}
		if pointerAlignment := targetData.ABITypeAlignment(i8ptrType); pointerAlignment < alignment {
			// Use min(alignment, alignof(void*)) as the alignment.
			alignment = pointerAlignment
		}

		// Insert alloca in the entry block. Do it here so that mem2reg can
		// promote it to a SSA value.
		fn := bitcast.InstructionParent().Parent()
		builder.SetInsertPointBefore(fn.EntryBasicBlock().FirstInstruction())
		allocaType := llvm.ArrayType(mod.Context().Int8Type(), int(size))
		alloca := builder.CreateAlloca(allocaType, "stackalloc.alloca")
		alloca.SetAlignment(alignment)

		// Zero the allocation inside the block where the value was originally allocated.
		zero := llvm.ConstNull(alloca.Type().ElementType())
		builder.SetInsertPointBefore(bitcast)
		store := builder.CreateStore(zero, alloca)
		store.SetAlignment(alignment)

		// Replace heap alloc bitcast with stack alloc bitcast.
		stackalloc := builder.CreateBitCast(alloca, bitcast.Type(), "stackalloc")
		bitcast.ReplaceAllUsesWith(stackalloc)
		if heapalloc != bitcast {
			bitcast.EraseFromParentAsInstruction()
		}
		heapalloc.EraseFromParentAsInstruction()
	}
}

// valueEscapesAt returns the instruction where the given value may escape and a
// nil llvm.Value if it definitely doesn't. The value must be an instruction.
func valueEscapesAt(value llvm.Value) llvm.Value {
	uses := getUses(value)
	for _, use := range uses {
		if use.IsAInstruction().IsNil() {
			panic("expected instruction use")
		}
		switch use.InstructionOpcode() {
		case llvm.GetElementPtr:
			if at := valueEscapesAt(use); !at.IsNil() {
				return at
			}
		case llvm.BitCast:
			// A bitcast escapes if the casted-to value escapes.
			if at := valueEscapesAt(use); !at.IsNil() {
				return at
			}
		case llvm.Load:
			// Load does not escape.
		case llvm.Store:
			// Store only escapes when the value is stored to, not when the
			// value is stored into another value.
			if use.Operand(0) == value {
				return use
			}
		case llvm.Call:
			if !hasFlag(use, value, "nocapture") {
				return use
			}
		case llvm.ICmp:
			// Comparing pointers don't let the pointer escape.
			// This is often a compiler-inserted nil check.
		default:
			// Unknown instruction, might escape.
			return use
		}
	}

	// Checked all uses, and none let the pointer value escape.
	return llvm.Value{}
}

// logAlloc prints a message to stderr explaining why the given object had to be
// allocated on the heap.
func logAlloc(logger func(token.Position, string), allocCall llvm.Value, reason string) {
	logger(getPosition(allocCall), "object allocated on the heap: "+reason)
}
