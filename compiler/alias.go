package compiler

// This file defines alias functions for functions that are normally defined in
// Go assembly.
//
// The Go toolchain defines many performance critical functions in assembly
// instead of plain Go. This is a problem for TinyGo as it currently (as of
// august 2021) is not able to compile these assembly files and even if it
// could, it would not be able to make use of them for many targets that are
// supported by TinyGo (baremetal RISC-V, AVR, etc). Therefore, many of these
// functions are aliased to their generic Go implementation.
// This results in slower than possible implementations, but at least they are
// usable.

import "tinygo.org/x/go-llvm"

var stdlibAliases = map[string]string{
	// crypto packages
	"crypto/ed25519/internal/edwards25519/field.feMul":    "crypto/ed25519/internal/edwards25519/field.feMulGeneric",
	"crypto/ed25519/internal/edwards25519/field.feSquare": "crypto/ed25519/internal/edwards25519/field.feSquareGeneric",
	"crypto/md5.block":         "crypto/md5.blockGeneric",
	"crypto/sha1.block":        "crypto/sha1.blockGeneric",
	"crypto/sha1.blockAMD64":   "crypto/sha1.blockGeneric",
	"crypto/sha256.block":      "crypto/sha256.blockGeneric",
	"crypto/sha512.blockAMD64": "crypto/sha512.blockGeneric",

	// math package
	"math.archHypot": "math.hypot",
	"math.archMax":   "math.max",
	"math.archMin":   "math.min",
	"math.archModf":  "math.modf",
}

// createAlias implements the function (in the builder) as a call to the alias
// function.
func (b *builder) createAlias(alias llvm.Value) {
	if !b.LTO {
		b.llvmFn.SetVisibility(llvm.HiddenVisibility)
	}
	b.llvmFn.SetUnnamedAddr(true)

	if b.Debug {
		if b.fn.Syntax() != nil {
			// Create debug info file if present.
			b.difunc = b.attachDebugInfo(b.fn)
		}
		pos := b.program.Fset.Position(b.fn.Pos())
		b.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), b.difunc, llvm.Metadata{})
	}
	entryBlock := b.ctx.AddBasicBlock(b.llvmFn, "entry")
	b.SetInsertPointAtEnd(entryBlock)
	if b.llvmFn.Type() != alias.Type() {
		b.addError(b.fn.Pos(), "alias function should have the same type as aliasee "+alias.Name())
		b.CreateUnreachable()
		return
	}
	result := b.CreateCall(alias.GlobalValueType(), alias, b.llvmFn.Params(), "")
	if result.Type().TypeKind() == llvm.VoidTypeKind {
		b.CreateRetVoid()
	} else {
		b.CreateRet(result)
	}
}
