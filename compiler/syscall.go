package compiler

// This file implements the syscall.Syscall and syscall.Syscall6 instructions as
// compiler builtins.

import (
	"go/constant"
	"strconv"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// emitSyscall emits an inline system call instruction, depending on the target
// OS/arch.
func (c *Compiler) emitSyscall(frame *Frame, call *ssa.CallCommon) (llvm.Value, error) {
	num, _ := constant.Uint64Val(call.Args[0].(*ssa.Const).Value)
	var syscallResult llvm.Value
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
		syscallResult = c.builder.CreateCall(target, args, "")
	case c.GOARCH == "arm" && c.GOOS == "linux":
		// Implement the EABI system call convention for Linux.
		// Source: syscall(2) man page.
		args := []llvm.Value{}
		argTypes := []llvm.Type{}
		// Constraints will look something like:
		//   ={r0},0,{r1},{r2},{r7},~{r3}
		constraints := "={r0}"
		for i, arg := range call.Args[1:] {
			constraints += "," + [...]string{
				"0", // tie to output
				"{r1}",
				"{r2}",
				"{r3}",
				"{r4}",
				"{r5}",
				"{r6}",
			}[i]
			llvmValue, err := c.parseExpr(frame, arg)
			if err != nil {
				return llvm.Value{}, err
			}
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		args = append(args, llvm.ConstInt(c.uintptrType, num, false))
		argTypes = append(argTypes, c.uintptrType)
		constraints += ",{r7}" // syscall number
		for i := len(call.Args) - 1; i < 4; i++ {
			// r0-r3 get clobbered after the syscall returns
			constraints += ",~{r" + strconv.Itoa(i) + "}"
		}
		fnType := llvm.FunctionType(c.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "svc #0", constraints, true, false, 0)
		syscallResult = c.builder.CreateCall(target, args, "")
	default:
		return llvm.Value{}, c.makeError(call.Pos(), "unknown GOOS/GOARCH for syscall: "+c.GOOS+"/"+c.GOARCH)
	}
	// Return values: r0, r1, err uintptr
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
}
