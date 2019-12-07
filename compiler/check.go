package compiler

// This file implements a set of sanity checks for the IR that is generated.
// It can catch some mistakes that LLVM's verifier cannot.

import (
	"errors"
	"fmt"

	"tinygo.org/x/go-llvm"
)

func (c *Compiler) checkType(t llvm.Type, checked map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) error {
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
		return fmt.Errorf("type %q uses context %v instead of the main context %v", t.Context(), c.ctx)
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
		// check underlying type
		if err := c.checkType(t.ElementType(), checked, specials); err != nil {
			return fmt.Errorf("failed to verify underlying type of pointer type %s: %s", t.String(), err.Error())
		}
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

func (c *Compiler) checkValue(v llvm.Value, types map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) error {
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

func (c *Compiler) checkInstruction(inst llvm.Value, types map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) error {
	// check value properties
	if err := c.checkValue(inst, types, specials); err != nil {
		return errorAt(inst, err.Error())
	}

	// check operands
	for i := 0; i < inst.OperandsCount(); i++ {
		if err := c.checkValue(inst.Operand(i), types, specials); err != nil {
			return errorAt(inst, fmt.Sprintf("failed to validate operand %d of instruction %q: %s", i, inst.Name(), err.Error()))
		}
	}

	return nil
}

func (c *Compiler) checkBasicBlock(bb llvm.BasicBlock, types map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) []error {
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

func (c *Compiler) checkFunction(fn llvm.Value, types map[llvm.Type]struct{}, specials map[llvm.TypeKind]llvm.Type) []error {
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

func (c *Compiler) checkModule() []error {
	// check for any context mismatches
	var errs []error
	switch {
	case c.mod.Context() == c.ctx:
		// this is correct
	case c.mod.Context() == llvm.GlobalContext():
		// somewhere we accidentally used the global context instead of a real context
		errs = append(errs, errors.New("module uses global context"))
	default:
		// we used some other context by accident
		errs = append(errs, fmt.Errorf("module uses context %v instead of the main context %v", c.mod.Context(), c.ctx))
	}

	types := map[llvm.Type]struct{}{}
	specials := map[llvm.TypeKind]llvm.Type{}
	for fn := c.mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		errs = append(errs, c.checkFunction(fn, types, specials)...)
	}
	for g := c.mod.FirstGlobal(); !g.IsNil(); g = llvm.NextGlobal(g) {
		if err := c.checkValue(g, types, specials); err != nil {
			errs = append(errs, fmt.Errorf("failed to verify global %s of module: %s", g.Name(), err.Error()))
		}
	}

	return errs
}
