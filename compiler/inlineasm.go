package compiler

// This file implements inline asm support by calling special functions.

import (
	"fmt"
	"go/constant"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// This is a compiler builtin, which emits a piece of inline assembly with no
// operands or return values. It is useful for trivial instructions, like wfi in
// ARM or sleep in AVR.
//
//     func Asm(asm string)
//
// The provided assembly must be a constant.
func (b *builder) createInlineAsm(args []ssa.Value) (llvm.Value, error) {
	// Magic function: insert inline assembly instead of calling it.
	fnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{}, false)
	asm := constant.StringVal(args[0].(*ssa.Const).Value)
	target := llvm.InlineAsm(fnType, asm, "", true, false, 0, false)
	return b.CreateCall(target, nil, ""), nil
}

// This is a compiler builtin, which allows assembly to be called in a flexible
// way.
//
//     func AsmFull(asm string, regs map[string]interface{}) uintptr
//
// The asm parameter must be a constant string. The regs parameter must be
// provided immediately. For example:
//
//     arm.AsmFull(
//         "str {value}, {result}",
//         map[string]interface{}{
//             "value":  1
//             "result": &dest,
//         })
func (b *builder) createInlineAsmFull(instr *ssa.CallCommon) (llvm.Value, error) {
	asmString := constant.StringVal(instr.Args[0].(*ssa.Const).Value)
	registers := map[string]llvm.Value{}
	if registerMap, ok := instr.Args[1].(*ssa.MakeMap); ok {
		for _, r := range *registerMap.Referrers() {
			switch r := r.(type) {
			case *ssa.DebugRef:
				// ignore
			case *ssa.MapUpdate:
				if r.Block() != registerMap.Block() {
					return llvm.Value{}, b.makeError(instr.Pos(), "register value map must be created in the same basic block")
				}
				key := constant.StringVal(r.Key.(*ssa.Const).Value)
				registers[key] = b.getValue(r.Value.(*ssa.MakeInterface).X)
			case *ssa.Call:
				if r.Common() == instr {
					break
				}
			default:
				return llvm.Value{}, b.makeError(instr.Pos(), "don't know how to handle argument to inline assembly: "+r.String())
			}
		}
	}
	// TODO: handle dollar signs in asm string
	registerNumbers := map[string]int{}
	var err error
	argTypes := []llvm.Type{}
	args := []llvm.Value{}
	constraints := []string{}
	hasOutput := false
	asmString = regexp.MustCompile(`\{\}`).ReplaceAllStringFunc(asmString, func(s string) string {
		hasOutput = true
		return "$0"
	})
	if hasOutput {
		constraints = append(constraints, "=&r")
		registerNumbers[""] = 0
	}
	asmString = regexp.MustCompile(`\{[a-zA-Z]+\}`).ReplaceAllStringFunc(asmString, func(s string) string {
		// TODO: skip strings like {r4} etc. that look like ARM push/pop
		// instructions.
		name := s[1 : len(s)-1]
		if _, ok := registers[name]; !ok {
			if err == nil {
				err = b.makeError(instr.Pos(), "unknown register name: "+name)
			}
			return s
		}
		if _, ok := registerNumbers[name]; !ok {
			registerNumbers[name] = len(registerNumbers)
			argTypes = append(argTypes, registers[name].Type())
			args = append(args, registers[name])
			switch registers[name].Type().TypeKind() {
			case llvm.IntegerTypeKind:
				constraints = append(constraints, "r")
			case llvm.PointerTypeKind:
				constraints = append(constraints, "*m")
			default:
				err = b.makeError(instr.Pos(), "unknown type in inline assembly for value: "+name)
				return s
			}
		}
		return fmt.Sprintf("${%v}", registerNumbers[name])
	})
	if err != nil {
		return llvm.Value{}, err
	}
	var outputType llvm.Type
	if hasOutput {
		outputType = b.uintptrType
	} else {
		outputType = b.ctx.VoidType()
	}
	fnType := llvm.FunctionType(outputType, argTypes, false)
	target := llvm.InlineAsm(fnType, asmString, strings.Join(constraints, ","), true, false, 0, false)
	result := b.CreateCall(target, args, "")
	if hasOutput {
		return result, nil
	} else {
		// Make sure we return something valid.
		return llvm.ConstInt(b.uintptrType, 0, false), nil
	}
}

// This is a compiler builtin which emits an inline SVCall instruction. It can
// be one of:
//
//     func SVCall0(num uintptr) uintptr
//     func SVCall1(num uintptr, a1 interface{}) uintptr
//     func SVCall2(num uintptr, a1, a2 interface{}) uintptr
//     func SVCall3(num uintptr, a1, a2, a3 interface{}) uintptr
//     func SVCall4(num uintptr, a1, a2, a3, a4 interface{}) uintptr
//
// The num parameter must be a constant. All other parameters may be any scalar
// value supported by LLVM inline assembly.
func (b *builder) emitSVCall(args []ssa.Value) (llvm.Value, error) {
	num, _ := constant.Uint64Val(args[0].(*ssa.Const).Value)
	llvmArgs := []llvm.Value{}
	argTypes := []llvm.Type{}
	asm := "svc #" + strconv.FormatUint(num, 10)
	constraints := "={r0}"
	for i, arg := range args[1:] {
		arg = arg.(*ssa.MakeInterface).X
		if i == 0 {
			constraints += ",0"
		} else {
			constraints += ",{r" + strconv.Itoa(i) + "}"
		}
		llvmValue := b.getValue(arg)
		llvmArgs = append(llvmArgs, llvmValue)
		argTypes = append(argTypes, llvmValue.Type())
	}
	// Implement the ARM calling convention by marking r1-r3 as
	// clobbered. r0 is used as an output register so doesn't have to be
	// marked as clobbered.
	constraints += ",~{r1},~{r2},~{r3}"
	fnType := llvm.FunctionType(b.uintptrType, argTypes, false)
	target := llvm.InlineAsm(fnType, asm, constraints, true, false, 0, false)
	return b.CreateCall(target, llvmArgs, ""), nil
}

// This is a compiler builtin which emits an inline SVCall instruction. It can
// be one of:
//
//     func SVCall0(num uintptr) uintptr
//     func SVCall1(num uintptr, a1 interface{}) uintptr
//     func SVCall2(num uintptr, a1, a2 interface{}) uintptr
//     func SVCall3(num uintptr, a1, a2, a3 interface{}) uintptr
//     func SVCall4(num uintptr, a1, a2, a3, a4 interface{}) uintptr
//
// The num parameter must be a constant. All other parameters may be any scalar
// value supported by LLVM inline assembly.
// Same as emitSVCall but for AArch64
func (b *builder) emitSV64Call(args []ssa.Value) (llvm.Value, error) {
	num, _ := constant.Uint64Val(args[0].(*ssa.Const).Value)
	llvmArgs := []llvm.Value{}
	argTypes := []llvm.Type{}
	asm := "svc #" + strconv.FormatUint(num, 10)
	constraints := "={x0}"
	for i, arg := range args[1:] {
		arg = arg.(*ssa.MakeInterface).X
		if i == 0 {
			constraints += ",0"
		} else {
			constraints += ",{x" + strconv.Itoa(i) + "}"
		}
		llvmValue := b.getValue(arg)
		llvmArgs = append(llvmArgs, llvmValue)
		argTypes = append(argTypes, llvmValue.Type())
	}
	// Implement the ARM64 calling convention by marking x1-x7 as
	// clobbered. x0 is used as an output register so doesn't have to be
	// marked as clobbered.
	constraints += ",~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7}"
	fnType := llvm.FunctionType(b.uintptrType, argTypes, false)
	target := llvm.InlineAsm(fnType, asm, constraints, true, false, 0, false)
	return b.CreateCall(target, llvmArgs, ""), nil
}

// This is a compiler builtin which emits CSR instructions. It can be one of:
//
//     func (csr CSR) Get() uintptr
//     func (csr CSR) Set(uintptr)
//     func (csr CSR) SetBits(uintptr) uintptr
//     func (csr CSR) ClearBits(uintptr) uintptr
//
// The csr parameter (method receiver) must be a constant. Other parameter can
// be any value.
func (b *builder) emitCSROperation(call *ssa.CallCommon) (llvm.Value, error) {
	csrConst, ok := call.Args[0].(*ssa.Const)
	if !ok {
		return llvm.Value{}, b.makeError(call.Pos(), "CSR must be constant")
	}
	csr := csrConst.Uint64()
	switch name := call.StaticCallee().Name(); name {
	case "Get":
		// Note that this instruction may have side effects, and thus must be
		// marked as such.
		fnType := llvm.FunctionType(b.uintptrType, nil, false)
		asm := fmt.Sprintf("csrr $0, %d", csr)
		target := llvm.InlineAsm(fnType, asm, "=r", true, false, 0, false)
		return b.CreateCall(target, nil, ""), nil
	case "Set":
		fnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{b.uintptrType}, false)
		asm := fmt.Sprintf("csrw %d, $0", csr)
		target := llvm.InlineAsm(fnType, asm, "r", true, false, 0, false)
		return b.CreateCall(target, []llvm.Value{b.getValue(call.Args[1])}, ""), nil
	case "SetBits":
		// Note: it may be possible to optimize this to csrrsi in many cases.
		fnType := llvm.FunctionType(b.uintptrType, []llvm.Type{b.uintptrType}, false)
		asm := fmt.Sprintf("csrrs $0, %d, $1", csr)
		target := llvm.InlineAsm(fnType, asm, "=r,r", true, false, 0, false)
		return b.CreateCall(target, []llvm.Value{b.getValue(call.Args[1])}, ""), nil
	case "ClearBits":
		// Note: it may be possible to optimize this to csrrci in many cases.
		fnType := llvm.FunctionType(b.uintptrType, []llvm.Type{b.uintptrType}, false)
		asm := fmt.Sprintf("csrrc $0, %d, $1", csr)
		target := llvm.InlineAsm(fnType, asm, "=r,r", true, false, 0, false)
		return b.CreateCall(target, []llvm.Value{b.getValue(call.Args[1])}, ""), nil
	default:
		return llvm.Value{}, b.makeError(call.Pos(), "unknown CSR operation: "+name)
	}
}
