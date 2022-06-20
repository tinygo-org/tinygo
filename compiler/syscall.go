package compiler

// This file implements the syscall.Syscall and syscall.Syscall6 instructions as
// compiler builtins.

import (
	"strconv"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// createRawSyscall creates a system call with the provided system call number
// and returns the result as a single integer (the system call result). The
// result is not further interpreted.
func (b *builder) createRawSyscall(call *ssa.CallCommon) (llvm.Value, error) {
	num := b.getValue(call.Args[0])
	switch {
	case b.GOARCH == "amd64" && b.GOOS == "linux":
		// Sources:
		//   https://stackoverflow.com/a/2538212
		//   https://en.wikibooks.org/wiki/X86_Assembly/Interfacing_with_Linux#syscall
		args := []llvm.Value{num}
		argTypes := []llvm.Type{b.uintptrType}
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
			llvmValue := b.getValue(arg)
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		constraints += ",~{rcx},~{r11}"
		fnType := llvm.FunctionType(b.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "syscall", constraints, true, false, llvm.InlineAsmDialectIntel, false)
		return b.CreateCall(target, args, ""), nil
	case b.GOARCH == "386" && b.GOOS == "linux":
		// Sources:
		//   syscall(2) man page
		//   https://stackoverflow.com/a/2538212
		//   https://en.wikibooks.org/wiki/X86_Assembly/Interfacing_with_Linux#int_0x80
		args := []llvm.Value{num}
		argTypes := []llvm.Type{b.uintptrType}
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
			llvmValue := b.getValue(arg)
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		fnType := llvm.FunctionType(b.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "int 0x80", constraints, true, false, llvm.InlineAsmDialectIntel, false)
		return b.CreateCall(target, args, ""), nil
	case b.GOARCH == "arm" && b.GOOS == "linux":
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
			llvmValue := b.getValue(arg)
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		args = append(args, num)
		argTypes = append(argTypes, b.uintptrType)
		constraints += ",{r7}" // syscall number
		for i := len(call.Args) - 1; i < 4; i++ {
			// r0-r3 get clobbered after the syscall returns
			constraints += ",~{r" + strconv.Itoa(i) + "}"
		}
		fnType := llvm.FunctionType(b.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "svc #0", constraints, true, false, 0, false)
		return b.CreateCall(target, args, ""), nil
	case b.GOARCH == "arm64" && b.GOOS == "linux":
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
			llvmValue := b.getValue(arg)
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		args = append(args, num)
		argTypes = append(argTypes, b.uintptrType)
		constraints += ",{x8}" // syscall number
		for i := len(call.Args) - 1; i < 8; i++ {
			// x0-x7 may get clobbered during the syscall following the aarch64
			// calling convention.
			constraints += ",~{x" + strconv.Itoa(i) + "}"
		}
		constraints += ",~{x16},~{x17}" // scratch registers
		fnType := llvm.FunctionType(b.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "svc #0", constraints, true, false, 0, false)
		return b.CreateCall(target, args, ""), nil
	default:
		return llvm.Value{}, b.makeError(call.Pos(), "unknown GOOS/GOARCH for syscall: "+b.GOOS+"/"+b.GOARCH)
	}
}

// createSyscall emits instructions for the syscall.Syscall* family of
// functions, depending on the target OS/arch.
func (b *builder) createSyscall(call *ssa.CallCommon) (llvm.Value, error) {
	switch b.GOOS {
	case "linux":
		syscallResult, err := b.createRawSyscall(call)
		if err != nil {
			return syscallResult, err
		}
		// Return values: r0, r1 uintptr, err Errno
		// Pseudocode:
		//     var err uintptr
		//     if syscallResult < 0 && syscallResult > -4096 {
		//         err = -syscallResult
		//     }
		//     return syscallResult, 0, err
		zero := llvm.ConstInt(b.uintptrType, 0, false)
		inrange1 := b.CreateICmp(llvm.IntSLT, syscallResult, llvm.ConstInt(b.uintptrType, 0, false), "")
		inrange2 := b.CreateICmp(llvm.IntSGT, syscallResult, llvm.ConstInt(b.uintptrType, 0xfffffffffffff000, true), "") // -4096
		hasError := b.CreateAnd(inrange1, inrange2, "")
		errResult := b.CreateSelect(hasError, b.CreateSub(zero, syscallResult, ""), zero, "syscallError")
		retval := llvm.Undef(b.ctx.StructType([]llvm.Type{b.uintptrType, b.uintptrType, b.uintptrType}, false))
		retval = b.CreateInsertValue(retval, syscallResult, 0, "")
		retval = b.CreateInsertValue(retval, zero, 1, "")
		retval = b.CreateInsertValue(retval, errResult, 2, "")
		return retval, nil
	case "windows":
		// On Windows, syscall.Syscall* is basically just a function pointer
		// call. This is complicated in gc because of stack switching and the
		// different ABI, but easy in TinyGo: just call the function pointer.
		// The signature looks like this:
		//   func Syscall(trap, nargs, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)

		// Prepare input values.
		var paramTypes []llvm.Type
		var params []llvm.Value
		for _, val := range call.Args[2:] {
			param := b.getValue(val)
			params = append(params, param)
			paramTypes = append(paramTypes, param.Type())
		}
		llvmType := llvm.FunctionType(b.uintptrType, paramTypes, false)
		fn := b.getValue(call.Args[0])
		fnPtr := b.CreateIntToPtr(fn, llvm.PointerType(llvmType, 0), "")

		// Prepare some functions that will be called later.
		setLastError := b.mod.NamedFunction("SetLastError")
		if setLastError.IsNil() {
			llvmType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{b.ctx.Int32Type()}, false)
			setLastError = llvm.AddFunction(b.mod, "SetLastError", llvmType)
		}
		getLastError := b.mod.NamedFunction("GetLastError")
		if getLastError.IsNil() {
			llvmType := llvm.FunctionType(b.ctx.Int32Type(), nil, false)
			getLastError = llvm.AddFunction(b.mod, "GetLastError", llvmType)
		}

		// Now do the actual call. Pseudocode:
		//     SetLastError(0)
		//     r1 = trap(a1, a2, a3, ...)
		//     err = uintptr(GetLastError())
		//     return r1, 0, err
		// Note that SetLastError/GetLastError could be replaced with direct
		// access to the thread control block, which is probably smaller and
		// faster. The Go runtime does this in assembly.
		b.CreateCall(setLastError, []llvm.Value{llvm.ConstNull(b.ctx.Int32Type())}, "")
		syscallResult := b.CreateCall(fnPtr, params, "")
		errResult := b.CreateCall(getLastError, nil, "err")
		if b.uintptrType != b.ctx.Int32Type() {
			errResult = b.CreateZExt(errResult, b.uintptrType, "err.uintptr")
		}

		// Return r1, 0, err
		retval := llvm.ConstNull(b.ctx.StructType([]llvm.Type{b.uintptrType, b.uintptrType, b.uintptrType}, false))
		retval = b.CreateInsertValue(retval, syscallResult, 0, "")
		retval = b.CreateInsertValue(retval, errResult, 2, "")
		return retval, nil
	default:
		return llvm.Value{}, b.makeError(call.Pos(), "unknown GOOS/GOARCH for syscall: "+b.GOOS+"/"+b.GOARCH)
	}
}

// createRawSyscallNoError emits instructions for the Linux-specific
// syscall.rawSyscallNoError function.
func (b *builder) createRawSyscallNoError(call *ssa.CallCommon) (llvm.Value, error) {
	syscallResult, err := b.createRawSyscall(call)
	if err != nil {
		return syscallResult, err
	}
	retval := llvm.ConstNull(b.ctx.StructType([]llvm.Type{b.uintptrType, b.uintptrType}, false))
	retval = b.CreateInsertValue(retval, syscallResult, 0, "")
	retval = b.CreateInsertValue(retval, llvm.ConstInt(b.uintptrType, 0, false), 1, "")
	return retval, nil
}
