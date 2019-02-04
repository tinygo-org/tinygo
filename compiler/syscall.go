package compiler

// This file implements the syscall.Syscall and syscall.Syscall6 instructions as
// compiler builtins.

import (
	"go/constant"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// emitSyscall emits an inline system call instruction, depending on the target
// OS/arch.
func (c *Compiler) emitSyscall(frame *Frame, call *ssa.CallCommon) (llvm.Value, error) {
	num, _ := constant.Uint64Val(call.Args[0].(*ssa.Const).Value)
	switch {
	case c.GOARCH == "amd64" && c.GOOS == "linux":
		// Sources:
		//   https://stackoverflow.com/a/2538212
		//   https://en.wikibooks.org/wiki/X86_Assembly/Interfacing_with_Linux#syscall
		args := []llvm.Value{llvm.ConstInt(c.uintptrType, num, false)}
		argTypes := []llvm.Type{c.uintptrType}
		// Constraints will look something like:
		//   "={rax},0,{rdi},{rsi},{rdx},{r10},{r8},{r9},~{rcx},~{r11}"
		constraints := "={rax},0"
		for i, arg := range call.Args[1:] {
			constraints += "," + [...]string{
				"{rdi}",
				"{rsi}",
				"{rdx}",
				"{r10}",
				"{r8}",
				"{r9}",
			}[i]
			llvmValue, err := c.parseExpr(frame, arg)
			if err != nil {
				return llvm.Value{}, err
			}
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		constraints += ",~{rcx},~{r11}"
		fnType := llvm.FunctionType(c.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "syscall", constraints, true, false, llvm.InlineAsmDialectIntel)
		syscallResult := c.builder.CreateCall(target, args, "")
		// Return values: r1, r1, err uintptr
		// Pseudocode:
		//     var err uintptr
		//     if syscallResult < 0 && syscallResult > -4096 {
		//         err = -syscallResult
		//     }
		//     return syscallResult, 0, err
		zero := llvm.ConstInt(c.uintptrType, 0, false)
		inrange1 := c.builder.CreateICmp(llvm.IntSLT, syscallResult, llvm.ConstInt(c.uintptrType, 0, false), "")
		inrange2 := c.builder.CreateICmp(llvm.IntSGT, syscallResult, llvm.ConstInt(c.uintptrType, 0xfffffffffffff000, true), "") // -4096
		hasError := c.builder.CreateAnd(inrange1, inrange2, "")
		errResult := c.builder.CreateSelect(hasError, c.builder.CreateNot(syscallResult, ""), zero, "syscallError")
		retval := llvm.Undef(llvm.StructType([]llvm.Type{c.uintptrType, c.uintptrType, c.uintptrType}, false))
		retval = c.builder.CreateInsertValue(retval, syscallResult, 0, "")
		retval = c.builder.CreateInsertValue(retval, zero, 1, "")
		retval = c.builder.CreateInsertValue(retval, errResult, 2, "")
		return retval, nil
	default:
		return llvm.Value{}, c.makeError(call.Pos(), "unknown GOOS/GOARCH for syscall: "+c.GOOS+"/"+c.GOARCH)
	}
}
