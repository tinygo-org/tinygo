package compiler

// This file implements capturing variables by value. The ssa package keeps a
// variable that is captured by a closure in a heap allocated cell, even when
// it is never changed after it is initialized. Every use then loads from that
// cell, which LLVM often cannot see through. If the variable is stored exactly
// once, before any closure is created, the cell can be left out and the value
// is passed to the closure directly.

import (
	"go/token"
	"go/types"

	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// capturedStore returns the single store that initializes alloc if the
// variable can be captured by value, or nil otherwise.
func (b *builder) capturedStore(alloc *ssa.Alloc) *ssa.Store {
	if store, ok := b.captureStores[alloc]; ok {
		return store
	}
	store := b.findCapturedStore(alloc)
	if b.captureStores == nil {
		b.captureStores = make(map[*ssa.Alloc]*ssa.Store)
	}
	b.captureStores[alloc] = store
	return store
}

func (b *builder) findCapturedStore(alloc *ssa.Alloc) *ssa.Store {
	if !alloc.Heap {
		return nil // not heap allocated
	}
	elemType := alloc.Type().Underlying().(*types.Pointer).Elem()
	if b.isAggregateValue(elemType) || b.targetData.TypeAllocSize(b.getLLVMType(elemType)) == 0 {
		return nil // aggregate or zero-size type
	}
	refs := alloc.Referrers()
	if refs == nil {
		return nil // no referrers
	}

	// Find the one store and check that nothing else uses the address.
	var store *ssa.Store
	for _, ref := range *refs {
		switch ref := ref.(type) {
		case *ssa.Store:
			if ref.Addr != ssa.Value(alloc) || ref.Val == ssa.Value(alloc) || store != nil {
				return nil // stored more than once or address stored
			}
			store = ref
		case *ssa.UnOp:
			if ref.Op != token.MUL {
				return nil // unexpected unary operation
			}
		case *ssa.MakeClosure, *ssa.DebugRef:
		default:
			return nil // address used by another instruction
		}
	}
	// The store must run once for every execution of the allocation.
	if store == nil || store.Block() != alloc.Block() || instrIndex(store) < instrIndex(alloc) {
		return nil // no single initializing store after the allocation
	}

	closures := 0
	for _, ref := range *refs {
		switch ref := ref.(type) {
		case *ssa.UnOp:
			if !dominates(store, ref) {
				return nil // load not dominated by the store
			}
		case *ssa.MakeClosure:
			if !dominates(store, ref) {
				return nil // closure not dominated by the store
			}
			fn := ref.Fn.(*ssa.Function)
			if b.closureSite(fn) != ref {
				// The closure function must know how its free variables are
				// passed, so it must be created in exactly one place.
				return nil // closure is not a function literal created in one place
			}
			if len(ref.Bindings) == 1 && b.packedAsInt(elemType) {
				return nil // single integer packed into the context, needs LLVM 22
			}
			for i, binding := range ref.Bindings {
				if binding != ssa.Value(alloc) {
					continue
				}
				closures++
				if !onlyLoaded(fn.FreeVars[i]) {
					return nil // closure assigns or takes the address of it
				}
			}
		}
	}
	if closures == 0 {
		return nil // not bound by a closure
	}
	return store
}

// packedAsInt returns whether a closure whose only bound value has type typ
// gets that value packed into the context pointer with inttoptr. The closure
// then reads it back with ptrtoint, which LLVM counts as capturing the
// context, so the escape analysis would send the context of every indirect
// call of that type to the heap. Since LLVM 22 it is read with ptrtoaddr,
// which does not capture (see unpackInt in llvm.go).
func (b *builder) packedAsInt(typ types.Type) bool {
	if llvmutil.HasPtrToAddr() {
		return false
	}
	llvmType := b.getLLVMType(typ)
	return llvmType.TypeKind() == llvm.IntegerTypeKind &&
		b.targetData.TypeAllocSize(llvmType) <= uint64(b.ptrBits)/8
}

// onlyLoaded returns whether the pointer v is only ever dereferenced.
func onlyLoaded(v ssa.Value) bool {
	refs := v.Referrers()
	if refs == nil {
		return false
	}
	for _, ref := range *refs {
		switch ref := ref.(type) {
		case *ssa.UnOp:
			if ref.Op != token.MUL {
				return false
			}
		case *ssa.DebugRef:
		default:
			return false
		}
	}
	return true
}

// instrIndex returns the index of instr in its basic block.
func instrIndex(instr ssa.Instruction) int {
	for i, other := range instr.Block().Instrs {
		if other == instr {
			return i
		}
	}
	return -1
}

// dominates returns whether a always runs before b.
func dominates(a, b ssa.Instruction) bool {
	if a.Block() == b.Block() {
		return instrIndex(a) < instrIndex(b)
	}
	return a.Block().Dominates(b.Block())
}

// closureSite returns the only MakeClosure that creates the closure function
// fn, or nil if fn is not a function literal created in exactly one place.
func (b *builder) closureSite(fn *ssa.Function) *ssa.MakeClosure {
	if site, ok := b.closureSites[fn]; ok {
		return site
	}
	var site *ssa.MakeClosure
	if parent := fn.Parent(); parent != nil {
	search:
		for _, block := range parent.Blocks {
			for _, instr := range block.Instrs {
				if mc, ok := instr.(*ssa.MakeClosure); ok && mc.Fn == ssa.Value(fn) {
					if site != nil {
						site = nil
						break search
					}
					site = mc
				}
			}
		}
	}
	if b.closureSites == nil {
		b.closureSites = make(map[*ssa.Function]*ssa.MakeClosure)
	}
	b.closureSites[fn] = site
	return site
}

// bindingByValue returns whether the closure created by mc gets binding i as
// a value instead of a pointer to a heap cell. Both parseMakeClosure and the
// closure function itself use it, so they always agree on the layout of the
// context.
func (b *builder) bindingByValue(mc *ssa.MakeClosure, i int) bool {
	alloc, ok := mc.Bindings[i].(*ssa.Alloc)
	// capturedStore only succeeds if mc is the only site of its function.
	return ok && b.capturedStore(alloc) != nil
}

// freeVarByValue returns whether the closure free variable fv holds the
// value itself instead of a pointer to it.
func (b *builder) freeVarByValue(fv *ssa.FreeVar) bool {
	fn := fv.Parent()
	mc := b.closureSite(fn)
	if mc == nil {
		return false
	}
	for i, other := range fn.FreeVars {
		if other == fv {
			return b.bindingByValue(mc, i)
		}
	}
	return false
}
