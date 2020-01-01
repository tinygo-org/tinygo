package compiler

import (
	"strconv"
	"strings"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// emitInterruptGlobal creates a new runtime/interrupt.Interrupt struct that
// will be lowered to a real interrupt during interrupt lowering.
//
// This two-stage approach allows unused interrupts to be optimized away if
// necessary.
func (c *Compiler) emitInterruptGlobal(frame *Frame, instr *ssa.CallCommon) (llvm.Value, error) {
	// Get the interrupt number, which must be a compile-time constant.
	id, ok := instr.Args[0].(*ssa.Const)
	if !ok {
		return llvm.Value{}, c.makeError(instr.Pos(), "interrupt ID is not a constant")
	}

	// Get the func value, which also must be a compile time constant.
	// Note that bound functions are allowed if the function has a pointer
	// receiver and is a global. This is rather strict but still allows for
	// idiomatic Go code.
	funcValue := c.getValue(frame, instr.Args[1])
	if funcValue.IsAConstant().IsNil() {
		// Try to determine the cause of the non-constantness for a nice error
		// message.
		switch instr.Args[1].(type) {
		case *ssa.MakeClosure:
			// This may also be a bound method.
			return llvm.Value{}, c.makeError(instr.Pos(), "closures are not supported in interrupt.New")
		}
		// Fall back to a generic error.
		return llvm.Value{}, c.makeError(instr.Pos(), "interrupt function must be constant")
	}

	// Create a new global of type runtime/interrupt.handle. Globals of this
	// type are lowered in the interrupt lowering pass.
	globalType := c.ir.Program.ImportedPackage("runtime/interrupt").Type("handle").Type()
	globalLLVMType := c.getLLVMType(globalType)
	globalName := "runtime/interrupt.$interrupt" + strconv.FormatInt(id.Int64(), 10)
	if global := c.mod.NamedGlobal(globalName); !global.IsNil() {
		return llvm.Value{}, c.makeError(instr.Pos(), "interrupt redeclared in this program")
	}
	global := llvm.AddGlobal(c.mod, globalLLVMType, globalName)
	global.SetLinkage(llvm.PrivateLinkage)
	global.SetGlobalConstant(true)
	initializer := llvm.ConstNull(globalLLVMType)
	initializer = llvm.ConstInsertValue(initializer, funcValue, []uint32{0})
	initializer = llvm.ConstInsertValue(initializer, llvm.ConstInt(c.intType, uint64(id.Int64()), true), []uint32{1, 0})
	global.SetInitializer(initializer)

	// Add debug info to the interrupt global.
	if c.Debug() {
		pos := c.ir.Program.Fset.Position(instr.Pos())
		diglobal := c.dibuilder.CreateGlobalVariableExpression(c.difiles[pos.Filename], llvm.DIGlobalVariableExpression{
			Name:        "interrupt" + strconv.FormatInt(id.Int64(), 10),
			LinkageName: globalName,
			File:        c.getDIFile(pos.Filename),
			Line:        pos.Line,
			Type:        c.getDIType(globalType),
			Expr:        c.dibuilder.CreateExpression(nil),
			LocalToUnit: false,
		})
		global.AddMetadata(0, diglobal)
	}

	// Create the runtime/interrupt.Interrupt type. It is a struct with a single
	// member of type int.
	num := llvm.ConstPtrToInt(global, c.intType)
	interrupt := llvm.ConstNamedStruct(c.mod.GetTypeByName("runtime/interrupt.Interrupt"), []llvm.Value{num})

	// Add dummy "use" call for AVR, because interrupts may be used even though
	// they are never referenced again. This is unlike Cortex-M or the RISC-V
	// PLIC where each interrupt must be enabled using the interrupt number, and
	// thus keeps the Interrupt object alive.
	// This call is removed during interrupt lowering.
	if strings.HasPrefix(c.Triple(), "avr") {
		useFn := c.mod.NamedFunction("runtime/interrupt.use")
		if useFn.IsNil() {
			useFnType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{interrupt.Type()}, false)
			useFn = llvm.AddFunction(c.mod, "runtime/interrupt.use", useFnType)
		}
		c.builder.CreateCall(useFn, []llvm.Value{interrupt}, "")
	}

	return interrupt, nil
}
