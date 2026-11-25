package compiler

// Lower loops that zero every element of a slice to a memset, like gc's
// arrayRangeClear in cmd/compile/internal/walk/range.go.

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

type clearLoop struct {
	header       *ssa.BasicBlock
	body         *ssa.BasicBlock // equals header for range over int
	exit         *ssa.BasicBlock
	slice        ssa.Value
	headerInstrs []ssa.Instruction // compute slice in a three-clause header
	store        *ssa.Store
}

// findClearLoops returns the slice zeroing loops in fn, keyed by both header
// and body block.
func findClearLoops(fn *ssa.Function) map[*ssa.BasicBlock]*clearLoop {
	var loops map[*ssa.BasicBlock]*clearLoop
	for _, block := range fn.Blocks {
		var loop *clearLoop
		switch block.Comment {
		case "rangeindex.loop":
			loop = matchRangeIndexClear(block)
		case "rangeint.body":
			loop = matchRangeIntClear(block)
		case "for.loop":
			loop = matchForClear(block)
		}
		if loop == nil || !loop.valid() {
			continue
		}
		if loops == nil {
			loops = make(map[*ssa.BasicBlock]*clearLoop)
		}
		if loops[loop.header] != nil || loops[loop.body] != nil {
			continue
		}
		loops[loop.header] = loop
		loops[loop.body] = loop
	}
	return loops
}

// matchRangeIndexClear matches for i := range X { X[i] = 0 }.
func matchRangeIndexClear(header *ssa.BasicBlock) *clearLoop {
	body, exit, ok := headerShape(header)
	if !ok || body == header {
		return nil
	}
	preIndex, bodyIndex := predIndexes(header, body)
	if preIndex < 0 {
		return nil
	}
	if len(body.Preds) != 1 || len(body.Succs) != 1 || body.Succs[0] != header {
		return nil
	}

	instrs := nonDebugInstrs(header)
	if len(instrs) != 4 {
		return nil
	}
	phi, ok := instrs[0].(*ssa.Phi)
	if !ok || !isIntConst(phi.Edges[preIndex], -1) {
		return nil
	}
	next, ok := instrs[1].(*ssa.BinOp)
	if !ok || !isIncrement(next, phi) || phi.Edges[bodyIndex] != next {
		return nil
	}
	cond, ok := instrs[2].(*ssa.BinOp)
	if !ok || cond.Op != token.LSS || cond.X != next {
		return nil
	}
	if ifInstr, ok := instrs[3].(*ssa.If); !ok || ifInstr.Cond != cond {
		return nil
	}

	pre := header.Preds[preIndex]
	length, slice := lenCall(cond.Y)
	if slice == nil || length.Block() != pre || !preheaderSafe(pre, slice, length) {
		return nil
	}

	store := matchBodyStore(body, slice, next)
	if store == nil {
		return nil
	}
	return &clearLoop{
		header: header,
		body:   body,
		exit:   exit,
		slice:  slice,
		store:  store,
	}
}

// matchRangeIntClear matches for i := range len(X) { X[i] = 0 }.
func matchRangeIntClear(body *ssa.BasicBlock) *clearLoop {
	_, exit, ok := headerShape(body)
	if !ok || body.Succs[0] != body {
		return nil
	}
	preIndex, bodyIndex := predIndexes(body, body)
	if preIndex < 0 {
		return nil
	}
	pre := body.Preds[preIndex]

	instrs := nonDebugInstrs(body)
	if len(instrs) < 6 {
		return nil
	}
	phi, ok := instrs[0].(*ssa.Phi)
	if !ok || !isIntConst(phi.Edges[preIndex], 0) {
		return nil
	}
	n := len(instrs)
	ifInstr, ok := instrs[n-1].(*ssa.If)
	if !ok {
		return nil
	}
	cond, ok := instrs[n-2].(*ssa.BinOp)
	if !ok || cond.Op != token.LSS || ifInstr.Cond != cond {
		return nil
	}
	next, ok := instrs[n-3].(*ssa.BinOp)
	if !ok || !isIncrement(next, phi) || phi.Edges[bodyIndex] != next || cond.X != next {
		return nil
	}

	// The preheader must enter the loop only if 0 < len(X).
	preIf, ok := pre.Instrs[len(pre.Instrs)-1].(*ssa.If)
	if !ok || pre.Succs[0] != body || pre.Succs[1] != exit {
		return nil
	}
	guard, ok := preIf.Cond.(*ssa.BinOp)
	if !ok || guard.Op != token.LSS || !isIntConst(guard.X, 0) || guard.Y != cond.Y {
		return nil
	}

	length, slice := lenCall(cond.Y)
	if slice == nil || length.Block() != pre || !preheaderSafe(pre, slice, length, guard) {
		return nil
	}

	store := matchStoreInstrs(instrs[1:n-3], body, slice, phi)
	if store == nil {
		return nil
	}
	return &clearLoop{
		header: body,
		body:   body,
		exit:   exit,
		slice:  slice,
		store:  store,
	}
}

// matchForClear matches for i := 0; i < len(X); i++ { X[i] = 0 }.
func matchForClear(header *ssa.BasicBlock) *clearLoop {
	body, exit, ok := headerShape(header)
	if !ok || body == header {
		return nil
	}
	preIndex, bodyIndex := predIndexes(header, body)
	if preIndex < 0 {
		return nil
	}
	if len(body.Preds) != 1 || len(body.Succs) != 1 || body.Succs[0] != header {
		return nil
	}

	instrs := nonDebugInstrs(header)
	if len(instrs) < 4 {
		return nil
	}
	phi, ok := instrs[0].(*ssa.Phi)
	if !ok || !isIntConst(phi.Edges[preIndex], 0) {
		return nil
	}
	n := len(instrs)
	ifInstr, ok := instrs[n-1].(*ssa.If)
	if !ok {
		return nil
	}
	cond, ok := instrs[n-2].(*ssa.BinOp)
	if !ok || cond.Op != token.LSS || cond.X != phi || ifInstr.Cond != cond {
		return nil
	}
	length, slice := lenCall(cond.Y)
	if slice == nil || length != instrs[n-3] {
		return nil
	}

	chain, root := chainInstrs(slice)
	if !sameInstrs(chain, instrs[1:n-3]) || definedIn(root, header) || definedIn(root, body) {
		return nil
	}

	bodyInstrs := nonDebugInstrs(body)
	m := len(bodyInstrs)
	if m < 4 {
		return nil
	}
	if _, ok := bodyInstrs[m-1].(*ssa.Jump); !ok {
		return nil
	}
	next, ok := bodyInstrs[m-2].(*ssa.BinOp)
	if !ok || !isIncrement(next, phi) || phi.Edges[bodyIndex] != next {
		return nil
	}
	store := matchStoreInstrs(bodyInstrs[:m-2], body, slice, phi)
	if store == nil {
		return nil
	}
	return &clearLoop{
		header:       header,
		body:         body,
		exit:         exit,
		slice:        slice,
		headerInstrs: chain,
		store:        store,
	}
}

// valid checks that the store writes a zero value of nonzero size and that no
// loop value is used after the loop.
func (loop *clearLoop) valid() bool {
	sliceType, ok := loop.slice.Type().Underlying().(*types.Slice)
	if !ok || isZeroSize(sliceType.Elem()) {
		return false
	}
	if !isZeroConst(loop.store.Val) {
		return false
	}
	if loop.exit == loop.header || loop.exit == loop.body {
		return false
	}
	inLoop := func(b *ssa.BasicBlock) bool {
		return b == loop.header || b == loop.body
	}
	for _, block := range []*ssa.BasicBlock{loop.header, loop.body} {
		for _, instr := range block.Instrs {
			value, ok := instr.(ssa.Value)
			if !ok {
				continue
			}
			refs := value.Referrers()
			if refs == nil {
				continue
			}
			for _, ref := range *refs {
				if !inLoop(ref.Block()) {
					// These are still emitted.
					if loop.isHeaderInstr(instr) {
						continue
					}
					return false
				}
			}
		}
	}
	return true
}

func (loop *clearLoop) isHeaderInstr(instr ssa.Instruction) bool {
	for _, h := range loop.headerInstrs {
		if h == instr {
			return true
		}
	}
	return false
}

// headerShape checks that block has two predecessors and ends in an If with
// two distinct successors.
func headerShape(block *ssa.BasicBlock) (body, exit *ssa.BasicBlock, ok bool) {
	if len(block.Instrs) == 0 || len(block.Preds) != 2 || len(block.Succs) != 2 {
		return nil, nil, false
	}
	if _, ok := block.Instrs[len(block.Instrs)-1].(*ssa.If); !ok {
		return nil, nil, false
	}
	if block.Succs[0] == block.Succs[1] {
		return nil, nil, false
	}
	return block.Succs[0], block.Succs[1], true
}

// predIndexes returns the indexes of the preheader and latch edges in
// header.Preds, or -1, -1.
func predIndexes(header, latch *ssa.BasicBlock) (preIndex, latchIndex int) {
	switch {
	case header.Preds[0] == latch && header.Preds[1] != latch:
		return 1, 0
	case header.Preds[1] == latch && header.Preds[0] != latch:
		return 0, 1
	}
	return -1, -1
}

func nonDebugInstrs(block *ssa.BasicBlock) []ssa.Instruction {
	instrs := make([]ssa.Instruction, 0, len(block.Instrs))
	for _, instr := range block.Instrs {
		if _, ok := instr.(*ssa.DebugRef); ok {
			continue
		}
		instrs = append(instrs, instr)
	}
	return instrs
}

func isIntConst(v ssa.Value, want int64) bool {
	c, ok := v.(*ssa.Const)
	if !ok || c.Value == nil || c.Value.Kind() != constant.Int {
		return false
	}
	if basic, ok := c.Type().Underlying().(*types.Basic); !ok || basic.Info()&types.IsInteger == 0 {
		return false
	}
	got, exact := constant.Int64Val(c.Value)
	return exact && got == want
}

// isIncrement reports whether next is phi + 1.
func isIncrement(next *ssa.BinOp, phi *ssa.Phi) bool {
	return next.Op == token.ADD && next.X == phi && isIntConst(next.Y, 1)
}

// lenCall returns the call and X if v is len(X) for a slice X.
func lenCall(v ssa.Value) (*ssa.Call, ssa.Value) {
	call, ok := v.(*ssa.Call)
	if !ok {
		return nil, nil
	}
	builtin, ok := call.Call.Value.(*ssa.Builtin)
	if !ok || builtin.Name() != "len" || len(call.Call.Args) != 1 {
		return nil, nil
	}
	arg := call.Call.Args[0]
	if _, ok := arg.Type().Underlying().(*types.Slice); !ok {
		return nil, nil
	}
	return call, arg
}

// definedIn reports whether v is an instruction in block.
func definedIn(v ssa.Value, block *ssa.BasicBlock) bool {
	instr, ok := v.(ssa.Instruction)
	return ok && instr.Block() == block
}

// preheaderSafe checks that nothing in pre writes memory between computing
// slice and entering the loop, so the reload in the body sees the same slice.
func preheaderSafe(pre *ssa.BasicBlock, slice ssa.Value, extra ...ssa.Instruction) bool {
	chain, _ := chainInstrs(slice)
	allowed := make(map[ssa.Instruction]bool)
	for _, instr := range chain {
		if instr.Block() != pre {
			return false
		}
		allowed[instr] = true
	}
	for _, instr := range extra {
		allowed[instr] = true
	}
	start := len(pre.Instrs) - 1
	for i, instr := range pre.Instrs {
		if allowed[instr] {
			start = i
			break
		}
	}
	for _, instr := range pre.Instrs[start : len(pre.Instrs)-1] {
		if _, ok := instr.(*ssa.DebugRef); ok {
			continue
		}
		if !allowed[instr] {
			return false
		}
	}
	return true
}

// chainInstrs returns the loads and field addresses that compute v, root first,
// and the root value.
func chainInstrs(v ssa.Value) (chain []ssa.Instruction, root ssa.Value) {
	for {
		switch x := v.(type) {
		case *ssa.UnOp:
			if x.Op != token.MUL || x.CommaOk {
				return reversed(chain), v
			}
			chain = append(chain, x)
			v = x.X
		case *ssa.FieldAddr:
			chain = append(chain, x)
			v = x.X
		default:
			return reversed(chain), v
		}
	}
}

func reversed(s []ssa.Instruction) []ssa.Instruction {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func sameInstrs(a, b []ssa.Instruction) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// sameSafeExpr reports whether reload b in body computes the same value as a,
// like SameSafeExpr in cmd/compile/internal/ir/expr.go.
func sameSafeExpr(a, b ssa.Value, body *ssa.BasicBlock) bool {
	if a == b {
		return true
	}
	if !definedIn(b, body) {
		return false
	}
	switch a := a.(type) {
	case *ssa.UnOp:
		b, ok := b.(*ssa.UnOp)
		return ok && a.Op == token.MUL && b.Op == token.MUL && !a.CommaOk && !b.CommaOk &&
			types.Identical(a.Type(), b.Type()) && sameSafeExpr(a.X, b.X, body)
	case *ssa.FieldAddr:
		b, ok := b.(*ssa.FieldAddr)
		return ok && a.Field == b.Field && types.Identical(a.X.Type(), b.X.Type()) && sameSafeExpr(a.X, b.X, body)
	}
	return false
}

// matchBodyStore matches a body of only the slice reload, the store and a jump.
func matchBodyStore(body *ssa.BasicBlock, slice, index ssa.Value) *ssa.Store {
	instrs := nonDebugInstrs(body)
	if len(instrs) < 3 {
		return nil
	}
	if _, ok := instrs[len(instrs)-1].(*ssa.Jump); !ok {
		return nil
	}
	return matchStoreInstrs(instrs[:len(instrs)-1], body, slice, index)
}

// matchStoreInstrs matches instrs as the reload of slice in body followed by
// a zero store to &X[index].
func matchStoreInstrs(instrs []ssa.Instruction, body *ssa.BasicBlock, slice, index ssa.Value) *ssa.Store {
	if len(instrs) < 2 {
		return nil
	}
	store, ok := instrs[len(instrs)-1].(*ssa.Store)
	if !ok {
		return nil
	}
	addr, ok := instrs[len(instrs)-2].(*ssa.IndexAddr)
	if !ok || store.Addr != addr || addr.Index != index {
		return nil
	}
	if _, ok := addr.X.Type().Underlying().(*types.Slice); !ok {
		return nil
	}
	if !sameSafeExpr(slice, addr.X, body) {
		return nil
	}
	chain, _ := chainInstrs(addr.X)
	var inBody []ssa.Instruction
	for _, instr := range chain {
		if instr.Block() == body {
			inBody = append(inBody, instr)
		}
	}
	if !sameInstrs(inBody, instrs[:len(instrs)-2]) {
		return nil
	}
	if refs := addr.Referrers(); refs == nil || len(*refs) != 1 {
		return nil
	}
	return store
}

// isZeroConst reports whether v is the constant zero value of its type, which
// is all zero bits in TinyGo.
func isZeroConst(v ssa.Value) bool {
	c, ok := v.(*ssa.Const)
	if !ok {
		return false
	}
	if _, ok := c.Type().(*types.TypeParam); ok {
		return false
	}
	if c.Value == nil {
		return true
	}
	basic, ok := c.Type().Underlying().(*types.Basic)
	if !ok {
		return false
	}
	info := basic.Info()
	switch {
	case info&types.IsBoolean != 0:
		return !constant.BoolVal(c.Value)
	case info&types.IsString != 0:
		return constant.StringVal(c.Value) == ""
	case info&types.IsNumeric != 0:
		return constant.Sign(c.Value) == 0
	}
	return false
}

// isZeroSize reports whether values of typ take no memory.
func isZeroSize(typ types.Type) bool {
	switch typ := typ.Underlying().(type) {
	case *types.Struct:
		for i := 0; i < typ.NumFields(); i++ {
			if !isZeroSize(typ.Field(i).Type()) {
				return false
			}
		}
		return true
	case *types.Array:
		return typ.Len() == 0 || isZeroSize(typ.Elem())
	}
	return false
}

// createClearLoop emits the clear in the loop header. Like gc, it assumes the
// stores never change X, which testdata/clearloop.go shows can matter.
func (b *builder) createClearLoop(loop *clearLoop, block *ssa.BasicBlock) {
	if block != loop.header {
		if b.dumpSSA {
			fmt.Printf("\t(clear loop body)\n")
		}
		b.CreateUnreachable()
		return
	}
	for _, instr := range loop.headerInstrs {
		if b.dumpSSA {
			fmt.Printf("\t%s = %s\n", instr.(ssa.Value).Name(), instr.String())
		}
		b.createInstruction(instr)
	}
	if b.dumpSSA {
		fmt.Printf("\tclear(%s) (clear loop)\n", loop.slice.Name())
	}
	pos := loop.store.Pos()
	if b.Debug {
		b.setDebugLocation(pos)
	}
	value := b.getValue(loop.slice, pos)
	elemType := loop.slice.Type().Underlying().(*types.Slice).Elem()
	b.createSliceClear(value, elemType)
	b.CreateBr(b.blockInfo[loop.exit.Index].entry)
}
