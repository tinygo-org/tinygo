package interp

import (
	"errors"
	"fmt"
	"go/token"
	"math/bits"
	"path/filepath"
	"strings"

	"tinygo.org/x/go-llvm"
)

func compile(c *constParser, sig signature, fn llvm.Value) (*runnable, error) {
	if fn.IsDeclaration() {
		// This is not known.
		return nil, errRuntime
	}

	// List the blocks in reverse postorder.
	entry := fn.FirstBasicBlock()
	blocks := listBlocks(entry, sig.ret == nil, make(map[llvm.BasicBlock]struct{}), nil)
	for i := range blocks[:len(blocks)/2] {
		blocks[i], blocks[len(blocks)-i-1] = blocks[len(blocks)-i-1], blocks[i]
	}
	order := make(map[llvm.BasicBlock]uint, len(blocks))
	for i := range blocks {
		order[blocks[i].b] = uint(i)
	}

	// Compute predecessors.
	preds := make([][]uint, len(blocks))
	{
		dedup := make(map[uint]struct{})
		for i := range blocks {
			to := blocks[i].to
			if len(to) > 0 {
				for _, dst := range blocks[i].to {
					j := order[dst]
					if _, ok := dedup[j]; ok {
						continue
					}

					preds[j] = append(preds[j], uint(i))
				}
				for j := range dedup {
					delete(dedup, j)
				}
			}
		}
	}
	if len(preds[0]) != 0 {
		return nil, errors.New("malformed IR: branch to entry block")
	}

	// Compute immediate dominators.
	// This is a tweaked version of the "engineered algorithm" from https://www.cs.rice.edu/~keith/EMBED/dom.pdf.
	// The main tweaks are:
	// - This uses reverse postorder numbering to match the loop direction instead of postorder numbering (simpler).
	// - This initializes doms with the first listed predecessor instead of a sentinel (simpler, and usually allows it to terminate in one pass).
	// ("A Simple, Fast Dominance Algorithm", Keith D. Cooper, Timothy J. Harvey, and Ken Kennedy)
	{
		doms := make([]uint, len(blocks))
		for i, preds := range preds {
			if i == 0 {
				// The entry block is not dominated by anything.
				doms[i] = 0
			} else {
				// Choose the first listed predecessor.
				// This ensures that the path down the tree always leads to the entry block.
				doms[i] = preds[0]
			}
		}
		for {
			var changed bool
			for b, dom := range doms {
				for _, p := range preds[b] {
					// Update dom to the common dominator with p.
					for p != dom {
						for p > dom {
							p = doms[p]
						}
						for dom > p {
							dom = doms[dom]
							changed = true
						}
					}
				}
				doms[b] = dom
			}
			if !changed {
				break
			}
		}
		for i := range blocks {
			if i != 0 {
				blocks[i].idom = blocks[doms[i]].b
			}
		}
	}

	bmap := make(map[llvm.BasicBlock]blockInfo, len(blocks))
	for _, b := range blocks {
		bmap[b.b] = b
	}
	builder := builder{
		sig: sig,
	}
	ivals := make(map[llvm.Value]value)
	for i, param := range fn.Params() {
		ivals[param] = builder.pushValue(sig.args[i].t)
	}
	comp := compiler{
		sig:         sig,
		blocks:      bmap,
		constParser: c,
		b:           builder,
		ivals:       ivals,
		labels:      make(map[llvm.BasicBlock]uint),
		heights:     make(map[llvm.BasicBlock]uint),
		pending:     make(map[llvm.BasicBlock][]llvm.BasicBlock),
		built:       make(map[llvm.BasicBlock]struct{}),
	}
	err := comp.compileBlock(entry)
	if err != nil {
		return nil, err
	}

	return comp.b.finish(), nil
}

// listBlocks in postorder, appending them to dst.
func listBlocks(b llvm.BasicBlock, void bool, visited map[llvm.BasicBlock]struct{}, dst []blockInfo) []blockInfo {
	// Parse the terminator.
	term := b.LastInstruction()
	termKind := termRevert
	var to []llvm.BasicBlock
	var ret llvm.Value
	switch term.InstructionOpcode() {
	case llvm.Ret:
		termKind = termRet
		if !void {
			ret = term.Operand(0)
		}
	case llvm.Br:
		switch term.OperandsCount() {
		case 1:
			termKind, to = termBr, []llvm.BasicBlock{term.Operand(0).AsBasicBlock()}
		case 3:
			termKind, to = termCondBr, []llvm.BasicBlock{term.Operand(1).AsBasicBlock(), term.Operand(2).AsBasicBlock()}
		}
	case llvm.Switch:
		termKind = termSwitch
		to = make([]llvm.BasicBlock, term.OperandsCount()/2)
		for i := range to {
			to[i] = term.Operand(2*i + 1).AsBasicBlock()
		}
	case llvm.Unreachable:
		termKind = termUB
	}

	// Parse phi nodes.
	var phis []llPhi
	inst := b.FirstInstruction()
	for ; !inst.IsAPHINode().IsNil(); inst = llvm.NextInstruction(inst) {
		n := inst.IncomingCount()
		values := make(map[llvm.BasicBlock]llvm.Value, n)
		for i := 0; i < n; i++ {
			values[inst.IncomingBlock(i)] = inst.IncomingValue(i)
		}
		phis = append(phis, llPhi{inst, values})
	}

	// Mark this block as visited.
	visited[b] = struct{}{}

	// List destination blocks.
	for _, to := range to {
		if _, ok := visited[to]; ok {
			// This destination has already been visited.
			continue
		}
		dst = listBlocks(to, void, visited, dst)
	}

	// Append this block.
	return append(dst, blockInfo{
		b: b,

		phis: phis,

		first: inst,

		term:     term,
		termKind: termKind,
		to:       to,
		ret:      ret,
	})
}

type compiler struct {
	sig signature

	blocks map[llvm.BasicBlock]blockInfo

	*constParser

	b builder

	ivals  map[llvm.Value]value
	istack []llvm.Value

	labels  map[llvm.BasicBlock]uint
	heights map[llvm.BasicBlock]uint
	pending map[llvm.BasicBlock][]llvm.BasicBlock
	built   map[llvm.BasicBlock]struct{}
}

func (c *compiler) compileBlock(b llvm.BasicBlock) error {
	if _, ok := c.built[b]; ok {
		panic("already built " + b.AsValue().Name())
	}
	c.built[b] = struct{}{}

	defer func(i int) {
		for _, v := range c.istack[i:] {
			delete(c.ivals, v)
		}
		c.istack = c.istack[:i]
	}(len(c.istack))

	// Map phi values.
	info := c.blocks[b]
	expectHeight := c.b.stackHeight
	if idom := c.blocks[b].idom; !idom.IsNil() {
		c.b.stackHeight = c.heights[idom]
	}
	for _, phi := range info.phis {
		c.istack = append(c.istack, phi.phi)
		t, err := c.typ(phi.phi.Type())
		switch err.(type) {
		case nil:
		case errRevert:
			c.b.insertInst(&failInst{err, phi.phi.InstructionDebugLoc()})
			return nil
		default:
			return err
		}
		c.ivals[phi.phi] = c.b.pushValue(t)
	}
	if c.b.stackHeight != expectHeight {
		panic("height mismatch")
	}

	// Compile block body.
	for inst := info.first; inst != info.term; inst = llvm.NextInstruction(inst) {
		c.istack = append(c.istack, inst)
		switch err := c.emitInst(inst).(type) {
		case nil:
		case errRevert:
			c.b.insertInst(&failInst{err, inst.InstructionDebugLoc()})
			return nil
		default:
			switch err {
			case errUB:
				c.b.insertInst(&failInst{err, inst.InstructionDebugLoc()})
				return nil
			default:
				return err
			}
		}
	}

	// Compile terminator.
	c.heights[b] = c.b.stackHeight
	switch err := c.emitTerminator(info.termKind, info.term, b, info.to).(type) {
	case nil:
	case errRevert:
		c.b.insertInst(&failInst{err, info.term.InstructionDebugLoc()})
	default:
		switch err {
		case errUB:
			c.b.insertInst(&failInst{err, info.term.InstructionDebugLoc()})
		default:
			return err
		}
	}

	// Compile pending blocks.
	for {
		// Pop the next pending block.
		pending := c.pending[b]
		if len(pending) == 0 {
			break
		}
		next := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		c.pending[b] = pending

		// Compile it.
		if _, ok := c.built[next]; ok {
			// This block has already been compiled.
			continue
		}
		c.b.startLabel(c.labels[next])
		err := c.compileBlock(next)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *compiler) emitInst(inst llvm.Value) error {
	dbg := inst.InstructionDebugLoc()
	op := inst.InstructionOpcode()
	switch op {
	case llvm.Add, llvm.Sub, llvm.Mul, llvm.UDiv, llvm.SDiv,
		llvm.Shl, llvm.LShr, llvm.AShr,
		llvm.And, llvm.Or, llvm.Xor,
		llvm.GetElementPtr, llvm.SExt, llvm.ICmp, llvm.Select:
		expr, err := parseExpr(op, inst, c)
		if err != nil {
			return err
		}
		v, err := expr.create(&c.b, dbg)
		if err != nil {
			return err
		}
		c.ivals[inst] = v
		return nil

	case llvm.ZExt, llvm.Trunc, llvm.BitCast, llvm.IntToPtr, llvm.PtrToInt:
		to, err := c.typ(inst.Type())
		if err != nil {
			return err
		}
		from, err := c.value(inst.Operand(0))
		if err != nil {
			return err
		}
		c.ivals[inst] = cast(to.(nonAggTyp), from)
		return nil

	case llvm.Load:
		from := inst.Operand(0)
		fromVal, err := c.value(from)
		if err != nil {
			return err
		}
		rawResTy := inst.Type()
		resTy, err := c.typ(rawResTy)
		if err != nil {
			return err
		}
		align := inst.Alignment()
		if align == 0 {
			// Assume ABI type alignment?
			align = c.td.ABITypeAlignment(rawResTy)
		}
		c.ivals[inst] = c.b.insertInst(&loadInst{
			from:       fromVal,
			ty:         resTy,
			alignScale: uint(bits.TrailingZeros(uint(align))),
			volatile:   inst.IsVolatile(),
			order:      inst.Ordering(),
			rawTy:      from.Type(),
			dbg:        dbg,
		})
		return nil

	case llvm.Store:
		toVal := inst.Operand(1)
		to, err := c.value(toVal)
		if err != nil {
			return err
		}
		vVal := inst.Operand(0)
		v, err := c.value(vVal)
		if err != nil {
			return err
		}
		vt := vVal.Type()
		align := inst.Alignment()
		if align == 0 {
			// Assume ABI type alignment?
			align = c.td.ABITypeAlignment(vt)
		}
		c.b.insertInst(&storeInst{
			to:         to,
			v:          v,
			alignScale: uint(bits.TrailingZeros(uint(align))),
			volatile:   inst.IsVolatile(),
			order:      inst.Ordering(),
			rawTy:      vt,
			dbg:        dbg,
		})
		return nil

	case llvm.Call:
		sig, err := c.parseCallSignature(inst)
		if err != nil {
			return err
		}
		called, err := c.value(inst.CalledValue())
		if err != nil {
			return err
		}
		args := make([]value, len(sig.args))[:0]
		for i := range sig.args {
			v, err := c.value(inst.Operand(i))
			if err != nil {
				return err
			}
			args = unpack(v, args)
		}
		call := callInst{
			called: called,
			args:   args,
			sig:    sig,
			dbg:    dbg,
		}
		i, err := parseAsIntrinsic(c.constParser, call)
		switch err {
		case nil:
		case errRuntime:
			i = &call
		default:
			return err
		}
		if i != nil {
			v := c.b.insertInst(i)
			if sig.ret != nil {
				c.ivals[inst] = v
			}
		}
		return nil

	case llvm.ExtractValue:
		v, err := parseExtractValue(inst, c)
		if err != nil {
			return err
		}
		c.ivals[inst] = v
		return nil

	case llvm.InsertValue:
		v, err := parseInsertValue(inst, c)
		if err != nil {
			return err
		}
		c.ivals[inst] = v
		return nil

	case llvm.Alloca:
		t := inst.Type()
		ptrTy, err := c.typ(t)
		if err != nil {
			return err
		}
		ty, err := c.typ(t.ElementType())
		if err != nil {
			return err
		}
		n, err := c.value(inst.Operand(0))
		if err != nil {
			return err
		}
		alignScale := uint(bits.TrailingZeros(uint(inst.Alignment())))
		c.ivals[inst] = c.b.insertInst(&allocaInst{
			ty:         ty,
			n:          n,
			ptrTy:      ptrTy.(ptrType),
			alignScale: alignScale,
			dbg:        dbg,
		})
		return nil
	}

	return todo("emit instruction with opcode " + opString(op))
}

func (c *compiler) emitTerminator(kind termKind, inst llvm.Value, from llvm.BasicBlock, to []llvm.BasicBlock) error {
	dbg := inst.InstructionDebugLoc()
	switch kind {
	case termRet:
		var v value
		if c.sig.ret != nil {
			var err error
			v, err = c.value(c.blocks[from].ret)
			if err != nil {
				return err
			}
		}
		c.b.insertInst(&brInst{c.retEdge(v), dbg})
		return nil
	case termBr:
		return c.doMkBr(from, to[0], dbg)
	case termCondBr:
		cond, err := c.value(inst.Operand(0))
		if err != nil {
			return err
		}
		if _, ok := cond.val.(smallInt); ok {
			return c.doMkBr(from, to[cond.raw], dbg)
		}
		elseEdge, err := c.mkEdge(from, to[0])
		if err != nil {
			return err
		}
		c.b.insertInst(&switchInst{cond, map[uint64]brEdge{0: elseEdge}, dbg})
		return c.doMkBr(from, to[1], dbg)
	case termSwitch:
		v, err := c.value(inst.Operand(0))
		if err != nil {
			return err
		}
		if v.typ().(iType) > i64 {
			return todo("switch on a value wider than 64 bits")
		}
		if len(to) != 1 {
			if _, ok := v.val.(smallInt); ok {
				for i, to := range to[1:] {
					j := inst.Operand(2*i + 2).ZExtValue()
					if j == v.raw {
						return c.doMkBr(from, to, dbg)
					}
				}
			} else {
				edges := make(map[uint64]brEdge, len(to)-1)
				for i, to := range to[1:] {
					j := inst.Operand(2*i + 2).ZExtValue()
					edge, err := c.mkEdge(from, to)
					if err != nil {
						return err
					}
					edges[j] = edge
				}
				c.b.insertInst(&switchInst{v, edges, dbg})
			}
		}
		return c.doMkBr(from, to[0], dbg)
	case termUB:
		c.b.insertInst(&failInst{errUB, dbg})
		return nil
	default:
		return todo("unknown terminator")
	}
}

func (c *compiler) doMkBr(from, to llvm.BasicBlock, dbg llvm.Metadata) error {
	edge, err := c.mkEdge(from, to)
	if err != nil {
		return err
	}
	if len(edge.phis) == 0 && edge.to != ^uint(0) {
		if _, ok := c.built[to]; !ok && c.dominated(from, c.blocks[to].idom) {
			// Compile the block right here instead of creating a branch.
			c.b.startLabel(edge.to)
			return c.compileBlock(to)
		}
	}

	c.b.insertInst(&brInst{edge, dbg})
	return nil
}

func (c *compiler) dominated(block, by llvm.BasicBlock) bool {
	for !block.IsNil() {
		if block == by {
			return true
		}
		block = c.blocks[block].idom
	}

	return false
}

func (c *compiler) mkEdge(from, to llvm.BasicBlock) (brEdge, error) {
	toInfo := c.blocks[to]
	if toInfo.termKind == termBr && len(toInfo.phis) == 0 && toInfo.idom == from && toInfo.first == toInfo.term {
		// When compiling a lookup switch, a branch through an empty block may be used to change a phi value in LLVM IR.
		// Since interp stores phi values inside of the branch edge instead of the target block, this can be merged.
		from, to = to, toInfo.to[0]
		toInfo = c.blocks[to]
	}
	if toInfo.termKind == termRet && toInfo.first == toInfo.term {
		// Combine a branch to a return into a return edge.
		for _, phi := range toInfo.phis {
			if phi.phi == toInfo.ret {
				v, ok := phi.value[from]
				if !ok {
					return brEdge{}, errors.New("missing phi value")
				}
				val, err := c.value(v)
				if err != nil {
					return brEdge{}, err
				}
				return c.retEdge(val), nil
			}
		}

		// If this returns an error, the error should not actually be returned from mkEdge.
		// Instead, this will transform into a branch to a thing that produces the error.
		var v value
		var err error
		if c.sig.ret != nil {
			v, err = c.value(toInfo.ret)
		}
		if err == nil {
			return c.retEdge(v), nil
		}
	}

	// Look up phi values.
	phiVals := make([]value, len(toInfo.phis))[:0]
	for _, phi := range toInfo.phis {
		v, ok := phi.value[from]
		if !ok {
			return brEdge{}, errors.New("missing phi value")
		}
		val, err := c.value(v)
		if err != nil {
			return brEdge{}, err
		}
		phiVals = unpack(val, phiVals)
	}

	label, ok := c.labels[to]
	var height uint
	if ok {
		height = c.b.labels[label].height
	} else {
		// Create the label and request block creation.
		height, ok = c.heights[toInfo.idom]
		if !ok {
			return brEdge{}, errors.New("branch to a block without available height info")
		}
		height += uint(len(phiVals))
		label = c.b.createLabel(height)
		c.labels[to] = label
		c.pending[toInfo.idom] = append(c.pending[toInfo.idom], to)
	}
	height -= uint(len(phiVals))

	// Simplify phi copies.
simpliphi:
	if len(phiVals) > 0 {
		if _, ok := phiVals[0].val.(runtimeValue); ok && phiVals[0].raw == uint64(height) {
			phiVals = phiVals[1:]
			height++
			goto simpliphi
		}
	}

	return brEdge{label, phiVals}, nil
}

func unpack(v value, dst []value) []value {
	if t, ok := v.typ().(aggTyp); ok {
		n := t.elems()
		for i := uint32(0); i < n; i++ {
			dst = unpack(extractValue(v, i), dst)
		}
		return dst
	}

	return append(dst, v)
}

func (c *compiler) retEdge(v value) brEdge {
	var phis []value
	if c.sig.ret != nil {
		phis = unpack(v, nil)
	}
	return brEdge{^uint(0), phis}
}

func (c *compiler) value(v llvm.Value) (value, error) {
	if val, ok := c.ivals[v]; ok {
		return val, nil
	}

	return c.constParser.value(v)
}

type blockInfo struct {
	b llvm.BasicBlock

	phis []llPhi

	first llvm.Value

	term     llvm.Value
	termKind termKind
	to       []llvm.BasicBlock
	ret      llvm.Value

	idom llvm.BasicBlock
}

type termKind uint8

const (
	termRevert termKind = iota
	termUB
	termBr
	termCondBr
	termSwitch
	termRet
)

type llPhi struct {
	phi   llvm.Value
	value map[llvm.BasicBlock]llvm.Value
}

type runnable struct {
	sig    signature
	instrs []instruction
	labels []label
}

func (r *runnable) String() string {
	instrs, labels := r.instrs, r.labels
	lidx := make(map[uint][]uint, len(labels))
	for i, l := range labels {
		lidx[l.start] = append(lidx[l.start], uint(i))
	}
	lines := make([]string, 2+len(instrs)+len(labels)+2)
	lines[0] = r.sig.String() + " {"
	var height uint
	for _, a := range r.sig.args {
		_, height = createDecomposedIndices(height, a.t)
	}
	lines[1] = fmt.Sprintf("entry: (height: %d)", height)
	j := 2
	for i, inst := range instrs {
		for _, i := range lidx[uint(i)] {
			height = labels[i].height
			lines[j] = fmt.Sprintf("%d: (height: %d)", i, height)
			j++
		}

		str := inst.String()
		if res := inst.result(); res != nil {
			v, end := createDecomposedIndices(height, res)
			str = v.String() + " = " + str
			height = end
		}
		lines[j] = "\t" + str
		j++
	}
	height = 0
	if r.sig.ret != nil {
		_, height = createDecomposedIndices(height, r.sig.ret)
	}
	lines[len(lines)-2] = fmt.Sprintf("ret: (height: %d)", height)
	lines[len(lines)-1] = "}"
	return strings.Join(lines, "\n")
}

type builder struct {
	sig         signature
	instrs      []instruction
	labels      []label
	stackHeight uint
}

func (b *builder) createLabel(height uint) uint {
	idx := uint(len(b.labels))
	b.labels = append(b.labels, label{height: height})
	return idx
}

func (b *builder) startLabel(idx uint) {
	b.stackHeight = b.labels[idx].height
	b.labels[idx].start = uint(len(b.instrs))
}

type label struct {
	start  uint
	height uint
}

func (b *builder) insertInst(inst instruction) value {
	res := inst.result()
	b.instrs = append(b.instrs, inst)

	if res == nil {
		return value{}
	}
	return b.pushValue(res)
}

// pushValue allocates space to push a value of the specified type to the stack.
// Aggregates are assumed to be decomposed.
func (b *builder) pushValue(ty typ) value {
	v, end := createDecomposedIndices(b.stackHeight, ty)
	b.stackHeight = end
	return v
}

func createDecomposedIndices(start uint, ty typ) (v value, end uint) {
	switch ty := ty.(type) {
	case nonAggTyp:
		return runtime(ty, start), start + 1
	case arrType:
		elems := make([]value, ty.n)
		for i := range elems {
			elems[i], start = createDecomposedIndices(start, ty.of)
		}
		return arrayValue(ty.of, elems...), start
	case *structType:
		fields := make([]value, len(ty.fields))
		for i, f := range ty.fields {
			fields[i], start = createDecomposedIndices(start, f.ty)
		}
		return structValue(ty, fields...), start
	default:
		panic("cannot handle type " + ty.String())
	}
}

// revert removes the block from the builder.
// This assumes there are no other blocks after this.
func (b *builder) revert(to uint) {
	height := b.labels[to].height
	b.instrs = b.instrs[:b.labels[to].start]
	b.stackHeight = height
	b.labels = b.labels[:to]
}

func (b *builder) finish() *runnable {
	return &runnable{b.sig, b.instrs, b.labels}
}

type instruction interface {
	// result is the result type of the instruction (before decomposition).
	result() typ

	// execute the instruction.
	// This may return an errRevert to revert the function.
	exec(state *execState) error

	// runtime runs the instruction at runtime.
	runtime(gen *rtGen) error

	fmt.Stringer
}

func dbgSuffix(dbg llvm.Metadata) string {
	pos := parseDbg(dbg)
	if !pos.IsValid() {
		return ""
	}

	return " (" + pos.String() + ")"
}

func parseDbg(dbg llvm.Metadata) token.Position {
	if dbg.IsNil() {
		return token.Position{}
	}
	file := dbg.LocationScope().ScopeFile()
	return token.Position{
		Filename: filepath.Join(file.FileDirectory(), file.FileFilename()),
		Line:     int(dbg.LocationLine()),
		Column:   int(dbg.LocationColumn()),
	}
}

func opString(op llvm.Opcode) string {
	if str, ok := instructionNameMap[op]; ok {
		return str
	}

	return fmt.Sprintf("opcode %d", op)
}

// instructionNameMap maps from instruction opcodes to instruction names. This
// can be useful for debug logging.
var instructionNameMap = map[llvm.Opcode]string{
	llvm.Ret:         "ret",
	llvm.Br:          "br",
	llvm.Switch:      "switch",
	llvm.IndirectBr:  "indirectbr",
	llvm.Invoke:      "invoke",
	llvm.Unreachable: "unreachable",

	// Standard Binary Operators
	llvm.Add:  "add",
	llvm.FAdd: "fadd",
	llvm.Sub:  "sub",
	llvm.FSub: "fsub",
	llvm.Mul:  "mul",
	llvm.FMul: "fmul",
	llvm.UDiv: "udiv",
	llvm.SDiv: "sdiv",
	llvm.FDiv: "fdiv",
	llvm.URem: "urem",
	llvm.SRem: "srem",
	llvm.FRem: "frem",

	// Logical Operators
	llvm.Shl:  "shl",
	llvm.LShr: "lshr",
	llvm.AShr: "ashr",
	llvm.And:  "and",
	llvm.Or:   "or",
	llvm.Xor:  "xor",

	// Memory Operators
	llvm.Alloca:        "alloca",
	llvm.Load:          "load",
	llvm.Store:         "store",
	llvm.GetElementPtr: "getelementptr",

	// Cast Operators
	llvm.Trunc:    "trunc",
	llvm.ZExt:     "zext",
	llvm.SExt:     "sext",
	llvm.FPToUI:   "fptoui",
	llvm.FPToSI:   "fptosi",
	llvm.UIToFP:   "uitofp",
	llvm.SIToFP:   "sitofp",
	llvm.FPTrunc:  "fptrunc",
	llvm.FPExt:    "fpext",
	llvm.PtrToInt: "ptrtoint",
	llvm.IntToPtr: "inttoptr",
	llvm.BitCast:  "bitcast",

	// Other Operators
	llvm.ICmp:           "icmp",
	llvm.FCmp:           "fcmp",
	llvm.PHI:            "phi",
	llvm.Call:           "call",
	llvm.Select:         "select",
	llvm.VAArg:          "vaarg",
	llvm.ExtractElement: "extractelement",
	llvm.InsertElement:  "insertelement",
	llvm.ShuffleVector:  "shufflevector",
	llvm.ExtractValue:   "extractvalue",
	llvm.InsertValue:    "insertvalue",
}
