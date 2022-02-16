package interp

import (
	"fmt"
	"strings"

	"tinygo.org/x/go-llvm"
)

type expr interface {
	eval() (value, error)

	create(dst *builder, dbg llvm.Metadata) (value, error)

	fmt.Stringer
}

func parseExpr(op llvm.Opcode, expr llvm.Value, parser parser) (expr, error) {
	switch op {
	case llvm.Add, llvm.Sub, llvm.Mul, llvm.UDiv:
		typ, err := parser.typ(expr.Type())
		if err != nil {
			return nil, err
		}
		switch typ := typ.(type) {
		case iType:
			bin, err := parseBinIntExpr(typ, expr, parser)
			if err != nil {
				return nil, err
			}
			if typ <= i64 {
				switch op {
				case llvm.Add:
					return smallIntAddExpr{bin}, nil
				case llvm.Sub:
					return smallIntSubExpr{bin}, nil
				case llvm.Mul:
					return smallIntMulExpr{bin}, nil
				case llvm.UDiv:
					return smallUIntDivExpr{bin}, nil
				default:
					panic("missing int bin op")
				}
			}
		}

		// Revert when processing any other types (small ints are not the only valid types here).
		return nil, todo("parse int bin op with type " + typ.String())

	case llvm.GetElementPtr:
		rawBase := expr.Operand(0)
		base, err := parser.value(rawBase)
		if err != nil {
			return nil, err
		}
		switch t := base.typ().(type) {
		case ptrType:
			over, err := parser.typ(rawBase.Type().ElementType())
			if err != nil {
				return nil, err
			}
			outer, err := parser.value(expr.Operand(1))
			if err != nil {
				return nil, err
			}
			indices := make([]value, expr.OperandsCount()-2)
			for i := range indices {
				idx, err := parser.value(expr.Operand(i + 2))
				if err != nil {
					return nil, err
				}
				indices[i] = idx
			}
			return gepExpr{base, over, outer, indices}, nil

		default:
			return nil, todo("gep of type " + t.String())
		}

	case llvm.SExt:
		src, err := parser.value(expr.Operand(0))
		if err != nil {
			return nil, err
		}
		to, err := parser.typ(expr.Type())
		if err != nil {
			return nil, err
		}
		switch to := to.(type) {
		case iType:
			if to <= i64 {
				return signExtendSmallExpr{src, to}, nil
			}
		}

		// Revert when processing any other types (small ints are not the only valid types here).
		return nil, todo("sign extend to " + to.String())

	case llvm.ICmp:
		pred := expr.IntPredicate()
		x, err := parser.value(expr.Operand(0))
		if err != nil {
			return nil, err
		}
		y, err := parser.value(expr.Operand(1))
		if err != nil {
			return nil, err
		}
		ty := x.typ()
		switch pred {
		case llvm.IntUGT:
			pred, x, y = llvm.IntULT, y, x
		case llvm.IntUGE:
			pred, x, y = llvm.IntULE, y, x
		case llvm.IntSGT:
			pred, x, y = llvm.IntSLT, y, x
		case llvm.IntSGE:
			pred, x, y = llvm.IntSLE, y, x
		}
		switch ty := ty.(type) {
		case iType:
			if ty <= i64 {
				return smallIntCmpExpr{x, y, ty, pred}, nil
			}
		case ptrType:
			iTy := ty.idxTy()
			return smallIntCmpExpr{cast(iTy, x), cast(iTy, y), iTy, pred}, nil
		}
		return nil, todo("icmp " + ty.String())

	default:
		return nil, todo("parse expr with op: " + opString(op))
	}
}

// smallIntAddExpr is an expression for the sum of two small integers.
type smallIntAddExpr struct {
	binIntExpr
}

func (e smallIntAddExpr) eval() (value, error) {
	if assert {
		if e.ty > 64 {
			panic("input too big")
		}
		if xTy := e.x.typ(); xTy != e.ty {
			return value{}, typeError{e.ty, xTy}
		}
		if yTy := e.y.typ(); yTy != e.ty {
			return value{}, typeError{e.ty, yTy}
		}
	}

	type kind uint8
	const (
		idk kind = iota
		num
		ptr
		whatever
	)

	var x, y uint64
	var xObj, yObj *memObj
	xKind, yKind := idk, idk
	switch val := e.x.val.(type) {
	case smallInt:
		x, xKind = e.x.raw, num
	case undef:
		xKind = whatever
	case *offAddr:
		x, xObj, xKind = e.x.raw, val.obj(), ptr
	}
	switch val := e.y.val.(type) {
	case smallInt:
		y, yKind = e.y.raw, num
	case undef:
		yKind = whatever
	case *offAddr:
		y, yObj, yKind = e.y.raw, val.obj(), ptr
	}

	type kinds struct{ x, y kind }
	switch (kinds{xKind, yKind}) {
	case kinds{num, num}:
		// Evaluate directly.
		return smallIntValue(e.ty, x+y), nil

	case kinds{ptr, num}:
		// Add to the offset, retaining overflow behavior.
		return xObj.addr(x + y), nil

	case kinds{num, ptr}:
		// Add to the offset, retaining overflow behavior.
		return yObj.addr(x + y), nil

	case kinds{num, idk}:
		if x == 0 {
			// 0 + y = y
			return e.y, nil
		}

	case kinds{idk, num}:
		if y == 0 {
			// x + 0 = x
			return e.x, nil
		}
	}
	if xKind == whatever || yKind == whatever {
		// The result is undefined.
		return undefValue(e.ty), nil
	}

	return value{}, errRuntime
}

func (e smallIntAddExpr) resolve(stack []value) (smallIntAddExpr, error) {
	res, err := e.binIntExpr.resolve(stack)
	if err != nil {
		return smallIntAddExpr{}, err
	}
	return smallIntAddExpr{res}, nil
}

func (e smallIntAddExpr) String() string {
	return e.binIntExpr.String("add")
}

func (e smallIntAddExpr) create(dst *builder, dbg llvm.Metadata) (value, error) {
	switch v, err := e.eval(); err {
	case nil:
		return v, nil
	case errRuntime:
		return dst.insertInst(&smallIntAddInst{e, dbg}), nil
	default:
		return value{}, err
	}
}

type smallIntAddInst struct {
	expr smallIntAddExpr
	dbg  llvm.Metadata
}

func (i *smallIntAddInst) result() typ {
	return i.expr.ty
}

func (i *smallIntAddInst) String() string {
	return i.expr.String() + dbgSuffix(i.dbg)
}

func (i *smallIntAddInst) exec(state *execState) error {
	// This would be a great use for generics. . .

	// Resolve the expression.
	expr, err := i.expr.resolve(state.locals())
	if err != nil {
		return err
	}

	// Evaluate the expression.
	v, err := expr.eval()
	switch err {
	case nil:
	case errRuntime:
		// Create a runtime instruction to evaluate the expression.
		v, err = i.expr.create(&state.rt, i.dbg)
		if err != nil {
			return err
		}
	default:
		return err
	}

	// Push the result onto the stack.
	state.stack = append(state.stack, v)

	return nil
}

func (i *smallIntAddInst) runtime(gen *rtGen) error {
	return i.expr.runtime(gen, i.dbg, (*llvm.Builder).CreateAdd)
}

// smallIntSubExpr is an expression for the difference of two small integers.
type smallIntSubExpr struct {
	binIntExpr
}

func (e smallIntSubExpr) eval() (value, error) {
	if assert {
		if e.ty > 64 {
			panic("input too big")
		}
		if xTy := e.x.typ(); xTy != e.ty {
			return value{}, typeError{e.ty, xTy}
		}
		if yTy := e.y.typ(); yTy != e.ty {
			return value{}, typeError{e.ty, yTy}
		}
	}

	type kind uint8
	const (
		idk kind = iota
		num
		ptr
		whatever
	)

	var x, y uint64
	var xObj, yObj *memObj
	xKind, yKind := idk, idk
	switch val := e.x.val.(type) {
	case smallInt:
		x, xKind = e.x.raw, num
	case undef:
		xKind = whatever
	case *offAddr:
		x, xObj, xKind = e.x.raw, val.obj(), ptr
	}
	switch val := e.y.val.(type) {
	case smallInt:
		y, yKind = e.y.raw, num
	case undef:
		yKind = whatever
	case *offAddr:
		y, yObj, yKind = e.y.raw, val.obj(), ptr
	}

	type kinds struct{ x, y kind }
	switch (kinds{xKind, yKind}) {
	case kinds{num, num}:
		// Evaluate directly.
		return smallIntValue(e.ty, x-y), nil

	case kinds{ptr, ptr}:
		if xObj == yObj {
			// (ptr + x) - (ptr + y) = x - y
			return smallIntValue(e.ty, x-y), nil
		}

	case kinds{ptr, num}:
		// Add to the offset, retaining overflow behavior.
		return xObj.addr(x - y), nil

	case kinds{idk, num}:
		if y == 0 {
			// x - 0 = x
			return e.x, nil
		}
	}
	if xKind == whatever || yKind == whatever {
		// The result is undefined.
		return undefValue(e.ty), nil
	}

	return value{}, errRuntime
}

func (e smallIntSubExpr) resolve(stack []value) (smallIntSubExpr, error) {
	res, err := e.binIntExpr.resolve(stack)
	if err != nil {
		return smallIntSubExpr{}, err
	}
	return smallIntSubExpr{res}, nil
}

func (e smallIntSubExpr) String() string {
	return e.binIntExpr.String("sub")
}

func (e smallIntSubExpr) create(dst *builder, dbg llvm.Metadata) (value, error) {
	switch v, err := e.eval(); err {
	case nil:
		return v, nil
	case errRuntime:
		return dst.insertInst(&smallIntSubInst{e, dbg}), nil
	default:
		return value{}, err
	}
}

type smallIntSubInst struct {
	expr smallIntSubExpr
	dbg  llvm.Metadata
}

func (i *smallIntSubInst) result() typ {
	return i.expr.ty
}

func (i *smallIntSubInst) String() string {
	return i.expr.String() + dbgSuffix(i.dbg)
}

func (i *smallIntSubInst) exec(state *execState) error {
	// This would be a great use for generics. . .

	// Resolve the expression.
	expr, err := i.expr.resolve(state.locals())
	if err != nil {
		return err
	}

	// Evaluate the expression.
	v, err := expr.eval()
	switch err {
	case nil:
	case errRuntime:
		// Create a runtime instruction to evaluate the expression.
		v, err = i.expr.create(&state.rt, i.dbg)
		if err != nil {
			return err
		}
	default:
		return err
	}

	// Push the result onto the stack.
	state.stack = append(state.stack, v)

	return nil
}

func (i *smallIntSubInst) runtime(gen *rtGen) error {
	return i.expr.runtime(gen, i.dbg, (*llvm.Builder).CreateSub)
}

// smallIntMulExpr is an expression for the product of two small integers.
type smallIntMulExpr struct {
	binIntExpr
}

func (e smallIntMulExpr) eval() (value, error) {
	if assert {
		if e.ty > 64 {
			panic("input too big")
		}
		if xTy := e.x.typ(); xTy != e.ty {
			return value{}, typeError{xTy, e.ty}
		}
		if yTy := e.y.typ(); yTy != e.ty {
			return value{}, typeError{yTy, e.ty}
		}
	}

	type kind uint8
	const (
		idk kind = iota
		num
		whatever
	)

	var x, y uint64
	xKind, yKind := idk, idk
	switch e.x.val.(type) {
	case smallInt:
		x, xKind = e.x.raw, num
	case undef:
		xKind = whatever
	}
	switch e.y.val.(type) {
	case smallInt:
		y, yKind = e.y.raw, num
	case undef:
		yKind = whatever
	}

	type kinds struct{ x, y kind }
	switch (kinds{xKind, yKind}) {
	case kinds{num, num}:
		// Evaluate directly.
		return smallIntValue(e.ty, x*y), nil

	case kinds{num, idk}:
		switch x {
		case 0:
			// 0 * y = 0
			return smallIntValue(e.ty, 0), nil

		case 1:
			// 1 * y = y
			return e.y, nil
		}

	case kinds{idk, num}:
		switch y {
		case 0:
			// x * 0 = 0
			return smallIntValue(e.ty, 0), nil

		case 1:
			// x * 1 = x
			return e.x, nil
		}

	case kinds{num, whatever}:
		switch {
		case x == 0:
			// 0 * undef = 0
			return smallIntValue(e.ty, 0), nil

		case x%2 != 0:
			// x * undef = undef if x is coprime to the base (2^n).
			return undefValue(e.ty), nil
		}

	case kinds{whatever, num}:
		switch {
		case y == 0:
			// undef * 0 = 0
			return smallIntValue(e.ty, 0), nil

		case y%2 != 0:
			// undef * y = undef if y is coprime to the base (2^n).
			return undefValue(e.ty), nil
		}
	}

	return value{}, errRuntime
}

func (e smallIntMulExpr) resolve(stack []value) (smallIntMulExpr, error) {
	res, err := e.binIntExpr.resolve(stack)
	if err != nil {
		return smallIntMulExpr{}, err
	}
	return smallIntMulExpr{res}, nil
}

func (e smallIntMulExpr) String() string {
	return e.binIntExpr.String("mul")
}

func (e smallIntMulExpr) create(dst *builder, dbg llvm.Metadata) (value, error) {
	switch v, err := e.eval(); err {
	case nil:
		return v, nil
	case errRuntime:
		return dst.insertInst(&smallIntMulInst{e, dbg}), nil
	default:
		return value{}, err
	}
}

type smallIntMulInst struct {
	expr smallIntMulExpr
	dbg  llvm.Metadata
}

func (i *smallIntMulInst) result() typ {
	return i.expr.ty
}

func (i *smallIntMulInst) String() string {
	return i.expr.String() + dbgSuffix(i.dbg)
}

func (i *smallIntMulInst) exec(state *execState) error {
	// This would be a great use for generics. . .

	// Resolve the expression.
	expr, err := i.expr.resolve(state.locals())
	if err != nil {
		return err
	}

	// Evaluate the expression.
	v, err := expr.eval()
	switch err {
	case nil:
	case errRuntime:
		// Create a runtime instruction to evaluate the expression.
		v, err = i.expr.create(&state.rt, i.dbg)
		if err != nil {
			return err
		}
	default:
		return err
	}

	// Push the result onto the stack.
	state.stack = append(state.stack, v)

	return nil
}

func (i *smallIntMulInst) runtime(gen *rtGen) error {
	return i.expr.runtime(gen, i.dbg, (*llvm.Builder).CreateMul)
}

// smallUIntDivExpr is an expression for the quotient of two small unsigned integers.
type smallUIntDivExpr struct {
	binIntExpr
}

func (e smallUIntDivExpr) eval() (value, error) {
	if assert {
		if e.ty > 64 {
			panic("input too big")
		}
		if xTy := e.x.typ(); xTy != e.ty {
			return value{}, typeError{xTy, e.ty}
		}
		if yTy := e.y.typ(); yTy != e.ty {
			return value{}, typeError{yTy, e.ty}
		}
	}

	type kind uint8
	const (
		idk kind = iota
		num
		whatever
	)

	var x, y uint64
	xKind, yKind := idk, idk
	switch e.x.val.(type) {
	case smallInt:
		x, xKind = e.x.raw, num
	case undef:
		xKind = whatever
	}
	switch e.y.val.(type) {
	case smallInt:
		y, yKind = e.y.raw, num
	case undef:
		yKind = whatever
	}

	type kinds struct{ x, y kind }
	switch (kinds{xKind, yKind}) {
	case kinds{num, num}:
		// Evaluate directly.
		if y == 0 {
			// Division by zero is undefined behavior.
			return value{}, errUB
		}
		return smallIntValue(e.ty, x/y), nil

	case kinds{num, idk}:
		if x == 0 {
			// 0 / y = 0 (for y != 0)
			return smallIntValue(e.ty, 0), nil
		}

	case kinds{idk, num}:
		switch y {
		case 0:
			// Division by zero is undefined behavior.
			return value{}, errUB

		case 1:
			// x / 1 = x
			return e.x, nil
		}

	case kinds{num, whatever}:
		// Division by undef is undefined behavior.
		return value{}, errUB

	case kinds{whatever, num}:
		switch {
		case y == 0:
			// Division by zero is undefined behavior.
			return value{}, errUB

		case y%2 != 0:
			// undef / y = undef if y is coprime to the base (2^n).
			return undefValue(e.ty), nil
		}
	}

	return value{}, errRuntime
}

func (e smallUIntDivExpr) resolve(stack []value) (smallUIntDivExpr, error) {
	res, err := e.binIntExpr.resolve(stack)
	if err != nil {
		return smallUIntDivExpr{}, err
	}
	return smallUIntDivExpr{res}, nil
}

func (e smallUIntDivExpr) String() string {
	return e.binIntExpr.String("udiv")
}

func (e smallUIntDivExpr) create(dst *builder, dbg llvm.Metadata) (value, error) {
	switch v, err := e.eval(); err {
	case nil:
		return v, nil
	case errRuntime:
		return dst.insertInst(&smallUIntDivInst{e, dbg}), nil
	default:
		return value{}, err
	}
}

type smallUIntDivInst struct {
	expr smallUIntDivExpr
	dbg  llvm.Metadata
}

func (i *smallUIntDivInst) result() typ {
	return i.expr.ty
}

func (i *smallUIntDivInst) String() string {
	return i.expr.String() + dbgSuffix(i.dbg)
}

func (i *smallUIntDivInst) exec(state *execState) error {
	// This would be a great use for generics. . .

	// Resolve the expression.
	expr, err := i.expr.resolve(state.locals())
	if err != nil {
		return err
	}

	// Evaluate the expression.
	v, err := expr.eval()
	switch err {
	case nil:
	case errRuntime:
		// Create a runtime instruction to evaluate the expression.
		v, err = i.expr.create(&state.rt, i.dbg)
		if err != nil {
			return err
		}
	default:
		return err
	}

	// Push the result onto the stack.
	state.stack = append(state.stack, v)

	return nil
}

func (i *smallUIntDivInst) runtime(gen *rtGen) error {
	return i.expr.runtime(gen, i.dbg, (*llvm.Builder).CreateUDiv)
}

func parseBinIntExpr(typ iType, expr llvm.Value, parser parser) (binIntExpr, error) {
	x, err := parser.value(expr.Operand(0))
	if err != nil {
		return binIntExpr{}, err
	}
	if assert {
		xTyp := x.typ()
		if xTyp != typ {
			return binIntExpr{}, typeError{typ, xTyp}
		}
	}
	y, err := parser.value(expr.Operand(1))
	if err != nil {
		return binIntExpr{}, err
	}
	if assert {
		yTyp := y.typ()
		if yTyp != typ {
			return binIntExpr{}, typeError{typ, yTyp}
		}
	}
	return binIntExpr{x, y, typ}, nil
}

type binIntExpr struct {
	x, y value
	ty   iType
}

func (e binIntExpr) resolve(stack []value) (binIntExpr, error) {
	x := e.x.resolve(stack)
	if assert {
		xTy := x.typ()
		if xTy != e.ty {
			return binIntExpr{}, typeError{e.ty, xTy}
		}
	}
	y := e.y.resolve(stack)
	if assert {
		yTy := y.typ()
		if yTy != e.ty {
			return binIntExpr{}, typeError{e.ty, yTy}
		}
	}
	return binIntExpr{x, y, e.ty}, nil
}

func (e binIntExpr) String(op string) string {
	return e.ty.String() + " " + op + " " + e.x.String() + ", " + e.y.String()
}

func (e binIntExpr) runtime(gen *rtGen, dbg llvm.Metadata, fn func(*llvm.Builder, llvm.Value, llvm.Value, string) llvm.Value) error {
	oldDbg := gen.dbg
	gen.dbg = dbg
	defer func() { gen.dbg = oldDbg }()

	t := gen.iType(e.ty)
	res := fn(&gen.builder, gen.value(t, e.x), gen.value(t, e.y), "")
	gen.applyDebug(res)
	gen.stack = append(gen.stack, res)
	return nil
}

type smallIntCmpExpr struct {
	x, y value
	ty   iType
	pred llvm.IntPredicate
}

func (e smallIntCmpExpr) eval() (value, error) {
	type kind uint8
	const (
		idk kind = iota
		num
		ptr
		whatever
	)

	var x, y uint64
	var xObj, yObj *memObj
	xKind, yKind := idk, idk
	switch val := e.x.val.(type) {
	case smallInt:
		x, xKind = e.x.raw, num
	case undef:
		xKind = whatever
	case *offAddr:
		x, xObj, xKind = e.x.raw, val.obj(), ptr
	}
	switch val := e.y.val.(type) {
	case smallInt:
		y, yKind = e.y.raw, num
	case undef:
		yKind = whatever
	case *offAddr:
		y, yObj, yKind = e.y.raw, val.obj(), ptr
	}

	type kinds struct{ x, y kind }
	switch (kinds{xKind, yKind}) {
	case kinds{ptr, ptr}:
		switch e.pred {
		case llvm.IntEQ:
			if xObj != yObj && xObj.unique && yObj.unique && (x == y || (x < xObj.size && y < yObj.size)) {
				return boolValue(false), nil
			}

		case llvm.IntNE:
			if xObj != yObj && xObj.unique && yObj.unique && (x == y || (x < xObj.size && y < yObj.size)) {
				return boolValue(true), nil
			}
		}
		if xObj != yObj || x >= xObj.size || y >= yObj.size {
			break
		}
		fallthrough
	case kinds{num, num}:
		switch e.pred {
		case llvm.IntEQ:
			return boolValue(x == y), nil
		case llvm.IntNE:
			return boolValue(x != y), nil
		case llvm.IntULT:
			return boolValue(x < y), nil
		case llvm.IntULE:
			return boolValue(x <= y), nil
		case llvm.IntSLT:
			shift := i64 - e.ty
			return boolValue(int64(x<<shift)>>shift < int64(y<<shift)>>shift), nil
		case llvm.IntSLE:
			shift := i64 - e.ty
			return boolValue(int64(x<<shift)>>shift <= int64(y<<shift)>>shift), nil
		}

	case kinds{ptr, num}:
		if x < xObj.size && y == 0 {
			switch e.pred {
			case llvm.IntEQ:
				return boolValue(false), nil
			case llvm.IntNE:
				return boolValue(true), nil
			}
		}

	case kinds{num, ptr}:
		if y < yObj.size && x == 0 {
			switch e.pred {
			case llvm.IntEQ:
				return boolValue(false), nil
			case llvm.IntNE:
				return boolValue(true), nil
			}
		}

	case kinds{whatever, whatever}:
		return undefValue(i1), nil
	}

	return value{}, errRuntime
}

func (e smallIntCmpExpr) resolve(stack []value) smallIntCmpExpr {
	return smallIntCmpExpr{e.x.resolve(stack), e.y.resolve(stack), e.ty, e.pred}
}

func (e smallIntCmpExpr) String() string {
	var pred string
	switch e.pred {
	case llvm.IntEQ:
		pred = "eq"
	case llvm.IntNE:
		pred = "ne"
	case llvm.IntULT:
		pred = "ult"
	case llvm.IntULE:
		pred = "ule"
	case llvm.IntSLT:
		pred = "slt"
	case llvm.IntSLE:
		pred = "sle"
	default:
		pred = "idk"
	}
	return "icmp " + pred + " " + e.x.String() + ", " + e.y.String()
}

func (e smallIntCmpExpr) create(dst *builder, dbg llvm.Metadata) (value, error) {
	switch v, err := e.eval(); err {
	case nil:
		return v, nil
	case errRuntime:
		return dst.insertInst(&smallIntCmpInst{e, dbg}), nil
	default:
		return value{}, err
	}
}

type smallIntCmpInst struct {
	expr smallIntCmpExpr
	dbg  llvm.Metadata
}

func (i *smallIntCmpInst) result() typ {
	return i1
}

func (i *smallIntCmpInst) String() string {
	return i.expr.String() + dbgSuffix(i.dbg)
}

func (i *smallIntCmpInst) exec(state *execState) error {
	// This would be a great use for generics. . .

	// Resolve the expression.
	expr := i.expr.resolve(state.locals())

	// Evaluate the expression.
	v, err := expr.eval()
	switch err {
	case nil:
	case errRuntime:
		// Create a runtime instruction to evaluate the expression.
		v, err = i.expr.create(&state.rt, i.dbg)
		if err != nil {
			return err
		}
	default:
		return err
	}

	// Push the result onto the stack.
	state.stack = append(state.stack, v)

	return nil
}

func (i *smallIntCmpInst) runtime(gen *rtGen) error {
	oldDbg := gen.dbg
	gen.dbg = i.dbg
	defer func() { gen.dbg = oldDbg }()

	ty := gen.iType(i.expr.ty)
	x := gen.value(ty, i.expr.x)
	y := gen.value(ty, i.expr.y)
	res := gen.builder.CreateICmp(i.expr.pred, x, y, "")
	gen.applyDebug(res)
	gen.stack = append(gen.stack, res)
	return nil
}

// signExtendSmallExpr is an expression to sign-extend a small integer value by copying the sign bit.
type signExtendSmallExpr struct {
	in value
	to iType
}

func (e signExtendSmallExpr) eval() (value, error) {
	switch val := e.in.val.(type) {
	case smallInt:
		shift := 64 - val
		return smallIntValue(e.to, uint64(int64(e.in.raw<<shift)>>shift)), nil

	case undef:
		// TODO: is this a bad idea?
		return e.to.zero(), nil
	}

	return value{}, errRuntime
}

func (e signExtendSmallExpr) resolve(stack []value) (signExtendSmallExpr, error) {
	return signExtendSmallExpr{e.in.resolve(stack), e.to}, nil
}

func (e signExtendSmallExpr) String() string {
	return e.to.String() + " sext " + e.in.String()
}

func (e signExtendSmallExpr) create(dst *builder, dbg llvm.Metadata) (value, error) {
	return dst.insertInst(&signExtendInst{e, dbg}), nil
}

type signExtendInst struct {
	expr signExtendSmallExpr
	dbg  llvm.Metadata
}

func (i *signExtendInst) result() typ {
	return i.expr.to
}

func (i *signExtendInst) String() string {
	return i.expr.String() + dbgSuffix(i.dbg)
}

func (i *signExtendInst) exec(state *execState) error {
	// This would be a great use for generics. . .

	// Resolve the expression.
	expr, err := i.expr.resolve(state.locals())
	if err != nil {
		return err
	}

	// Evaluate the expression.
	v, err := expr.eval()
	switch err {
	case nil:
	case errRuntime:
		// Create a runtime instruction to evaluate the expression.
		v, err = i.expr.create(&state.rt, i.dbg)
		if err != nil {
			return err
		}
	default:
		return err
	}

	// Push the result onto the stack.
	state.stack = append(state.stack, v)

	return nil
}

func (i *signExtendInst) runtime(gen *rtGen) error {
	oldDbg := gen.dbg
	gen.dbg = i.dbg
	defer func() { gen.dbg = oldDbg }()

	from := gen.value(gen.iType(i.expr.in.typ().(iType)), i.expr.in)
	res := gen.builder.CreateSExt(from, gen.iType(i.expr.to), "")
	gen.applyDebug(res)
	gen.stack = append(gen.stack, res)
	return nil
}

// gepExpr is an expression that to get a pointer to an element of an object.
type gepExpr struct {
	base    value
	over    typ
	outer   value
	indices []value
}

func (e gepExpr) eval() (value, error) {
	ty := e.base.typ()
	idxTy := ty.(ptrType).idxTy()

	// Evaluate the outer offset.
	outer, err := signCast(idxTy, e.outer)
	if err != nil {
		return value{}, err
	}
	off, err := smallIntMulExpr{binIntExpr{outer, smallIntValue(idxTy, e.over.bytes()), idxTy}}.eval()
	if err != nil {
		return value{}, err
	}

	// Apply additional offsets.
	into := e.over
	for _, idx := range e.indices {
		var term value
		var next typ
		switch into := into.(type) {
		case *structType:
			if assert {
				if iTy := idx.typ(); iTy != i32 {
					return value{}, fmt.Errorf("bad struct index: %w", typeError{i32, iTy})
				}
				if _, ok := idx.val.(smallInt); !ok {
					return value{}, fmt.Errorf("struct index %s is not an integer constant", idx.String())
				}
			}
			i := int32(idx.raw)
			if i < 0 || int(i) >= len(into.fields) {
				return value{}, fmt.Errorf("struct index out of bounds [0, %d): %d", len(into.fields), i)
			}
			term = smallIntValue(idxTy, into.fields[i].offset)
			next = into.fields[i].ty

		case arrType:
			idx, err := signCast(idxTy, idx)
			if err != nil {
				return value{}, err
			}
			term, err = smallIntMulExpr{binIntExpr{idx, smallIntValue(idxTy, into.of.bytes()), idxTy}}.eval()
			if err != nil {
				return value{}, err
			}
			next = into.of

		default:
			return value{}, errRuntime
		}
		off, err = smallIntAddExpr{binIntExpr{off, term, idxTy}}.eval()
		if err != nil {
			return value{}, err
		}
		into = next
	}

	if _, ok := off.val.(smallInt); ok && off.raw == 0 {
		// There is no applied offset.
		return e.base, nil
	}

	switch val := e.base.val.(type) {
	case *offPtr:
		// Convert gep(gep(x, a), b) to gep(x, a+b).
		mergeOff, err := smallIntAddExpr{binIntExpr{smallIntValue(idxTy, e.base.raw), off, idxTy}}.eval()
		if err != nil {
			return value{}, err
		}
		return val.obj().gep(mergeOff), nil
	case uglyGEP:
		// Convert gep(gep(x, a), b) to gep(x, a+b).
		mergeOff, err := smallIntAddExpr{binIntExpr{value{val.val, e.base.raw}, off, idxTy}}.eval()
		if err != nil {
			return value{}, err
		}
		return val.obj.gep(mergeOff), nil
	case castVal:
		from := value{val.val, e.base.raw}
		if from.typ() == idxTy {
			// Convert gep(inttoptr(x), off) to inttoptr(x + off).
			v, err := smallIntAddExpr{binIntExpr{from, off, idxTy}}.eval()
			if err != nil {
				return value{}, err
			}
			return cast(val.to, v), nil
		}
	case undef:
		return undefValue(ty), nil
	}

	return value{}, errRuntime
}

func (e gepExpr) String() string {
	indices := make([]string, len(e.indices))
	for i, idx := range e.indices {
		indices[i] = ", " + idx.String()
	}
	return e.base.typ().String() + " gep " + e.over.String() + ", " + e.base.String() + ", " + e.outer.String() + strings.Join(indices, "")
}

func (e gepExpr) create(dst *builder, dbg llvm.Metadata) (value, error) {
	ty := e.base.typ().(ptrType)
	idxTy := ty.idxTy()

	// Generate offset terms.
	terms := make([]value, len(e.indices)+1)
	{
		outer, err := createSignCast(dst, idxTy, e.outer, dbg)
		if err != nil {
			return value{}, err
		}
		outer, err = smallIntMulExpr{binIntExpr{outer, smallIntValue(idxTy, e.over.bytes()), idxTy}}.create(dst, dbg)
		if err != nil {
			return value{}, err
		}
		terms[0] = outer
	}
	into := e.over
	for i, idx := range e.indices {
		var term value
		var next typ
		switch into := into.(type) {
		case *structType:
			if assert {
				if iTy := idx.typ(); iTy != i32 {
					return value{}, fmt.Errorf("bad struct index: %w", typeError{i32, iTy})
				}
				if _, ok := idx.val.(smallInt); !ok {
					return value{}, fmt.Errorf("struct index %s is not an integer constant", idx.String())
				}
			}
			i := int32(idx.raw)
			if i < 0 || int(i) >= len(into.fields) {
				return value{}, fmt.Errorf("struct index out of bounds [0, %d): %d", len(into.fields), i)
			}
			term = smallIntValue(idxTy, into.fields[i].offset)
			next = into.fields[i].ty

		case arrType:
			idx, err := createSignCast(dst, idxTy, idx, dbg)
			if err != nil {
				return value{}, err
			}
			term, err = smallIntMulExpr{binIntExpr{idx, smallIntValue(idxTy, into.of.bytes()), idxTy}}.create(dst, dbg)
			if err != nil {
				return value{}, err
			}
			next = into.of

		default:
			return value{}, todo("gep index into " + idx.typ().String())
		}
		terms[i+1] = term
		into = next
	}

	// Merge constant terms.
	var obj *memObj
	{
		var c uint64
		switch val := e.base.val.(type) {
		case *offPtr:
			obj, c = val.obj(), e.base.raw
		case uglyGEP:
			obj = val.obj
			terms = append(terms, value{val.val, e.base.raw})
		}
		tmp := terms
		terms = terms[:0]
		for _, v := range tmp {
			if _, ok := v.val.(smallInt); ok {
				c += v.raw
				continue
			}

			terms = append(terms, v)
		}
		if c != 0 {
			terms = append(terms, smallIntValue(idxTy, c))
		}
	}
	if len(terms) == 0 {
		return e.base, nil
	}

	// Construct the offset.
	off := terms[0]
	for _, v := range terms[1:] {
		var err error
		off, err = smallIntAddExpr{binIntExpr{off, v, idxTy}}.create(dst, dbg)
		if err != nil {
			return value{}, err
		}
	}

	// Generate the result value.
	if obj != nil {
		return obj.gep(off), nil
	}
	return dst.insertInst(&uglyGEPInst{e.base, off, ty, dbg}), nil
}

type uglyGEPInst struct {
	base value
	off  value
	ty   ptrType
	dbg  llvm.Metadata
}

func (i *uglyGEPInst) result() typ {
	return i.base.typ()
}

func (i *uglyGEPInst) String() string {
	return i.result().String() + " uglygep " + i.base.String() + ", " + i.off.String() + dbgSuffix(i.dbg)
}

func (i *uglyGEPInst) exec(state *execState) error {
	// Resolve the arguments.
	locals := state.locals()
	base := i.base.resolve(locals)
	off := i.off.resolve(locals)

	// Do the GEP.
	var v value
	var ok bool
	switch val := base.val.(type) {
	case *offPtr:
		offExpr := smallIntAddExpr{binIntExpr{smallIntValue(i.ty.idxTy(), base.raw), off, i.ty.idxTy()}}
		mergeOff, err := offExpr.eval()
		switch err {
		case nil:
		case errRuntime:
			mergeOff, err = offExpr.create(&state.rt, i.dbg)
			if err != nil {
				return err
			}
		default:
			return err
		}
		v = val.obj().gep(mergeOff)
		ok = true

	case uglyGEP:
		offExpr := smallIntAddExpr{binIntExpr{value{base.val, base.raw}, off, i.ty.idxTy()}}
		mergeOff, err := offExpr.eval()
		switch err {
		case nil:
		case errRuntime:
			mergeOff, err = offExpr.create(&state.rt, i.dbg)
			if err != nil {
				return err
			}
		default:
			return err
		}
		v = val.obj.gep(mergeOff)
		ok = true

	case undef:
		v = undefValue(i.ty)
		ok = true

	case castVal:
		from := value{val.val, base.raw}
		if from.typ() == i.ty.idxTy() {
			// Convert gep(inttoptr(x), off) to inttoptr(x + off).
			var err error
			v, err = smallIntAddExpr{binIntExpr{from, off, i.ty.idxTy()}}.eval()
			if err != nil {
				return err
			}
			v = cast(val.to, v)
			ok = true
		}
	}
	if !ok {
		v = state.rt.insertInst(&uglyGEPInst{base, off, i.ty, i.dbg})
	}

	// Push the result onto the stack.
	state.stack = append(state.stack, v)

	return nil
}

func (i *uglyGEPInst) runtime(gen *rtGen) error {
	oldDbg := gen.dbg
	gen.dbg = i.dbg
	defer func() { gen.dbg = oldDbg }()

	base := gen.value(gen.ptr(gen.iType(i8), i.ty.in()), i.base)
	off := gen.value(gen.iType(i.ty.idxTy()), i.off)
	res := gen.builder.CreateGEP(base, []llvm.Value{off}, "")
	gen.applyDebug(res)
	gen.stack = append(gen.stack, res)
	return nil
}

func createSignCast(dst *builder, to iType, v value, dbg llvm.Metadata) (value, error) {
	from := v.typ().(iType)
	switch {
	case to > from:
		expr := signExtendSmallExpr{v, to}
		v, err := expr.eval()
		switch err {
		case nil:
			return v, nil
		case errRuntime:
			return expr.create(dst, dbg)
		default:
			return value{}, err
		}
	case to < from:
		return cast(to, v), nil
	default:
		return v, nil
	}
}

func signCast(to iType, v value) (value, error) {
	from := v.typ().(iType)
	switch {
	case to > from:
		return signExtendSmallExpr{v, to}.eval()
	case to < from:
		return cast(to, v), nil
	default:
		return v, nil
	}
}
