package interp

// This file compiles the LLVM IR to a form that's easy to efficiently
// interpret.

import (
	"strings"

	"tinygo.org/x/go-llvm"
)

// A function is a compiled LLVM function, which means that interpreting it
// avoids most CGo calls necessary. This is done in a separate step so the
// result can be cached.
// Functions are in SSA form, just like the LLVM version if it. The first block
// (blocks[0]) is the entry block.
type function struct {
	llvmFn llvm.Value
	name   string       // precalculated llvmFn.Name()
	params []llvm.Value // precalculated llvmFn.Params()
	blocks []*basicBlock
	locals map[llvm.Value]int
}

// basicBlock represents a LLVM basic block and contains a slice of
// instructions. The last instruction must be a terminator instruction.
type basicBlock struct {
	phiNodes     []instruction
	instructions []instruction
}

// instruction is a precompiled LLVM IR instruction. The operands can be either
// an already known value (such as literalValue or pointerValue) but can also be
// the special localValue, which means that the value is a function parameter or
// is produced by another instruction in the function. In that case, the
// interpreter will replace the operand with that local value.
type instruction struct {
	opcode     llvm.Opcode
	localIndex int
	operands   []value
	llvmInst   llvm.Value
	name       string
}

// String returns a nice human-readable version of this instruction.
func (inst *instruction) String() string {
	operands := make([]string, len(inst.operands))
	for i, op := range inst.operands {
		operands[i] = op.String()
	}

	name := ""
	if int(inst.opcode) < len(instructionNameMap) {
		name = instructionNameMap[inst.opcode]
	}
	if name == "" {
		name = "<unknown op>"
	}
	return name + " " + strings.Join(operands, " ")
}

// compileFunction compiles a given LLVM function to an easier to interpret
// version of the function. As far as possible, all operands are preprocessed so
// that the interpreter doesn't have to call into LLVM.
func (r *runner) compileFunction(llvmFn llvm.Value) *function {
	fn := &function{
		llvmFn: llvmFn,
		name:   llvmFn.Name(),
		params: llvmFn.Params(),
		locals: make(map[llvm.Value]int),
	}
	if llvmFn.IsDeclaration() {
		// Nothing to do.
		return fn
	}

	for i, param := range fn.params {
		fn.locals[param] = i
	}

	// Make a map of all the blocks, to quickly find the block number for a
	// given branch instruction.
	blockIndices := make(map[llvm.Value]int)
	for llvmBB := llvmFn.FirstBasicBlock(); !llvmBB.IsNil(); llvmBB = llvm.NextBasicBlock(llvmBB) {
		index := len(blockIndices)
		blockIndices[llvmBB.AsValue()] = index
	}

	// Compile every block.
	for llvmBB := llvmFn.FirstBasicBlock(); !llvmBB.IsNil(); llvmBB = llvm.NextBasicBlock(llvmBB) {
		bb := &basicBlock{}
		fn.blocks = append(fn.blocks, bb)

		// Compile every instruction in the block.
		for llvmInst := llvmBB.FirstInstruction(); !llvmInst.IsNil(); llvmInst = llvm.NextInstruction(llvmInst) {
			// Create instruction skeleton.
			opcode := llvmInst.InstructionOpcode()
			inst := instruction{
				opcode:     opcode,
				localIndex: len(fn.locals),
				llvmInst:   llvmInst,
			}
			fn.locals[llvmInst] = len(fn.locals)

			// Add operands specific for this instruction.
			switch opcode {
			case llvm.Ret:
				// Return instruction, which can either be a `ret void` (no
				// return value) or return a value.
				numOperands := llvmInst.OperandsCount()
				if numOperands != 0 {
					inst.operands = []value{
						r.getValue(llvmInst.Operand(0)),
					}
				}
			case llvm.Br:
				// Branch instruction. Can be either a conditional branch (with
				// 3 operands) or unconditional branch (with just one basic
				// block operand).
				numOperands := llvmInst.OperandsCount()
				switch numOperands {
				case 3:
					// Conditional jump to one of two blocks. Comparable to an
					// if/else in procedural languages.
					thenBB := llvmInst.Operand(2)
					elseBB := llvmInst.Operand(1)
					inst.operands = []value{
						r.getValue(llvmInst.Operand(0)),
						literalValue{uint32(blockIndices[thenBB])},
						literalValue{uint32(blockIndices[elseBB])},
					}
				case 1:
					// Unconditional jump to a target basic block. Comparable to
					// a jump in C and Go.
					jumpBB := llvmInst.Operand(0)
					inst.operands = []value{
						literalValue{uint32(blockIndices[jumpBB])},
					}
				default:
					panic("unknown number of operands")
				}
			case llvm.Switch:
				// A switch is an array of (value, label) pairs, of which the
				// first one indicates the to-switch value and the default
				// label.
				numOperands := llvmInst.OperandsCount()
				for i := 0; i < numOperands; i += 2 {
					inst.operands = append(inst.operands, r.getValue(llvmInst.Operand(i)))
					inst.operands = append(inst.operands, literalValue{uint32(blockIndices[llvmInst.Operand(i+1)])})
				}
			case llvm.PHI:
				inst.name = llvmInst.Name()
				incomingCount := inst.llvmInst.IncomingCount()
				for i := 0; i < incomingCount; i++ {
					incomingBB := inst.llvmInst.IncomingBlock(i)
					incomingValue := inst.llvmInst.IncomingValue(i)
					inst.operands = append(inst.operands,
						literalValue{uint32(blockIndices[incomingBB.AsValue()])},
						r.getValue(incomingValue),
					)
				}
			case llvm.Select:
				// Select is a special instruction that is much like a ternary
				// operator. It produces operand 1 or 2 based on the boolean
				// that is operand 0.
				inst.name = llvmInst.Name()
				inst.operands = []value{
					r.getValue(llvmInst.Operand(0)),
					r.getValue(llvmInst.Operand(1)),
					r.getValue(llvmInst.Operand(2)),
				}
			case llvm.Call:
				// Call is a regular function call but could also be a runtime
				// intrinsic. Some runtime intrinsics are treated specially by
				// the interpreter, such as runtime.alloc. We don't
				// differentiate between them here because these calls may also
				// need to be run at runtime, in which case they should all be
				// created in the same way.
				llvmCalledValue := llvmInst.CalledValue()
				if !llvmCalledValue.IsAFunction().IsNil() {
					name := llvmCalledValue.Name()
					if name == "llvm.dbg.value" || strings.HasPrefix(name, "llvm.lifetime.") {
						// These intrinsics should not be interpreted, they are not
						// relevant to the execution of this function.
						continue
					}
				}
				inst.name = llvmInst.Name()
				numOperands := llvmInst.OperandsCount()
				inst.operands = append(inst.operands, r.getValue(llvmCalledValue))
				for i := 0; i < numOperands-1; i++ {
					inst.operands = append(inst.operands, r.getValue(llvmInst.Operand(i)))
				}
			case llvm.Load:
				// Load instruction. The interpreter will load from the
				// appropriate memory view.
				// Also provide the memory size to be loaded, which is necessary
				// with a lack of type information.
				inst.name = llvmInst.Name()
				inst.operands = []value{
					r.getValue(llvmInst.Operand(0)),
					literalValue{r.targetData.TypeAllocSize(llvmInst.Type())},
				}
			case llvm.Store:
				// Store instruction. The interpreter will create a new object
				// in the memory view of the function invocation and store to
				// that, to make it possible to roll back this store.
				inst.operands = []value{
					r.getValue(llvmInst.Operand(0)),
					r.getValue(llvmInst.Operand(1)),
				}
			case llvm.Alloca:
				// Alloca allocates stack space for local variables.
				numElements := r.getValue(inst.llvmInst.Operand(0)).(literalValue).value.(uint32)
				elementSize := r.targetData.TypeAllocSize(inst.llvmInst.AllocatedType())
				inst.operands = []value{
					literalValue{elementSize * uint64(numElements)},
				}
			case llvm.GetElementPtr:
				// GetElementPtr does pointer arithmetic.
				inst.name = llvmInst.Name()
				ptr := llvmInst.Operand(0)
				n := llvmInst.OperandsCount()
				elementType := llvmInst.GEPSourceElementType()
				// gep: [source ptr, dest value size, pairs of indices...]
				inst.operands = []value{
					r.getValue(ptr),
					r.getValue(llvmInst.Operand(1)),
					literalValue{r.targetData.TypeAllocSize(elementType)},
				}
				for i := 2; i < n; i++ {
					operand := r.getValue(llvmInst.Operand(i))
					switch elementType.TypeKind() {
					case llvm.StructTypeKind:
						index := operand.(literalValue).value.(uint32)
						elementOffset := r.targetData.ElementOffset(elementType, int(index))
						// Encode operands in a special way. The elementOffset
						// is just the offset in bytes. The elementSize is a
						// negative number (when cast to a int64) by flipping
						// all the bits. This allows the interpreter to detect
						// this is a struct field and that it should not
						// multiply it with the elementOffset to get the offset.
						// It is important for the interpreter to know the
						// struct field index for when the GEP must be done at
						// runtime.
						inst.operands = append(inst.operands, literalValue{elementOffset}, literalValue{^uint64(index)})
						elementType = elementType.StructElementTypes()[index]
					case llvm.ArrayTypeKind:
						elementType = elementType.ElementType()
						elementSize := r.targetData.TypeAllocSize(elementType)
						elementSizeOperand := literalValue{elementSize}
						// Add operand * elementSizeOperand bytes to the pointer.
						inst.operands = append(inst.operands, operand, elementSizeOperand)
					default:
						// This should be unreachable.
						panic("unknown type: " + elementType.String())
					}
				}
			case llvm.BitCast, llvm.IntToPtr, llvm.PtrToInt:
				// Bitcasts are usually used to cast a pointer from one type to
				// another leaving the pointer itself intact.
				inst.name = llvmInst.Name()
				inst.operands = []value{
					r.getValue(llvmInst.Operand(0)),
				}
			case llvm.ExtractValue:
				inst.name = llvmInst.Name()
				agg := llvmInst.Operand(0)
				var offset uint64
				indexingType := agg.Type()
				for _, index := range inst.llvmInst.Indices() {
					switch indexingType.TypeKind() {
					case llvm.StructTypeKind:
						offset += r.targetData.ElementOffset(indexingType, int(index))
						indexingType = indexingType.StructElementTypes()[index]
					case llvm.ArrayTypeKind:
						indexingType = indexingType.ElementType()
						elementSize := r.targetData.TypeAllocSize(indexingType)
						offset += elementSize * uint64(index)
					default:
						panic("unknown type kind") // unreachable
					}
				}
				size := r.targetData.TypeAllocSize(inst.llvmInst.Type())
				// extractvalue [agg, byteOffset, byteSize]
				inst.operands = []value{
					r.getValue(agg),
					literalValue{offset},
					literalValue{size},
				}
			case llvm.InsertValue:
				inst.name = llvmInst.Name()
				agg := llvmInst.Operand(0)
				var offset uint64
				indexingType := agg.Type()
				for _, index := range inst.llvmInst.Indices() {
					switch indexingType.TypeKind() {
					case llvm.StructTypeKind:
						offset += r.targetData.ElementOffset(indexingType, int(index))
						indexingType = indexingType.StructElementTypes()[index]
					case llvm.ArrayTypeKind:
						indexingType = indexingType.ElementType()
						elementSize := r.targetData.TypeAllocSize(indexingType)
						offset += elementSize * uint64(index)
					default:
						panic("unknown type kind") // unreachable
					}
				}
				// insertvalue [agg, elt, byteOffset]
				inst.operands = []value{
					r.getValue(agg),
					r.getValue(llvmInst.Operand(1)),
					literalValue{offset},
				}
			case llvm.ICmp:
				inst.name = llvmInst.Name()
				inst.operands = []value{
					r.getValue(llvmInst.Operand(0)),
					r.getValue(llvmInst.Operand(1)),
					literalValue{uint8(llvmInst.IntPredicate())},
				}
			case llvm.FCmp:
				inst.name = llvmInst.Name()
				inst.operands = []value{
					r.getValue(llvmInst.Operand(0)),
					r.getValue(llvmInst.Operand(1)),
					literalValue{uint8(llvmInst.FloatPredicate())},
				}
			case llvm.Add, llvm.Sub, llvm.Mul, llvm.UDiv, llvm.SDiv, llvm.URem, llvm.SRem, llvm.Shl, llvm.LShr, llvm.AShr, llvm.And, llvm.Or, llvm.Xor:
				// Integer binary operations.
				inst.name = llvmInst.Name()
				inst.operands = []value{
					r.getValue(llvmInst.Operand(0)),
					r.getValue(llvmInst.Operand(1)),
				}
			case llvm.SExt, llvm.ZExt, llvm.Trunc:
				// Extend or shrink an integer size.
				// No sign extension going on so easy to do.
				// zext: [value, bitwidth]
				// trunc: [value, bitwidth]
				inst.name = llvmInst.Name()
				inst.operands = []value{
					r.getValue(llvmInst.Operand(0)),
					literalValue{uint64(llvmInst.Type().IntTypeWidth())},
				}
			case llvm.SIToFP, llvm.UIToFP:
				// Convert an integer to a floating point instruction.
				// opcode: [value, bitwidth]
				inst.name = llvmInst.Name()
				inst.operands = []value{
					r.getValue(llvmInst.Operand(0)),
					literalValue{uint64(r.targetData.TypeAllocSize(llvmInst.Type()) * 8)},
				}
			default:
				// Unknown instruction, which is already set in inst.opcode so
				// is detectable.
				// This error is handled when actually trying to interpret this
				// instruction (to not trigger on code that won't be executed).
			}
			if inst.opcode == llvm.PHI {
				// PHI nodes need to be treated specially, see the comment in
				// interpreter.go for an explanation.
				bb.phiNodes = append(bb.phiNodes, inst)
			} else {
				bb.instructions = append(bb.instructions, inst)
			}
		}
	}
	return fn
}

// instructionNameMap maps from instruction opcodes to instruction names. This
// can be useful for debug logging.
var instructionNameMap = [...]string{
	llvm.Ret:         "ret",
	llvm.Br:          "br",
	llvm.Switch:      "switch",
	llvm.IndirectBr:  "indirectbr",
	llvm.Invoke:      "invoke",
	llvm.Unreachable: "unreachable",

	// Standard Binary Operators
	llvm.Add:  "add",
	llvm.FAdd: "fadd",
	llvm.Sub:  "sub",
	llvm.FSub: "fsub",
	llvm.Mul:  "mul",
	llvm.FMul: "fmul",
	llvm.UDiv: "udiv",
	llvm.SDiv: "sdiv",
	llvm.FDiv: "fdiv",
	llvm.URem: "urem",
	llvm.SRem: "srem",
	llvm.FRem: "frem",

	// Logical Operators
	llvm.Shl:  "shl",
	llvm.LShr: "lshr",
	llvm.AShr: "ashr",
	llvm.And:  "and",
	llvm.Or:   "or",
	llvm.Xor:  "xor",

	// Memory Operators
	llvm.Alloca:        "alloca",
	llvm.Load:          "load",
	llvm.Store:         "store",
	llvm.GetElementPtr: "getelementptr",

	// Cast Operators
	llvm.Trunc:    "trunc",
	llvm.ZExt:     "zext",
	llvm.SExt:     "sext",
	llvm.FPToUI:   "fptoui",
	llvm.FPToSI:   "fptosi",
	llvm.UIToFP:   "uitofp",
	llvm.SIToFP:   "sitofp",
	llvm.FPTrunc:  "fptrunc",
	llvm.FPExt:    "fpext",
	llvm.PtrToInt: "ptrtoint",
	llvm.IntToPtr: "inttoptr",
	llvm.BitCast:  "bitcast",

	// Other Operators
	llvm.ICmp:           "icmp",
	llvm.FCmp:           "fcmp",
	llvm.PHI:            "phi",
	llvm.Call:           "call",
	llvm.Select:         "select",
	llvm.VAArg:          "vaarg",
	llvm.ExtractElement: "extractelement",
	llvm.InsertElement:  "insertelement",
	llvm.ShuffleVector:  "shufflevector",
	llvm.ExtractValue:   "extractvalue",
	llvm.InsertValue:    "insertvalue",
}
