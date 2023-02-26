package compiler

import (
	"strconv"
	"strings"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// createInterruptGlobal creates a new runtime/interrupt.Interrupt struct that
// will be lowered to a real interrupt during interrupt lowering.
//
// This two-stage approach allows unused interrupts to be optimized away if
// necessary.
func (b *builder) createInterruptGlobal(instr *ssa.CallCommon) (llvm.Value, error) {
	// Get the interrupt number, which must be a compile-time constant.
	id, ok := instr.Args[0].(*ssa.Const)
	if !ok {
		return llvm.Value{}, b.makeError(instr.Pos(), "interrupt ID is not a constant")
	}

	// Get the func value, which also must be a compile time constant.
	// Note that bound functions are allowed if the function has a pointer
	// receiver and is a global. This is rather strict but still allows for
	// idiomatic Go code.
	funcValue := b.getValue(instr.Args[1])
	if funcValue.IsAConstant().IsNil() {
		// Try to determine the cause of the non-constantness for a nice error
		// message.
		switch instr.Args[1].(type) {
		case *ssa.MakeClosure:
			// This may also be a bound method.
			return llvm.Value{}, b.makeError(instr.Pos(), "closures are not supported in interrupt.New")
		}
		// Fall back to a generic error.
		return llvm.Value{}, b.makeError(instr.Pos(), "interrupt function must be constant")
	}
	_, funcRawPtr, funcContext := b.decodeFuncValue(funcValue, nil)
	funcPtr := llvm.ConstPtrToInt(funcRawPtr, b.uintptrType)

	// Create a new global of type runtime/interrupt.handle. Globals of this
	// type are lowered in the interrupt lowering pass.
	globalType := b.program.ImportedPackage("runtime/interrupt").Type("handle").Type()
	globalLLVMType := b.getLLVMType(globalType)
	globalName := b.fn.Package().Pkg.Path() + "$interrupt" + strconv.FormatInt(id.Int64(), 10)
	global := llvm.AddGlobal(b.mod, globalLLVMType, globalName)
	if !b.LTO {
		global.SetVisibility(llvm.HiddenVisibility)
	}
	global.SetGlobalConstant(true)
	global.SetUnnamedAddr(true)
	initializer := llvm.ConstNull(globalLLVMType)
	initializer = b.CreateInsertValue(initializer, funcContext, 0, "")
	initializer = b.CreateInsertValue(initializer, funcPtr, 1, "")
	initializer = b.CreateInsertValue(initializer, llvm.ConstNamedStruct(globalLLVMType.StructElementTypes()[2], []llvm.Value{
		llvm.ConstInt(b.intType, uint64(id.Int64()), true),
	}), 2, "")
	global.SetInitializer(initializer)

	// Add debug info to the interrupt global.
	if b.Debug {
		pos := b.program.Fset.Position(instr.Pos())
		diglobal := b.dibuilder.CreateGlobalVariableExpression(b.getDIFile(pos.Filename), llvm.DIGlobalVariableExpression{
			Name:        "interrupt" + strconv.FormatInt(id.Int64(), 10),
			LinkageName: globalName,
			File:        b.getDIFile(pos.Filename),
			Line:        pos.Line,
			Type:        b.getDIType(globalType),
			Expr:        b.dibuilder.CreateExpression(nil),
			LocalToUnit: false,
		})
		global.AddMetadata(0, diglobal)
	}

	// Create the runtime/interrupt.Interrupt type. It is a struct with a single
	// member of type int.
	num := llvm.ConstPtrToInt(global, b.intType)
	interrupt := llvm.ConstNamedStruct(b.mod.GetTypeByName("runtime/interrupt.Interrupt"), []llvm.Value{num})

	// Add dummy "use" call for AVR, because interrupts may be used even though
	// they are never referenced again. This is unlike Cortex-M or the RISC-V
	// PLIC where each interrupt must be enabled using the interrupt number, and
	// thus keeps the Interrupt object alive.
	// This call is removed during interrupt lowering.
	if strings.HasPrefix(b.Triple, "avr") {
		useFn := b.mod.NamedFunction("runtime/interrupt.use")
		if useFn.IsNil() {
			useFnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{interrupt.Type()}, false)
			useFn = llvm.AddFunction(b.mod, "runtime/interrupt.use", useFnType)
		}
		b.CreateCall(useFn.GlobalValueType(), useFn, []llvm.Value{interrupt}, "")
	}

	return interrupt, nil
}
