package compiler

// This file implements the syscall.Syscall and syscall.Syscall6 instructions as
// compiler builtins.

import (
	"strconv"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// emitSyscall emits an inline system call instruction, depending on the target
// OS/arch.
func (c *Compiler) emitSyscall(frame *Frame, call *ssa.CallCommon) (llvm.Value, error) {
	num := c.getValue(frame, call.Args[0])
	var syscallResult llvm.Value
	switch {
	case c.GOARCH() == "amd64":
		if c.GOOS() == "darwin" {
			// Darwin adds this magic number to system call numbers:
			//
			// > Syscall classes for 64-bit system call entry.
			// > For 64-bit users, the 32-bit syscall number is partitioned
			// > with the high-order bits representing the class and low-order
			// > bits being the syscall number within that class.
			// > The high-order 32-bits of the 64-bit syscall number are unused.
			// > All system classes enter the kernel via the syscall instruction.
			//
			// Source: https://opensource.apple.com/source/xnu/xnu-792.13.8/osfmk/mach/i386/syscall_sw.h
			num = c.builder.CreateOr(num, llvm.ConstInt(c.uintptrType, 0x2000000, false), "")
		}
		// Sources:
		//   https://stackoverflow.com/a/2538212
		//   https://en.wikibooks.org/wiki/X86_Assembly/Interfacing_with_Linux#syscall
		args := []llvm.Value{num}
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
				"{r11}",
				"{r12}",
				"{r13}",
			}[i]
			llvmValue := c.getValue(frame, arg)
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		constraints += ",~{rcx},~{r11}"
		fnType := llvm.FunctionType(c.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "syscall", constraints, true, false, llvm.InlineAsmDialectIntel)
		syscallResult = c.builder.CreateCall(target, args, "")
	case c.GOARCH() == "386" && c.GOOS() == "linux":
		// Sources:
		//   syscall(2) man page
		//   https://stackoverflow.com/a/2538212
		//   https://en.wikibooks.org/wiki/X86_Assembly/Interfacing_with_Linux#int_0x80
		args := []llvm.Value{num}
		argTypes := []llvm.Type{c.uintptrType}
		// Constraints will look something like:
		//   "={eax},0,{ebx},{ecx},{edx},{esi},{edi},{ebp}"
		constraints := "={eax},0"
		for i, arg := range call.Args[1:] {
			constraints += "," + [...]string{
				"{ebx}",
				"{ecx}",
				"{edx}",
				"{esi}",
				"{edi}",
				"{ebp}",
			}[i]
			llvmValue := c.getValue(frame, arg)
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		fnType := llvm.FunctionType(c.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "int 0x80", constraints, true, false, llvm.InlineAsmDialectIntel)
		syscallResult = c.builder.CreateCall(target, args, "")
	case c.GOARCH() == "arm" && c.GOOS() == "linux":
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
			llvmValue := c.getValue(frame, arg)
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		args = append(args, num)
		argTypes = append(argTypes, c.uintptrType)
		constraints += ",{r7}" // syscall number
		for i := len(call.Args) - 1; i < 4; i++ {
			// r0-r3 get clobbered after the syscall returns
			constraints += ",~{r" + strconv.Itoa(i) + "}"
		}
		fnType := llvm.FunctionType(c.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "svc #0", constraints, true, false, 0)
		syscallResult = c.builder.CreateCall(target, args, "")
	case c.GOARCH() == "arm64" && c.GOOS() == "linux":
		// Source: syscall(2) man page.
		args := []llvm.Value{}
		argTypes := []llvm.Type{}
		// Constraints will look something like:
		//   ={x0},0,{x1},{x2},{x8},~{x3},~{x4},~{x5},~{x6},~{x7},~{x16},~{x17}
		constraints := "={x0}"
		for i, arg := range call.Args[1:] {
			constraints += "," + [...]string{
				"0", // tie to output
				"{x1}",
				"{x2}",
				"{x3}",
				"{x4}",
				"{x5}",
			}[i]
			llvmValue := c.getValue(frame, arg)
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		args = append(args, num)
		argTypes = append(argTypes, c.uintptrType)
		constraints += ",{x8}" // syscall number
		for i := len(call.Args) - 1; i < 8; i++ {
			// x0-x7 may get clobbered during the syscall following the aarch64
			// calling convention.
			constraints += ",~{x" + strconv.Itoa(i) + "}"
		}
		constraints += ",~{x16},~{x17}" // scratch registers
		fnType := llvm.FunctionType(c.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "svc #0", constraints, true, false, 0)
		syscallResult = c.builder.CreateCall(target, args, "")
	default:
		return llvm.Value{}, c.makeError(call.Pos(), "unknown GOOS/GOARCH for syscall: "+c.GOOS()+"/"+c.GOARCH())
	}
	switch c.GOOS() {
	case "linux", "freebsd":
		// Return values: r0, r1 uintptr, err Errno
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
		errResult := c.builder.CreateSelect(hasError, c.builder.CreateSub(zero, syscallResult, ""), zero, "syscallError")
		retval := llvm.Undef(c.ctx.StructType([]llvm.Type{c.uintptrType, c.uintptrType, c.uintptrType}, false))
		retval = c.builder.CreateInsertValue(retval, syscallResult, 0, "")
		retval = c.builder.CreateInsertValue(retval, zero, 1, "")
		retval = c.builder.CreateInsertValue(retval, errResult, 2, "")
		return retval, nil
	case "darwin":
		// Return values: r0, r1 uintptr, err Errno
		// Pseudocode:
		//     var err uintptr
		//     if syscallResult != 0 {
		//         err = syscallResult
		//     }
		//     return syscallResult, 0, err
		zero := llvm.ConstInt(c.uintptrType, 0, false)
		hasError := c.builder.CreateICmp(llvm.IntNE, syscallResult, llvm.ConstInt(c.uintptrType, 0, false), "")
		errResult := c.builder.CreateSelect(hasError, syscallResult, zero, "syscallError")
		retval := llvm.Undef(c.ctx.StructType([]llvm.Type{c.uintptrType, c.uintptrType, c.uintptrType}, false))
		retval = c.builder.CreateInsertValue(retval, syscallResult, 0, "")
		retval = c.builder.CreateInsertValue(retval, zero, 1, "")
		retval = c.builder.CreateInsertValue(retval, errResult, 2, "")
		return retval, nil
	default:
		return llvm.Value{}, c.makeError(call.Pos(), "unknown GOOS/GOARCH for syscall: "+c.GOOS()+"/"+c.GOARCH())
	}
}
