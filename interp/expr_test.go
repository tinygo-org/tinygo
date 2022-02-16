package interp

import (
	"testing"
)

func TestEvalSmallAdd(t *testing.T) {
	t.Parallel()

	ptrTy := pointer(defaultAddrSpace, i32)
	x := memObj{
		ptrTy: ptrTy,
		name:  "x",
	}
	testEval(t,
		evalCase{
			name:   "Const",
			expr:   smallIntAddExpr{binIntExpr{smallIntValue(i8, 1), smallIntValue(i8, 2), i8}},
			expect: "i8 3",
		},
		evalCase{
			name:   "Overflow",
			expr:   smallIntAddExpr{binIntExpr{smallIntValue(i8, 255), smallIntValue(i8, 3), i8}},
			expect: "i8 2",
		},
		evalCase{
			name:   "PtrOff",
			expr:   smallIntAddExpr{binIntExpr{x.addr((1 << 32) - 1), smallIntValue(i32, 5), i32}},
			expect: "i32 @x + 0x4",
		},
		evalCase{
			name:   "Undef",
			expr:   smallIntAddExpr{binIntExpr{undefValue(i32), runtime(i32, 4), i32}},
			expect: "i32 undef",
		},
		evalCase{
			name:   "SimplifyZero",
			expr:   smallIntAddExpr{binIntExpr{i32.zero(), runtime(i32, 1), i32}},
			expect: "i32 %1",
		},
	)
}

func TestEvalSmallMul(t *testing.T) {
	t.Parallel()

	testEval(t,
		evalCase{
			name:   "Const",
			expr:   smallIntMulExpr{binIntExpr{smallIntValue(i32, 5), smallIntValue(i32, 6), i32}},
			expect: "i32 30",
		},
		evalCase{
			name:   "Overflow",
			expr:   smallIntMulExpr{binIntExpr{smallIntValue(i8, 17), smallIntValue(i8, 17), i8}},
			expect: "i8 33",
		},
		evalCase{
			name:   "Short",
			expr:   smallIntMulExpr{binIntExpr{smallIntValue(i8, 0), runtime(i8, 2), i8}},
			expect: "i8 0",
		},
		evalCase{
			name:   "SimplifyOne",
			expr:   smallIntMulExpr{binIntExpr{smallIntValue(i64, 1), runtime(i64, 5), i64}},
			expect: "i64 %5",
		},
		evalCase{
			name:   "Undef",
			expr:   smallIntMulExpr{binIntExpr{undefValue(i8), smallIntValue(i8, 3), i8}},
			expect: "i8 undef",
		},
		evalCase{
			name: "PartialUndef",
			expr: smallIntMulExpr{binIntExpr{undefValue(i8), smallIntValue(i8, 128), i8}},
		},
		evalCase{
			name:   "ShortUndef",
			expr:   smallIntMulExpr{binIntExpr{undefValue(i8), smallIntValue(i8, 0), i8}},
			expect: "i8 0",
		},
	)
}

func testEval(t *testing.T, cases ...evalCase) {
	t.Helper()

	for _, c := range cases {
		t.Run(c.name, c.run)
		if r, ok := c.expr.(symmetricExpr); ok {
			reversed := evalCase{c.name + "Reversed", r.reverse(), c.expect}
			t.Run(reversed.name, reversed.run)
		}
	}
}

type evalCase struct {
	name   string
	expr   expr
	expect string
}

type symmetricExpr interface {
	expr
	reverse() expr
}

func (expr smallIntAddExpr) reverse() expr {
	return smallIntAddExpr{binIntExpr{expr.y, expr.x, expr.ty}}
}

func (expr smallIntMulExpr) reverse() expr {
	return smallIntMulExpr{binIntExpr{expr.y, expr.x, expr.ty}}
}

func (c evalCase) run(t *testing.T) {
	t.Parallel()

	v, err := c.expr.eval()
	switch err {
	case nil:
		if c.expect == "" {
			t.Errorf("expected failure but got %s", v.String())
			return
		}

		str := v.String()
		if str != c.expect {
			t.Errorf("expected %s but got %s", c.expect, str)
		}

	case errRuntime:
		if c.expect != "" {
			t.Errorf("expected %s but deferred to runtime", c.expect)
		}

	default:
		t.Errorf("unexpected error: %s", err.Error())
	}
}
