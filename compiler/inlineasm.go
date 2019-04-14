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

// This is a compiler builtin, which reads the given register by name:
//
//     func ReadRegister(name string) uintptr
//
// The register name must be a constant, for example "sp".
func (c *Compiler) emitReadRegister(args []ssa.Value) (llvm.Value, error) {
	fnType := llvm.FunctionType(c.uintptrType, []llvm.Type{}, false)
	regname := constant.StringVal(args[0].(*ssa.Const).Value)
	target := llvm.InlineAsm(fnType, "mov $0, "+regname, "=r", false, false, 0)
	return c.builder.CreateCall(target, nil, ""), nil
}

// This is a compiler builtin, which emits a piece of inline assembly with no
// operands or return values. It is useful for trivial instructions, like wfi in
// ARM or sleep in AVR.
//
//     func Asm(asm string)
//
// The provided assembly must be a constant.
func (c *Compiler) emitAsm(args []ssa.Value) (llvm.Value, error) {
	// Magic function: insert inline assembly instead of calling it.
	fnType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{}, false)
	asm := constant.StringVal(args[0].(*ssa.Const).Value)
	target := llvm.InlineAsm(fnType, asm, "", true, false, 0)
	return c.builder.CreateCall(target, nil, ""), nil
}

// This is a compiler builtin, which allows assembly to be called in a flexible
// way.
//
//     func AsmFull(asm string, regs map[string]interface{})
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
func (c *Compiler) emitAsmFull(frame *Frame, instr *ssa.CallCommon) (llvm.Value, error) {
	asmString := constant.StringVal(instr.Args[0].(*ssa.Const).Value)
	registers := map[string]llvm.Value{}
	registerMap := instr.Args[1].(*ssa.MakeMap)
	for _, r := range *registerMap.Referrers() {
		switch r := r.(type) {
		case *ssa.DebugRef:
			// ignore
		case *ssa.MapUpdate:
			if r.Block() != registerMap.Block() {
				return llvm.Value{}, c.makeError(instr.Pos(), "register value map must be created in the same basic block")
			}
			key := constant.StringVal(r.Key.(*ssa.Const).Value)
			//println("value:", r.Value.(*ssa.MakeInterface).X.String())
			value, err := c.parseExpr(frame, r.Value.(*ssa.MakeInterface).X)
			if err != nil {
				return llvm.Value{}, err
			}
			registers[key] = value
		case *ssa.Call:
			if r.Common() == instr {
				break
			}
		default:
			return llvm.Value{}, c.makeError(instr.Pos(), "don't know how to handle argument to inline assembly: "+r.String())
		}
	}
	// TODO: handle dollar signs in asm string
	registerNumbers := map[string]int{}
	var err error
	argTypes := []llvm.Type{}
	args := []llvm.Value{}
	constraints := []string{}
	asmString = regexp.MustCompile("\\{[a-zA-Z]+\\}").ReplaceAllStringFunc(asmString, func(s string) string {
		// TODO: skip strings like {r4} etc. that look like ARM push/pop
		// instructions.
		name := s[1 : len(s)-1]
		if _, ok := registers[name]; !ok {
			if err == nil {
				err = c.makeError(instr.Pos(), "unknown register name: "+name)
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
				err = c.makeError(instr.Pos(), "unknown type in inline assembly for value: "+name)
				return s
			}
		}
		return fmt.Sprintf("${%v}", registerNumbers[name])
	})
	if err != nil {
		return llvm.Value{}, err
	}
	fnType := llvm.FunctionType(c.ctx.VoidType(), argTypes, false)
	target := llvm.InlineAsm(fnType, asmString, strings.Join(constraints, ","), true, false, 0)
	return c.builder.CreateCall(target, args, ""), nil
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
func (c *Compiler) emitSVCall(frame *Frame, args []ssa.Value) (llvm.Value, error) {
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
		llvmValue, err := c.parseExpr(frame, arg)
		if err != nil {
			return llvm.Value{}, err
		}
		llvmArgs = append(llvmArgs, llvmValue)
		argTypes = append(argTypes, llvmValue.Type())
	}
	// Implement the ARM calling convention by marking r1-r3 as
	// clobbered. r0 is used as an output register so doesn't have to be
	// marked as clobbered.
	constraints += ",~{r1},~{r2},~{r3}"
	fnType := llvm.FunctionType(c.uintptrType, argTypes, false)
	target := llvm.InlineAsm(fnType, asm, constraints, true, false, 0)
	return c.builder.CreateCall(target, llvmArgs, ""), nil
}
