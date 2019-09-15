package compiler

import (
	"errors"
	"fmt"

	"tinygo.org/x/go-llvm"
)

func (c *Compiler) checkType(t llvm.Type, checked map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) {
	if t.IsNil() {
		panic(t)
	}

	// prevent infinite recursion for self-referential types
	if _, ok := checked[t]; ok {
		return
	}
	checked[t] = struct{}{}

	// check for any context mismatches
	switch {
	case t.Context() == c.ctx:
		// this is correct
	case t.Context() == llvm.GlobalContext():
		// somewhere we accidentally used the global context instead of a real context
		panic(fmt.Errorf("type %q uses global context", t.String()))
	default:
		// we used some other context by accident
		panic(fmt.Errorf("type %q uses context %v instead of the main context %v", t.Context(), c.ctx))
	}

	// if this is a composite type, check the components of the type
	switch t.TypeKind() {
	case llvm.VoidTypeKind, llvm.LabelTypeKind, llvm.TokenTypeKind, llvm.MetadataTypeKind:
		// there should only be one of any of these
		if s, ok := specials[t.TypeKind()]; !ok {
			specials[t.TypeKind()] = t
		} else if s != t {
			panic(fmt.Errorf("duplicate special type %q: %v and %v", t.TypeKind().String(), t, s))
		}
	case llvm.FloatTypeKind, llvm.DoubleTypeKind, llvm.X86_FP80TypeKind, llvm.FP128TypeKind, llvm.PPC_FP128TypeKind:
		// floating point numbers are primitives - nothing to recurse
	case llvm.IntegerTypeKind:
		// integers are primitives - nothing to recurse
	case llvm.FunctionTypeKind:
		// check arguments and return(s)
		for _, v := range t.ParamTypes() {
			c.checkType(v, checked, specials)
		}
		c.checkType(t.ReturnType(), checked, specials)
	case llvm.StructTypeKind:
		// check all elements
		for _, v := range t.StructElementTypes() {
			c.checkType(v, checked, specials)
		}
	case llvm.ArrayTypeKind:
		// check element type
		c.checkType(t.ElementType(), checked, specials)
	case llvm.PointerTypeKind:
		// check underlying type
		c.checkType(t.ElementType(), checked, specials)
	case llvm.VectorTypeKind:
		// check element type
		c.checkType(t.ElementType(), checked, specials)
	}
}

func (c *Compiler) checkValue(v llvm.Value, checked map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) {
	// check type
	c.checkType(v.Type(), checked, specials)
}

func (c *Compiler) checkInstruction(inst llvm.Value, checked map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) {
	// check value properties
	c.checkValue(inst, checked, specials)
}

func (c *Compiler) checkBasicBlock(bb llvm.BasicBlock, checked map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) {
	// check basic block value and type
	c.checkValue(bb.AsValue(), checked, specials)

	// check instructions
	for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
		c.checkInstruction(inst, checked, specials)
	}
}

func (c *Compiler) checkFunction(fn llvm.Value, checked map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) {
	// check function value and type
	c.checkValue(fn, checked, specials)

	// check basic blocks
	for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
		c.checkBasicBlock(bb, checked, specials)
	}
}

func (c *Compiler) check() {
	// check for any context mismatches
	switch {
	case c.mod.Context() == c.ctx:
		// this is correct
	case c.mod.Context() == llvm.GlobalContext():
		// somewhere we accidentally used the global context instead of a real context
		panic(errors.New("module uses global context"))
	default:
		// we used some other context by accident
		panic(fmt.Errorf("module uses context %v instead of the main context %v", c.mod.Context(), c.ctx))
	}

	// base of type-check stack
	checked := map[llvm.Type]struct{}{}
	specials := map[llvm.TypeKind]llvm.Type{}
	for fn := c.mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		c.checkFunction(fn, checked, specials)
	}
	for g := c.mod.FirstGlobal(); !g.IsNil(); g = llvm.NextGlobal(g) {
		c.checkValue(g, checked, specials)
	}
}
