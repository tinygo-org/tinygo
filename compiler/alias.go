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
	"crypto/md5.block":         "crypto/md5.blockGeneric",
	"crypto/sha1.block":        "crypto/sha1.blockGeneric",
	"crypto/sha1.blockAMD64":   "crypto/sha1.blockGeneric",
	"crypto/sha256.block":      "crypto/sha256.blockGeneric",
	"crypto/sha512.blockAMD64": "crypto/sha512.blockGeneric",

	// math package
	"math.Asin":      "math.asin",
	"math.Asinh":     "math.asinh",
	"math.Acos":      "math.acos",
	"math.Acosh":     "math.acosh",
	"math.Atan":      "math.atan",
	"math.Atanh":     "math.atanh",
	"math.Atan2":     "math.atan2",
	"math.Cbrt":      "math.cbrt",
	"math.Ceil":      "math.ceil",
	"math.archCeil":  "math.ceil",
	"math.Cos":       "math.cos",
	"math.Cosh":      "math.cosh",
	"math.Erf":       "math.erf",
	"math.Erfc":      "math.erfc",
	"math.Exp":       "math.exp",
	"math.archExp":   "math.exp",
	"math.Expm1":     "math.expm1",
	"math.Exp2":      "math.exp2",
	"math.archExp2":  "math.exp2",
	"math.Floor":     "math.floor",
	"math.archFloor": "math.floor",
	"math.Frexp":     "math.frexp",
	"math.Hypot":     "math.hypot",
	"math.archHypot": "math.hypot",
	"math.Ldexp":     "math.ldexp",
	"math.Log":       "math.log",
	"math.archLog":   "math.log",
	"math.Log1p":     "math.log1p",
	"math.Log10":     "math.log10",
	"math.Log2":      "math.log2",
	"math.Max":       "math.max",
	"math.archMax":   "math.max",
	"math.Min":       "math.min",
	"math.archMin":   "math.min",
	"math.Mod":       "math.mod",
	"math.Modf":      "math.modf",
	"math.archModf":  "math.modf",
	"math.Pow":       "math.pow",
	"math.Remainder": "math.remainder",
	"math.Sin":       "math.sin",
	"math.Sinh":      "math.sinh",
	"math.Sqrt":      "math.sqrt",
	"math.archSqrt":  "math.sqrt",
	"math.Tan":       "math.tan",
	"math.Tanh":      "math.tanh",
	"math.Trunc":     "math.trunc",
	"math.archTrunc": "math.trunc",
}

// createAlias implements the function (in the builder) as a call to the alias
// function.
func (b *builder) createAlias(alias llvm.Value) {
	b.llvmFn.SetVisibility(llvm.HiddenVisibility)
	b.llvmFn.SetUnnamedAddr(true)

	if b.Debug {
		if b.fn.Syntax() != nil {
			// Create debug info file if present.
			b.difunc = b.attachDebugInfo(b.fn)
		}
		pos := b.program.Fset.Position(b.fn.Pos())
		b.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), b.difunc, llvm.Metadata{})
	}
	entryBlock := llvm.AddBasicBlock(b.llvmFn, "entry")
	b.SetInsertPointAtEnd(entryBlock)
	if b.llvmFn.Type() != alias.Type() {
		b.addError(b.fn.Pos(), "alias function should have the same type as aliasee "+alias.Name())
		b.CreateUnreachable()
		return
	}
	result := b.CreateCall(alias, b.llvmFn.Params(), "")
	if result.Type().TypeKind() == llvm.VoidTypeKind {
		b.CreateRetVoid()
	} else {
		b.CreateRet(result)
	}
}
