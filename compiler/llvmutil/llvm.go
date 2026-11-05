// Package llvmutil contains utility functions used across multiple compiler
// packages. For example, they may be used by both the compiler package and
// transformation packages.
//
// Normally, utility packages are avoided. However, in this case, the utility
// functions are non-trivial and hard to get right. Copying them to multiple
// places would be a big risk if only one of them is updated.
package llvmutil

import (
	"encoding/binary"
	"math"
	"strings"

	"tinygo.org/x/go-llvm"
)

// CreateEntryBlockAlloca creates a new alloca in the entry block, even though
// the IR builder is located elsewhere. It assumes that the insert point is
// at the end of the current block.
func CreateEntryBlockAlloca(builder llvm.Builder, t llvm.Type, name string) llvm.Value {
	currentBlock := builder.GetInsertBlock()
	entryBlock := currentBlock.Parent().EntryBasicBlock()
	if entryBlock.FirstInstruction().IsNil() {
		builder.SetInsertPointAtEnd(entryBlock)
	} else {
		builder.SetInsertPointBefore(entryBlock.FirstInstruction())
	}
	alloca := builder.CreateAlloca(t, name)
	builder.SetInsertPointAtEnd(currentBlock)
	return alloca
}

// CreateTemporaryAlloca creates a new alloca in the entry block and adds
// lifetime start information in the IR signalling that the alloca won't be used
// before this point.
//
// This is useful for creating temporary allocas for intrinsics. Don't forget to
// end the lifetime using emitLifetimeEnd after you're done with it.
func CreateTemporaryAlloca(builder llvm.Builder, mod llvm.Module, t llvm.Type, name string) (alloca, size llvm.Value) {
	ctx := t.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	alloca = CreateEntryBlockAlloca(builder, t, name)
	size = llvm.ConstInt(ctx.Int64Type(), targetData.TypeAllocSize(t), false)
	fnType, fn := getLifetimeStartFunc(mod)
	builder.CreateCall(fnType, fn, lifetimeCallArgs(size, alloca), "")
	return
}

// EmitLifetimeEnd signals the end of an (alloca) lifetime by calling the
// llvm.lifetime.end intrinsic. It is commonly used together with
// createTemporaryAlloca.
func EmitLifetimeEnd(builder llvm.Builder, mod llvm.Module, ptr, size llvm.Value) {
	fnType, fn := getLifetimeEndFunc(mod)
	builder.CreateCall(fnType, fn, lifetimeCallArgs(size, ptr), "")
}

// lifetimeCallArgs returns the arguments to pass to a call of the
// llvm.lifetime.start/end intrinsics. LLVM 22 dropped the (redundant,
// already required to match the alloca size) i64 size argument, so the
// intrinsic now only takes the pointer.
func lifetimeCallArgs(size, ptr llvm.Value) []llvm.Value {
	if llvm.VersionMajor >= 22 {
		return []llvm.Value{ptr}
	}
	return []llvm.Value{size, ptr}
}

// lifetimeFuncType returns the function type of the llvm.lifetime.start/end
// intrinsics, which lost their i64 size parameter in LLVM 22.
func lifetimeFuncType(ctx llvm.Context, ptrType llvm.Type) llvm.Type {
	if llvm.VersionMajor >= 22 {
		return llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType}, false)
	}
	return llvm.FunctionType(ctx.VoidType(), []llvm.Type{ctx.Int64Type(), ptrType}, false)
}

// getLifetimeStartFunc returns the llvm.lifetime.start intrinsic and creates it
// first if it doesn't exist yet.
func getLifetimeStartFunc(mod llvm.Module) (llvm.Type, llvm.Value) {
	fnName := "llvm.lifetime.start.p0"
	fn := mod.NamedFunction(fnName)
	ctx := mod.Context()
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	fnType := lifetimeFuncType(ctx, ptrType)
	if fn.IsNil() {
		fn = llvm.AddFunction(mod, fnName, fnType)
	}
	return fnType, fn
}

// getLifetimeEndFunc returns the llvm.lifetime.end intrinsic and creates it
// first if it doesn't exist yet.
func getLifetimeEndFunc(mod llvm.Module) (llvm.Type, llvm.Value) {
	fnName := "llvm.lifetime.end.p0"
	fn := mod.NamedFunction(fnName)
	ctx := mod.Context()
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	fnType := lifetimeFuncType(ctx, ptrType)
	if fn.IsNil() {
		fn = llvm.AddFunction(mod, fnName, fnType)
	}
	return fnType, fn
}

// AppendToGlobal appends the given values to a global array like llvm.used. The global might
// not exist yet. The values can be any pointer type, they will be cast to i8*.
func AppendToGlobal(mod llvm.Module, globalName string, values ...llvm.Value) {
	// Read the existing values in the llvm.used array (if it exists).
	var usedValues []llvm.Value
	if used := mod.NamedGlobal(globalName); !used.IsNil() {
		builder := mod.Context().NewBuilder()
		defer builder.Dispose()
		usedInitializer := used.Initializer()
		num := usedInitializer.Type().ArrayLength()
		for i := range num {
			usedValues = append(usedValues, builder.CreateExtractValue(usedInitializer, i, ""))
		}
		used.EraseFromParentAsGlobal()
	}

	// Add the new values.
	ptrType := llvm.PointerType(mod.Context().Int8Type(), 0)
	for _, value := range values {
		// Note: the bitcast is necessary to cast AVR function pointers to
		// address space 0 pointer types.
		usedValues = append(usedValues, llvm.ConstPointerCast(value, ptrType))
	}

	// Create a new array (with the old and new values).
	usedInitializer := llvm.ConstArray(ptrType, usedValues)
	used := llvm.AddGlobal(mod, usedInitializer.Type(), globalName)
	used.SetInitializer(usedInitializer)
	used.SetLinkage(llvm.AppendingLinkage)
}

// RemoveFromGlobal removes values matching the predicate from an appending
// global such as llvm.used.
func RemoveFromGlobal(mod llvm.Module, globalName string, remove func(llvm.Value) bool) {
	global := mod.NamedGlobal(globalName)
	if global.IsNil() {
		return
	}

	builder := mod.Context().NewBuilder()
	defer builder.Dispose()
	initializer := global.Initializer()
	var kept []llvm.Value
	for i := 0; i < initializer.Type().ArrayLength(); i++ {
		value := builder.CreateExtractValue(initializer, i, "")
		base := value
		for !base.IsAConstantExpr().IsNil() && base.OperandsCount() == 1 {
			base = base.Operand(0)
		}
		if !remove(base) {
			kept = append(kept, value)
		}
	}
	global.EraseFromParentAsGlobal()
	if len(kept) != 0 {
		AppendToGlobal(mod, globalName, kept...)
	}
}

// RemoveGlobalReferences removes one occurrence from targetGlobal for each
// value listed in referenceGlobal, then removes referenceGlobal itself.
func RemoveGlobalReferences(mod llvm.Module, targetGlobal, referenceGlobal string) {
	references := mod.NamedGlobal(referenceGlobal)
	if references.IsNil() {
		return
	}

	builder := mod.Context().NewBuilder()
	defer builder.Dispose()
	initializer := references.Initializer()
	values := make(map[llvm.Value]int, initializer.Type().ArrayLength())
	for i := 0; i < initializer.Type().ArrayLength(); i++ {
		value := builder.CreateExtractValue(initializer, i, "")
		for !value.IsAConstantExpr().IsNil() && value.OperandsCount() == 1 {
			value = value.Operand(0)
		}
		values[value]++
	}
	references.EraseFromParentAsGlobal()
	RemoveFromGlobal(mod, targetGlobal, func(value llvm.Value) bool {
		if values[value] == 0 {
			return false
		}
		values[value]--
		return true
	})
}

// NoCaptureAttrName returns the name of the LLVM attribute kind that marks a
// pointer parameter as guaranteed not to escape by capture.
//
// LLVM 21 removed the old boolean 'nocapture' enum attribute in favor of the
// more expressive 'captures' int attribute, where 'captures(none)' (encoded
// as the value 0) is the equivalent of the old 'nocapture'. LLVM 20 supports
// both, but its own optimizer still emits 'nocapture', so the cutoff for
// reading/writing the new name is LLVM 21.
func NoCaptureAttrName() string {
	if llvm.VersionMajor >= 21 {
		return "captures"
	}
	return "nocapture"
}

// IsNoCapture reports whether attr (looked up using the kind returned by
// NoCaptureAttrName) indicates that the pointer it is attached to does not
// escape by capture. It returns false for a nil attribute.
func IsNoCapture(attr llvm.Attribute) bool {
	if attr.IsNil() {
		return false
	}
	if llvm.VersionMajor >= 21 {
		// captures(none) is encoded as the value 0; any other value permits
		// some form of capture.
		return attr.GetEnumValue() == 0
	}
	return true
}

// UnsignedRangeAttribute returns a range attribute that limits an integer of
// the given width to [0, max]. It returns a nil attribute before LLVM 19 or when max+1 does not fit.
func UnsignedRangeAttribute(ctx llvm.Context, bits uint, max uint64) llvm.Attribute {
	if max >= math.MaxUint64>>(64-bits) {
		return llvm.Attribute{}
	}
	return ctx.CreateConstantRangeAttribute(llvm.AttributeKindID("range"), bits, []uint64{0}, []uint64{max + 1})
}

// ByteOrder returns the byte order for the given target triple. Most targets are little
// endian, but for example MIPS can be big-endian.
func ByteOrder(target string) binary.ByteOrder {
	if strings.HasPrefix(target, "mips-") {
		return binary.BigEndian
	} else {
		return binary.LittleEndian
	}
}

// WritesArgsOnlyAttrName is a string function attribute set by the compiler on
// runtime functions that write memory only through pointer parameters, and
// never through ones marked readonly. LLVM ignores it. The transform passes
// use it where LLVM cannot infer this, for example across indirect calls.
const WritesArgsOnlyAttrName = "tinygo-writes-args-only"
