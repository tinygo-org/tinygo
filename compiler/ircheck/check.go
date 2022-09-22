// Package ircheck implements a checker for LLVM IR, that goes a bit further
// than the regular LLVM IR verifier. Note that it checks different things, so
// this is not a replacement for the LLVM verifier but does catch things that
// the LLVM verifier doesn't catch.
package ircheck

import (
	"errors"
	"fmt"

	"tinygo.org/x/go-llvm"
)

type checker struct {
	ctx llvm.Context
}

func (c *checker) checkType(t llvm.Type, checked map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) error {
	// prevent infinite recursion for self-referential types
	if _, ok := checked[t]; ok {
		return nil
	}
	checked[t] = struct{}{}

	// check for any context mismatches
	switch {
	case t.Context() == c.ctx:
		// this is correct
	case t.Context() == llvm.GlobalContext():
		// somewhere we accidentally used the global context instead of a real context
		return fmt.Errorf("type %q uses global context", t.String())
	default:
		// we used some other context by accident
		return fmt.Errorf("type %q uses context %v instead of the main context %v", t.String(), t.Context(), c.ctx)
	}

	// if this is a composite type, check the components of the type
	switch t.TypeKind() {
	case llvm.VoidTypeKind, llvm.LabelTypeKind, llvm.TokenTypeKind, llvm.MetadataTypeKind:
		// there should only be one of any of these
		if s, ok := specials[t.TypeKind()]; !ok {
			specials[t.TypeKind()] = t
		} else if s != t {
			return fmt.Errorf("duplicate special type %q: %v and %v", t.TypeKind().String(), t, s)
		}
	case llvm.FloatTypeKind, llvm.DoubleTypeKind, llvm.X86_FP80TypeKind, llvm.FP128TypeKind, llvm.PPC_FP128TypeKind:
		// floating point numbers are primitives - nothing to recurse
	case llvm.IntegerTypeKind:
		// integers are primitives - nothing to recurse
	case llvm.FunctionTypeKind:
		// check arguments and return(s)
		for i, v := range t.ParamTypes() {
			if err := c.checkType(v, checked, specials); err != nil {
				return fmt.Errorf("failed to verify argument %d of type %s: %s", i, t.String(), err.Error())
			}
		}
		if err := c.checkType(t.ReturnType(), checked, specials); err != nil {
			return fmt.Errorf("failed to verify return type of type %s: %s", t.String(), err.Error())
		}
	case llvm.StructTypeKind:
		// check all elements
		for i, v := range t.StructElementTypes() {
			if err := c.checkType(v, checked, specials); err != nil {
				return fmt.Errorf("failed to verify type of field %d of struct type %s: %s", i, t.String(), err.Error())
			}
		}
	case llvm.ArrayTypeKind:
		// check element type
		if err := c.checkType(t.ElementType(), checked, specials); err != nil {
			return fmt.Errorf("failed to verify element type of array type %s: %s", t.String(), err.Error())
		}
	case llvm.PointerTypeKind:
		// Pointers can't be checked in an opaque pointer world.
	case llvm.VectorTypeKind:
		// check element type
		if err := c.checkType(t.ElementType(), checked, specials); err != nil {
			return fmt.Errorf("failed to verify element type of vector type %s: %s", t.String(), err.Error())
		}
	default:
		return fmt.Errorf("unrecognized kind %q of type %s", t.TypeKind(), t.String())
	}

	return nil
}

func (c *checker) checkValue(v llvm.Value, types map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) error {
	// check type
	if err := c.checkType(v.Type(), types, specials); err != nil {
		return fmt.Errorf("failed to verify type of value: %s", err.Error())
	}

	// check if this is an undefined void
	if v.IsUndef() && v.Type().TypeKind() == llvm.VoidTypeKind {
		return errors.New("encountered undefined void value")
	}

	return nil
}

func (c *checker) checkInstruction(inst llvm.Value, types map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) error {
	// check value properties
	if err := c.checkValue(inst, types, specials); err != nil {
		return errorAt(inst, err.Error())
	}

	// The alloca instruction can be present in every basic block. However,
	// allocas in basic blocks other than the entry basic block have a number of
	// problems:
	//   * They are hard to optimize, leading to potential missed optimizations.
	//   * They may cause stack overflows in loops that would otherwise be
	//     innocent.
	//   * They cause extra code to be generated, because it requires the use of
	//     a frame pointer.
	//   * Perhaps most importantly, the coroutine lowering pass of LLVM (as of
	//     LLVM 9) cannot deal with these allocas:
	//     https://llvm.org/docs/Coroutines.html
	// Therefore, alloca instructions should be limited to the entry block.
	if !inst.IsAAllocaInst().IsNil() {
		if inst.InstructionParent() != inst.InstructionParent().Parent().EntryBasicBlock() {
			return errorAt(inst, "internal error: non-static alloca")
		}
	}

	// check operands
	for i := 0; i < inst.OperandsCount(); i++ {
		if err := c.checkValue(inst.Operand(i), types, specials); err != nil {
			return errorAt(inst, fmt.Sprintf("failed to validate operand %d of instruction %q: %s", i, inst.Name(), err.Error()))
		}
	}

	return nil
}

func (c *checker) checkBasicBlock(bb llvm.BasicBlock, types map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) []error {
	// check basic block value and type
	var errs []error
	if err := c.checkValue(bb.AsValue(), types, specials); err != nil {
		errs = append(errs, errorAt(bb.Parent(), fmt.Sprintf("failed to validate value of basic block %s: %v", bb.AsValue().Name(), err)))
	}

	// check instructions
	for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
		if err := c.checkInstruction(inst, types, specials); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (c *checker) checkFunction(fn llvm.Value, types map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) []error {
	// check function value and type
	var errs []error
	if err := c.checkValue(fn, types, specials); err != nil {
		errs = append(errs, fmt.Errorf("failed to validate value of function %s: %s", fn.Name(), err.Error()))
	}

	// check basic blocks
	for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
		errs = append(errs, c.checkBasicBlock(bb, types, specials)...)
	}

	return errs
}

// Module checks the given module and returns a slice of error, if there are
// any.
func Module(mod llvm.Module) []error {
	// check for any context mismatches
	var errs []error
	c := checker{
		ctx: mod.Context(),
	}
	if c.ctx == llvm.GlobalContext() {
		// somewhere we accidentally used the global context instead of a real context
		errs = append(errs, errors.New("module uses global context"))
	}

	types := map[llvm.Type]struct{}{}
	specials := map[llvm.TypeKind]llvm.Type{}
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		errs = append(errs, c.checkFunction(fn, types, specials)...)
	}
	for g := mod.FirstGlobal(); !g.IsNil(); g = llvm.NextGlobal(g) {
		if err := c.checkValue(g, types, specials); err != nil {
			errs = append(errs, fmt.Errorf("failed to verify global %s of module: %s", g.Name(), err.Error()))
		}
	}

	return errs
}
