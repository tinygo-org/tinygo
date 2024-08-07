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
// result is not further interpreted (with the exception of MIPS to use the same
// return value everywhere).
func (b *builder) createRawSyscall(call *ssa.CallCommon) (llvm.Value, error) {
	num := b.getValue(call.Args[0], getPos(call))
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
			}[i]
			llvmValue := b.getValue(arg, getPos(call))
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		// rcx and r11 are clobbered by the syscall, so make sure they are not used
		constraints += ",~{rcx},~{r11}"
		fnType := llvm.FunctionType(b.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "syscall", constraints, true, false, llvm.InlineAsmDialectIntel, false)
		return b.CreateCall(fnType, target, args, ""), nil

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
			llvmValue := b.getValue(arg, getPos(call))
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		fnType := llvm.FunctionType(b.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "int 0x80", constraints, true, false, llvm.InlineAsmDialectIntel, false)
		return b.CreateCall(fnType, target, args, ""), nil

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
			llvmValue := b.getValue(arg, getPos(call))
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
		return b.CreateCall(fnType, target, args, ""), nil

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
			llvmValue := b.getValue(arg, getPos(call))
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
		return b.CreateCall(fnType, target, args, ""), nil

	case (b.GOARCH == "mips" || b.GOARCH == "mipsle") && b.GOOS == "linux":
		// Implement the system call convention for Linux.
		// Source: syscall(2) man page and musl:
		// https://git.musl-libc.org/cgit/musl/tree/arch/mips/syscall_arch.h
		// Also useful:
		// https://web.archive.org/web/20220529105937/https://www.linux-mips.org/wiki/Syscall
		// The syscall number goes in r2, the result also in r2.
		// Register r7 is both an input parameter and an output parameter: if it
		// is non-zero, the system call failed and r2 is the error code.
		// The code below implements the O32 syscall ABI, not the N32 ABI. It
		// could implement both at the same time if needed (like what appears to
		// be done in musl) by forcing arg5-arg7 into the right registers but
		// letting the compiler decide the registers should result in _slightly_
		// faster and smaller code.
		args := []llvm.Value{num}
		argTypes := []llvm.Type{b.uintptrType}
		constraints := "={$2},={$7},0"
		syscallParams := call.Args[1:]

		if len(syscallParams) > 7 {
			// There is one syscall that uses 7 parameters: sync_file_range.
			// But only 7, not more. Go however only has Syscall6 and Syscall9.
			// Therefore, we can ignore the remaining parameters.
			syscallParams = syscallParams[:7]
		}
		for i, arg := range syscallParams {
			constraints += "," + [...]string{
				"{$4}", // arg1
				"{$5}", // arg2
				"{$6}", // arg3
				"1",    // arg4, error return
				"r",    // arg5 on the stack
				"r",    // arg6 on the stack
				"r",    // arg7 on the stack
			}[i]
			llvmValue := b.getValue(arg, getPos(call))
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		// Create assembly code.
		// Parameters beyond the first 4 are passed on the stack instead of in
		// registers in the O32 syscall ABI.
		// We need ".set noat" because LLVM might pick register $1 ($at) as the
		// register for a parameter and apparently this is not allowed on MIPS
		// unless you use this specific pragma.
		asm := "syscall"
		switch len(syscallParams) {
		case 5:
			asm = "" +
				".set noat\n" +
				"subu $$sp, $$sp, 32\n" +
				"sw $7, 16($$sp)\n" + // arg5
				"syscall\n" +
				"addu $$sp, $$sp, 32\n" +
				".set at\n"
		case 6:
			asm = "" +
				".set noat\n" +
				"subu $$sp, $$sp, 32\n" +
				"sw $7, 16($$sp)\n" + // arg5
				"sw $8, 20($$sp)\n" + // arg6
				"syscall\n" +
				"addu $$sp, $$sp, 32\n" +
				".set at\n"
		case 7:
			asm = "" +
				".set noat\n" +
				"subu $$sp, $$sp, 32\n" +
				"sw $7, 16($$sp)\n" + // arg5
				"sw $8, 20($$sp)\n" + // arg6
				"sw $9, 24($$sp)\n" + // arg7
				"syscall\n" +
				"addu $$sp, $$sp, 32\n" +
				".set at\n"
		}
		constraints += ",~{$3},~{$4},~{$5},~{$6},~{$8},~{$9},~{$10},~{$11},~{$12},~{$13},~{$14},~{$15},~{$24},~{$25},~{hi},~{lo},~{memory}"
		returnType := b.ctx.StructType([]llvm.Type{b.uintptrType, b.uintptrType}, false)
		fnType := llvm.FunctionType(returnType, argTypes, false)
		target := llvm.InlineAsm(fnType, asm, constraints, true, true, 0, false)
		call := b.CreateCall(fnType, target, args, "")
		resultCode := b.CreateExtractValue(call, 0, "") // r2
		errorFlag := b.CreateExtractValue(call, 1, "")  // r7
		// Pseudocode to return the result with the same convention as other
		// archs:
		//    return (errorFlag != 0) ? -resultCode : resultCode;
		// At least on QEMU with the O32 ABI, the error code is always positive.
		zero := llvm.ConstInt(b.uintptrType, 0, false)
		isError := b.CreateICmp(llvm.IntNE, errorFlag, zero, "")
		negativeResult := b.CreateSub(zero, resultCode, "")
		result := b.CreateSelect(isError, negativeResult, resultCode, "")
		return result, nil

	case (b.GOARCH == "riscv" || b.GOARCH == "riscv64") && b.GOOS == "linux":
		// https://stackoverflow.com/questions/59800430/risc-v-ecall-syscall-calling-convention-on-pk-linux
		// https://git.kernel.org/pub/scm/docs/man-pages/man-pages.git/tree/man2/syscall.2?h=man-pages-5.04#n200
		// https://pdos.csail.mit.edu/6.S081/2021/slides/6s081-lec-syscall.pdf
		// https://git.musl-libc.org/cgit/musl/tree/arch/riscv64/syscall_arch.h
		// https://riscv.org/wp-content/uploads/2015/01/riscv-calling.pdf
		args := []llvm.Value{num}
		argTypes := []llvm.Type{}
		constraints := "={a7}" // syscall number
		for i, arg := range call.Args[1:] {
			constraints += "," + [...]string{
				"{a0}",
				"{a1}",
				"{a2}",
				"{a3}",
				"{a4}",
				"{a5}",
			}[i]
			llvmValue := b.getValue(arg, getPos(call))
			args = append(args, llvmValue)
			argTypes = append(argTypes, llvmValue.Type())
		}
		args = append(args, num)
		argTypes = append(argTypes, b.uintptrType)

		// constrain registers used for syscall/ecall
		for i := len(call.Args) - 1; i < 4; i++ {
			constraints += ",~{a" + strconv.Itoa(i) + "}"
		}

		// constrain caller responsible registers
		// See Table 18.2: RISC-V calling convention register usage
		// TODO: does llvm take their stack usage into account?
		// TODO: are only the registers needed saved here or all? can llvm determine this itself?

		// temporary integer registers
		for i := 0; i < 8; i++ {
			constraints += ",~{t" + strconv.Itoa(i) + "}"
		}

		// temporary floating-point registers
		for i := 0; i < 12; i++ {
			constraints += ",~{ft" + strconv.Itoa(i) + "}"
		}

		// generate function
		fnType := llvm.FunctionType(b.uintptrType, argTypes, false)
		target := llvm.InlineAsm(fnType, "ecall", constraints, true, false, 0, false)
		return b.CreateCall(fnType, target, args, ""), nil

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
			param := b.getValue(val, getPos(call))
			params = append(params, param)
			paramTypes = append(paramTypes, param.Type())
		}
		llvmType := llvm.FunctionType(b.uintptrType, paramTypes, false)
		fn := b.getValue(call.Args[0], getPos(call))
		fnPtr := b.CreateIntToPtr(fn, b.dataPtrType, "")

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
		b.CreateCall(setLastError.GlobalValueType(), setLastError, []llvm.Value{llvm.ConstNull(b.ctx.Int32Type())}, "")
		syscallResult := b.CreateCall(llvmType, fnPtr, params, "")
		errResult := b.CreateCall(getLastError.GlobalValueType(), getLastError, nil, "err")
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
