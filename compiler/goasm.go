package compiler

import (
	"fmt"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"tinygo.org/x/go-llvm"
)

func (b *builder) createGoAsmWrapper(forwardName string) {
	// Obtain architecture specific things.
	var offsetField types.Type
	var callAsm, callConstraints, stackSubAsm, stackAddAsm string
	var stackAlign int64
	var fr *goasmFrame
	switch b.archFamily() {
	case "i386":
		offsetField = types.NewStruct(nil, nil)
		stackAlign = 16
		stackSubAsm = "subl $$%d, %%esp"
		stackAddAsm = "addl $$%d, %%esp"
		callAsm = "calll %#v"
		callConstraints = "~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory}"
		fr = &goasmFrame{
			builder:      b,
			i8Registers:  []string{"dil", "sil", "dl", "cl", "al", "bl", "bpl"},
			i16Registers: []string{"di", "si", "dx", "cx", "ax", "bx", "bp"},
			i32Registers: []string{"edi", "esi", "edx", "ecx", "eax", "ebx", "ebp"},
			integerLoad: func(size int64, signed bool) string {
				if signed {
					return map[int64]string{
						1: "movsbl %[2]d(%%esp), %%%[1]s",
						2: "movswl %[2]d(%%esp), %%%[1]s",
						4: "movl %[2]d(%%esp), %%%[1]s",
					}[size]
				}
				return map[int64]string{
					1: "movzbl %[2]d(%%esp), %%%[1]s",
					2: "movzwl %[2]d(%%esp), %%%[1]s",
					4: "movl %[2]d(%%esp), %%%[1]s",
				}[size]
			},
			integerStore: func(size int64) string {
				return map[int64]string{
					1: "movb %%%s, %d(%%esp)",
					2: "movw %%%s, %d(%%esp)",
					4: "movl %%%s, %d(%%esp)",
				}[size]
			},
			float32Registers: []string{"xmm0", "xmm1", "xmm2", "xmm3", "xmm4", "xmm5", "xmm6", "xmm7"},
			float64Registers: []string{"xmm0", "xmm1", "xmm2", "xmm3", "xmm4", "xmm5", "xmm6", "xmm7"},
			f32Store:         "movss %%%s, %d(%%esp)",
			f64Store:         "movsd %%%s, %d(%%esp)",
			f32Load:          "movss %[2]d(%%esp), %%%[1]s",
			f64Load:          "movsd %[2]d(%%esp), %%%[1]s",
		}
	case "x86_64":
		offsetField = types.NewStruct(nil, nil)
		stackAlign = 16
		stackSubAsm = "subq $$%d, %%rsp"
		stackAddAsm = "addq $$%d, %%rsp"
		callAsm = "callq %#v"
		callConstraints = "~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory}"
		// Registers are ordered in a way that they should match the System-V
		// ABI for the first few registers. This avoids unnecessary register
		// copies.
		fr = &goasmFrame{
			builder:      b,
			i8Registers:  []string{"dil", "sil", "dl", "cl", "r8b", "r9b", "al", "bl", "bpl", "r10b", "r11b", "r12b", "r13b", "r14b", "r15b"},
			i16Registers: []string{"di", "si", "dx", "cx", "r8w", "r9w", "ax", "bx", "bp", "r10w", "r11w", "r12w", "r13w", "r14w", "r15w"},
			i32Registers: []string{"edi", "esi", "edx", "ecx", "r8d", "r9d", "eax", "ebx", "ebp", "r10d", "r11d", "r12d", "r13d", "r14d", "r15d"},
			i64Registers: []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9", "rax", "rbx", "rbp", "r10", "r11", "r12", "r13", "r14", "r15"},
			integerLoad: func(size int64, signed bool) string {
				if signed {
					return map[int64]string{
						1: "movsbl %[2]d(%%rsp), %%%[1]s",
						2: "movswl %[2]d(%%rsp), %%%[1]s",
						4: "movl %[2]d(%%rsp), %%%[1]s",
						8: "movq %[2]d(%%rsp), %%%[1]s",
					}[size]
				}
				return map[int64]string{
					1: "movzbl %[2]d(%%rsp), %%%[1]s",
					2: "movzwl %[2]d(%%rsp), %%%[1]s",
					4: "movl %[2]d(%%rsp), %%%[1]s",
					8: "movq %[2]d(%%rsp), %%%[1]s",
				}[size]
			},
			integerStore: func(size int64) string {
				return map[int64]string{
					1: "movb %%%s, %d(%%rsp)",
					2: "movw %%%s, %d(%%rsp)",
					4: "movl %%%s, %d(%%rsp)",
					8: "movq %%%s, %d(%%rsp)",
				}[size]
			},
			float32Registers: []string{"xmm0", "xmm1", "xmm2", "xmm3", "xmm4", "xmm5", "xmm6", "xmm7", "xmm8", "xmm9", "xmm10", "xmm11", "xmm12", "xmm13", "xmm14", "xmm15", "xmm16", "xmm17", "xmm18", "xmm19", "xmm20", "xmm21", "xmm22", "xmm23", "xmm24", "xmm25", "xmm26", "xmm27", "xmm28", "xmm29", "xmm30", "xmm31"},
			float64Registers: []string{"xmm0", "xmm1", "xmm2", "xmm3", "xmm4", "xmm5", "xmm6", "xmm7", "xmm8", "xmm9", "xmm10", "xmm11", "xmm12", "xmm13", "xmm14", "xmm15", "xmm16", "xmm17", "xmm18", "xmm19", "xmm20", "xmm21", "xmm22", "xmm23", "xmm24", "xmm25", "xmm26", "xmm27", "xmm28", "xmm29", "xmm30", "xmm31"},
			f32Store:         "movss %%%s, %d(%%rsp)",
			f64Store:         "movsd %%%s, %d(%%rsp)",
			f32Load:          "movss %[2]d(%%rsp), %%%[1]s",
			f64Load:          "movsd %[2]d(%%rsp), %%%[1]s",
		}
	case "arm":
		// Storing and reloading R11 because it appears that R11 is reserved on
		// Linux. We store it in the area that's empty anyway (it seems to be
		// used for the return pointer).
		offsetField = types.Typ[types.Uintptr]
		stackAlign = 8
		stackSubAsm = "sub sp, sp, #%d"
		stackAddAsm = "add sp, sp, #%d"
		callAsm = "str r11, [sp]\n\tbl %#v\n\tldr r11, [sp]"
		callConstraints = "~{cpsr},~{memory}"
		fr = &goasmFrame{
			builder:      b,
			i32Registers: []string{"r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7", "r8", "r9", "r10", "r12", "lr"},
			integerLoad: func(size int64, signed bool) string {
				if signed {
					return map[int64]string{
						1: "ldrsb %s, [sp, #%d]",
						2: "ldrsh %s, [sp, #%d]",
						4: "ldr %s, [sp, #%d]",
					}[size]
				}
				return map[int64]string{
					1: "ldrb %s, [sp, #%d]",
					2: "ldrh %s, [sp, #%d]",
					4: "ldr %s, [sp, #%d]",
				}[size]
			},
			integerStore: func(size int64) string {
				return map[int64]string{
					1: "ldrsb %s, [sp, #%d]",
					2: "ldrsh %s, [sp, #%d]",
					4: "ldr %s, [sp, #%d]",
				}[size]
			},
			float32Registers: []string{"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9", "s10", "s11", "s12", "s13", "s14", "s15"},
			float64Registers: []string{"d0", "d1", "d2", "d3", "d4", "d5", "d6", "d7", "d8", "d9", "d10", "d11", "d12", "d13", "d14", "d15"},
			f32Store:         "vstr %s, [sp, #%d]",
			f64Store:         "vstr %s, [sp, #%d]",
			f32Load:          "vldr %s, [sp, #%d]",
			f64Load:          "vldr %s, [sp, #%d]",
		}
	case "aarch64":
		offsetField = types.Typ[types.Uintptr]
		stackAlign = 16
		stackSubAsm = "sub sp, sp, #%d"
		stackAddAsm = "add sp, sp, #%d"
		callAsm = "bl %#v"
		callConstraints = "~{x0},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{lr},~{nzcv},~{ffr},~{vg},~{memory}"
		fr = &goasmFrame{
			builder:      b,
			i32Registers: []string{"w0", "w1", "w2", "w3", "w4", "w5", "w6", "w7", "w8", "w9", "w10", "w11", "w12", "w13", "w14", "w15", "w16", "w17", "w19", "w20", "w21", "w22", "w23", "w24", "w25", "w26", "w27", "w28"},
			i64Registers: []string{"x0", "x1", "x2", "x3", "x4", "x5", "x6", "x7", "x8", "x9", "x10", "x11", "x12", "x13", "x14", "x15", "x16", "x17", "x19", "x20", "x21", "x22", "x23", "x24", "x25", "x26", "x27", "x28"},
			integerLoad: func(size int64, signed bool) string {
				if signed {
					return map[int64]string{
						1: "ldrsb %s, [sp, #%d]",
						2: "ldrsh %s, [sp, #%d]",
						4: "ldr %s, [sp, #%d]",
						8: "ldr %s, [sp, #%d]",
					}[size]
				}
				return map[int64]string{
					1: "ldrb %s, [sp, #%d]",
					2: "ldrh %s, [sp, #%d]",
					4: "ldr %s, [sp, #%d]",
					8: "ldr %s, [sp, #%d]",
				}[size]
			},
			integerStore: func(size int64) string {
				return map[int64]string{
					1: "ldrsb %s, [sp, #%d]",
					2: "ldrsh %s, [sp, #%d]",
					4: "ldr %s, [sp, #%d]",
					8: "ldr %s, [sp, #%d]",
				}[size]
			},
			float32Registers: []string{"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9", "s10", "s11", "s12", "s13", "s14", "s15", "s16", "s17", "s18", "s19", "s20", "s21", "s22", "s23", "s24", "s25", "s26", "s27", "s28", "s29", "s30"},
			float64Registers: []string{"d0", "d1", "d2", "d3", "d4", "d5", "d6", "d7", "d8", "d9", "d10", "d11", "d12", "d13", "d14", "d15", "d16", "d17", "d18", "d19", "d20", "d21", "d22", "d23", "d24", "d25", "d26", "d27", "d28", "d29", "d30"},
			f32Store:         "str %s, [sp, #%d]",
			f64Store:         "str %s, [sp, #%d]",
			f32Load:          "ldr %s, [sp, #%d]",
			f64Load:          "ldr %s, [sp, #%d]",
		}
		if b.GOOS != "darwin" && b.GOOS != "windows" {
			callConstraints += ",~{x18}"
		}
	default:
		b.addError(b.fn.Pos(), "unknown architecture for Go assembly: "+b.archFamily())
	}

	// Initialize state that is used while assigning registers to parameters.
	fr.integerInputs = make([]bool, len(fr.i32Registers))
	fr.integerOutputs = make([]bool, len(fr.i32Registers))
	fr.floatInputs = make([]bool, len(fr.float32Registers))
	fr.floatOutputs = make([]bool, len(fr.float32Registers))
	fr.sizes = types.SizesFor("gc", b.GOARCH)

	// Initialize parameters, create entry block, etc.
	b.createFunctionStart(true)

	// Determine the stack layout that's used for the Go ABI.
	// The layout roughly follows this convention (from low to high address):
	//   - empty space (for the return pointer?)
	//   - parameters
	//   - return values
	// More information can be found here (ABI0 is equivalent to the regabi
	// without any integer or floating point registers):
	// https://go.googlesource.com/go/+/refs/heads/master/src/cmd/compile/abi-internal.md
	// We need to use size calculations as used by gc (the regular Go compiler)
	// because that's what the assembly expects. It's usually the same as LLVM,
	// but importantly differs on ARM where int64 is 32-bit aligned in gc but
	// 64-bit aligned according to LLVM (and the ARM AAPCS).
	var paramValues []llvm.Value
	var paramFields []*types.Var
	for _, param := range b.fn.Params {
		value := b.getValue(param, getPos(param))
		paramValues = append(paramValues, value)
		paramFields = append(paramFields, types.NewField(token.NoPos, nil, param.Name(), param.Type(), false))
	}
	var resultFields []*types.Var
	results := b.fn.Signature.Results()
	for i := 0; i < results.Len(); i++ {
		field := results.At(i)
		resultFields = append(resultFields, types.NewField(token.NoPos, nil, "result_"+strconv.Itoa(i), field.Type(), false))
	}
	paramStruct := types.NewStruct(paramFields, nil)
	stackStructFields := []*types.Var{
		types.NewField(token.NoPos, nil, "offset", offsetField, false),
		types.NewField(token.NoPos, nil, "params", paramStruct, false),
		types.NewField(token.NoPos, nil, "align", types.NewArray(types.Typ[types.Uintptr], 0), false),
		types.NewField(token.NoPos, nil, "results", types.NewStruct(resultFields, nil), false),
	}
	stackStruct := types.NewStruct(stackStructFields, nil)
	stackStructOffsets := fr.sizes.Offsetsof(stackStructFields)

	// Assign registers to return values.
	resultOffsets := fr.sizes.Offsetsof(resultFields)
	for i := range resultFields {
		offset := stackStructOffsets[2] + resultOffsets[i]
		fr.assignResult(offset, resultFields[i].Type())
	}

	// Assign registers to parameters.
	paramOffsets := fr.sizes.Offsetsof(paramFields)
	for i, param := range paramValues {
		offset := stackStructOffsets[1] + paramOffsets[i]
		fr.assignParameter(offset, param, paramFields[i].Type())
	}

	// Mark all unused registers as clobbered.
	if fr.i64Registers != nil {
		// 64-bit systems. Assume 64-bit registers overlap all 32-bit registers.
		for i, reg := range fr.i64Registers {
			if !fr.integerOutputs[i] {
				callConstraints += fmt.Sprintf(",~{%s}", reg)
			}
		}
	} else {
		// 32-bit systems.
		for i, reg := range fr.i32Registers {
			if !fr.integerOutputs[i] {
				callConstraints += fmt.Sprintf(",~{%s}", reg)
			}
		}
	}
	for i, reg := range fr.float64Registers {
		if !fr.floatOutputs[i] {
			callConstraints += fmt.Sprintf(",~{%s}", reg)
		}
	}

	// Determine the LLVM function signature of the inline assembly.
	var paramTypes []llvm.Type
	for _, value := range fr.asmParams {
		paramTypes = append(paramTypes, value.Type())
	}
	var resultType llvm.Type
	if len(fr.resultTypes) == 0 {
		resultType = b.ctx.VoidType()
	} else if len(fr.resultTypes) == 1 {
		// LLVM doesn't like a single field in a struct return like `{ double }`.
		// Therefore this is special-cased.
		resultType = fr.resultTypes[0]
	} else {
		resultType = b.ctx.StructType(fr.resultTypes, false)
	}
	asmType := llvm.FunctionType(resultType, paramTypes, false)

	// Call the Go assembly!
	// This is done in inline assembly because ABI0 clobbers more registers than
	// a call would in the C calling convention.
	stackStructSize := align(fr.sizes.Sizeof(stackStruct), stackAlign)
	var asms []string
	asms = append(asms, fmt.Sprintf(stackSubAsm, stackStructSize))
	asms = append(asms, fr.stackStoreAsms...)
	asms = append(asms, fmt.Sprintf(callAsm, forwardName))
	asms = append(asms, fr.stackLoadAsms...)
	asms = append(asms, fmt.Sprintf(stackAddAsm, stackStructSize))
	asmString := strings.Join(asms, "\n\t")
	fr.constraints = append(fr.constraints, callConstraints)
	constraintsString := strings.Join(fr.constraints, ",")
	inlineAsm := llvm.InlineAsm(asmType, asmString, constraintsString, true, true, 0, false)
	asm := b.CreateCall(asmType, inlineAsm, fr.asmParams, "")

	// Collect the results from the inline assembly to construct a return value.
	// We need to do this because the return value is just a long list of
	// registers while the function return type may actually be a more complex
	// type. Example: { i16, i32, i32 } vs { i16, { i32, i32 }}.
	var asmResults []llvm.Value
	if len(fr.resultTypes) == 1 {
		asmResults = append(asmResults, asm)
	} else {
		for i := range fr.resultTypes {
			result := fr.CreateExtractValue(asm, i, "")
			asmResults = append(asmResults, result)
		}
	}
	var resultValues []llvm.Value
	var index int
	for _, field := range resultFields {
		result := fr.createResult(field.Type(), asmResults, &index)
		resultValues = append(resultValues, result)
	}

	// Return the resulting value.
	b.createReturn(resultValues)
}

// Struct to keep track of all the information regarding a call to a Go assembly
// function.
type goasmFrame struct {
	*builder

	// The types.Size not for TinyGo, but for regular Go.
	// It differs in some specific cases, and Go assembly assumes it uses the
	// regular Go variant.
	sizes types.Sizes

	asmParams      []llvm.Value // parameters to the inline assembly
	resultTypes    []llvm.Type  // result types of the inline assembly (made a struct if more than one)
	constraints    []string     // inline assembly constraints
	stackStoreAsms []string     // assembly lines to store parameters to the stack
	stackLoadAsms  []string     // assembly lines to load results from the stack

	// Constant architecture details regarding integer registers.
	// It is assumed that each register at the same index overlaps, and others
	// don't. For example, on arm64 w0 and x0 overlap.
	i8Registers  []string
	i16Registers []string
	i32Registers []string
	i64Registers []string
	integerLoad  func(int64, bool) string
	integerStore func(int64) string

	// Slice of the same length as i32Registers above. It is used to keep track
	// of which registers are already assigned as an input or output register.
	integerInputs  []bool
	integerOutputs []bool

	// Constant architecture details regarding floating point registers, just
	// like the integer registers above.
	float32Registers []string
	float64Registers []string
	f32Store         string
	f64Store         string
	f32Load          string
	f64Load          string

	// Slice of the same length as float32Registers above. It is used to keep
	// track of which floating point registers are already assigned as an input
	// or output register.
	floatInputs  []bool
	floatOutputs []bool
}

// Determine for the given parameter which registers will be used for calling
// into the inline assembly and the instructions for storing it to the stack.
// In short, the value is broken down into its primitive values and each
// primitive value is assigned a register.
func (fr *goasmFrame) assignParameter(offset int64, value llvm.Value, typ types.Type) {
	switch typ := typ.Underlying().(type) {
	case *types.Basic:
		switch typ.Kind() {
		case types.String:
			ptrSize := fr.sizes.Sizeof(types.Typ[types.Uintptr])
			ptr := fr.CreateExtractValue(value, 0, "")
			len := fr.CreateExtractValue(value, 1, "")
			fr.assignParameter(offset, ptr, types.NewPointer(types.Typ[types.Byte]))
			fr.assignParameter(offset+ptrSize, len, types.Typ[types.Uintptr])
		case types.Complex64:
			real := fr.CreateExtractValue(value, 0, "")
			imag := fr.CreateExtractValue(value, 1, "")
			fr.assignParameter(offset+0, real, types.Typ[types.Float32])
			fr.assignParameter(offset+4, imag, types.Typ[types.Float32])
		case types.Complex128:
			real := fr.CreateExtractValue(value, 0, "")
			imag := fr.CreateExtractValue(value, 1, "")
			fr.assignParameter(offset+0, real, types.Typ[types.Float64])
			fr.assignParameter(offset+8, imag, types.Typ[types.Float64])
		case types.Float32, types.Float64:
			var reg, inst string
			if typ.Kind() == types.Float32 {
				reg = fr.findRegister(fr.floatInputs, fr.float32Registers)
				inst = fmt.Sprintf(fr.f32Store, reg, offset)
			} else {
				reg = fr.findRegister(fr.floatInputs, fr.float64Registers)
				inst = fmt.Sprintf(fr.f64Store, reg, offset)
			}
			fr.stackStoreAsms = append(fr.stackStoreAsms, inst)
			fr.constraints = append(fr.constraints, fmt.Sprintf("{%s}", reg))
			fr.asmParams = append(fr.asmParams, value)
		default:
			fr.assignIntegerParameter(offset, value, typ)
		}
	case *types.Chan, *types.Map, *types.Pointer:
		fr.assignIntegerParameter(offset, value, typ)
	case *types.Array:
		elemSize := fr.sizes.Sizeof(typ.Elem())
		for i := int64(0); i < typ.Len(); i++ {
			elem := fr.CreateExtractValue(value, int(i), "")
			fr.assignParameter(offset+elemSize*i, elem, typ.Elem())
		}
	case *types.Slice:
		ptrSize := fr.sizes.Sizeof(types.Typ[types.Uintptr])
		// assign buffer
		ptr := fr.CreateExtractValue(value, 0, "")
		fr.assignParameter(offset+ptrSize*0, ptr, types.NewPointer(typ.Elem()))
		// assign len
		len := fr.CreateExtractValue(value, 1, "")
		fr.assignParameter(offset+ptrSize*1, len, types.Typ[types.Uintptr])
		// assign cap
		cap := fr.CreateExtractValue(value, 1, "")
		fr.assignParameter(offset+ptrSize*2, cap, types.Typ[types.Uintptr])
	case *types.Struct:
		var fields []*types.Var
		for i := 0; i < typ.NumFields(); i++ {
			fields = append(fields, typ.Field(i))
		}
		offsets := fr.sizes.Offsetsof(fields)
		for i := 0; i < typ.NumFields(); i++ {
			elem := fr.CreateExtractValue(value, i, "")
			fr.assignParameter(offset+offsets[i], elem, typ.Field(i).Type())
		}
	default:
		panic("unknown type: " + typ.String()) // TODO: handle func and interface gracefully
	}
}

// See assignParameter. This function assigns a primitive (integer or pointer)
// value to a register and creates the instruction to store it to the Go
// parameter area.
func (fr *goasmFrame) assignIntegerParameter(offset int64, value llvm.Value, typ types.Type) {
	var reg string
	switch fr.sizes.Sizeof(typ) {
	case 1:
		if fr.i8Registers != nil { // 386, amd64
			reg = fr.findRegister(fr.integerInputs, fr.i8Registers)
		} else {
			reg = fr.findRegister(fr.integerInputs, fr.i32Registers)
		}
	case 2:
		if fr.i16Registers != nil { // 386, amd64
			reg = fr.findRegister(fr.integerInputs, fr.i16Registers)
		} else {
			reg = fr.findRegister(fr.integerInputs, fr.i32Registers)
		}
	case 4:
		reg = fr.findRegister(fr.integerInputs, fr.i32Registers)
	case 8:
		if fr.i64Registers == nil {
			// Special case for 64-bit values on a 32-bit system.
			lobits := fr.CreateTrunc(value, fr.ctx.Int32Type(), "")
			shr := fr.CreateLShr(value, llvm.ConstInt(fr.ctx.Int64Type(), 32, false), "")
			hibits := fr.CreateTrunc(shr, fr.ctx.Int32Type(), "")
			fr.assignIntegerParameter(offset+0, lobits, types.Typ[types.Uint32])
			fr.assignIntegerParameter(offset+4, hibits, types.Typ[types.Uint32])
			return
		}
		reg = fr.findRegister(fr.integerInputs, fr.i64Registers)
	default:
		panic("unknown parameter size") // should be unreachable
	}
	inst := fmt.Sprintf(fr.integerStore(fr.sizes.Sizeof(typ)), reg, offset)
	fr.stackStoreAsms = append(fr.stackStoreAsms, inst)
	fr.constraints = append(fr.constraints, fmt.Sprintf("{%s}", reg))
	fr.asmParams = append(fr.asmParams, value)
}

// assignResult assings result values to registers and determines the
// instructions to use for loading the result from the Go function stack frame.
// It doesn't reconstruct the resulting value from the inline assembly call,
// this is done in createResult.
func (fr *goasmFrame) assignResult(offset int64, typ types.Type) {
	switch typ := typ.Underlying().(type) {
	case *types.Basic:
		switch typ.Kind() {
		case types.Float32, types.Float64:
			var reg, inst string
			if typ.Kind() == types.Float32 {
				reg = fr.findRegister(fr.floatOutputs, fr.float32Registers)
				inst = fmt.Sprintf(fr.f32Load, reg, offset)
			} else {
				reg = fr.findRegister(fr.floatOutputs, fr.float64Registers)
				inst = fmt.Sprintf(fr.f64Load, reg, offset)
			}
			fr.constraints = append(fr.constraints, fmt.Sprintf("={%s}", reg))
			fr.stackLoadAsms = append(fr.stackLoadAsms, inst)
			fr.resultTypes = append(fr.resultTypes, fr.getLLVMType(typ))
		case types.Complex64:
			fr.assignResult(offset+0, types.Typ[types.Float32])
			fr.assignResult(offset+4, types.Typ[types.Float32])
		case types.Complex128:
			fr.assignResult(offset+0, types.Typ[types.Float64])
			fr.assignResult(offset+8, types.Typ[types.Float64])
		case types.String:
			ptrSize := fr.sizes.Sizeof(types.Typ[types.Uintptr])
			fr.assignResult(offset+ptrSize*0, types.NewPointer(types.Typ[types.Byte])) // buffer
			fr.assignResult(offset+ptrSize*1, types.Typ[types.Uintptr])                // len
		default:
			signed := false
			if typ.Info()&types.IsInteger != 0 {
				signed = typ.Info()&types.IsUnsigned == 0
			}
			fr.assignIntegerResult(offset, typ, signed)
		}
	case *types.Chan, *types.Map, *types.Pointer:
		fr.assignIntegerResult(offset, typ, false)
	case *types.Array:
		elemSize := fr.sizes.Sizeof(typ.Elem())
		for i := int64(0); i < typ.Len(); i++ {
			fr.assignResult(offset+elemSize*i, typ.Elem())
		}
	case *types.Slice:
		ptrSize := fr.sizes.Sizeof(types.Typ[types.Uintptr])
		fr.assignResult(offset+ptrSize*0, types.NewPointer(typ.Elem())) // buffer
		fr.assignResult(offset+ptrSize*1, types.Typ[types.Uintptr])     // len
		fr.assignResult(offset+ptrSize*2, types.Typ[types.Uintptr])     // cap
	case *types.Struct:
		var fields []*types.Var
		for i := 0; i < typ.NumFields(); i++ {
			fields = append(fields, typ.Field(i))
		}
		offsets := fr.sizes.Offsetsof(fields)
		for i := 0; i < typ.NumFields(); i++ {
			fr.assignResult(offset+offsets[i], typ.Field(i).Type())
		}
	default:
		panic("unknown type: " + typ.String())
	}
}

func (fr *goasmFrame) assignIntegerResult(offset int64, typ types.Type, signed bool) {
	var reg string
	switch fr.sizes.Sizeof(typ) {
	case 1, 2, 4:
		reg = fr.findRegister(fr.integerOutputs, fr.i32Registers)
	case 8:
		if fr.i64Registers == nil {
			// For 32-bit architectures.
			// Split up the value into two 32-bit registers.
			fr.assignIntegerResult(offset+0, types.Typ[types.Uint32], false)
			fr.assignIntegerResult(offset+4, types.Typ[types.Uint32], false)
			return
		}
		reg = fr.findRegister(fr.integerOutputs, fr.i64Registers)
	default:
		panic("unknown parameter size") // should be unreachable
	}
	inst := fmt.Sprintf(fr.integerLoad(fr.sizes.Sizeof(typ), signed), reg, offset)
	fr.constraints = append(fr.constraints, fmt.Sprintf("={%s}", reg))
	fr.stackLoadAsms = append(fr.stackLoadAsms, inst)
	fr.resultTypes = append(fr.resultTypes, fr.getLLVMType(typ))
}

// createResult constructs a proper return value from the list of results from
// the assembly expression: it converts a list of values in registers back to
// structs and such.
func (fr *goasmFrame) createResult(typ types.Type, asmResults []llvm.Value, index *int) (result llvm.Value) {
	origType := typ
	switch typ := typ.Underlying().(type) {
	case *types.Basic:
		switch typ.Kind() {
		case types.Complex64, types.Complex128:
			real := asmResults[*index+0]
			imag := asmResults[*index+1]
			result = llvm.ConstNull(fr.getLLVMType(typ))
			result = fr.CreateInsertValue(result, real, 0, "")
			result = fr.CreateInsertValue(result, imag, 1, "")
			*index += 2
		case types.String:
			buf := asmResults[*index+0]
			len := asmResults[*index+1]
			result = llvm.ConstNull(fr.getLLVMType(typ))
			result = fr.CreateInsertValue(result, buf, 0, "")
			result = fr.CreateInsertValue(result, len, 1, "")
			*index += 2
		default:
			if fr.i64Registers == nil && typ.Info()&types.IsInteger != 0 && fr.sizes.Sizeof(typ) == 8 {
				// 64-bit value on a 32-bit system. We need to join the two
				// values together again.
				lobits := asmResults[*index+0]
				lobits = fr.CreateZExt(lobits, fr.ctx.Int64Type(), "")
				hibits := asmResults[*index+1]
				hibits = fr.CreateZExt(hibits, fr.ctx.Int64Type(), "")
				hibits = fr.CreateShl(hibits, llvm.ConstInt(fr.ctx.Int64Type(), 32, false), "")
				result = fr.CreateOr(lobits, hibits, "")
				*index += 2
				return
			}
			result = asmResults[*index]
			*index++
		}
	case *types.Array:
		result = llvm.ConstNull(fr.getLLVMType(typ))
		for i := 0; i < int(typ.Len()); i++ {
			elem := fr.createResult(typ.Elem(), asmResults, index)
			result = fr.CreateInsertValue(result, elem, i, "")
		}
	case *types.Chan, *types.Map, *types.Pointer:
		result = asmResults[*index]
		*index++
	case *types.Slice:
		buf := asmResults[*index+0]
		len := asmResults[*index+1]
		cap := asmResults[*index+2]
		result = llvm.ConstNull(fr.getLLVMType(typ))
		result = fr.CreateInsertValue(result, buf, 0, "")
		result = fr.CreateInsertValue(result, len, 1, "")
		result = fr.CreateInsertValue(result, cap, 2, "")
		*index += 3
	case *types.Struct:
		result = llvm.ConstNull(fr.getLLVMType(origType))
		for i := 0; i < typ.NumFields(); i++ {
			elem := fr.createResult(typ.Field(i).Type(), asmResults, index)
			result = fr.CreateInsertValue(result, elem, i, "")
		}
	default:
		panic("unknown type: " + typ.String())
	}
	return
}

func (fr *goasmFrame) findRegister(used []bool, regs []string) string {
	for i, r := range regs {
		if used[i] {
			continue
		}
		used[i] = true
		return r
	}
	// Ran out of registers.
	// TODO: handle this correctly by using an alloca instead as a fallback.
	panic(fmt.Sprintf("todo: ran out of registers (all %d registers in use)", len(regs)))
}

func (b *builder) createGoAsmExport(forwardName string) {
	// Obtain some information about the target.
	stackAlignment := uint64(0)
	switch b.archFamily() {
	case "x86_64":
		// Go uses an 8-byte stack alignment. Increase this to a 16-byte
		// alignment.
		stackAlignment = 16
	case "arm":
		// Go appears to use a 4-byte stack alignment. The AAPCS requires an
		// 8-byte alignment, so increase the alignment to 8 bytes.
		stackAlignment = 8
	default:
		// - 386: not sure, need to check this.
		// - arm64: always uses a 16-byte stack alignment
	}

	// Create function that reads incoming parameters from the stack.
	llvmFnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{b.dataPtrType}, false)
	llvmFn := llvm.AddFunction(b.mod, b.info.linkName+"$goasmwrapper", llvmFnType)
	llvmFn.SetLinkage(llvm.InternalLinkage)
	noinline := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("noinline"), 0)
	llvmFn.AddFunctionAttr(noinline)
	if stackAlignment != 0 {
		alignstack := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("alignstack"), stackAlignment)
		llvmFn.AddFunctionAttr(alignstack)
	}
	b.addStandardDeclaredAttributes(llvmFn)
	b.addStandardDefinedAttributes(llvmFn)
	bb := llvm.AddBasicBlock(llvmFn, "entry")
	b.SetInsertPointAtEnd(bb)
	if b.Debug {
		pos := b.program.Fset.Position(b.fn.Pos())
		difunc := b.attachDebugInfoRaw(b.fn, llvmFn, "$goasmwrapper", pos.Filename, pos.Line)
		b.SetCurrentDebugLocation(uint(pos.Line), 0, difunc, llvm.Metadata{})
	}

	// Determine the stack layout that's used for the Go ABI.
	// See createGoAsmWrapper for details.
	sizes := types.SizesFor("gc", b.GOARCH)
	var paramFields []*types.Var
	for _, param := range b.fn.Params {
		paramFields = append(paramFields, types.NewField(token.NoPos, nil, param.Name(), param.Type(), false))
	}
	var resultFields []*types.Var
	for i := 0; i < b.fn.Signature.Results().Len(); i++ {
		resultFields = append(resultFields, b.fn.Signature.Results().At(i))
	}
	paramStruct := types.NewStruct(paramFields, nil)
	stackStructFields := []*types.Var{
		types.NewField(token.NoPos, nil, "offset", types.Typ[types.Uintptr], false),
		types.NewField(token.NoPos, nil, "params", paramStruct, false),
		types.NewField(token.NoPos, nil, "align", types.NewArray(types.Typ[types.Uintptr], 0), false),
		types.NewField(token.NoPos, nil, "results", types.NewStruct(resultFields, nil), false),
	}
	stackStructOffsets := sizes.Offsetsof(stackStructFields)

	// Read parameters from the stack.
	sp := llvmFn.Param(0)
	var params []llvm.Value
	paramOffsets := sizes.Offsetsof(paramFields)
	for i, param := range paramFields {
		offset := stackStructOffsets[1] + paramOffsets[i]
		value := b.loadUsingGoLayout(b.fn.Pos(), sp, sizes, offset, param.Type())
		params = append(params, value)
	}

	// Call the Go function!
	params = append(params, llvm.ConstNull(b.dataPtrType)) // context
	resultValue := b.createCall(b.llvmFnType, b.llvmFn, params, "result")

	// Split the result value into a slice, to match resultFields.
	var resultValues []llvm.Value
	if len(resultFields) == 1 {
		resultValues = []llvm.Value{resultValue}
	} else if len(resultFields) > 1 {
		for i := range resultFields {
			value := b.CreateExtractValue(resultValue, i, "")
			resultValues = append(resultValues, value)
		}
	}

	// Store the result in the stack space reserved by the Go assembly.
	resultOffsets := sizes.Offsetsof(resultFields)
	for i, result := range resultFields {
		offset := stackStructOffsets[2] + resultOffsets[i]
		b.storeUsingGoLayout(b.fn.Pos(), sp, sizes, offset, result.Type(), resultValues[i])
	}

	// Values are returned by passing them in a special way on the stack, not
	// via a conventional return.
	b.CreateRetVoid()

	// TODO: use llvm.sponentry when available (ARM and AArch64 as of LLVM 15).
	b.createGoAsmSPForward(forwardName, llvmFnType, llvmFn)
}

// Create a stub function that captures the stack pointer and passes it to
// another function.
// We should really be using llvm.sponentry when available, or even port it to
// new architectures as needed.
func (b *builder) createGoAsmSPForward(forwardName string, forwardFunctionType llvm.Type, forwardFunction llvm.Value) {
	// Create function that forwards the stack pointer.
	llvmFnType := llvm.FunctionType(b.ctx.VoidType(), nil, false)
	llvmFn := llvm.AddFunction(b.mod, forwardName, llvmFnType)
	b.addStandardDeclaredAttributes(llvmFn)
	b.addStandardDefinedAttributes(llvmFn)
	bb := llvm.AddBasicBlock(llvmFn, "entry")
	b.Builder.SetInsertPointAtEnd(bb)
	if b.Debug {
		pos := b.program.Fset.Position(b.fn.Pos())
		b.SetCurrentDebugLocation(uint(pos.Line), 0, b.difunc, llvm.Metadata{})
	}
	asmType := llvm.FunctionType(b.dataPtrType, nil, false)
	var asmString string
	switch b.archFamily() {
	case "x86_64":
		asmString = "mov %rsp, $0"
	case "arm":
		asmString = "mov r0, sp"
	case "aarch64":
		asmString = "mov x0, sp"
	default:
		b.addError(b.fn.Pos(), "cannot wrap Go assembly: unknown architecture")
		b.CreateUnreachable()
		return
	}
	asm := llvm.InlineAsm(asmType, asmString, "=r", false, false, 0, false)
	sp := b.CreateCall(asmType, asm, nil, "sp")
	call := b.CreateCall(forwardFunctionType, forwardFunction, []llvm.Value{sp}, "")
	call.SetTailCall(true)
	b.CreateRetVoid()

	// Make sure no prologue/epilogue is created (that would change the stack
	// pointer).
	naked := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("naked"), 0)
	llvmFn.AddFunctionAttr(naked)
}

// Load a value from the given pointer with the given offset, assuming the
// memory layout in the sizes parameter.
func (b *builder) loadUsingGoLayout(pos token.Pos, ptr llvm.Value, sizes types.Sizes, offset int64, typ types.Type) llvm.Value {
	typ = typ.Underlying()
	switch typ := typ.(type) {
	case *types.Basic, *types.Pointer, *types.Slice:
		gep := b.CreateGEP(b.ctx.Int8Type(), ptr, []llvm.Value{llvm.ConstInt(b.ctx.Int32Type(), uint64(offset), false)}, "")
		valueType := b.getLLVMType(typ)
		return b.CreateLoad(valueType, gep, "")
	default:
		b.addError(pos, "todo: unknown type to load: "+typ.String())
		return llvm.Undef(b.getLLVMType(typ))
	}
}

// Store a value at the address given by ptr with the given offset, assuming the
// memory layout in the sizes parameter. The ptr must be of type *i8.
func (b *builder) storeUsingGoLayout(pos token.Pos, ptr llvm.Value, sizes types.Sizes, offset int64, typ types.Type, value llvm.Value) {
	typ = typ.Underlying()
	switch typ := typ.(type) {
	case *types.Basic, *types.Pointer, *types.Slice:
		gep := b.CreateGEP(b.ctx.Int8Type(), ptr, []llvm.Value{llvm.ConstInt(b.ctx.Int32Type(), uint64(offset), false)}, "")
		b.CreateStore(value, gep)
	default:
		b.addError(pos, "todo: unknown type to store: "+typ.String())
	}
}
