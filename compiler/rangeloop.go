package compiler

import (
	"go/constant"
	"go/token"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// rangeBound returns the index and bound of a range loop body block.
// go/ssa range loops never produce a negative index, so k < n holds as unsigned.
func rangeBound(block *ssa.BasicBlock) (ssa.Value, ssa.Value, bool) {
	if block.Comment != "rangeindex.body" && block.Comment != "rangeint.body" {
		return nil, nil, false
	}
	var k, n ssa.Value
	for _, pred := range block.Preds {
		ifInstr, ok := pred.Instrs[len(pred.Instrs)-1].(*ssa.If)
		if !ok || pred.Succs[0] != block {
			return nil, nil, false
		}
		cond, ok := ifInstr.Cond.(*ssa.BinOp)
		if !ok || cond.Op != token.LSS {
			return nil, nil, false
		}
		if n == nil {
			n = cond.Y
		} else if !sameValue(n, cond.Y) {
			return nil, nil, false
		}
		if len(block.Preds) == 1 {
			k = cond.X
		}
	}
	if k == nil {
		for _, instr := range block.Instrs {
			phi, ok := instr.(*ssa.Phi)
			if !ok {
				break
			}
			if phi.Comment == "rangeint.iter" {
				k = phi
			}
		}
	}
	if k == nil || n == nil || !usedInBody(k) {
		return nil, nil, false
	}
	return k, n, true
}

// usedInBody reports whether k has uses other than the loop phi, increment and test.
func usedInBody(k ssa.Value) bool {
	for _, ref := range *k.Referrers() {
		switch ref := ref.(type) {
		case *ssa.Phi:
			continue
		case *ssa.BinOp:
			if ref.Op == token.ADD || ref.Op == token.LSS {
				continue
			}
		}
		return true
	}
	return false
}

func sameValue(a, b ssa.Value) bool {
	if a == b {
		return true
	}
	ca, okA := a.(*ssa.Const)
	cb, okB := b.(*ssa.Const)
	return okA && okB && ca.Value != nil && cb.Value != nil && constant.Compare(ca.Value, token.EQL, cb.Value)
}

// createRangeAssume tells LLVM that the index of a range loop body is below its bound.
func (b *builder) createRangeAssume(k, n ssa.Value) {
	kv := b.getValue(k, k.Pos())
	nv := b.getValue(n, k.Pos())
	if kv.Type() != nv.Type() {
		return
	}
	cmp := b.CreateICmp(llvm.IntULT, kv, nv, "")
	fnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{b.ctx.Int1Type()}, false)
	fn := b.mod.NamedFunction("llvm.assume")
	if fn.IsNil() {
		fn = llvm.AddFunction(b.mod, "llvm.assume", fnType)
	}
	b.CreateCall(fnType, fn, []llvm.Value{cmp}, "")
}
