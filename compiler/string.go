package compiler

import (
	"strconv"

	"tinygo.org/x/go-llvm"
)

func (b *builder) createStringEqual(lhs, rhs llvm.Value) llvm.Value {
	// Compare the lengths.
	lhsLen := b.CreateExtractValue(lhs, 1, "streq.lhs.len")
	rhsLen := b.CreateExtractValue(rhs, 1, "streq.rhs.len")
	lenCmp := b.CreateICmp(llvm.IntEQ, lhsLen, rhsLen, "streq.len.eq")

	// Branch on the length comparison.
	bodyCmpBlock := b.ctx.AddBasicBlock(b.llvmFn, "streq.body")
	nextBlock := b.ctx.AddBasicBlock(b.llvmFn, "streq.next")
	b.CreateCondBr(lenCmp, bodyCmpBlock, nextBlock)

	// Use memcmp to compare the contents if the lengths are equal.
	b.SetInsertPointAtEnd(bodyCmpBlock)
	lhsPtr := b.CreateExtractValue(lhs, 0, "streq.lhs.ptr")
	rhsPtr := b.CreateExtractValue(rhs, 0, "streq.rhs.ptr")
	memcmp := b.createMemCmp(lhsPtr, rhsPtr, lhsLen, "streq.memcmp")
	memcmpEq := b.CreateICmp(llvm.IntEQ, memcmp, llvm.ConstNull(b.cIntType), "streq.memcmp.eq")
	b.CreateBr(nextBlock)

	// Create a phi to join the results.
	b.SetInsertPointAtEnd(nextBlock)
	result := b.CreatePHI(b.ctx.Int1Type(), "")
	result.AddIncoming([]llvm.Value{llvm.ConstNull(b.ctx.Int1Type()), memcmpEq}, []llvm.BasicBlock{b.currentBlockInfo.exit, bodyCmpBlock})
	b.currentBlockInfo.exit = nextBlock // adjust outgoing block for phi nodes

	return result
}

func (b *builder) createStringLess(lhs, rhs llvm.Value) llvm.Value {
	// Calculate the minimum of the two string lengths.
	lhsLen := b.CreateExtractValue(lhs, 1, "strlt.lhs.len")
	rhsLen := b.CreateExtractValue(rhs, 1, "strlt.rhs.len")
	minFnName := "llvm.umin.i" + strconv.Itoa(b.uintptrType.IntTypeWidth())
	minFn := b.mod.NamedFunction(minFnName)
	if minFn.IsNil() {
		fnType := llvm.FunctionType(b.uintptrType, []llvm.Type{b.uintptrType, b.uintptrType}, false)
		minFn = llvm.AddFunction(b.mod, minFnName, fnType)
	}
	minLen := b.CreateCall(minFn.GlobalValueType(), minFn, []llvm.Value{lhsLen, rhsLen}, "strlt.min.len")

	// Compare the common-length body.
	lhsPtr := b.CreateExtractValue(lhs, 0, "strlt.lhs.ptr")
	rhsPtr := b.CreateExtractValue(rhs, 0, "strlt.rhs.ptr")
	memcmp := b.createMemCmp(lhsPtr, rhsPtr, minLen, "strlt.memcmp")

	// Evaluate the result as: memcmp == 0 ? lhsLen < rhsLen : memcmp < 0
	zero := llvm.ConstNull(b.cIntType)
	memcmpEQ := b.CreateICmp(llvm.IntEQ, memcmp, zero, "strlt.memcmp.eq")
	lenLT := b.CreateICmp(llvm.IntULT, lhsLen, rhsLen, "strlt.len.lt")
	memcmpLT := b.CreateICmp(llvm.IntSLT, memcmp, zero, "strlt.memcmp.lt")
	return b.CreateSelect(memcmpEQ, lenLT, memcmpLT, "strlt.result")
}

// createMemCmp compares memory by calling the libc function memcmp.
// This function is handled specially by LLVM:
// - It can be constant-folded in some trivial cases (e.g. len 0)
// - It can be replaced with loads and compares when the length is small and known
func (b *builder) createMemCmp(lhs, rhs, len llvm.Value, name string) llvm.Value {
	memcmp := b.mod.NamedFunction("memcmp")
	if memcmp.IsNil() {
		fnType := llvm.FunctionType(b.cIntType, []llvm.Type{b.dataPtrType, b.dataPtrType, b.uintptrType}, false)
		memcmp = llvm.AddFunction(b.mod, "memcmp", fnType)

		// The memcmp call does not capture the string.
		nocapture := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("nocapture"), 0)
		memcmp.AddAttributeAtIndex(1, nocapture)
		memcmp.AddAttributeAtIndex(2, nocapture)

		// The memcmp call does not modify the string.
		readonly := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("readonly"), 0)
		memcmp.AddAttributeAtIndex(1, readonly)
		memcmp.AddAttributeAtIndex(2, readonly)
	}
	return b.CreateCall(memcmp.GlobalValueType(), memcmp, []llvm.Value{lhs, rhs, len}, name)
}
