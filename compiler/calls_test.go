package compiler

import (
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/ssa"
)

func TestBinOpCallProperties(t *testing.T) {
	t.Parallel()

	intType := types.Typ[types.Int]
	uintType := types.Typ[types.Uint]
	floatType := types.Typ[types.Float64]
	stringType := types.Typ[types.String]
	anyType := types.NewInterfaceType(nil, nil)
	intArray := types.NewArray(intType, 2)
	anyArray := types.NewArray(anyType, 2)

	for _, tc := range []struct {
		op        token.Token
		x, y      types.Type
		mayUnwind bool
	}{
		{token.ADD, intType, intType, false},
		{token.SUB, intType, intType, false},
		{token.MUL, intType, intType, false},
		{token.AND, intType, intType, false},
		{token.OR, intType, intType, false},
		{token.XOR, intType, intType, false},
		{token.AND_NOT, intType, intType, false},
		{token.LSS, intType, intType, false},
		{token.LEQ, intType, intType, false},
		{token.GTR, intType, intType, false},
		{token.GEQ, intType, intType, false},
		{token.EQL, intType, intType, false},
		{token.NEQ, intType, intType, false},
		{token.ADD, stringType, stringType, false},
		{token.EQL, stringType, stringType, false},
		{token.QUO, floatType, floatType, false},
		{token.SHL, intType, uintType, false},
		{token.SHR, intType, uintType, false},
		{token.EQL, intArray, intArray, false},
		{token.QUO, intType, intType, true},
		{token.REM, intType, intType, true},
		{token.QUO, uintType, uintType, true},
		{token.SHL, intType, intType, true},
		{token.SHR, intType, intType, true},
		{token.EQL, anyType, anyType, true},
		{token.NEQ, anyType, anyType, true},
		{token.EQL, anyArray, anyArray, true},
	} {
		instr := &ssa.BinOp{Op: tc.op, X: ssa.NewConst(nil, tc.x), Y: ssa.NewConst(nil, tc.y)}
		want := callProperties(0)
		if tc.mayUnwind {
			want = callMayUnwind
		}
		if got := instructionCallProperties(instr); got != want {
			t.Errorf("%s %s %s: got properties %d, want %d", tc.x, tc.op, tc.y, got, want)
		}
	}
}
